package communityuc

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/sunibgram/application/helper"
	communityruleuc "github.com/sunibgram/application/usecase/community_rule"
	storageuc "github.com/sunibgram/application/usecase/storage"
	communitymodel "github.com/sunibgram/domain/model/community"
	communitydiscovermodel "github.com/sunibgram/domain/model/community_discover"
	communitymembermodel "github.com/sunibgram/domain/model/community_member"
	communityrulemodel "github.com/sunibgram/domain/model/community_rule"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
	cacherepoiface "github.com/sunibgram/domain/repository/cache"
	categoryrepoiface "github.com/sunibgram/domain/repository/category"
	communityrepoiface "github.com/sunibgram/domain/repository/community"
	communitymemberrepoiface "github.com/sunibgram/domain/repository/community_member"
	communitypinrepoiface "github.com/sunibgram/domain/repository/community_pin"
	communityrulerepoiface "github.com/sunibgram/domain/repository/community_rule"
	followrepoiface "github.com/sunibgram/domain/repository/follow"
	userrepoiface "github.com/sunibgram/domain/repository/user"
	userprofilerepoiface "github.com/sunibgram/domain/repository/user_profile"
	communitysvciface "github.com/sunibgram/domain/service/community"
	communityactionconfirmationsvciface "github.com/sunibgram/domain/service/community_action_confirmation"
)

type CommunityService struct {
	communityRepo  communityrepoiface.ICommunityRepository
	memberRepo     communitymemberrepoiface.ICommunityMemberRepository
	userRepo       userrepoiface.IUserRepository
	profileRepo    userprofilerepoiface.IUserProfileRepository
	followRepo     followrepoiface.IFollowRepository
	categoryRepo   categoryrepoiface.ICategoryRepository
	ruleRepo       communityrulerepoiface.ICommunityRuleRepository
	pinRepo        communitypinrepoiface.ICommunityPinRepository
	cacheRepo      cacherepoiface.ICacheRepository
	storageUsecase *storageuc.StorageUsecase
	auditSvc       communityactionconfirmationsvciface.ICommunityActionConfirmationService
}

func NewCommunityService(
	communityRepo communityrepoiface.ICommunityRepository,
	memberRepo communitymemberrepoiface.ICommunityMemberRepository,
	userRepo userrepoiface.IUserRepository,
	profileRepo userprofilerepoiface.IUserProfileRepository,
	followRepo followrepoiface.IFollowRepository,
	categoryRepo categoryrepoiface.ICategoryRepository,
	ruleRepo communityrulerepoiface.ICommunityRuleRepository,
	pinRepo communitypinrepoiface.ICommunityPinRepository,
	cacheRepo cacherepoiface.ICacheRepository,
	storageUsecase *storageuc.StorageUsecase,
	auditSvc communityactionconfirmationsvciface.ICommunityActionConfirmationService,
) communitysvciface.ICommunityService {
	return &CommunityService{
		communityRepo:  communityRepo,
		memberRepo:     memberRepo,
		userRepo:       userRepo,
		profileRepo:    profileRepo,
		followRepo:     followRepo,
		categoryRepo:   categoryRepo,
		ruleRepo:       ruleRepo,
		pinRepo:        pinRepo,
		cacheRepo:      cacheRepo,
		storageUsecase: storageUsecase,
		auditSvc:       auditSvc,
	}
}

func (s *CommunityService) Create(ctx context.Context, userID uint, name, description string, rules []string, categoryID *uuid.UUID, visibility communitymodel.CommunityVisibility, banner *multipart.FileHeader) (*communitymodel.Community, []communityrulemodel.CommunityRule, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, nil, errors.New("user not found")
	}
	if !user.IsPremium {
		return nil, nil, errors.New("premium subscription required to create a community")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil, errors.New("name is required")
	}
	if utf8.RuneCountInString(name) > 100 {
		return nil, nil, errors.New("name must be 100 characters or less")
	}
	description = strings.TrimSpace(description)
	if utf8.RuneCountInString(description) > 2000 {
		return nil, nil, errors.New("description must be 2000 characters or less")
	}
	if err := s.ensureCommunityNameAvailable(name, nil); err != nil {
		return nil, nil, err
	}
	if visibility != communitymodel.CommunityVisibilityPublic && visibility != communitymodel.CommunityVisibilityPrivate {
		visibility = communitymodel.CommunityVisibilityPublic
	}
	if err := s.validateCategoryID(categoryID); err != nil {
		return nil, nil, err
	}
	cleanedRules, err := communityruleuc.SanitizeCommunityRules(rules)
	if err != nil {
		return nil, nil, err
	}

	community := &communitymodel.Community{
		UserID:      userID,
		Name:        name,
		Description: description,
		CategoryID:  categoryID,
		Visibility:  visibility,
		MemberCount: 1,
	}

	if banner != nil {
		objName, err := s.uploadImage(ctx, banner, "community_banner")
		if err != nil {
			return nil, nil, fmt.Errorf("banner upload failed: %w", err)
		}
		community.BannerObjectName = objName
	}

	var savedRules []communityrulemodel.CommunityRule
	if err := s.communityRepo.Transaction(func(tx *gorm.DB) error {
		if err := s.communityRepo.SaveWithTx(tx, community); err != nil {
			return err
		}

		member := &communitymembermodel.CommunityMember{
			CommunityID: community.CommunityID,
			UserID:      userID,
			Role:        communitymembermodel.CommunityMemberRoleAdmin,
		}
		if err := s.memberRepo.SaveWithTx(tx, member); err != nil {
			return err
		}

		savedRules = communityruleuc.BuildCommunityRules(community.CommunityID, cleanedRules)
		return s.ruleRepo.BulkCreateWithTx(tx, savedRules)
	}); err != nil {
		return nil, nil, err
	}

	InvalidateDiscoverCommunitiesCache(ctx, s.cacheRepo)
	s.hydrateCommunityCategory(community)
	return community, savedRules, nil
}

