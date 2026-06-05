package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sunibgram/infrastructure/cache"
	authctrl "github.com/sunibgram/webapi/controller/auth"
	bookmarkctrl "github.com/sunibgram/webapi/controller/bookmark"
	bookmarkcollectionctrl "github.com/sunibgram/webapi/controller/bookmark_collection"
	cachectrl "github.com/sunibgram/webapi/controller/cache"
	categoryctrl "github.com/sunibgram/webapi/controller/category"
	chatctrl "github.com/sunibgram/webapi/controller/chat"
	chatroomctrl "github.com/sunibgram/webapi/controller/chat_room"
	communityctrl "github.com/sunibgram/webapi/controller/community"
	communityeventctrl "github.com/sunibgram/webapi/controller/community_event"
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
	facultylocationctrl "github.com/sunibgram/webapi/controller/faculty_location"
	feedbackctrl "github.com/sunibgram/webapi/controller/feedback"
	followctrl "github.com/sunibgram/webapi/controller/follow"
	foxchatctrl "github.com/sunibgram/webapi/controller/fox_chat"
	foxsessionctrl "github.com/sunibgram/webapi/controller/fox_session"
	hashtagctrl "github.com/sunibgram/webapi/controller/hashtag"
	likectrl "github.com/sunibgram/webapi/controller/like"
	notificationctrl "github.com/sunibgram/webapi/controller/notification"
	notificationusermutectrl "github.com/sunibgram/webapi/controller/notification_user_mute"
	orderctrl "github.com/sunibgram/webapi/controller/order"
	premiumsubscriptionctrl "github.com/sunibgram/webapi/controller/premium_subscription"
	reportctrl "github.com/sunibgram/webapi/controller/report"
	repostctrl "github.com/sunibgram/webapi/controller/repost"
	rolectrl "github.com/sunibgram/webapi/controller/role"
	schemactrl "github.com/sunibgram/webapi/controller/schema"
	sharectrl "github.com/sunibgram/webapi/controller/share"
	threadctrl "github.com/sunibgram/webapi/controller/thread"
	threadpinctrl "github.com/sunibgram/webapi/controller/thread_pin"
	threadpollctrl "github.com/sunibgram/webapi/controller/thread_poll"
	threadsummaryctrl "github.com/sunibgram/webapi/controller/thread_summary"
	trendreportctrl "github.com/sunibgram/webapi/controller/trend_report"
	trendingthreadctrl "github.com/sunibgram/webapi/controller/trending_thread"
	trendinguserctrl "github.com/sunibgram/webapi/controller/trending_user"
	userctrl "github.com/sunibgram/webapi/controller/user"
	userblockctrl "github.com/sunibgram/webapi/controller/user_block"
	userprofilectrl "github.com/sunibgram/webapi/controller/user_profile"
	usersettingsctrl "github.com/sunibgram/webapi/controller/user_settings"
	verificationctrl "github.com/sunibgram/webapi/controller/verification"
	violationctrl "github.com/sunibgram/webapi/controller/violation"
	"github.com/sunibgram/webapi/middleware"
	adminroute "github.com/sunibgram/webapi/router/admin"
	authroute "github.com/sunibgram/webapi/router/auth"
	bookmarkcollectionroute "github.com/sunibgram/webapi/router/bookmark_collection"
	categoryroute "github.com/sunibgram/webapi/router/category"
	chatroute "github.com/sunibgram/webapi/router/chat"
	communityroute "github.com/sunibgram/webapi/router/community"
	communityeventroute "github.com/sunibgram/webapi/router/community_event"
	devroute "github.com/sunibgram/webapi/router/dev"
	emailroute "github.com/sunibgram/webapi/router/email"
	eventroute "github.com/sunibgram/webapi/router/event"
	exploreroute "github.com/sunibgram/webapi/router/explore"
	facultyroute "github.com/sunibgram/webapi/router/faculty"
	facultylocationroute "github.com/sunibgram/webapi/router/faculty_location"
	feedbackroute "github.com/sunibgram/webapi/router/feedback"
	foxchatroute "github.com/sunibgram/webapi/router/fox_chat"
	foxsessionroute "github.com/sunibgram/webapi/router/fox_session"
	hashtagroute "github.com/sunibgram/webapi/router/hashtag"
	notificationroute "github.com/sunibgram/webapi/router/notification"
	orderroute "github.com/sunibgram/webapi/router/order"
	premiumsubscriptionroute "github.com/sunibgram/webapi/router/premium_subscription"
	reportroute "github.com/sunibgram/webapi/router/report"
	shareroute "github.com/sunibgram/webapi/router/share"
	threadroute "github.com/sunibgram/webapi/router/thread"
	threadpollroute "github.com/sunibgram/webapi/router/thread_poll"
	trendreportroute "github.com/sunibgram/webapi/router/trend_report"
	trendingthreadroute "github.com/sunibgram/webapi/router/trending_thread"
	trendinguserroute "github.com/sunibgram/webapi/router/trending_user"
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
	bookmarkCollectionCtrl *bookmarkcollectionctrl.BookmarkCollectionController,
	threadPinCtrl *threadpinctrl.ThreadPinController,
	repostCtrl *repostctrl.RepostController,
	summaryCtrl *threadsummaryctrl.ThreadSummaryController,
	foxChatCtrl *foxchatctrl.FoxChatController,
	foxSessionCtrl *foxsessionctrl.FoxSessionController,
	emailCtrl *emailctrl.EmailController,
	verificationCtrl *verificationctrl.VerificationController,
	followCtrl *followctrl.FollowController,
	userBlockCtrl *userblockctrl.UserBlockController,
	muteCtrl *notificationusermutectrl.NotificationUserMuteController,
	communityCtrl *communityctrl.CommunityController,
	communityMemberCtrl *communityctrl.CommunityMemberController,
	communityEventCtrl *communityeventctrl.CommunityEventController,
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
	facultyLocationCtrl *facultylocationctrl.FacultyLocationController,
	feedbackCtrl *feedbackctrl.FeedbackController,
	orderCtrl *orderctrl.OrderController,
	premiumSubscriptionCtrl *premiumsubscriptionctrl.PremiumSubscriptionController,
	devSubscriptionCtrl *premiumsubscriptionctrl.DevSubscriptionController,
	chatRoomCtrl *chatroomctrl.ChatRoomController,
	chatCtrl *chatctrl.ChatController,
	shareCtrl *sharectrl.ShareController,
	notificationCtrl *notificationctrl.NotificationController,
	threadPollCtrl *threadpollctrl.ThreadPollController,
	trendingUserCtrl *trendinguserctrl.TrendingUserController,
	trendingThreadCtrl *trendingthreadctrl.TrendingThreadController,
	trendReportCtrl *trendreportctrl.TrendReportController,
	schemaCtrl *schemactrl.SchemaController,
	c *cache.Cache,
) {
	api := app.Group("/api")
	auth := middleware.Auth()

	authroute.RegisterAuthRoutes(api, authCtrl, auth, c)
	verificationroute.RegisterVerificationRoutes(api, verificationCtrl, c)
	usersettingsroute.RegisterUserSettingsRoutes(api, userSettingsCtrl, auth)
	userprofileroute.RegisterProfileRoutes(api, profileCtrl, auth)
	exploreroute.RegisterExploreRoutes(api, exploreCtrl, auth)
	hashtagroute.RegisterHashtagRoutes(api, hashtagCtrl, auth)
	threadroute.RegisterThreadRoutes(api, threadCtrl, likeCtrl, bookmarkCtrl, threadPinCtrl, repostCtrl, summaryCtrl, auth)
	foxchatroute.RegisterFoxChatRoutes(api, foxChatCtrl, auth)
	foxsessionroute.RegisterFoxSessionRoutes(api, foxSessionCtrl, auth)
	bookmarkcollectionroute.RegisterBookmarkCollectionRoutes(api, bookmarkCollectionCtrl, auth)
	eventroute.RegisterEventRoutes(api, eventCtrl, eventPinCtrl, eventParticipantCtrl, eventThreadPinCtrl, auth)
	emailroute.RegisterEmailRoutes(api, emailCtrl, auth)
	userroute.RegisterUserRoutes(api, followCtrl, exploreCtrl, userBlockCtrl, muteCtrl, userCtrl, auth)
	adminroute.RegisterAdminRoutes(api, roleCtrl, userCtrl, cacheCtrl, schemaCtrl)
	communityroute.RegisterCommunityRoutes(api, communityCtrl, communityMemberCtrl, communityRequestCtrl, communityThreadCtrl, communityRuleCtrl, communityThreadPinCtrl, communityPinCtrl, auth)
	communityeventroute.RegisterCommunityEventRoutes(api, communityEventCtrl, auth)
	categoryroute.RegisterCategoryRoutes(api, categoryCtrl)
	violationroute.RegisterViolationRoutes(api, violationCtrl, auth)
	reportroute.RegisterReportRoutes(api, reportCtrl, auth)
	facultyroute.RegisterFacultyRoutes(api, facultyCtrl, profileCtrl, auth)
	facultylocationroute.RegisterFacultyLocationRoutes(api, facultyLocationCtrl)
	feedbackroute.RegisterFeedbackRoutes(api, feedbackCtrl, auth)
	orderroute.RegisterOrderRoutes(api, orderCtrl, auth, c)
	premiumsubscriptionroute.RegisterPremiumSubscriptionRoutes(api, premiumSubscriptionCtrl, auth)
	devroute.RegisterDevRoutes(api, devSubscriptionCtrl, auth)
	chatroute.RegisterChatRoutes(api, chatRoomCtrl, chatCtrl, auth)
	shareroute.RegisterShareRoutes(api, shareCtrl, auth)
	notificationroute.RegisterNotificationRoutes(api, notificationCtrl, auth)
	threadpollroute.RegisterThreadPollRoutes(api, threadPollCtrl, auth)
	trendinguserroute.RegisterTrendingUserRoutes(api, trendingUserCtrl, auth)
	trendingthreadroute.RegisterTrendingThreadRoutes(api, trendingThreadCtrl, auth)
	trendreportroute.RegisterTrendReportRoutes(api, trendReportCtrl, auth)
}
