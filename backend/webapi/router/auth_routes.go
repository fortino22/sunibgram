package router

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/infrastructure/ratelimit"
	"github.com/sunibgram/webapi/controller"
)

func registerAuthRoutes(api fiber.Router, ctrl *controller.AuthController, verificationCtrl *controller.VerificationController, authMW fiber.Handler) {
	otpSendLimiter := ratelimit.New(3, 10*time.Minute)
	otpSendRateLimit := func(c *fiber.Ctx) error {
		if err := otpSendLimiter.Allow(c.IP()); err != nil {
			return c.Status(fiber.StatusTooManyRequests).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return c.Next()
	}

	auth := api.Group("/auth")
	auth.Post("/register", ctrl.Register)
	auth.Post("/login/otp/send", otpSendRateLimit, verificationCtrl.SendLoginOTP)
	auth.Post("/login", ctrl.LoginWithOTP)
	auth.Post("/login/prod", ctrl.LoginProd)
	auth.Post("/refresh", ctrl.Refresh)
	auth.Post("/logout", authMW, ctrl.Logout)
	auth.Put("/change-password", authMW, ctrl.ChangePassword)
	auth.Post("/forgot-password", ctrl.ForgotPassword)
}