func (s *CommunityService) Update(ctx context.Context, id, userID uint, name, description string, keepRuleIDs []uint, newRules []string, categoryID *uuid.UUID, visibility communitymodel.CommunityVisibility, banner *multipart.FileHeader, clearBanner bool) (*communitymodel.Community, []communityrulemodel.CommunityRule, error) {
	community, err := s.communityRepo.FindByID(id)
	if err != nil {
		return nil, nil, errors.New("community not found")
	}
	if !community.IsWritable() {
		return nil, nil, errors.New("community is inactive")
	}
	actor, err := s.memberRepo.FindMember(id, userID)
	if err != nil || !IsCommunityAdmin(actor.Role) {
		return nil, nil, errors.New("forbidden")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil, errors.New("name is required")
	}
	if utf8.RuneCountInString(name) > 100 {
		return nil, nil, errors.New("name must be 100 characters or less")
	}
	description = strings.TrimSpace(description)
	if utf8.RuneCountInString(description) > 2000 {
		return nil, nil, errors.New("description must be 2000 characters or less")
	}
	if err := s.ensureCommunityNameAvailable(name, &id); err != nil {
		return nil, nil, err
	}
	if visibility != communitymodel.CommunityVisibilityPublic && visibility != communitymodel.CommunityVisibilityPrivate {
		visibility = communitymodel.CommunityVisibilityPublic
	}
	if err := s.validateCategoryID(categoryID); err != nil {
		return nil, nil, err
	}
	dedupedKeepIDs := communityruleuc.DedupeRuleIDs(keepRuleIDs)
	cleanedNewRules, err := communityruleuc.SanitizeCommunityRules(newRules)
	if err != nil {
		return nil, nil, err
	}
	if len(dedupedKeepIDs)+len(cleanedNewRules) > communityruleuc.MaxCommunityRules {
		return nil, nil, errors.New("maximum 10 rules allowed")
	}
	if len(dedupedKeepIDs) > 0 {
		existing, err := s.ruleRepo.ListByIDsAndCommunity(community.CommunityID, dedupedKeepIDs)
		if err != nil {
			return nil, nil, err
		}
		if len(existing) != len(dedupedKeepIDs) {
			return nil, nil, errors.New("one or more keep_rule_ids do not belong to community")
		}
	}

	community.Name = name
	community.Description = description
	community.CategoryID = categoryID
	community.Visibility = visibility

	if clearBanner && community.BannerObjectName != "" {
		_ = s.storageUsecase.Delete(ctx, community.BannerObjectName)
		community.BannerObjectName = ""
	} else if banner != nil {
		if community.BannerObjectName != "" {
			_ = s.storageUsecase.Delete(ctx, community.BannerObjectName)
		}
		objName, err := s.uploadImage(ctx, banner, "community_banner")
		if err != nil {
			return nil, nil, fmt.Errorf("banner upload failed: %w", err)
		}
		community.BannerObjectName = objName
	}

	var savedRules []communityrulemodel.CommunityRule
	if err := s.communityRepo.Transaction(func(tx *gorm.DB) error {
		if err := s.communityRepo.UpdateWithTx(tx, community); err != nil {
			return err
		}
		if err := s.ruleRepo.DeleteByCommunityNotInWithTx(tx, community.CommunityID, dedupedKeepIDs); err != nil {
			return err
		}
		// Two-phase reorder: shift kept rules to negative temp order to avoid unique idx conflict.
		for i, ruleID := range dedupedKeepIDs {
			if err := s.ruleRepo.UpdateOrderWithTx(tx, ruleID, -(i + 1)); err != nil {
				return err
			}
		}
		for i, ruleID := range dedupedKeepIDs {
			if err := s.ruleRepo.UpdateOrderWithTx(tx, ruleID, i+1); err != nil {
				return err
			}
		}
		newRuleRecords := make([]communityrulemodel.CommunityRule, len(cleanedNewRules))
		for i, text := range cleanedNewRules {
			newRuleRecords[i] = communityrulemodel.CommunityRule{
				CommunityID: community.CommunityID,
				OrderNum:    len(dedupedKeepIDs) + i + 1,
				Text:        text,
			}
		}
		return s.ruleRepo.BulkCreateWithTx(tx, newRuleRecords)
	}); err != nil {
		return nil, nil, err
	}
	final, err := s.ruleRepo.ListByCommunity(community.CommunityID)
	if err != nil {
		return nil, nil, err
	}
	savedRules = final

	InvalidateDiscoverCommunitiesCache(ctx, s.cacheRepo)
	if s.pinRepo != nil {
		s.pinRepo.InvalidateByCommunity(community.CommunityID)
	}
	s.hydrateCommunityCategory(community)
	return community, savedRules, nil
}

