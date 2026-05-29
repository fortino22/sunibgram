package trendreportroute

import (
	"github.com/gofiber/fiber/v2"

	trendreportctrl "github.com/sunibgram/webapi/controller/trend_report"
	"github.com/sunibgram/webapi/middleware"
)

func RegisterTrendReportRoutes(api fiber.Router, ctrl *trendreportctrl.TrendReportController, auth fiber.Handler) {
	api.Get("/trend-reports/reasons", ctrl.ListReasons)
	api.Get("/trend-reports/stats", auth, middleware.AdminOnly(), ctrl.Stats)
	api.Post("/trend-reports", auth, ctrl.Create)
}
