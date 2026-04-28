package controller

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/webapi/ctxutil"
)

type ThreadMediaController struct {
	mediaService service.IThreadMediaService
}

func NewThreadMediaController(mediaService service.IThreadMediaService) *ThreadMediaController {
	return &ThreadMediaController{mediaService: mediaService}
}

func (c *ThreadMediaController) Upload(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid thread id"})
	}

	form, err := ctx.MultipartForm()
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid multipart form"})
	}

	files := form.File["files"]
	if len(files) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "at least one file is required"})
	}

	if _, err := c.mediaService.Upload(ctx.Context(), uint(id), claims.UserID, files); err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "media uploaded successfully"})
}

func (c *ThreadMediaController) Update(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid thread id"})
	}

	form, err := ctx.MultipartForm()
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid multipart form"})
	}

	files := form.File["files"]
	if len(files) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "at least one file is required"})
	}

	if err := c.mediaService.Update(ctx.Context(), uint(id), claims.UserID, files); err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "media updated successfully"})
}

func (c *ThreadMediaController) Delete(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid thread id"})
	}

	var req dto.DeleteMediaRequest
	if err := ctx.BodyParser(&req); err != nil || len(req.MediaIDs) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "media_ids is required"})
	}

	if err := c.mediaService.Delete(ctx.Context(), uint(id), claims.UserID, req.MediaIDs); err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "media deleted successfully"})
}
