package controller

import (
	"errors"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/dto/utils"
	webapiutils "github.com/sunibgram/webapi/utils"
)

type CommunityController struct {
	communityService service.ICommunityService
	ruleService      service.ICommunityRuleService
	profileService   service.IUserProfileService
}

func NewCommunityController(communityService service.ICommunityService, ruleService service.ICommunityRuleService, profileService service.IUserProfileService) *CommunityController {
	return &CommunityController{communityService: communityService, ruleService: ruleService, profileService: profileService}
}

func (c *CommunityController) Create(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
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
	visibility := model.CommunityVisibility(ctx.FormValue("visibility"))

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
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "community created successfully", "community_id": community.CommunityID})
}

func (c *CommunityController) Update(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	name := ctx.FormValue("name")
	description := ctx.FormValue("description")
	keepRuleIDs, err := parseKeepRuleIDs(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	newRules := readNewRulesFromForm(ctx)
	categoryID, ok := parseRequiredFormUUID(ctx, "category_id", "category")
	if !ok {
		return nil
	}
	visibility := model.CommunityVisibility(ctx.FormValue("visibility"))
	clearBanner := ctx.FormValue("clear_banner") == "true"

	var banner *multipart.FileHeader
	if f, err := ctx.FormFile("banner"); err == nil {
		banner = f
	}
	if banner != nil && clearBanner {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "cannot upload banner and clear_banner simultaneously"})
	}

	if _, _, err := c.communityService.Update(
		ctx.Context(), id, claims.UserID,
		name, description, keepRuleIDs, newRules, categoryID, visibility,
		banner, clearBanner,
	); err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "community updated successfully"})
}

func (c *CommunityController) Delete(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	if err := c.communityService.Delete(ctx.Context(), id, claims.UserID); err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "community deleted successfully"})
}

func (c *CommunityController) GetByID(ctx *fiber.Ctx) error {
	id, ok := ParseIDParam(ctx, "id", "community")
	if !ok {
		return nil
	}

	community, err := c.communityService.GetByID(id)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: "community not found"})
	}

	rules, _ := c.ruleService.ListByCommunity(id)
	return ctx.Status(fiber.StatusOK).JSON(c.toResponseWithRules(community, rules))
}

func (c *CommunityController) Discover(ctx *fiber.Ctx) error {
	var categoryID *uuid.UUID
	if raw := ctx.Query("category"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid category id"})
		}
		categoryID = &id
	}

	communities, err := c.communityService.Discover(ctx.Context(), categoryID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": c.toDiscoverResponses(communities)})
}

func (c *CommunityController) SearchByName(ctx *fiber.Ctx) error {
	communities, err := c.communityService.SearchByName(ctx.Context(), ctx.Query("community_name"))
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: "internal server error"})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": c.toDiscoverResponses(communities)})
}

func (c *CommunityController) GetMyCreated(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	page, limit := ParsePagination(ctx)

	communities, total, err := c.communityService.GetMyCreated(claims.UserID, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: "internal server error"})
	}

	slim := make([]dto.MyCommunityResponse, len(communities))
	for i := range communities {
		slim[i] = dto.MyCommunityResponse{
			CommunityID:   communities[i].CommunityID,
			CommunityName: communities[i].Name,
			BannerURL:     BuildMediaURL(communities[i].BannerObjectName, communities[i].BannerBucket, c.communityService.GetURL),
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": slim,
		"meta": dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *CommunityController) GetMyJoined(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	page, limit := ParsePagination(ctx)

	communities, total, err := c.communityService.GetMyJoined(claims.UserID, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: "internal server error"})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedCommunityResponse{
		Data: c.toResponses(communities),
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *CommunityController) GetAbout(ctx *fiber.Ctx) error {
	id, ok := ParseIDParam(ctx, "id", "community")
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
			return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: "internal server error"})
	}

	return ctx.Status(fiber.StatusOK).JSON(c.buildAboutResponse(about))
}

