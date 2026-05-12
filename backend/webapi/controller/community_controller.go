package controller

import (
	"mime/multipart"

	"github.com/gofiber/fiber/v2"
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
	category := ctx.FormValue("category")
	visibility := model.CommunityVisibility(ctx.FormValue("visibility"))

	var avatar, banner *multipart.FileHeader
	if f, err := ctx.FormFile("avatar"); err == nil {
		avatar = f
	}
	if f, err := ctx.FormFile("banner"); err == nil {
		banner = f
	}

	_, err := c.communityService.Create(
		ctx.Context(), claims.UserID,
		name, description, rules, category, visibility,
		avatar, banner,
	)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "community created successfully"})
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
	category := ctx.FormValue("category")
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
		name, description, rules, category, visibility,
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

func (c *CommunityController) GetAll(ctx *fiber.Ctx) error {
	page, limit := ParsePagination(ctx)
	category := ctx.Query("category", "")

	communities, total, err := c.communityService.GetAll(page, limit, category)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedCommunityResponse{
		Data: c.toResponses(communities),
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
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

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedCommunityResponse{
		Data: c.toResponses(communities),
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
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
	return dto.CommunityResponse{
		CommunityID: comm.CommunityID,
		UserID:      comm.UserID,
		Username:    comm.User.UserName,
		Name:        comm.Name,
		Description: comm.Description,
		Rules:       comm.Rules,
		Category:    comm.Category,
		Visibility:  string(comm.Visibility),
		AvatarURL:   avatarURL,
		BannerURL:   bannerURL,
		MemberCount: comm.MemberCount,
		CreatedAt:   utils.Wrap(comm.CreatedAt),
		UpdatedAt:   utils.Wrap(comm.UpdatedAt),
	}
}

func (c *CommunityController) toResponses(communities []model.Community) []dto.CommunityResponse {
	out := make([]dto.CommunityResponse, len(communities))
	for i := range communities {
		out[i] = c.toResponse(&communities[i])
	}
	return out
}
