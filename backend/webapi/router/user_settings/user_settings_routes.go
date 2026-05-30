package usersettingsroute

import (
	"github.com/gofiber/fiber/v2"

	usersettingsctrl "github.com/sunibgram/webapi/controller/user_settings"
)

func RegisterUserSettingsRoutes(api fiber.Router, ctrl *usersettingsctrl.UserSettingsController, auth fiber.Handler) {
	settings := api.Group("/user-settings", auth)
	settings.Get("/me", ctrl.GetUserSettingByUserID)
	settings.Put("/me", ctrl.UpdateUserSetting)
	settings.Put("/me/theme", ctrl.UpdateUserTheme)
}
