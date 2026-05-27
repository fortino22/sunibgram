package eventroute

import (
	"github.com/gofiber/fiber/v2"

	eventctrl "github.com/sunibgram/webapi/controller/event"
	eventparticipantctrl "github.com/sunibgram/webapi/controller/event_participant"
	eventpinctrl "github.com/sunibgram/webapi/controller/event_pin"
	eventthreadpinctrl "github.com/sunibgram/webapi/controller/event_thread_pin"
)

func RegisterEventRoutes(
	api fiber.Router,
	eventCtrl *eventctrl.EventController,
	pinCtrl *eventpinctrl.EventPinController,
	participantCtrl *eventparticipantctrl.EventParticipantController,
	threadPinCtrl *eventthreadpinctrl.EventThreadPinController,
	auth fiber.Handler,
) {
	events := api.Group("/events", auth)
	events.Post("/", eventCtrl.Create)
	events.Get("/", eventCtrl.GetAll)
	events.Get("/trending", eventCtrl.GetTrending)
	events.Get("/admin", eventCtrl.GetPending)
	events.Get("/me", eventCtrl.GetMyEvents)
	events.Get("/pins/me", pinCtrl.GetMyPins)
	events.Get("/joined/me", participantCtrl.GetMyJoined)
	events.Get("/:id", eventCtrl.GetByID)
	events.Put("/:id", eventCtrl.Update)
	events.Patch("/:id/approve", eventCtrl.ApproveEvent)
	events.Patch("/:id/decline", eventCtrl.DeclineEvent)
	events.Patch("/:id/revision", eventCtrl.RequestRevision)
	events.Delete("/:id", eventCtrl.Delete)
	events.Get("/:id/threads", eventCtrl.GetThreads)
	events.Post("/:id/threads/:threadId/pin", threadPinCtrl.Pin)
	events.Delete("/:id/threads/:threadId/unpin", threadPinCtrl.Unpin)
	events.Post("/:id/pin", pinCtrl.Pin)
	events.Delete("/:id/unpin", pinCtrl.Unpin)
	events.Post("/:id/join", participantCtrl.Join)
	events.Delete("/:id/leave", participantCtrl.Leave)
}
