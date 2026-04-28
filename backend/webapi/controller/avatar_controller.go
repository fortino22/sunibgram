package controller

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/webapi/ctxutil"
)

type AvatarController struct {
	avatarService service.IAvatarService
}

func NewAvatarController(avatarService service.IAvatarService) *AvatarController {
	return &AvatarController{avatarService: avatarService}
}

func (c *AvatarController) UploadAvatar(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "file is required"})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: "failed to open file"})
	}
	defer file.Close()

	avatar, err := c.avatarService.UploadAvatar(ctx.Context(), claims.UserID, file, fileHeader)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(avatarResponse(avatar))
}

func (c *AvatarController) UpdateAvatar(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "file is required"})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: "failed to open file"})
	}
	defer file.Close()

	avatar, err := c.avatarService.UploadAvatar(ctx.Context(), claims.UserID, file, fileHeader)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(avatarResponse(avatar))
}

func (c *AvatarController) GetAvatarByUserID(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("userId"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid user id"})
	}

	avatar, err := c.avatarService.GetAvatarByUserID(uint(id))
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: "avatar not found"})
	}

	return ctx.Status(fiber.StatusOK).JSON(avatarResponse(avatar))
}

func (c *AvatarController) DeleteAvatar(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	if err := c.avatarService.DeleteAvatarByUserID(claims.UserID); err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "avatar deleted"})
}

func avatarResponse(a *model.UserAvatar) dto.AvatarResponse {
	return dto.AvatarResponse{
		AvatarID:   a.AvatarID,
		UserID:     a.UserID,
		URL:        a.URL,
		ObjectName: a.ObjectName,
		Bucket:     a.Bucket,
		UpdatedAt:  a.UpdatedAt,
	}
}
