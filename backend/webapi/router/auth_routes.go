package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/webapi/controller"
)

func registerAuthRoutes(api fiber.Router, ctrl *controller.AuthController, authMW fiber.Handler) {
	auth := api.Group("/auth")
	auth.Post("/register", ctrl.Register)
	auth.Post("/login", ctrl.Login)
	auth.Post("/login/prod", ctrl.LoginProd)
	auth.Post("/refresh", ctrl.Refresh)
	auth.Post("/logout", authMW, ctrl.Logout)
	auth.Put("/change-password", authMW, ctrl.ChangePassword)
	auth.Post("/forgot-password", ctrl.ForgotPassword)
}
