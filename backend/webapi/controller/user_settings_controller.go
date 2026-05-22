package controller

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sunibgram/domain/service"
	authdto "github.com/sunibgram/webapi/dto/auth"
	usersettingsdto "github.com/sunibgram/webapi/dto/user_settings"
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
			return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(usersettingsdto.UserSettingsResponse{
		UserID:    setting.UserID,
		IsPrivate: setting.IsPrivate,
	})
}
func (c *UserSettingsController) UpdateUserSetting(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	var req usersettingsdto.UpdateUserSettingsRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}

	if err := c.settingsService.UpdateUserSetting(ctx.Context(), claims.UserID, req.IsPrivate); err != nil {
		if err.Error() == "only the owner can update this setting" {
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "user settings updated successfully"})
}
