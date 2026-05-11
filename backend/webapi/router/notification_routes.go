package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/webapi/controller"
)

func registerNotificationRoutes(api fiber.Router, notificationCtrl *controller.NotificationController, auth fiber.Handler) {
	notifications := api.Group("/notifications")
	notifications.Use(auth)
	notifications.Get("", notificationCtrl.GetMine)
	// notifications.Post("", notificationCtrl.Send)
	notifications.Delete("/:id", notificationCtrl.Delete)
}