func (s *CommunityService) Delete(ctx context.Context, id, userID uint) error {
	community, err := s.communityRepo.FindByID(id)
	if err != nil {
		return errors.New("community not found")
	}
	if !community.IsWritable() {
		return errors.New("community is inactive")
	}
	actor, err := s.memberRepo.FindMember(id, userID)
	if err != nil || !IsCommunityAdmin(actor.Role) {
		return errors.New("forbidden")
	}
	if community.BannerObjectName != "" {
		_ = s.storageUsecase.Delete(ctx, community.BannerObjectName)
	}
	if s.pinRepo != nil {
		s.pinRepo.InvalidateByCommunity(id)
	}
	if err := s.communityRepo.Delete(id); err != nil {
		return err
	}
	if s.auditSvc != nil {
		_ = s.auditSvc.RecordCommunityDeletion(ctx, id, userID, community.Name)
	}
	InvalidateDiscoverCommunitiesCache(ctx, s.cacheRepo)
	return nil
}

func (s *CommunityService) Discover(ctx context.Context, categoryID *uuid.UUID) ([]communitydiscovermodel.DiscoverCommunity, error) {
	if categoryID != nil && *categoryID != uuid.Nil {
		if _, err := s.categoryRepo.FindByID(*categoryID); err != nil {
			return nil, errors.New("categoryId doesn't exist")
		}
	}

	var cached []communitydiscovermodel.DiscoverCommunity
	if s.cacheRepo != nil {
		if err := s.cacheRepo.Get(ctx, discoverCommunitiesCacheKey, &cached); err == nil {
			return filterDiscoverByCategory(cached, categoryID), nil
		}
	}

	communities, err := s.communityRepo.FindDiscover()
	if err != nil {
		return nil, err
	}
	if s.cacheRepo != nil {
		_ = s.cacheRepo.Set(ctx, discoverCommunitiesCacheKey, communities, discoverCommunitiesTTL)
	}
	return filterDiscoverByCategory(communities, categoryID), nil
}

func (s *CommunityService) SearchByName(ctx context.Context, name string) ([]communitydiscovermodel.DiscoverCommunity, error) {
	name = strings.TrimSpace(name)
	if utf8.RuneCountInString(name) > 100 {
		name = string([]rune(name)[:100])
	}
	var cached []communitydiscovermodel.DiscoverCommunity
	if s.cacheRepo != nil {
		if err := s.cacheRepo.Get(ctx, communitySearchCacheKey(name), &cached); err == nil {
			return cached, nil
		}
	}

	communities, err := s.communityRepo.FindDiscoverByName(name)
	if err != nil {
		return nil, err
	}
	if s.cacheRepo != nil {
		_ = s.cacheRepo.Set(ctx, communitySearchCacheKey(name), communities, discoverCommunitiesTTL)
	}
	return communities, nil
}

func filterDiscoverByCategory(communities []communitydiscovermodel.DiscoverCommunity, categoryID *uuid.UUID) []communitydiscovermodel.DiscoverCommunity {
	if categoryID == nil || *categoryID == uuid.Nil {
		return communities
	}
	out := communities[:0:0]
	for _, c := range communities {
		if c.Category != nil && c.Category.ID == *categoryID {
			out = append(out, c)
		}
	}
	return out
}

func (s *CommunityService) GetByID(id uint) (*communitymodel.Community, error) {
	community, err := s.communityRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	s.hydrateCommunityCategory(community)
	return community, nil
}

