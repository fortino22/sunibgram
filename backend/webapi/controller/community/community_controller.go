package communityctrl

import (
	"errors"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	communitymodel "github.com/sunibgram/domain/model/community"
	communitydiscovermodel "github.com/sunibgram/domain/model/community_discover"
	communityrulemodel "github.com/sunibgram/domain/model/community_rule"
	communitysvciface "github.com/sunibgram/domain/service/community"
	communityrulesvciface "github.com/sunibgram/domain/service/community_rule"
	userprofilesvciface "github.com/sunibgram/domain/service/user_profile"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	utils "github.com/sunibgram/webapi/dto/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	communitydto "github.com/sunibgram/webapi/dto/community"
	communityruledto "github.com/sunibgram/webapi/dto/community_rule"
	threaddto "github.com/sunibgram/webapi/dto/thread"
	webapiutils "github.com/sunibgram/webapi/utils"
)

type CommunityController struct {
	communityService communitysvciface.ICommunityService
	ruleService      communityrulesvciface.ICommunityRuleService
	profileService   userprofilesvciface.IUserProfileService
}

func NewCommunityController(communityService communitysvciface.ICommunityService, ruleService communityrulesvciface.ICommunityRuleService, profileService userprofilesvciface.IUserProfileService) *CommunityController {
	return &CommunityController{communityService: communityService, ruleService: ruleService, profileService: profileService}
}

func (c *CommunityController) Create(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	name := ctx.FormValue("name")
	description := ctx.FormValue("description")
	rules := readRulesFromForm(ctx)
	categoryID, ok := parseRequiredFormUUID(ctx, "category_id", "category")
	if !ok {
		return nil
	}
	visibility := communitymodel.CommunityVisibility(ctx.FormValue("visibility"))

	var banner *multipart.FileHeader
	if f, err := ctx.FormFile("banner"); err == nil {
		banner = f
	}

	community, _, err := c.communityService.Create(
		ctx.Context(), claims.UserID,
		name, description, rules, categoryID, visibility,
		banner,
	)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "community created successfully", "community_id": community.CommunityID})
}

func (c *CommunityController) Update(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ctrlshared.ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	name := ctx.FormValue("name")
	description := ctx.FormValue("description")
	keepRuleIDs, err := parseKeepRuleIDs(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
	newRules := readNewRulesFromForm(ctx)
	categoryID, ok := parseRequiredFormUUID(ctx, "category_id", "category")
	if !ok {
		return nil
	}
	visibility := communitymodel.CommunityVisibility(ctx.FormValue("visibility"))
	clearBanner := ctx.FormValue("clear_banner") == "true"

	var banner *multipart.FileHeader
	if f, err := ctx.FormFile("banner"); err == nil {
		banner = f
	}
	if banner != nil && clearBanner {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "cannot upload banner and clear_banner simultaneously"})
	}

	if _, _, err := c.communityService.Update(
		ctx.Context(), id, claims.UserID,
		name, description, keepRuleIDs, newRules, categoryID, visibility,
		banner, clearBanner,
	); err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "community updated successfully"})
}

func (c *CommunityController) Delete(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ctrlshared.ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	if err := c.communityService.Delete(ctx.Context(), id, claims.UserID); err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "community deleted successfully"})
}

func (c *CommunityController) GetByID(ctx *fiber.Ctx) error {
	id, ok := ctrlshared.ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	community, err := c.communityService.GetByID(id)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: "community not found"})
	}

	rules, _ := c.ruleService.ListByCommunity(id)
	return ctx.Status(fiber.StatusOK).JSON(c.toResponseWithRules(community, rules))
}

func (c *CommunityController) Discover(ctx *fiber.Ctx) error {
	var categoryID *uuid.UUID
	if raw := ctx.Query("category"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid category id"})
		}
		categoryID = &id
	}

	communities, err := c.communityService.Discover(ctx.Context(), categoryID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": c.toDiscoverResponses(communities)})
}

func (c *CommunityController) SearchByName(ctx *fiber.Ctx) error {
	communities, err := c.communityService.SearchByName(ctx.Context(), ctx.Query("community_name"))
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": c.toDiscoverResponses(communities)})
}

