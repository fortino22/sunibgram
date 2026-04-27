package controller

import (
	"strconv"

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
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	result := make([]dto.RoleResponse, len(roles))
	for i, r := range roles {
		result[i] = dto.RoleResponse{RoleID: r.RoleID, Name: r.Name}
	}

	return ctx.Status(fiber.StatusOK).JSON(result)
}

func (c *RoleController) GetByID(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid role id"})
	}

	role, err := c.roleService.GetByID(uint(id))
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: "role not found"})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.RoleResponse{RoleID: role.RoleID, Name: role.Name})
}

func (c *RoleController) Update(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid role id"})
	}

	var req dto.UpdateRoleRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}

	role := &model.Role{RoleID: uint(id), Name: req.Name}
	if err := c.roleService.Update(role); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.RoleResponse{RoleID: role.RoleID, Name: role.Name})
}

func (c *RoleController) Delete(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid role id"})
	}

	if err := c.roleService.Delete(uint(id)); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "role deleted successfully"})
}