func (c *CommunityController) buildAboutResponse(about *service.CommunityAbout) dto.CommunityAboutResponse {
	comm := about.Community
	postingPolicy := "Only members can post."
	visibilityPolicy := "All Communities are publicly visible."
	joinPolicy := "Anyone can join this Community."
	if comm.Visibility == model.CommunityVisibilityPrivate {
		visibilityPolicy = "This Community is private."
		joinPolicy = "Membership requires an approved join request."
	}

	creatorAvatar := ""
	creatorBio := ""
	if about.Creator.Profile != nil {
		creatorAvatar = BuildMediaURL(about.Creator.Profile.AvatarObjectName, about.Creator.Profile.AvatarBucket, c.profileService.GetURL)
		creatorBio = about.Creator.Profile.Bio
	}

	creator := dto.CommunityAboutCreatorResponse{
		Name:        about.Creator.User.Name,
		UserID:      about.Creator.User.UserID,
		Username:    about.Creator.User.UserName,
		AvatarURL:   creatorAvatar,
		Bio:         creatorBio,
		IsPremium:   about.Creator.User.IsPremium,
		IsFollowing: about.Creator.IsFollowing,
		CreatedAt:   utils.Wrap(comm.CreatedAt),
	}

	return dto.CommunityAboutResponse{
		CommunityID: comm.CommunityID,
		Name:        comm.Name,
		Info: dto.CommunityAboutInfoResponse{
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

func (c *CommunityController) toAboutMemberResponses(users []service.CommunityAboutUser) []dto.CommunityAboutMemberResponse {
	out := make([]dto.CommunityAboutMemberResponse, 0, len(users))
	for i := range users {
		u := users[i]
		if u.User == nil {
			continue
		}
		avatarURL := ""
		bio := ""
		if u.Profile != nil {
			avatarURL = BuildMediaURL(u.Profile.AvatarObjectName, u.Profile.AvatarBucket, c.profileService.GetURL)
			bio = u.Profile.Bio
		}
		out = append(out, dto.CommunityAboutMemberResponse{
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

func (c *CommunityController) toResponse(comm *model.Community) dto.CommunityResponse {
	return c.buildResponse(comm, nil)
}

func (c *CommunityController) toResponseWithRules(comm *model.Community, rules []model.CommunityRule) dto.CommunityResponse {
	return c.buildResponse(comm, toRuleMiniResponses(rules))
}

func toRuleMiniResponses(rules []model.CommunityRule) []dto.CommunityRuleMiniResponse {
	out := make([]dto.CommunityRuleMiniResponse, len(rules))
	for i := range rules {
		out[i] = dto.CommunityRuleMiniResponse{
			ID:   rules[i].ID,
			Text: rules[i].Text,
		}
	}
	return out
}

func (c *CommunityController) buildResponse(comm *model.Community, rules []dto.CommunityRuleMiniResponse) dto.CommunityResponse {
	bannerURL := BuildMediaURL(comm.BannerObjectName, comm.BannerBucket, c.communityService.GetURL)
	topMembers := c.toTopMemberResponses(comm.TopMembers)
	var categoryName string
	if comm.Category != nil {
		categoryName = comm.Category.Name
	}
	return dto.CommunityResponse{
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

func (c *CommunityController) toResponses(communities []model.Community) []dto.CommunityResponse {
	out := make([]dto.CommunityResponse, len(communities))
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

func (c *CommunityController) toDiscoverResponse(comm *model.DiscoverCommunity) dto.DiscoverCommunityResponse {
	members := make([]dto.DiscoverCommunityMemberResponse, len(comm.TopMembers))
	for i := range comm.TopMembers {
		members[i] = dto.DiscoverCommunityMemberResponse{
			UserID: strconv.FormatUint(uint64(comm.TopMembers[i].UserID), 10),
			AvatarURL: BuildMediaURL(
				comm.TopMembers[i].AvatarObjectName,
				comm.TopMembers[i].AvatarBucket,
				c.communityService.GetURL,
			),
		}
	}

	var categoryName string
	if comm.Category != nil {
		categoryName = comm.Category.Name
	}
	return dto.DiscoverCommunityResponse{
		CommunityID:  comm.CommunityID,
		Name:         comm.Name,
		BannerURL:    BuildMediaURL(comm.BannerObjectName, comm.BannerBucket, c.communityService.GetURL),
		MemberCount:  comm.MemberCount,
		CategoryName: categoryName,
		TopMembers:   members,
	}
}

func (c *CommunityController) toTopMemberResponses(members []model.DiscoverCommunityMember) []dto.CommunityTopMemberResponse {
	out := make([]dto.CommunityTopMemberResponse, len(members))
	for i := range members {
		out[i] = dto.CommunityTopMemberResponse{
			UserID: strconv.FormatUint(uint64(members[i].UserID), 10),
			AvatarURL: BuildMediaURL(
				members[i].AvatarObjectName,
				members[i].AvatarBucket,
				c.communityService.GetURL,
			),
		}
	}
	return out
}

func (c *CommunityController) toDiscoverResponses(communities []model.DiscoverCommunity) []dto.DiscoverCommunityResponse {
	out := make([]dto.DiscoverCommunityResponse, len(communities))
	for i := range communities {
		out[i] = c.toDiscoverResponse(&communities[i])
	}
	return out
}

func parseRequiredFormUUID(ctx *fiber.Ctx, fieldName, entityName string) (*uuid.UUID, bool) {
	raw := ctx.FormValue(fieldName)
	if raw == "" {
		_ = ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: entityName + " is required"})
		return nil, false
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		_ = ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid " + entityName + " id"})
		return nil, false
	}
	return &id, true
}
