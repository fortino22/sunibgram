package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
)

type RoleController struct {
	roleService service.IRoleService
}

func NewRoleController(roleService service.IRoleService) *RoleController {
	return &RoleController{roleService: roleService}
}

func (c *RoleController) Create(ctx *fiber.Ctx) error {
	var req dto.CreateRoleRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}

	role := &model.Role{Name: req.Name}
	if err := c.roleService.Create(role); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusCreated).JSON(dto.RoleResponse{RoleID: role.RoleID, Name: role.Name})
}

func (c *RoleController) GetAll(ctx *fiber.Ctx) error {
	roles, err := c.roleService.GetAll()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: "internal server error"})
	}

	result := make([]dto.RoleResponse, len(roles))
	for i, r := range roles {
		result[i] = dto.RoleResponse{RoleID: r.RoleID, Name: r.Name}
	}

	return ctx.Status(fiber.StatusOK).JSON(result)
}

func (c *RoleController) GetByID(ctx *fiber.Ctx) error {
	id, ok := ParseIDParam(ctx, "id", "role")
	if !ok {
		return nil
	}

	role, err := c.roleService.GetByID(id)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: "role not found"})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.RoleResponse{RoleID: role.RoleID, Name: role.Name})
}

func (c *RoleController) Update(ctx *fiber.Ctx) error {
	id, ok := ParseIDParam(ctx, "id", "role")
	if !ok {
		return nil
	}

	var req dto.UpdateRoleRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}

	role := &model.Role{RoleID: id, Name: req.Name}
	if err := c.roleService.Update(role); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.RoleResponse{RoleID: role.RoleID, Name: role.Name})
}

func (c *RoleController) Delete(ctx *fiber.Ctx) error {
	id, ok := ParseIDParam(ctx, "id", "role")
	if !ok {
		return nil
	}

	if err := c.roleService.Delete(id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: "internal server error"})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "role deleted successfully"})
}
