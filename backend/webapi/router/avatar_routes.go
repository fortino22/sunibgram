package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/webapi/controller"
)

func registerAvatarRoutes(api fiber.Router, ctrl *controller.AvatarController, auth fiber.Handler) {
	api.Post("/avatar", auth, ctrl.UploadAvatar)
	api.Put("/avatar", auth, ctrl.UpdateAvatar)
	api.Get("/avatar", auth, ctrl.GetAvatar)
	api.Delete("/avatar", auth, ctrl.DeleteAvatar)
}
