package notificationroute

import (
	"github.com/gofiber/fiber/v2"

	notificationctrl "github.com/sunibgram/webapi/controller/notification"
)

func RegisterNotificationRoutes(api fiber.Router, notificationCtrl *notificationctrl.NotificationController, auth fiber.Handler) {
	notifications := api.Group("/notifications")
	notifications.Use(auth)
	notifications.Get("", notificationCtrl.GetMine)
	// notifications.Post("", notificationCtrl.Send)
	notifications.Delete("/:id", notificationCtrl.Delete)
}