func (c *CommunityController) GetMyCreated(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	page, limit := ctrlshared.ParsePagination(ctx)

	communities, total, err := c.communityService.GetMyCreated(claims.UserID, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}

	slim := make([]communitydto.MyCommunityResponse, len(communities))
	for i := range communities {
		slim[i] = communitydto.MyCommunityResponse{
			CommunityID:   communities[i].CommunityID,
			CommunityName: communities[i].Name,
			BannerURL:     ctrlshared.BuildMediaURL(communities[i].BannerObjectName, c.communityService.GetURL),
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": slim,
		"meta": threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *CommunityController) GetMyJoined(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	page, limit := ctrlshared.ParsePagination(ctx)

	communities, total, err := c.communityService.GetMyJoined(claims.UserID, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}

	return ctx.Status(fiber.StatusOK).JSON(communitydto.PaginatedCommunityResponse{
		Data: c.toResponses(communities),
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *CommunityController) GetAbout(ctx *fiber.Ctx) error {
	id, ok := ctrlshared.ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	var viewerID *uint
	if claims, ok := webapiutils.Get(ctx); ok {
		uid := claims.UserID
		viewerID = &uid
	}

	about, err := c.communityService.GetAbout(ctx.Context(), id, viewerID)
	if err != nil {
		if err.Error() == "community not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}

	return ctx.Status(fiber.StatusOK).JSON(c.buildAboutResponse(about))
}

func (c *CommunityController) buildAboutResponse(about *communitysvciface.CommunityAbout) communitydto.CommunityAboutResponse {
	comm := about.Community
	postingPolicy := "Only members can post."
	visibilityPolicy := "All Communities are publicly visible."
	joinPolicy := "Anyone can join this Community."
	if comm.Visibility == communitymodel.CommunityVisibilityPrivate {
		visibilityPolicy = "This Community is private."
		joinPolicy = "Membership requires an approved join request."
	}

	creatorAvatar := ""
	creatorBio := ""
	if about.Creator.Profile != nil {
		creatorAvatar = ctrlshared.BuildMediaURL(about.Creator.Profile.AvatarObjectName, c.profileService.GetURL)
		creatorBio = about.Creator.Profile.Bio
	}

	creator := communitydto.CommunityAboutCreatorResponse{
		Name:        about.Creator.User.Name,
		UserID:      about.Creator.User.UserID,
		Username:    about.Creator.User.UserName,
		AvatarURL:   creatorAvatar,
		Bio:         creatorBio,
		IsPremium:   about.Creator.User.IsPremium,
		IsFollowing: about.Creator.IsFollowing,
		CreatedAt:   utils.Wrap(comm.CreatedAt),
	}

	return communitydto.CommunityAboutResponse{
		CommunityID: comm.CommunityID,
		Name:        comm.Name,
		Info: communitydto.CommunityAboutInfoResponse{
			Description:      comm.Description,
			PostingPolicy:    postingPolicy,
			VisibilityPolicy: visibilityPolicy,
			JoinPolicy:       joinPolicy,
		},
		Creator:    creator,
		Rules:      toRuleMiniResponses(about.Rules),
		Admins:     c.toAboutMemberResponses(about.Admins),
		TopMembers: c.toAboutMemberResponses(about.TopMembers),
	}
}

func (c *CommunityController) toAboutMemberResponses(users []communitysvciface.CommunityAboutUser) []communitydto.CommunityAboutMemberResponse {
	out := make([]communitydto.CommunityAboutMemberResponse, 0, len(users))
	for i := range users {
		u := users[i]
		if u.User == nil {
			continue
		}
		avatarURL := ""
		bio := ""
		if u.Profile != nil {
			avatarURL = ctrlshared.BuildMediaURL(u.Profile.AvatarObjectName, c.profileService.GetURL)
			bio = u.Profile.Bio
		}
		out = append(out, communitydto.CommunityAboutMemberResponse{
			UserID:      u.User.UserID,
			Name:        u.User.Name,
			Username:    u.User.UserName,
			AvatarURL:   avatarURL,
			Bio:         bio,
			IsPremium:   u.User.IsPremium,
			IsFollowing: u.IsFollowing,
		})
	}
	return out
}

func (c *CommunityController) toResponse(comm *communitymodel.Community) communitydto.CommunityResponse {
	return c.buildResponse(comm, nil)
}

func (c *CommunityController) toResponseWithRules(comm *communitymodel.Community, rules []communityrulemodel.CommunityRule) communitydto.CommunityResponse {
	return c.buildResponse(comm, toRuleMiniResponses(rules))
}

func toRuleMiniResponses(rules []communityrulemodel.CommunityRule) []communityruledto.CommunityRuleMiniResponse {
	out := make([]communityruledto.CommunityRuleMiniResponse, len(rules))
	for i := range rules {
		out[i] = communityruledto.CommunityRuleMiniResponse{
			ID:   rules[i].ID,
			Text: rules[i].Text,
		}
	}
	return out
}

func (c *CommunityController) buildResponse(comm *communitymodel.Community, rules []communityruledto.CommunityRuleMiniResponse) communitydto.CommunityResponse {
	bannerURL := ctrlshared.BuildMediaURL(comm.BannerObjectName, c.communityService.GetURL)
	topMembers := c.toTopMemberResponses(comm.TopMembers)
	var categoryName string
	if comm.Category != nil {
		categoryName = comm.Category.Name
	}
	return communitydto.CommunityResponse{
		CommunityID:  comm.CommunityID,
		UserID:       comm.UserID,
		Username:     comm.User.UserName,
		Name:         comm.Name,
		Description:  comm.Description,
		Rules:        rules,
		CategoryName: categoryName,
		Visibility:   string(comm.Visibility),
		BannerURL:    bannerURL,
		MemberCount:  comm.MemberCount,
		TopMembers:   topMembers,
		CreatedAt:    utils.Wrap(comm.CreatedAt),
		UpdatedAt:    utils.Wrap(comm.UpdatedAt),
	}
}

func (c *CommunityController) toResponses(communities []communitymodel.Community) []communitydto.CommunityResponse {
	out := make([]communitydto.CommunityResponse, len(communities))
	for i := range communities {
		out[i] = c.toResponse(&communities[i])
	}
	return out
}

func readRulesFromForm(ctx *fiber.Ctx) []string {
	form, err := ctx.MultipartForm()
	if err != nil || form == nil {
		return nil
	}
	return form.Value["rules"]
}

func readNewRulesFromForm(ctx *fiber.Ctx) []string {
	form, err := ctx.MultipartForm()
	if err != nil || form == nil {
		return nil
	}
	return form.Value["new_rules"]
}

func parseKeepRuleIDs(ctx *fiber.Ctx) ([]uint, error) {
	var values []string
	form, err := ctx.MultipartForm()
	if err == nil && form != nil {
		values = append(values, form.Value["keep_rule_ids"]...)
	}
	if len(values) == 0 {
		if raw := ctx.FormValue("keep_rule_ids"); raw != "" {
			values = append(values, raw)
		}
	}
	if len(values) == 0 {
		return nil, nil
	}
	ids := make([]uint, 0, len(values))
	for _, v := range values {
		for _, part := range strings.Split(v, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			id, err := strconv.ParseUint(part, 10, 64)
			if err != nil || id == 0 {
				return nil, errors.New("invalid keep_rule_ids")
			}
			ids = append(ids, uint(id))
		}
	}
	return ids, nil
}

func (c *CommunityController) toDiscoverResponse(comm *communitydiscovermodel.DiscoverCommunity) communitydto.DiscoverCommunityResponse {
	members := make([]communitydto.DiscoverCommunityMemberResponse, len(comm.TopMembers))
	for i := range comm.TopMembers {
		members[i] = communitydto.DiscoverCommunityMemberResponse{
			UserID: strconv.FormatUint(uint64(comm.TopMembers[i].UserID), 10),
			AvatarURL: ctrlshared.BuildMediaURL(
				comm.TopMembers[i].AvatarObjectName,
				c.communityService.GetURL,
			),
		}
	}

	var categoryName string
	if comm.Category != nil {
		categoryName = comm.Category.Name
	}
	return communitydto.DiscoverCommunityResponse{
		CommunityID:  comm.CommunityID,
		Name:         comm.Name,
		BannerURL:    ctrlshared.BuildMediaURL(comm.BannerObjectName, c.communityService.GetURL),
		MemberCount:  comm.MemberCount,
		CategoryName: categoryName,
		TopMembers:   members,
	}
}

func (c *CommunityController) toTopMemberResponses(members []communitydiscovermodel.DiscoverCommunityMember) []communitydto.CommunityTopMemberResponse {
	out := make([]communitydto.CommunityTopMemberResponse, len(members))
	for i := range members {
		out[i] = communitydto.CommunityTopMemberResponse{
			UserID: strconv.FormatUint(uint64(members[i].UserID), 10),
			AvatarURL: ctrlshared.BuildMediaURL(
				members[i].AvatarObjectName,
				c.communityService.GetURL,
			),
		}
	}
	return out
}

func (c *CommunityController) toDiscoverResponses(communities []communitydiscovermodel.DiscoverCommunity) []communitydto.DiscoverCommunityResponse {
	out := make([]communitydto.DiscoverCommunityResponse, len(communities))
	for i := range communities {
		out[i] = c.toDiscoverResponse(&communities[i])
	}
	return out
}

func parseRequiredFormUUID(ctx *fiber.Ctx, fieldName, entityName string) (*uuid.UUID, bool) {
	raw := ctx.FormValue(fieldName)
	if raw == "" {
		_ = ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: entityName + " is required"})
		return nil, false
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		_ = ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid " + entityName + " id"})
		return nil, false
	}
	return &id, true
}
