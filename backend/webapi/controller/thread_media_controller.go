package controller

import (
	"mime/multipart"
	"strconv"
	"strings"

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

// Update expects multipart fields named "file_<mediaID>" (e.g. "file_3", "file_7").
// Each field replaces the image for that specific media_id, keeping the id unchanged.
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

	// Parse fields named "file_<mediaID>".
	replacements := make(map[uint]*multipart.FileHeader)
	for key, headers := range form.File {
		if !strings.HasPrefix(key, "file_") || len(headers) == 0 {
			continue
		}
		mediaID, err := strconv.ParseUint(strings.TrimPrefix(key, "file_"), 10, 64)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid field name: " + key})
		}
		replacements[uint(mediaID)] = headers[0]
	}

	if len(replacements) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "at least one file_<mediaID> field is required"})
	}

	if err := c.mediaService.Update(ctx.Context(), uint(id), claims.UserID, replacements); err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
		}
		if err.Error() == "media not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
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
