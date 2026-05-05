package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/webapi/controller"
	"github.com/sunibgram/webapi/middleware"
)

func registerOrderRoutes(api fiber.Router, orderCtrl *controller.OrderController, auth fiber.Handler) {
	orders := api.Group("/orders")

	orders.Post("/notification", orderCtrl.MidtransNotification)

	orders.Use(auth)
	orders.Post("", orderCtrl.Create)
	orders.Post("/premium", orderCtrl.BuyPremium)
	orders.Get("/me", orderCtrl.GetMyOrders)
	orders.Get("/:id", orderCtrl.GetByID)
	orders.Patch("/:id/cancel", orderCtrl.Cancel)

	orders.Get("", middleware.AdminOnly(), orderCtrl.GetAll)
}
