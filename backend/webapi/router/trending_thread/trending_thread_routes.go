package trendingthreadroute

import (
	"github.com/gofiber/fiber/v2"

	trendingthreadctrl "github.com/sunibgram/webapi/controller/trending_thread"
	"github.com/sunibgram/webapi/middleware"
)

func RegisterTrendingThreadRoutes(api fiber.Router, ctrl *trendingthreadctrl.TrendingThreadController, auth fiber.Handler) {
	// Admin-only: auth first (sets claims), then AdminOnly() gates by role_name.
	api.Get("/trending-threads", auth, middleware.AdminOnly(), ctrl.GetTopWeekly)
}
