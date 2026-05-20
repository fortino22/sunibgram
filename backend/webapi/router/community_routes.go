package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/webapi/controller"
)

func registerCommunityRoutes(
	api fiber.Router,
	communityCtrl *controller.CommunityController,
	memberCtrl *controller.CommunityMemberController,
	requestCtrl *controller.CommunityRequestController,
	communityThreadCtrl *controller.CommunityThreadController,
	ruleCtrl *controller.CommunityRuleController,
	auth fiber.Handler,
) {
	api.Get("/communities/:id/rules", ruleCtrl.List)

	communities := api.Group("/communities", auth)

	//communities
	communities.Post("/", communityCtrl.Create)
	communities.Get("/discover", communityCtrl.Discover)
	communities.Get("/search", communityCtrl.SearchByName)
	communities.Get("/me", communityCtrl.GetMyCreated)
	communities.Get("/joined", communityCtrl.GetMyJoined)

	//community threads (static paths before /:id wildcard)
	communities.Get("/me/threads", communityThreadCtrl.GetMyCommunityThreads)
	communities.Get("/threads", communityThreadCtrl.GetAllCommunityThreads)
	communities.Get("/threads/top", communityThreadCtrl.GetCommunityThreadTopByName)
	communities.Get("/threads/latest", communityThreadCtrl.GetCommunityThreadLatestByName)

	communities.Get("/:id", communityCtrl.GetByID)
	communities.Put("/:id", communityCtrl.Update)
	communities.Delete("/:id", communityCtrl.Delete)

	//community members
	communities.Post("/:id/join", memberCtrl.Join)
	communities.Delete("/:id/leave", memberCtrl.Leave)
	communities.Get("/:id/members", memberCtrl.GetMembers)
	communities.Delete("/:id/members/:userId", memberCtrl.RemoveMember)
	communities.Patch("/:id/members/:userId/role", memberCtrl.UpdateRole)

	//community join requests
	communities.Post("/:id/requests", requestCtrl.Send)
	communities.Delete("/:id/requests", requestCtrl.Cancel)
	communities.Get("/:id/requests", requestCtrl.GetPending)
	communities.Post("/:id/requests/:userId/approve", requestCtrl.Approve)
	communities.Post("/:id/requests/:userId/decline", requestCtrl.Decline)
	communities.Post("/:communityID/threads", communityThreadCtrl.CreateCommunityThread)
	communities.Get("/:id/threads/top", communityThreadCtrl.GetCommunityThreadTop)
	communities.Get("/:id/threads/latest", communityThreadCtrl.GetCommunityThreadLatest)
	communities.Get("/:id/threads/media", communityThreadCtrl.GetCommunityThreadMedia)
}
