package router

import (
	"github.com/gofiber/fiber/v2"

	authctrl "github.com/sunibgram/webapi/controller/auth"
	bookmarkctrl "github.com/sunibgram/webapi/controller/bookmark"
	cachectrl "github.com/sunibgram/webapi/controller/cache"
	categoryctrl "github.com/sunibgram/webapi/controller/category"
	chatctrl "github.com/sunibgram/webapi/controller/chat"
	chatroomctrl "github.com/sunibgram/webapi/controller/chat_room"
	communityctrl "github.com/sunibgram/webapi/controller/community"
	communitypinctrl "github.com/sunibgram/webapi/controller/community_pin"
	communityrequestctrl "github.com/sunibgram/webapi/controller/community_request"
	communityrulectrl "github.com/sunibgram/webapi/controller/community_rule"
	communitythreadpinctrl "github.com/sunibgram/webapi/controller/community_thread_pin"
	emailctrl "github.com/sunibgram/webapi/controller/email"
	eventctrl "github.com/sunibgram/webapi/controller/event"
	eventparticipantctrl "github.com/sunibgram/webapi/controller/event_participant"
	eventpinctrl "github.com/sunibgram/webapi/controller/event_pin"
	eventthreadpinctrl "github.com/sunibgram/webapi/controller/event_thread_pin"
	explorectrl "github.com/sunibgram/webapi/controller/explore"
	facultyctrl "github.com/sunibgram/webapi/controller/faculty"
	feedbackctrl "github.com/sunibgram/webapi/controller/feedback"
	followctrl "github.com/sunibgram/webapi/controller/follow"
	hashtagctrl "github.com/sunibgram/webapi/controller/hashtag"
	likectrl "github.com/sunibgram/webapi/controller/like"
	notificationctrl "github.com/sunibgram/webapi/controller/notification"
	orderctrl "github.com/sunibgram/webapi/controller/order"
	premiumsubscriptionctrl "github.com/sunibgram/webapi/controller/premium_subscription"
	reportctrl "github.com/sunibgram/webapi/controller/report"
	repostctrl "github.com/sunibgram/webapi/controller/repost"
	rolectrl "github.com/sunibgram/webapi/controller/role"
	threadctrl "github.com/sunibgram/webapi/controller/thread"
	threadpinctrl "github.com/sunibgram/webapi/controller/thread_pin"
	threadpollctrl "github.com/sunibgram/webapi/controller/thread_poll"
	threadsummaryctrl "github.com/sunibgram/webapi/controller/thread_summary"
	userctrl "github.com/sunibgram/webapi/controller/user"
	userprofilectrl "github.com/sunibgram/webapi/controller/user_profile"
	usersettingsctrl "github.com/sunibgram/webapi/controller/user_settings"
	verificationctrl "github.com/sunibgram/webapi/controller/verification"
	violationctrl "github.com/sunibgram/webapi/controller/violation"
	"github.com/sunibgram/infrastructure/cache"
	"github.com/sunibgram/webapi/middleware"
	adminroute "github.com/sunibgram/webapi/router/admin"
	authroute "github.com/sunibgram/webapi/router/auth"
	categoryroute "github.com/sunibgram/webapi/router/category"
	chatroute "github.com/sunibgram/webapi/router/chat"
	communityroute "github.com/sunibgram/webapi/router/community"
	devroute "github.com/sunibgram/webapi/router/dev"
	emailroute "github.com/sunibgram/webapi/router/email"
	eventroute "github.com/sunibgram/webapi/router/event"
	exploreroute "github.com/sunibgram/webapi/router/explore"
	facultyroute "github.com/sunibgram/webapi/router/faculty"
	feedbackroute "github.com/sunibgram/webapi/router/feedback"
	hashtagroute "github.com/sunibgram/webapi/router/hashtag"
	notificationroute "github.com/sunibgram/webapi/router/notification"
	orderroute "github.com/sunibgram/webapi/router/order"
	premiumsubscriptionroute "github.com/sunibgram/webapi/router/premium_subscription"
	reportroute "github.com/sunibgram/webapi/router/report"
	threadroute "github.com/sunibgram/webapi/router/thread"
	threadpollroute "github.com/sunibgram/webapi/router/thread_poll"
	userroute "github.com/sunibgram/webapi/router/user"
	userprofileroute "github.com/sunibgram/webapi/router/user_profile"
	usersettingsroute "github.com/sunibgram/webapi/router/user_settings"
	verificationroute "github.com/sunibgram/webapi/router/verification"
	violationroute "github.com/sunibgram/webapi/router/violation"
)

