package facultyctrl

import (
	"github.com/gofiber/fiber/v2"

	facultysvciface "github.com/sunibgram/domain/service/faculty"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	facultydto "github.com/sunibgram/webapi/dto/faculty"
	threaddto "github.com/sunibgram/webapi/dto/thread"
)

type FacultyController struct {
	facultyService facultysvciface.IFacultyService
}

func NewFacultyController(facultyService facultysvciface.IFacultyService) *FacultyController {
	return &FacultyController{facultyService: facultyService}
}

func (c *FacultyController) Create(ctx *fiber.Ctx) error {
	var req facultydto.CreateFacultyRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}
	if err := c.facultyService.Create(req.FacultyName); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "faculty created successfully"})
}

func (c *FacultyController) GetAll(ctx *fiber.Ctx) error {
	page, limit := ctrlshared.ParsePagination(ctx)
	faculties, total, err := c.facultyService.GetAll(page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
	data := make([]facultydto.FacultyResponse, len(faculties))
	for i, f := range faculties {
		data[i] = facultydto.FacultyResponse{
			FacultyID:   f.FacultyID,
			FacultyName: f.FacultyName,
			TotalUsers:  f.TotalUsers,
		}
	}
	return ctx.JSON(facultydto.PaginatedFacultyResponse{
		Data: data,
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *FacultyController) GetByID(ctx *fiber.Ctx) error {
	id, ok := ctrlshared.ParseIDParam(ctx, "id", "faculty")
	if !ok {
		return nil
	}
	faculty, err := c.facultyService.GetByID(id)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
	return ctx.JSON(facultydto.FacultyResponse{
		FacultyID:   faculty.FacultyID,
		FacultyName: faculty.FacultyName,
		TotalUsers:  faculty.TotalUsers,
	})
}

func (c *FacultyController) Update(ctx *fiber.Ctx) error {
	id, ok := ctrlshared.ParseIDParam(ctx, "id", "faculty")
	if !ok {
		return nil
	}
	var req facultydto.UpdateFacultyRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}
	if err := c.facultyService.Update(id, req.FacultyName); err != nil {
		if err.Error() == "faculty not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
	return ctx.JSON(fiber.Map{"message": "faculty updated successfully"})
}

func (c *FacultyController) Delete(ctx *fiber.Ctx) error {
	id, ok := ctrlshared.ParseIDParam(ctx, "id", "faculty")
	if !ok {
		return nil
	}
	if err := c.facultyService.Delete(id); err != nil {
		if err.Error() == "faculty not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
	return ctx.JSON(fiber.Map{"message": "faculty deleted successfully"})
}
