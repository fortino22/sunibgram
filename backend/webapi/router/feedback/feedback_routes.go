package feedbackroute

import (
	"github.com/gofiber/fiber/v2"

	feedbackctrl "github.com/sunibgram/webapi/controller/feedback"
	"github.com/sunibgram/webapi/middleware"
)

func RegisterFeedbackRoutes(api fiber.Router, ctrl *feedbackctrl.FeedbackController, auth fiber.Handler) {
	feedback := api.Group("/feedback", auth)

	feedback.Get("", ctrl.List)
	feedback.Get("/:id", ctrl.GetByID)
	feedback.Post("", ctrl.Create)
	feedback.Delete("/:id", ctrl.Delete)

	feedback.Post("/:id/like", ctrl.Like)
	feedback.Post("/:id/dislike", ctrl.Dislike)
	feedback.Delete("/:id/vote", ctrl.ClearVote)

	feedback.Patch("/:id/status", middleware.AdminOnly(), ctrl.UpdateStatus)
}
