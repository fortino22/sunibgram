package bootstrap

import (
	"os"

	authctrl "github.com/sunibgram/webapi/controller/auth"
	bookmarkctrl "github.com/sunibgram/webapi/controller/bookmark"
	bookmarkcollectionctrl "github.com/sunibgram/webapi/controller/bookmark_collection"
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
	notificationusermutectrl "github.com/sunibgram/webapi/controller/notification_user_mute"
	orderctrl "github.com/sunibgram/webapi/controller/order"
	premiumsubscriptionctrl "github.com/sunibgram/webapi/controller/premium_subscription"
	reportctrl "github.com/sunibgram/webapi/controller/report"
	repostctrl "github.com/sunibgram/webapi/controller/repost"
	rolectrl "github.com/sunibgram/webapi/controller/role"
	sharectrl "github.com/sunibgram/webapi/controller/share"
	threadctrl "github.com/sunibgram/webapi/controller/thread"
	threadpinctrl "github.com/sunibgram/webapi/controller/thread_pin"
	threadpollctrl "github.com/sunibgram/webapi/controller/thread_poll"
	threadsummaryctrl "github.com/sunibgram/webapi/controller/thread_summary"
	trendreportctrl "github.com/sunibgram/webapi/controller/trend_report"
	trendinguserctrl "github.com/sunibgram/webapi/controller/trending_user"
	userctrl "github.com/sunibgram/webapi/controller/user"
	userblockctrl "github.com/sunibgram/webapi/controller/user_block"
	userprofilectrl "github.com/sunibgram/webapi/controller/user_profile"
	usersettingsctrl "github.com/sunibgram/webapi/controller/user_settings"
	verificationctrl "github.com/sunibgram/webapi/controller/verification"
	violationctrl "github.com/sunibgram/webapi/controller/violation"
	"github.com/sunibgram/infrastructure/cache"
	webapisocket "github.com/sunibgram/webapi/socket"
)

type Controllers struct {
	Cache               *cachectrl.CacheController
	Auth                *authctrl.AuthController
	Role                *rolectrl.RoleController
	User                *userctrl.UserController
	UserSettings        *usersettingsctrl.UserSettingsController
	UserProfile         *userprofilectrl.UserProfileController
	Email               *emailctrl.EmailController
	Verification        *verificationctrl.VerificationController
	Follow               *followctrl.FollowController
	UserBlock            *userblockctrl.UserBlockController
	NotificationUserMute *notificationusermutectrl.NotificationUserMuteController
	Thread               *threadctrl.ThreadController
	Event               *eventctrl.EventController
	EventThreadPin      *eventthreadpinctrl.EventThreadPinController
	Hashtag             *hashtagctrl.HashtagController
	Explore             *explorectrl.ExploreController
	Like                *likectrl.LikeController
	Bookmark            *bookmarkctrl.BookmarkController
	BookmarkCollection  *bookmarkcollectionctrl.BookmarkCollectionController
	ThreadPin           *threadpinctrl.ThreadPinController
	Repost              *repostctrl.RepostController
	ThreadSummary       *threadsummaryctrl.ThreadSummaryController
	Category            *categoryctrl.CategoryController
	Violation           *violationctrl.ViolationController
	Report              *reportctrl.ReportController
	Community           *communityctrl.CommunityController
	CommunityMember     *communityctrl.CommunityMemberController
	CommunityRequest    *communityrequestctrl.CommunityRequestController
	CommunityRule       *communityrulectrl.CommunityRuleController
	CommunityThreadPin  *communitythreadpinctrl.CommunityThreadPinController
	CommunityPin        *communitypinctrl.CommunityPinController
	CommunityThread     *communityctrl.CommunityThreadController
	EventPin            *eventpinctrl.EventPinController
	EventParticipant    *eventparticipantctrl.EventParticipantController
	Faculty             *facultyctrl.FacultyController
	Feedback            *feedbackctrl.FeedbackController
	Order               *orderctrl.OrderController
	PremiumSubscription *premiumsubscriptionctrl.PremiumSubscriptionController
	DevSubscription     *premiumsubscriptionctrl.DevSubscriptionController
	ChatRoom            *chatroomctrl.ChatRoomController
	Chat                *chatctrl.ChatController
	Share               *sharectrl.ShareController
	Notification        *notificationctrl.NotificationController
	ThreadPoll          *threadpollctrl.ThreadPollController
	TrendingUser        *trendinguserctrl.TrendingUserController
	TrendReport         *trendreportctrl.TrendReportController
	NotificationHandler *webapisocket.NotificationHandler
}

