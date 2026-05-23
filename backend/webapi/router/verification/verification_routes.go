package verificationroute

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/sunibgram/infrastructure/cache"
	"github.com/sunibgram/infrastructure/ratelimit"
	verificationctrl "github.com/sunibgram/webapi/controller/verification"
	authdto "github.com/sunibgram/webapi/dto/auth"
)

func RegisterVerificationRoutes(api fiber.Router, ctrl *verificationctrl.VerificationController, c *cache.Cache) {
	ipLimiter := ratelimit.New(c, "verification_ip", 10, 10*time.Minute)
	ipRateLimit := func(c *fiber.Ctx) error {
		if err := ipLimiter.Allow(c.IP()); err != nil {
			return c.Status(fiber.StatusTooManyRequests).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		return c.Next()
	}

	v := api.Group("/verification")
	v.Post("/send", ipRateLimit, ctrl.SendCode)
	v.Post("/verify", ipRateLimit, ctrl.VerifyCode)
}
