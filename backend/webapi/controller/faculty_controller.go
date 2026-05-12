package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
)

type FacultyController struct {
	facultyService service.IFacultyService
}

func NewFacultyController(facultyService service.IFacultyService) *FacultyController {
	return &FacultyController{facultyService: facultyService}
}

func (c *FacultyController) Create(ctx *fiber.Ctx) error {
	var req dto.CreateFacultyRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}
	if err := c.facultyService.Create(req.FacultyName); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "faculty created successfully"})
}

func (c *FacultyController) GetAll(ctx *fiber.Ctx) error {
	page, limit := ParsePagination(ctx)
	faculties, total, err := c.facultyService.GetAll(page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	data := make([]dto.FacultyResponse, len(faculties))
	for i, f := range faculties {
		data[i] = dto.FacultyResponse{
			FacultyID:   f.FacultyID,
			FacultyName: f.FacultyName,
			TotalUsers:  f.TotalUsers,
		}
	}
	return ctx.JSON(dto.PaginatedFacultyResponse{
		Data: data,
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *FacultyController) GetByID(ctx *fiber.Ctx) error {
	id, ok := ParseIDParam(ctx, "id", "faculty")
	if !ok {
		return nil
	}
	faculty, err := c.facultyService.GetByID(id)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	return ctx.JSON(dto.FacultyResponse{
		FacultyID:   faculty.FacultyID,
		FacultyName: faculty.FacultyName,
		TotalUsers:  faculty.TotalUsers,
	})
}

func (c *FacultyController) Update(ctx *fiber.Ctx) error {
	id, ok := ParseIDParam(ctx, "id", "faculty")
	if !ok {
		return nil
	}
	var req dto.UpdateFacultyRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}
	if err := c.facultyService.Update(id, req.FacultyName); err != nil {
		if err.Error() == "faculty not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	return ctx.JSON(fiber.Map{"message": "faculty updated successfully"})
}

func (c *FacultyController) Delete(ctx *fiber.Ctx) error {
	id, ok := ParseIDParam(ctx, "id", "faculty")
	if !ok {
		return nil
	}
	if err := c.facultyService.Delete(id); err != nil {
		if err.Error() == "faculty not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	return ctx.JSON(fiber.Map{"message": "faculty deleted successfully"})
}
