package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/webapi/controller"
)

func registerUserSettingsRoutes(api fiber.Router, ctrl *controller.UserSettingsController, auth fiber.Handler) {
	settings := api.Group("/user-settings", auth)
	settings.Get("/me", ctrl.GetUserSettingByUserID)
	settings.Put("/me", ctrl.UpdateUserSetting)
}
