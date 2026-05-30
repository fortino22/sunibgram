package usersettingsctrl

import (
	"github.com/gofiber/fiber/v2"

	usersettingssvciface "github.com/sunibgram/domain/service/user_settings"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	usersettingsdto "github.com/sunibgram/webapi/dto/user_settings"
)

type UserSettingsController struct {
	settingsService usersettingssvciface.IUserSettingsService
}

func NewUserSettingsController(settingsService usersettingssvciface.IUserSettingsService) *UserSettingsController {
	return &UserSettingsController{settingsService: settingsService}
}

func (c *UserSettingsController) GetUserSettingByUserID(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	setting, err := c.settingsService.GetUserSettingByUserID(ctx.Context(), claims.UserID)
	if err != nil {
		if err.Error() == "user settings not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		return ctrlshared.HandleServiceError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(usersettingsdto.UserSettingsResponse{
		UserID:    setting.UserID,
		IsPrivate: setting.IsPrivate,
		UserTheme: string(setting.UserTheme),
	})
}
func (c *UserSettingsController) UpdateUserSetting(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
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
		return ctrlshared.HandleServiceError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "user settings updated successfully"})
}

func (c *UserSettingsController) UpdateUserTheme(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	var req usersettingsdto.UpdateUserThemeRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}

	if err := c.settingsService.UpdateUserTheme(ctx.Context(), claims.UserID, req.UserTheme); err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "user theme updated successfully"})
}
