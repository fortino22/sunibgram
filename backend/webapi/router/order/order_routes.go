package orderroute

import (
	"github.com/gofiber/fiber/v2"

	orderctrl "github.com/sunibgram/webapi/controller/order"
	"github.com/sunibgram/webapi/middleware"
)

func RegisterOrderRoutes(api fiber.Router, orderCtrl *orderctrl.OrderController, auth fiber.Handler) {
	orders := api.Group("/orders")

	orders.Post("/notification", orderCtrl.MidtransNotification)

	orders.Use(auth)
	// orders.Post("/premium", orderCtrl.BuyPremium)
	// orders.Get("/me", orderCtrl.GetMyOrders)
	// orders.Patch("/:id/cancel", orderCtrl.Cancel)

	orders.Get("", middleware.AdminOnly(), orderCtrl.GetAll)
	orders.Get("/:id", middleware.AdminOnly(), orderCtrl.GetByID)
}