func Register(
	app *fiber.App,
	authCtrl *authctrl.AuthController,
	roleCtrl *rolectrl.RoleController,
	userCtrl *userctrl.UserController,
	cacheCtrl *cachectrl.CacheController,
	userSettingsCtrl *usersettingsctrl.UserSettingsController,
	profileCtrl *userprofilectrl.UserProfileController,
	exploreCtrl *explorectrl.ExploreController,
	hashtagCtrl *hashtagctrl.HashtagController,
	threadCtrl *threadctrl.ThreadController,
	eventCtrl *eventctrl.EventController,
	likeCtrl *likectrl.LikeController,
	bookmarkCtrl *bookmarkctrl.BookmarkController,
	threadPinCtrl *threadpinctrl.ThreadPinController,
	repostCtrl *repostctrl.RepostController,
	summaryCtrl *threadsummaryctrl.ThreadSummaryController,
	emailCtrl *emailctrl.EmailController,
	verificationCtrl *verificationctrl.VerificationController,
	followCtrl *followctrl.FollowController,
	communityCtrl *communityctrl.CommunityController,
	communityMemberCtrl *communityctrl.CommunityMemberController,
	communityRequestCtrl *communityrequestctrl.CommunityRequestController,
	communityThreadCtrl *communityctrl.CommunityThreadController,
	communityRuleCtrl *communityrulectrl.CommunityRuleController,
	communityThreadPinCtrl *communitythreadpinctrl.CommunityThreadPinController,
	communityPinCtrl *communitypinctrl.CommunityPinController,
	eventPinCtrl *eventpinctrl.EventPinController,
	eventParticipantCtrl *eventparticipantctrl.EventParticipantController,
	eventThreadPinCtrl *eventthreadpinctrl.EventThreadPinController,
	categoryCtrl *categoryctrl.CategoryController,
	violationCtrl *violationctrl.ViolationController,
	reportCtrl *reportctrl.ReportController,
	facultyCtrl *facultyctrl.FacultyController,
	feedbackCtrl *feedbackctrl.FeedbackController,
	orderCtrl *orderctrl.OrderController,
	premiumSubscriptionCtrl *premiumsubscriptionctrl.PremiumSubscriptionController,
	devSubscriptionCtrl *premiumsubscriptionctrl.DevSubscriptionController,
	chatRoomCtrl *chatroomctrl.ChatRoomController,
	chatCtrl *chatctrl.ChatController,
	notificationCtrl *notificationctrl.NotificationController,
	threadPollCtrl *threadpollctrl.ThreadPollController,
	c *cache.Cache,
) {
	api := app.Group("/api")
	auth := middleware.Auth()

	authroute.RegisterAuthRoutes(api, authCtrl, verificationCtrl, auth, c)
	verificationroute.RegisterVerificationRoutes(api, verificationCtrl, c)
	usersettingsroute.RegisterUserSettingsRoutes(api, userSettingsCtrl, auth)
	userprofileroute.RegisterProfileRoutes(api, profileCtrl, auth)
	exploreroute.RegisterExploreRoutes(api, exploreCtrl, auth)
	hashtagroute.RegisterHashtagRoutes(api, hashtagCtrl, auth)
	threadroute.RegisterThreadRoutes(api, threadCtrl, likeCtrl, bookmarkCtrl, threadPinCtrl, repostCtrl, summaryCtrl, auth)
	eventroute.RegisterEventRoutes(api, eventCtrl, eventPinCtrl, eventParticipantCtrl, eventThreadPinCtrl, auth)
	emailroute.RegisterEmailRoutes(api, emailCtrl, auth)
	userroute.RegisterUserRoutes(api, followCtrl, exploreCtrl, auth)
	adminroute.RegisterAdminRoutes(api, roleCtrl, userCtrl, cacheCtrl)
	communityroute.RegisterCommunityRoutes(api, communityCtrl, communityMemberCtrl, communityRequestCtrl, communityThreadCtrl, communityRuleCtrl, communityThreadPinCtrl, communityPinCtrl, auth)
	categoryroute.RegisterCategoryRoutes(api, categoryCtrl)
	violationroute.RegisterViolationRoutes(api, violationCtrl, auth)
	reportroute.RegisterReportRoutes(api, reportCtrl, auth)
	facultyroute.RegisterFacultyRoutes(api, facultyCtrl, profileCtrl, auth)
	feedbackroute.RegisterFeedbackRoutes(api, feedbackCtrl, auth)
	orderroute.RegisterOrderRoutes(api, orderCtrl, auth, c)
	premiumsubscriptionroute.RegisterPremiumSubscriptionRoutes(api, premiumSubscriptionCtrl, auth)
	devroute.RegisterDevRoutes(api, devSubscriptionCtrl, auth)
	chatroute.RegisterChatRoutes(api, chatRoomCtrl, chatCtrl, auth)
	notificationroute.RegisterNotificationRoutes(api, notificationCtrl, auth)
	threadpollroute.RegisterThreadPollRoutes(api, threadPollCtrl, auth)
}
