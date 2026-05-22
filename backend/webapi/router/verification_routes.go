package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sunibgram/webapi/controller"
)

func registerVerificationRoutes(api fiber.Router, ctrl *controller.VerificationController) {
	v := api.Group("/verification")
	v.Post("/send", ctrl.SendCode)
	v.Post("/verify", ctrl.VerifyCode)
}
