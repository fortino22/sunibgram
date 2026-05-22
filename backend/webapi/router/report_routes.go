package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sunibgram/webapi/controller"
	"github.com/sunibgram/webapi/middleware"
)

func registerReportRoutes(api fiber.Router, reportCtrl *controller.ReportController, auth fiber.Handler) {
	api.Post("/reports", auth, reportCtrl.Create)

	reports := api.Group("/reports", auth, middleware.AdminOnly())
	reports.Get("", reportCtrl.GetAll)
	reports.Get("/:id", reportCtrl.GetByID)
	reports.Put("/:id/status", reportCtrl.UpdateStatus)
	reports.Delete("/:id", reportCtrl.Delete)
}
