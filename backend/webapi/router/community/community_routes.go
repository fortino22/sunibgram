package communityroute

import (
	"github.com/gofiber/fiber/v2"

	communityctrl "github.com/sunibgram/webapi/controller/community"
	communitypinctrl "github.com/sunibgram/webapi/controller/community_pin"
	communityrequestctrl "github.com/sunibgram/webapi/controller/community_request"
	communityrulectrl "github.com/sunibgram/webapi/controller/community_rule"
	communitythreadpinctrl "github.com/sunibgram/webapi/controller/community_thread_pin"
)

func RegisterCommunityRoutes(
	api fiber.Router,
	communityCtrl *communityctrl.CommunityController,
	memberCtrl *communityctrl.CommunityMemberController,
	requestCtrl *communityrequestctrl.CommunityRequestController,
	communityThreadCtrl *communityctrl.CommunityThreadController,
	ruleCtrl *communityrulectrl.CommunityRuleController,
	pinCtrl *communitythreadpinctrl.CommunityThreadPinController,
	communityPinCtrl *communitypinctrl.CommunityPinController,
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
	communities.Get("/pins/me", communityPinCtrl.GetMyPins)

	//community threads (static paths before /:id wildcard)
	communities.Get("/me/threads", communityThreadCtrl.GetMyCommunityThreads)
	communities.Get("/threads", communityThreadCtrl.GetAllCommunityThreads)
	communities.Get("/threads/top", communityThreadCtrl.GetCommunityThreadTopByName)
	communities.Get("/threads/latest", communityThreadCtrl.GetCommunityThreadLatestByName)

	communities.Get("/:id", communityCtrl.GetByID)
	communities.Get("/:id/about", communityCtrl.GetAbout)
	communities.Put("/:id", communityCtrl.Update)
	communities.Delete("/:id", communityCtrl.Delete)

	//community members
	communities.Post("/:id/join", memberCtrl.Join)
	communities.Delete("/:id/leave", memberCtrl.Leave)
	communities.Get("/:id/members", memberCtrl.GetMembers)
	communities.Get("/:id/users", memberCtrl.GetCommunityUsers)
	communities.Delete("/:id/members/:userId", memberCtrl.RemoveMember)
	communities.Post("/:id/members/:userId/kick", memberCtrl.Kick)
	communities.Patch("/:id/members/:userId/role", memberCtrl.UpdateRole)

	//community join requests
	communities.Post("/:id/requests", requestCtrl.Send)
	communities.Delete("/:id/requests", requestCtrl.Cancel)
	communities.Get("/:id/requests", requestCtrl.GetPending)
	communities.Post("/:id/requests/:userId/approve", requestCtrl.Approve)
	communities.Post("/:id/requests/:userId/decline", requestCtrl.Decline)
	communities.Post("/:communityID/threads", communityThreadCtrl.CreateCommunityThread)
	communities.Get("/:id/threads/sort", communityThreadCtrl.GetCommunityThreadsBySort)
	communities.Get("/:id/threads/top", communityThreadCtrl.GetCommunityThreadTop)
	communities.Get("/:id/threads/latest", communityThreadCtrl.GetCommunityThreadLatest)
	communities.Get("/:id/threads/media", communityThreadCtrl.GetCommunityThreadMedia)

	//community thread pin
	communities.Post("/:id/threads/:threadId/pin", pinCtrl.Pin)
	communities.Delete("/:id/threads/:threadId/unpin", pinCtrl.Unpin)

	//community pin (user-level favorites)
	communities.Post("/:id/pin", communityPinCtrl.Pin)
	communities.Delete("/:id/unpin", communityPinCtrl.Unpin)
}
