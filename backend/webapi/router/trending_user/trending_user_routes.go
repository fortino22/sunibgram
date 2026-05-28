package trendinguserroute

import (
	"github.com/gofiber/fiber/v2"

	trendinguserctrl "github.com/sunibgram/webapi/controller/trending_user"
)

func RegisterTrendingUserRoutes(api fiber.Router, ctrl *trendinguserctrl.TrendingUserController, auth fiber.Handler) {
	api.Get("/trending-users", auth, ctrl.GetWeeklyTrending)
}
