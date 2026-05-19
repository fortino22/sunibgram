package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/dto/utils"
)

type ViolationController struct {
	violationService service.IViolationService
}

func NewViolationController(violationService service.IViolationService) *ViolationController {
	return &ViolationController{violationService: violationService}
}

func (c *ViolationController) Create(ctx *fiber.Ctx) error {
	var req dto.CreateViolationRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}
	if err := c.violationService.Create(req.Name, req.Description); err != nil {
		return HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "violation created successfully"})
}

func (c *ViolationController) GetAll(ctx *fiber.Ctx) error {
	violations, err := c.violationService.GetAll()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": toViolationResponses(violations)})
}

func (c *ViolationController) GetByID(ctx *fiber.Ctx) error {
	id, ok := ParseUUIDParam(ctx, "id", "violation")
	if !ok {
		return nil
	}
	violation, err := c.violationService.GetByID(id)
	if err != nil {
		return HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(toViolationResponse(violation))
}

func (c *ViolationController) Update(ctx *fiber.Ctx) error {
	id, ok := ParseUUIDParam(ctx, "id", "violation")
	if !ok {
		return nil
	}

	var req dto.UpdateViolationRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}
	if err := c.violationService.Update(id, req.Name, req.Description); err != nil {
		return HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "violation updated successfully"})
}

func (c *ViolationController) Delete(ctx *fiber.Ctx) error {
	id, ok := ParseUUIDParam(ctx, "id", "violation")
	if !ok {
		return nil
	}
	if err := c.violationService.Delete(id); err != nil {
		return HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "violation deleted successfully"})
}

func toViolationResponse(violation *model.Violation) dto.ViolationResponse {
	return dto.ViolationResponse{
		ID:          violation.ID.String(),
		Name:        violation.Name,
		Description: violation.Description,
		CreatedAt:   utils.Wrap(violation.CreatedAt),
		UpdatedAt:   utils.Wrap(violation.UpdatedAt),
	}
}

func toViolationResponses(violations []model.Violation) []dto.ViolationResponse {
	responses := make([]dto.ViolationResponse, len(violations))
	for i := range violations {
		responses[i] = toViolationResponse(&violations[i])
	}
	return responses
}
