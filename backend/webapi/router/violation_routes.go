package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sunibgram/webapi/controller"
	"github.com/sunibgram/webapi/middleware"
)

func registerViolationRoutes(api fiber.Router, violationCtrl *controller.ViolationController, auth fiber.Handler) {
	api.Get("/violations", auth, violationCtrl.GetAll)
	api.Get("/violations/:id", auth, violationCtrl.GetByID)

	violations := api.Group("/violations", auth, middleware.AdminOnly())
	violations.Post("", violationCtrl.Create)
	violations.Put("/:id", violationCtrl.Update)
	violations.Delete("/:id", violationCtrl.Delete)
}
