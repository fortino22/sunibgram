package violationroute

import (
	"github.com/gofiber/fiber/v2"

	violationctrl "github.com/sunibgram/webapi/controller/violation"
	"github.com/sunibgram/webapi/middleware"
)

func RegisterViolationRoutes(api fiber.Router, violationCtrl *violationctrl.ViolationController, auth fiber.Handler) {
	api.Get("/violations", auth, violationCtrl.GetAll)
	api.Get("/violations/:id", auth, violationCtrl.GetByID)

	violations := api.Group("/violations", auth, middleware.AdminOnly())
	violations.Post("", violationCtrl.Create)
	violations.Put("/:id", violationCtrl.Update)
	violations.Delete("/:id", violationCtrl.Delete)
}
