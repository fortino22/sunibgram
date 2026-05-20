package controller

import (
	"mime/multipart"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/dto/utils"
)

type CommunityController struct {
	communityService service.ICommunityService
}

func NewCommunityController(communityService service.ICommunityService) *CommunityController {
	return &CommunityController{communityService: communityService}
}

func (c *CommunityController) Create(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	name := ctx.FormValue("name")
	description := ctx.FormValue("description")
	rules := ctx.FormValue("rules")
	categoryID, ok := parseRequiredFormUUID(ctx, "category_id", "category")
	if !ok {
		return nil
	}
	visibility := model.CommunityVisibility(ctx.FormValue("visibility"))

	var avatar, banner *multipart.FileHeader
	if f, err := ctx.FormFile("avatar"); err == nil {
		avatar = f
	}
	if f, err := ctx.FormFile("banner"); err == nil {
		banner = f
	}

	community, err := c.communityService.Create(
		ctx.Context(), claims.UserID,
		name, description, rules, categoryID, visibility,
		avatar, banner,
	)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusCreated).JSON(c.toResponse(community))
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
	rules := ctx.FormValue("rules")
	categoryID, ok := parseRequiredFormUUID(ctx, "category_id", "category")
	if !ok {
		return nil
	}
	visibility := model.CommunityVisibility(ctx.FormValue("visibility"))
	clearAvatar := ctx.FormValue("clear_avatar") == "true"
	clearBanner := ctx.FormValue("clear_banner") == "true"

	var avatar, banner *multipart.FileHeader
	if f, err := ctx.FormFile("avatar"); err == nil {
		avatar = f
	}
	if f, err := ctx.FormFile("banner"); err == nil {
		banner = f
	}

	if _, err := c.communityService.Update(
		ctx.Context(), id, claims.UserID,
		name, description, rules, categoryID, visibility,
		avatar, banner, clearAvatar, clearBanner,
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

	return ctx.Status(fiber.StatusOK).JSON(c.toResponse(community))
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
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": c.toDiscoverResponses(communities)})
}

func (c *CommunityController) SearchByName(ctx *fiber.Ctx) error {
	communities, err := c.communityService.SearchByName(ctx.Context(), ctx.Query("community_name"))
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
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
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
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
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedCommunityResponse{
		Data: c.toResponses(communities),
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *CommunityController) toResponse(comm *model.Community) dto.CommunityResponse {
	avatarURL := BuildMediaURL(comm.AvatarObjectName, comm.AvatarBucket, c.communityService.GetURL)
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
		Rules:        comm.Rules,
		CategoryName: categoryName,
		Visibility:   string(comm.Visibility),
		AvatarURL:    avatarURL,
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
