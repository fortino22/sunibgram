package controller

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/webapi/ctxutil"
)

type ThreadController struct {
	threadService service.IThreadService
	mediaService  service.IThreadMediaService
}

func NewThreadController(threadService service.IThreadService, mediaService service.IThreadMediaService) *ThreadController {
	return &ThreadController{threadService: threadService, mediaService: mediaService}
}

func (c *ThreadController) Create(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	var req dto.CreateThreadRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}

	thread, err := c.threadService.Create(ctx.Context(), claims.UserID, req.Content, req.ParentID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusCreated).JSON(toThreadResponse(thread, nil))
}

func (c *ThreadController) GetRoots(ctx *fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))

	threads, total, err := c.threadService.GetRoots(page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	data := make([]dto.ThreadResponse, len(threads))
	for i, t := range threads {
		data[i] = toThreadResponse(&t, nil)
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedThreadsResponse{
		Data:  data,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

func (c *ThreadController) GetByID(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid thread id"})
	}

	thread, err := c.threadService.GetByID(uint(id))
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: "thread not found"})
	}

	return ctx.Status(fiber.StatusOK).JSON(toThreadResponse(thread, nil))
}

func (c *ThreadController) GetChain(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid thread id"})
	}

	threads, err := c.threadService.GetChain(uint(id))
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	data := make([]dto.ThreadResponse, len(threads))
	for i, t := range threads {
		data[i] = toThreadResponse(&t, nil)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": data})
}

func (c *ThreadController) Update(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid thread id"})
	}

	var req dto.UpdateThreadRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}

	thread, err := c.threadService.Update(ctx.Context(), uint(id), claims.UserID, req.Content)
	if err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(toThreadResponse(thread, nil))
}

func (c *ThreadController) Delete(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid thread id"})
	}

	if err := c.threadService.Delete(ctx.Context(), uint(id), claims.UserID); err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
		}
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	_ = c.mediaService.DeleteByThread(ctx.Context(), uint(id))

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "thread deleted successfully"})
}

func toThreadResponse(t *model.Thread, media []model.ThreadMedia) dto.ThreadResponse {
	return dto.ThreadResponse{
		ThreadID:  t.ThreadID,
		UserID:    t.UserID,
		Username:  t.User.UserName,
		Content:   t.Content,
		ParentID:  t.ParentID,
		RootID:    t.RootID,
		Media:     toMediaResponses(media),
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func toMediaResponses(media []model.ThreadMedia) []dto.ThreadMediaResponse {
	result := make([]dto.ThreadMediaResponse, len(media))
	for i, m := range media {
		result[i] = dto.ThreadMediaResponse{MediaID: m.MediaID, URL: m.URL}
	}
	return result
}
