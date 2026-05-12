package controller

import (
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
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	setting, err := c.settingsService.GetUserSettingByUserID(ctx.Context(), claims.UserID)
	if err != nil {
		if err.Error() == "user settings not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.UserSettingsResponse{
		UserID:    setting.UserID,
		IsPrivate: setting.IsPrivate,
	})
}
func (c *UserSettingsController) UpdateUserSetting(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	var req dto.UpdateUserSettingsRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}

	if err := c.settingsService.UpdateUserSetting(ctx.Context(), claims.UserID, req.IsPrivate); err != nil {
		if err.Error() == "only the owner can update this setting" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "user settings updated successfully"})
}
