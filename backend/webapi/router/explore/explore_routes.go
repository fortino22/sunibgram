package exploreroute

import (
	"github.com/gofiber/fiber/v2"

	explorectrl "github.com/sunibgram/webapi/controller/explore"
)

func RegisterExploreRoutes(api fiber.Router, ctrl *explorectrl.ExploreController, auth fiber.Handler) {
	api.Get("/explore/search", auth, ctrl.ExploreSearch)
	api.Get("/explore/top", auth, ctrl.GetTop)
	api.Get("/explore/latest", auth, ctrl.GetLatest)
	api.Get("/explore/people", auth, ctrl.GetPeople)
	api.Get("/explore/media", auth, ctrl.GetMedia)
	api.Get("/explore/community", auth, ctrl.GetCommunities)
	api.Get("/explore/communities", auth, ctrl.GetCommunities)
}
