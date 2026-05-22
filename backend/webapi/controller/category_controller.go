package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	utils "github.com/sunibgram/webapi/dto/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	categorydto "github.com/sunibgram/webapi/dto/category"
)

type CategoryController struct {
	categoryService service.ICategoryService
}

func NewCategoryController(categoryService service.ICategoryService) *CategoryController {
	return &CategoryController{categoryService: categoryService}
}

func (c *CategoryController) Create(ctx *fiber.Ctx) error {
	var req categorydto.CreateCategoryRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}
	if err := c.categoryService.Create(req.Name); err != nil {
		return HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "category created successfully"})
}

func (c *CategoryController) GetAll(ctx *fiber.Ctx) error {
	categories, err := c.categoryService.GetAll()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": toCategoryResponses(categories)})
}

func (c *CategoryController) Update(ctx *fiber.Ctx) error {
	id, ok := ParseUUIDParam(ctx, "id", "category")
	if !ok {
		return nil
	}

	var req categorydto.UpdateCategoryRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}
	if err := c.categoryService.Update(id, req.Name); err != nil {
		return HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "category updated successfully"})
}

func (c *CategoryController) Delete(ctx *fiber.Ctx) error {
	id, ok := ParseUUIDParam(ctx, "id", "category")
	if !ok {
		return nil
	}

	if err := c.categoryService.Delete(id); err != nil {
		return HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "category deleted successfully"})
}

func ParseUUIDParam(ctx *fiber.Ctx, paramName string, entityName string) (uuid.UUID, bool) {
	id, err := uuid.Parse(ctx.Params(paramName))
	if err != nil {
		_ = ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid " + entityName + " id"})
		return uuid.Nil, false
	}
	return id, true
}

func toCategoryResponse(category *model.Category) categorydto.CategoryResponse {
	return categorydto.CategoryResponse{
		ID:        category.ID.String(),
		Name:      category.Name,
		CreatedAt: utils.Wrap(category.CreatedAt),
	}
}

func toCategoryResponses(categories []model.Category) []categorydto.CategoryResponse {
	responses := make([]categorydto.CategoryResponse, len(categories))
	for i := range categories {
		responses[i] = toCategoryResponse(&categories[i])
	}
	return responses
}
