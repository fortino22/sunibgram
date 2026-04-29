package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/webapi/controller"
)

func registerAuthRoutes(api fiber.Router, ctrl *controller.AuthController) {
	auth := api.Group("/auth")
	auth.Post("/register", ctrl.Register)
	auth.Post("/login", ctrl.Login)
	auth.Post("/login/prod", ctrl.LoginProd)
	auth.Post("/refresh", ctrl.Refresh)
}