func (s *CommunityService) GetAbout(ctx context.Context, communityID uint, viewerID *uint) (*communitysvciface.CommunityAbout, error) {
	community, err := s.communityRepo.FindByID(communityID)
	if err != nil {
		return nil, errors.New("community not found")
	}
	s.hydrateCommunityCategory(community)

	rules, err := s.ruleRepo.ListByCommunity(communityID)
	if err != nil {
		return nil, err
	}

	admins, err := s.memberRepo.FindAdmins(communityID)
	if err != nil {
		return nil, err
	}
	topMembers, err := s.memberRepo.FindTopMembers(communityID, 10)
	if err != nil {
		return nil, err
	}

	creatorUser, err := s.userRepo.FindByID(community.UserID)
	if err != nil {
		return nil, errors.New("community creator not found")
	}
	creatorProfile, _ := s.profileRepo.FindByUserID(community.UserID)

	targetIDs := make([]uint, 0, 1+len(admins)+len(topMembers))
	targetIDs = append(targetIDs, community.UserID)
	for i := range admins {
		targetIDs = append(targetIDs, admins[i].UserID)
	}
	for i := range topMembers {
		targetIDs = append(targetIDs, topMembers[i].UserID)
	}

	followSet := map[uint]bool{}
	if viewerID != nil && *viewerID != 0 && len(targetIDs) > 0 {
		fs, ferr := s.followRepo.GetFollowingSet(ctx, *viewerID, targetIDs)
		if ferr == nil {
			followSet = fs
		}
	}

	creatorAbout := communitysvciface.CommunityAboutUser{
		User:        creatorUser,
		Profile:     creatorProfile,
		IsFollowing: followSet[community.UserID],
	}
	if viewerID != nil && *viewerID == community.UserID {
		creatorAbout.IsFollowing = false
	}

	adminAbout := s.buildAboutUsers(admins, viewerID, followSet)
	memberAbout := s.buildAboutUsers(topMembers, viewerID, followSet)

	return &communitysvciface.CommunityAbout{
		Community:  community,
		Creator:    creatorAbout,
		Rules:      rules,
		Admins:     adminAbout,
		TopMembers: memberAbout,
	}, nil
}

func (s *CommunityService) buildAboutUsers(members []communitymembermodel.CommunityMember, viewerID *uint, followSet map[uint]bool) []communitysvciface.CommunityAboutUser {
	out := make([]communitysvciface.CommunityAboutUser, 0, len(members))
	for i := range members {
		m := &members[i]
		user := m.User
		var profile *userprofilemodel.UserProfile
		if m.Profile != nil {
			profile = m.Profile
		}
		isFollowing := followSet[m.UserID]
		if viewerID != nil && *viewerID == m.UserID {
			isFollowing = false
		}
		out = append(out, communitysvciface.CommunityAboutUser{
			User:        &user,
			Profile:     profile,
			IsFollowing: isFollowing,
		})
	}
	return out
}

func (s *CommunityService) GetMyCreated(userID uint, page, limit int) ([]communitymodel.Community, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	communities, total, err := s.communityRepo.FindByCreatorID(userID, page, limit)
	if err != nil {
		return nil, 0, err
	}
	s.hydrateCommunityCategories(communities)
	return communities, total, nil
}

func (s *CommunityService) GetMyJoined(userID uint, page, limit int) ([]communitymodel.Community, int64, error) {
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	communities, total, err := s.communityRepo.FindJoinedByUserID(userID, page, limit)
	if err != nil {
		return nil, 0, err
	}
	s.hydrateCommunityCategories(communities)
	return communities, total, nil
}

func (s *CommunityService) GetURL(objectName string) string {
	return s.storageUsecase.GetURL(objectName)
}

func (s *CommunityService) uploadImage(ctx context.Context, header *multipart.FileHeader, uploadType string) (string, error) {
	return helper.UploadImage(ctx, s.storageUsecase, header, uploadType)
}

func (s *CommunityService) validateCategoryID(categoryID *uuid.UUID) error {
	if categoryID == nil || *categoryID == uuid.Nil {
		return errors.New("category is required")
	}
	if _, err := s.categoryRepo.FindByID(*categoryID); err != nil {
		return errors.New("category not found")
	}
	return nil
}

func (s *CommunityService) ensureCommunityNameAvailable(name string, excludeID *uint) error {
	exists, err := s.communityRepo.ExistsByName(name, excludeID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("community name already taken")
	}
	return nil
}

func (s *CommunityService) hydrateCommunityCategories(communities []communitymodel.Community) {
	for i := range communities {
		s.hydrateCommunityCategory(&communities[i])
	}
}

func (s *CommunityService) hydrateCommunityCategory(community *communitymodel.Community) {
	if community == nil || community.CategoryID == nil {
		if community != nil {
			community.Category = nil
		}
		return
	}
	category, err := s.categoryRepo.FindByID(*community.CategoryID)
	if err != nil {
		community.Category = nil
		return
	}
	community.Category = category
}
