package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/webapi/controller"
)

func registerEventRoutes(
	api fiber.Router,
	eventCtrl *controller.EventController,
	auth fiber.Handler,
) {
	events := api.Group("/events", auth)
	events.Post("/", eventCtrl.Create)
	events.Get("/", eventCtrl.GetAll)
	events.Get("/pending", eventCtrl.GetPending)
	events.Get("/:id", eventCtrl.GetByID)
	events.Put("/:id", eventCtrl.Update)
	events.Patch("/:id/approve", eventCtrl.ApproveEvent)
	events.Patch("/:id/decline", eventCtrl.DeclineEvent)
	events.Delete("/:id", eventCtrl.Delete)
	events.Get("/:id/threads", eventCtrl.GetThreads)
}