func NewControllers(svcs *Services, ext *External, c *cache.Cache) *Controllers {
	return &Controllers{
		Cache:               cachectrl.NewCacheController(svcs.Cache),
		Auth:                authctrl.NewAuthController(svcs.Auth, svcs.Faculty),
		Role:                rolectrl.NewRoleController(svcs.Role),
		User:                userctrl.NewUserController(svcs.User, svcs.UserProfile),
		UserSettings:        usersettingsctrl.NewUserSettingsController(svcs.UserSettings),
		UserProfile:         userprofilectrl.NewUserProfileController(svcs.UserProfile, svcs.Faculty, ext.SocketHub, svcs.ChatRoom),
		Email:               emailctrl.NewEmailController(ext.Email),
		Verification:        verificationctrl.NewVerificationController(svcs.Verification, c),
		Follow:              followctrl.NewFollowController(svcs.Follow, svcs.UserProfile),
		UserBlock:            userblockctrl.NewUserBlockController(svcs.UserBlock, svcs.UserProfile),
		NotificationUserMute: notificationusermutectrl.NewNotificationUserMuteController(svcs.NotificationUserMute, svcs.UserProfile),
		Thread:              threadctrl.NewThreadController(svcs.Thread, svcs.ThreadMedia, svcs.CommunityMember, svcs.Like, svcs.Bookmark, svcs.ThreadPin, svcs.UserProfile, svcs.ThreadPoll),
		Event:               eventctrl.NewEventController(svcs.Event, svcs.Thread, svcs.ThreadMedia, svcs.Like, svcs.Bookmark, svcs.ThreadPin, svcs.UserProfile, svcs.ThreadPoll, svcs.EventThreadPin),
		EventThreadPin:      eventthreadpinctrl.NewEventThreadPinController(svcs.EventThreadPin),
		Hashtag:             hashtagctrl.NewHashtagController(svcs.Hashtag),
		Explore:             explorectrl.NewExploreController(svcs.Explore, svcs.UserProfile, svcs.ThreadMedia, svcs.Like, svcs.Bookmark, svcs.Thread, svcs.Community, svcs.ThreadPoll),
		Like:                likectrl.NewLikeController(svcs.Like, svcs.Bookmark, svcs.Thread, svcs.ThreadMedia, svcs.ThreadPoll, svcs.UserProfile),
		Bookmark:            bookmarkctrl.NewBookmarkController(svcs.Bookmark, svcs.Like, svcs.Thread, svcs.ThreadMedia, svcs.ThreadPoll, svcs.UserProfile),
		BookmarkCollection:  bookmarkcollectionctrl.NewBookmarkCollectionController(svcs.BookmarkCollection, svcs.Bookmark, svcs.Like, svcs.Thread, svcs.ThreadMedia, svcs.ThreadPoll, svcs.UserProfile),
		ThreadPin:           threadpinctrl.NewThreadPinController(svcs.ThreadPin),
		Repost:              repostctrl.NewRepostController(svcs.Thread, svcs.ThreadMedia),
		ThreadSummary:       threadsummaryctrl.NewThreadSummaryController(svcs.ThreadSummary, svcs.ThreadMedia),
		Category:            categoryctrl.NewCategoryController(svcs.Category),
		Violation:           violationctrl.NewViolationController(svcs.Violation),
		Report:              reportctrl.NewReportController(svcs.Report),
		Community:           communityctrl.NewCommunityController(svcs.Community, svcs.CommunityRule, svcs.UserProfile),
		CommunityMember:     communityctrl.NewCommunityMemberController(svcs.CommunityMember, svcs.UserProfile),
		CommunityRequest:    communityrequestctrl.NewCommunityRequestController(svcs.CommunityRequest, svcs.UserProfile),
		CommunityRule:       communityrulectrl.NewCommunityRuleController(svcs.CommunityRule),
		CommunityThreadPin:  communitythreadpinctrl.NewCommunityThreadPinController(svcs.CommunityThreadPin),
		CommunityPin:        communitypinctrl.NewCommunityPinController(svcs.CommunityPin, svcs.Community),
		CommunityThread:     communityctrl.NewCommunityThreadController(svcs.Thread, svcs.ThreadMedia, svcs.Like, svcs.Bookmark, svcs.ThreadPin, svcs.UserProfile, svcs.ThreadPoll, svcs.CommunityThreadPin),
		EventPin:            eventpinctrl.NewEventPinController(svcs.EventPin, svcs.Event),
		EventParticipant:    eventparticipantctrl.NewEventParticipantController(svcs.EventParticipant, svcs.Event),
		Faculty:             facultyctrl.NewFacultyController(svcs.Faculty),
		Feedback:            feedbackctrl.NewFeedbackController(svcs.Feedback, svcs.UserProfile, c),
		Order:               orderctrl.NewOrderController(svcs.Order, os.Getenv("MIDTRANS_SERVER_KEY")),
		PremiumSubscription: premiumsubscriptionctrl.NewPremiumSubscriptionController(svcs.PremiumSubscription),
		DevSubscription:     premiumsubscriptionctrl.NewDevSubscriptionController(svcs.PremiumSubscription),
		ChatRoom:            chatroomctrl.NewChatRoomController(svcs.ChatRoom, ext.SocketHub),
		Chat:                chatctrl.NewChatController(svcs.Chat, svcs.ChatMedia, svcs.ChatReaction, ext.SocketHub),
		Share:               sharectrl.NewShareController(svcs.Share, ext.Storage),
		Notification:        notificationctrl.NewNotificationController(svcs.Notification),
		ThreadPoll:          threadpollctrl.NewThreadPollController(svcs.ThreadPoll),
		TrendingUser:        trendinguserctrl.NewTrendingUserController(svcs.TrendingUser, svcs.UserProfile),
		TrendReport:         trendreportctrl.NewTrendReportController(svcs.TrendReport),
		NotificationHandler: webapisocket.NewNotificationHandler(ext.SocketHub, svcs.Notification),
	}
}
