package controller

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
)

type UserSettingsController struct {
	settingsService service.IUserSettingsService
}

func NewUserSettingsController(settingsService service.IUserSettingsService) *UserSettingsController {
	return &UserSettingsController{settingsService: settingsService}
}

func (c *UserSettingsController) GetUserSettingByUserID(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("userID"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid user id"})
	}

	setting, err := c.settingsService.GetUserSettingByUserID(ctx.Context(), uint(id))
	if err != nil {
		if err.Error() == "user settings not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"user_id":    setting.UserID,
		"is_private": setting.IsPrivate,
	})
}

func (c *UserSettingsController) UpdateUserSetting(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("userID"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid user id"})
	}

	var req struct {
		IsPrivate bool `json:"is_private"`
	}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}

	if err := c.settingsService.UpdateUserSetting(ctx.Context(), uint(id), req.IsPrivate); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "user settings updated successfully"})
}
