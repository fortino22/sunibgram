package orderroute

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/sunibgram/infrastructure/cache"
	"github.com/sunibgram/infrastructure/ratelimit"
	orderctrl "github.com/sunibgram/webapi/controller/order"
	authdto "github.com/sunibgram/webapi/dto/auth"
	"github.com/sunibgram/webapi/middleware"
)

func RegisterOrderRoutes(api fiber.Router, orderCtrl *orderctrl.OrderController, auth fiber.Handler, c *cache.Cache) {
	orders := api.Group("/orders")

	webhookLimiter := ratelimit.New(c, "midtrans_webhook", 60, time.Minute)
	webhookRateLimit := func(c *fiber.Ctx) error {
		if err := webhookLimiter.Allow(c.IP()); err != nil {
			return c.Status(fiber.StatusTooManyRequests).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		return c.Next()
	}

	orders.Post("/notification", webhookRateLimit, orderCtrl.MidtransNotification)

	orders.Use(auth)
	// orders.Post("/premium", orderCtrl.BuyPremium)
	// orders.Get("/me", orderCtrl.GetMyOrders)
	// orders.Patch("/:id/cancel", orderCtrl.Cancel)

	orders.Get("", middleware.AdminOnly(), orderCtrl.GetAll)
	orders.Get("/:id", middleware.AdminOnly(), orderCtrl.GetByID)
}
