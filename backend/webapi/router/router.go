package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/webapi/controller"
	"github.com/sunibgram/webapi/middleware"
)

func Register(
	app *fiber.App,
	authCtrl *controller.AuthController,
	roleCtrl *controller.RoleController,
	userCtrl *controller.UserController,
	cacheCtrl *controller.CacheController,
	userSettingsCtrl *controller.UserSettingsController,
	profileCtrl *controller.UserProfileController,
	exploreCtrl *controller.ExploreController,
	hashtagCtrl *controller.HashtagController,
	threadCtrl *controller.ThreadController,
	eventCtrl *controller.EventController,
	likeCtrl *controller.LikeController,
	bookmarkCtrl *controller.BookmarkController,
	threadPinCtrl *controller.ThreadPinController,
	repostCtrl *controller.RepostController,
	summaryCtrl *controller.ThreadSummaryController,
	emailCtrl *controller.EmailController,
	verificationCtrl *controller.VerificationController,
	followCtrl *controller.FollowController,
	communityCtrl *controller.CommunityController,
	communityMemberCtrl *controller.CommunityMemberController,
	communityRequestCtrl *controller.CommunityRequestController,
	communityThreadCtrl *controller.CommunityThreadController,
	communityRuleCtrl *controller.CommunityRuleController,
	communityThreadPinCtrl *controller.CommunityThreadPinController,
	communityPinCtrl *controller.CommunityPinController,
	eventPinCtrl *controller.EventPinController,
	eventParticipantCtrl *controller.EventParticipantController,
	eventThreadPinCtrl *controller.EventThreadPinController,
	categoryCtrl *controller.CategoryController,
	violationCtrl *controller.ViolationController,
	reportCtrl *controller.ReportController,
	facultyCtrl *controller.FacultyController,
	orderCtrl *controller.OrderController,
	premiumSubscriptionCtrl *controller.PremiumSubscriptionController,
	devSubscriptionCtrl *controller.DevSubscriptionController,
	chatRoomCtrl *controller.ChatRoomController,
	chatCtrl *controller.ChatController,
	notificationCtrl *controller.NotificationController,
	threadPollCtrl *controller.ThreadPollController,
) {
	api := app.Group("/api")
	auth := middleware.Auth()

	registerAuthRoutes(api, authCtrl, verificationCtrl, auth)
	registerVerificationRoutes(api, verificationCtrl)
	registerUserSettingsRoutes(api, userSettingsCtrl, auth)
	registerProfileRoutes(api, profileCtrl, auth)
	registerExploreRoutes(api, exploreCtrl, auth)
	registerHashtagRoutes(api, hashtagCtrl, auth)
	registerThreadRoutes(api, threadCtrl, likeCtrl, bookmarkCtrl, threadPinCtrl, repostCtrl, summaryCtrl, auth)
	registerEventRoutes(api, eventCtrl, eventPinCtrl, eventParticipantCtrl, eventThreadPinCtrl, auth)
	registerEmailRoutes(api, emailCtrl, auth)
	registerUserRoutes(api, followCtrl, exploreCtrl, auth)
	registerAdminRoutes(api, roleCtrl, userCtrl, cacheCtrl)
	registerCommunityRoutes(api, communityCtrl, communityMemberCtrl, communityRequestCtrl, communityThreadCtrl, communityRuleCtrl, communityThreadPinCtrl, communityPinCtrl, auth)
	registerCategoryRoutes(api, categoryCtrl)
	registerViolationRoutes(api, violationCtrl, auth)
	registerReportRoutes(api, reportCtrl, auth)
	registerFacultyRoutes(api, facultyCtrl, profileCtrl, auth)
	registerOrderRoutes(api, orderCtrl, auth)
	registerPremiumSubscriptionRoutes(api, premiumSubscriptionCtrl, auth)
	registerDevRoutes(api, devSubscriptionCtrl, auth)
	registerChatRoutes(api, chatRoomCtrl, chatCtrl, auth)
	registerNotificationRoutes(api, notificationCtrl, auth)
	registerThreadPollRoutes(api, threadPollCtrl, auth)
}
