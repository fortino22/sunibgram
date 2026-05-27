package authroute

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/sunibgram/infrastructure/cache"
	"github.com/sunibgram/infrastructure/ratelimit"
	authctrl "github.com/sunibgram/webapi/controller/auth"
	verificationctrl "github.com/sunibgram/webapi/controller/verification"
	authdto "github.com/sunibgram/webapi/dto/auth"
)

func RegisterAuthRoutes(api fiber.Router, ctrl *authctrl.AuthController, verificationCtrl *verificationctrl.VerificationController, authMW fiber.Handler, c *cache.Cache) {
	// otpSendLimiter := ratelimit.New(c, "auth_otp_send_ip", 3, 10*time.Minute)
	// otpSendRateLimit := func(c *fiber.Ctx) error {
	// 	if err := otpSendLimiter.Allow(c.IP()); err != nil {
	// 		return c.Status(fiber.StatusTooManyRequests).JSON(authdto.ErrorResponse{Message: err.Error()})
	// 	}
	// 	return c.Next()
	// }

	authIPLimiter := ratelimit.New(c, "auth_ip", 100, 5*time.Minute)
	authIPRateLimit := func(c *fiber.Ctx) error {
		if err := authIPLimiter.Allow(c.IP()); err != nil {
			return c.Status(fiber.StatusTooManyRequests).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		return c.Next()
	}

	auth := api.Group("/auth")
	auth.Post("/register", authIPRateLimit, ctrl.Register)
	auth.Post("/login/otp/send", verificationCtrl.SendLoginOTP)
	auth.Post("/login", ctrl.LoginWithOTP)
	auth.Post("/login/prod", authIPRateLimit, ctrl.LoginProd)
	auth.Post("/refresh", ctrl.Refresh)
	auth.Post("/logout", authMW, ctrl.Logout)
	auth.Put("/change-password", authMW, ctrl.ChangePassword)
	auth.Post("/forgot-password", authIPRateLimit, ctrl.ForgotPassword)
}
