package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sunibgram/webapi/controller"
)

func registerEmailRoutes(api fiber.Router, ctrl *controller.EmailController, auth fiber.Handler) {
	api.Post("/email/send", auth, ctrl.Send)
}
