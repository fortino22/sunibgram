package reportroute

import (
	"github.com/gofiber/fiber/v2"

	reportctrl "github.com/sunibgram/webapi/controller/report"
	"github.com/sunibgram/webapi/middleware"
)

func RegisterReportRoutes(api fiber.Router, reportCtrl *reportctrl.ReportController, auth fiber.Handler) {
	api.Post("/reports", auth, reportCtrl.Create)

	reports := api.Group("/reports", auth, middleware.AdminOnly())
	reports.Get("", reportCtrl.GetAll)
	// Static route registered before "/:id" so it isn't captured as an id.
	reports.Get("/reported", reportCtrl.GetReported)
	reports.Get("/:id", reportCtrl.GetByID)
	reports.Put("/:id/status", reportCtrl.UpdateStatus)
	reports.Delete("/:id", reportCtrl.Delete)
}
