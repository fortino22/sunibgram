package trendinguserroute

import (
	"github.com/gofiber/fiber/v2"

	trendinguserctrl "github.com/sunibgram/webapi/controller/trending_user"
)

func RegisterTrendingUserRoutes(api fiber.Router, ctrl *trendinguserctrl.TrendingUserController, auth fiber.Handler) {
	api.Get("/trending-users", auth, ctrl.GetWeeklyTrending)
	api.Get("/trending-users/my-current", auth, ctrl.GetMyCurrent)
	api.Get("/trending-users/my-history", auth, ctrl.GetMyHistory)
	api.Get("/trending-users/mutuals", auth, ctrl.SearchMutualUsers)
	api.Get("/trending-users/users/:userId/history", auth, ctrl.GetUserHistory)
	api.Get("/trending-users/snapshots/:snapshotId", auth, ctrl.GetMySnapshot)
}
