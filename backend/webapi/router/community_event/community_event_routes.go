package communityeventroute

import (
	"github.com/gofiber/fiber/v2"

	communityeventctrl "github.com/sunibgram/webapi/controller/community_event"
)

func RegisterCommunityEventRoutes(
	api fiber.Router,
	ctrl *communityeventctrl.CommunityEventController,
	auth fiber.Handler,
) {
	// Community-scoped: create + list events for one community.
	scoped := api.Group("/communities/:id/community-events", auth)
	scoped.Post("/", ctrl.Create)
	scoped.Get("/", ctrl.GetByCommunity)
	scoped.Get("/invitee-candidates", ctrl.SearchInviteeCandidates)

	// Standalone: event detail + comments + reactions (deep-linkable by event id).
	events := api.Group("/community-events", auth)
	events.Get("/:eventId", ctrl.GetByID)
	events.Put("/:eventId", ctrl.Update)
	events.Delete("/:eventId", ctrl.Delete)
	events.Get("/:eventId/comments", ctrl.GetComments)
	events.Post("/:eventId/comments", ctrl.AddComment)
	events.Put("/:eventId/comments/:commentId", ctrl.UpdateComment)
	events.Delete("/:eventId/comments/:commentId", ctrl.DeleteComment)
	events.Post("/:eventId/comments/:commentId/pin", ctrl.PinComment)
	events.Delete("/:eventId/comments/:commentId/pin", ctrl.UnpinComment)
	events.Put("/:eventId/reaction", ctrl.SetReaction)
	events.Delete("/:eventId/reaction", ctrl.RemoveReaction)
}
