package authroute

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/sunibgram/infrastructure/cache"
	"github.com/sunibgram/infrastructure/ratelimit"
	authctrl "github.com/sunibgram/webapi/controller/auth"
	authdto "github.com/sunibgram/webapi/dto/auth"
)

func RegisterAuthRoutes(api fiber.Router, ctrl *authctrl.AuthController, authMW fiber.Handler, c *cache.Cache) {
	authIPLimiter := ratelimit.New(c, "auth_ip", 100, 5*time.Minute)
	authIPRateLimit := func(c *fiber.Ctx) error {
		if err := authIPLimiter.Allow(c.IP()); err != nil {
			return c.Status(fiber.StatusTooManyRequests).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		return c.Next()
	}

	auth := api.Group("/auth")
	auth.Post("/register", authIPRateLimit, ctrl.Register)
	auth.Post("/login", authIPRateLimit, ctrl.Login)
	auth.Get("/login/verify", ctrl.ConfirmLogin)
	auth.Post("/refresh", ctrl.Refresh)
	auth.Post("/logout", authMW, ctrl.Logout)
	auth.Put("/change-password", authMW, ctrl.ChangePassword)
	auth.Post("/forgot-password", authIPRateLimit, ctrl.ForgotPassword)
	auth.Get("/username-available", authIPRateLimit, ctrl.CheckUsername)
}
