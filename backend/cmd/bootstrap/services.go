package bootstrap

import (
	authuc "github.com/sunibgram/application/usecase/auth"
	bookmarkuc "github.com/sunibgram/application/usecase/bookmark"
	cacheuc "github.com/sunibgram/application/usecase/cache"
	categoryuc "github.com/sunibgram/application/usecase/category"
	chatuc "github.com/sunibgram/application/usecase/chat"
	chatmediauc "github.com/sunibgram/application/usecase/chat_media"
	chatreactionuc "github.com/sunibgram/application/usecase/chat_reaction"
	chatroomuc "github.com/sunibgram/application/usecase/chat_room"
	communityuc "github.com/sunibgram/application/usecase/community"
	communitypinuc "github.com/sunibgram/application/usecase/community_pin"
	communityrequestuc "github.com/sunibgram/application/usecase/community_request"
	communityruleuc "github.com/sunibgram/application/usecase/community_rule"
	communitythreadpinuc "github.com/sunibgram/application/usecase/community_thread_pin"
	eventuc "github.com/sunibgram/application/usecase/event"
	eventparticipantuc "github.com/sunibgram/application/usecase/event_participant"
	eventpinuc "github.com/sunibgram/application/usecase/event_pin"
	eventthreadpinuc "github.com/sunibgram/application/usecase/event_thread_pin"
	exploreuc "github.com/sunibgram/application/usecase/explore"
	facultyuc "github.com/sunibgram/application/usecase/faculty"
	followuc "github.com/sunibgram/application/usecase/follow"
	hashtaguc "github.com/sunibgram/application/usecase/hashtag"
	likeuc "github.com/sunibgram/application/usecase/like"
	notificationuc "github.com/sunibgram/application/usecase/notification"
	orderuc "github.com/sunibgram/application/usecase/order"
	premiumsubscriptionuc "github.com/sunibgram/application/usecase/premium_subscription"
	reportuc "github.com/sunibgram/application/usecase/report"
	roleuc "github.com/sunibgram/application/usecase/role"
	threaduc "github.com/sunibgram/application/usecase/thread"
	threadmediauc "github.com/sunibgram/application/usecase/thread_media"
	threadpinuc "github.com/sunibgram/application/usecase/thread_pin"
	threadpolluc "github.com/sunibgram/application/usecase/thread_poll"
	threadsummaryuc "github.com/sunibgram/application/usecase/thread_summary"
	useruc "github.com/sunibgram/application/usecase/user"
	userprofileuc "github.com/sunibgram/application/usecase/user_profile"
	usersettingsuc "github.com/sunibgram/application/usecase/user_settings"
	verificationuc "github.com/sunibgram/application/usecase/verification"
	violationuc "github.com/sunibgram/application/usecase/violation"
	authsvciface "github.com/sunibgram/domain/service/auth"
	bookmarksvciface "github.com/sunibgram/domain/service/bookmark"
	cachesvciface "github.com/sunibgram/domain/service/cache"
	categorysvciface "github.com/sunibgram/domain/service/category"
	chatsvciface "github.com/sunibgram/domain/service/chat"
	chatmediasvciface "github.com/sunibgram/domain/service/chat_media"
	chatreactionsvciface "github.com/sunibgram/domain/service/chat_reaction"
	chatroomsvciface "github.com/sunibgram/domain/service/chat_room"
	communitysvciface "github.com/sunibgram/domain/service/community"
	communitymembersvciface "github.com/sunibgram/domain/service/community_member"
	communitypinsvciface "github.com/sunibgram/domain/service/community_pin"
	communityrequestsvciface "github.com/sunibgram/domain/service/community_request"
	communityrulesvciface "github.com/sunibgram/domain/service/community_rule"
	communitythreadpinsvciface "github.com/sunibgram/domain/service/community_thread_pin"
	eventsvciface "github.com/sunibgram/domain/service/event"
	eventparticipantsvciface "github.com/sunibgram/domain/service/event_participant"
	eventpinsvciface "github.com/sunibgram/domain/service/event_pin"
	eventthreadpinsvciface "github.com/sunibgram/domain/service/event_thread_pin"
	exploresvciface "github.com/sunibgram/domain/service/explore"
	facultysvciface "github.com/sunibgram/domain/service/faculty"
	followsvciface "github.com/sunibgram/domain/service/follow"
	hashtagsvciface "github.com/sunibgram/domain/service/hashtag"
	likesvciface "github.com/sunibgram/domain/service/like"
	notificationsvciface "github.com/sunibgram/domain/service/notification"
	ordersvciface "github.com/sunibgram/domain/service/order"
	premiumsubscriptionsvciface "github.com/sunibgram/domain/service/premium_subscription"
	reportsvciface "github.com/sunibgram/domain/service/report"
	rolesvciface "github.com/sunibgram/domain/service/role"
	threadsvciface "github.com/sunibgram/domain/service/thread"
	threadmediasvciface "github.com/sunibgram/domain/service/thread_media"
	threadpinsvciface "github.com/sunibgram/domain/service/thread_pin"
	threadpollsvciface "github.com/sunibgram/domain/service/thread_poll"
	threadsummarysvciface "github.com/sunibgram/domain/service/thread_summary"
	usersvciface "github.com/sunibgram/domain/service/user"
	userprofilesvciface "github.com/sunibgram/domain/service/user_profile"
	usersettingssvciface "github.com/sunibgram/domain/service/user_settings"
	verificationsvciface "github.com/sunibgram/domain/service/verification"
	violationsvciface "github.com/sunibgram/domain/service/violation"
	"github.com/sunibgram/infrastructure/cache"
)

type Services struct {
	Cache                cachesvciface.ICacheService
	Auth                 authsvciface.IAuthService
	Role                 rolesvciface.IRoleService
	User                 usersvciface.IUserService
	UserSettings         usersettingssvciface.IUserSettingsService
	UserProfile          userprofilesvciface.IUserProfileService
	Verification         verificationsvciface.IVerificationService
	Notification         notificationsvciface.INotificationService
	Follow               followsvciface.IFollowService
	Hashtag              hashtagsvciface.IHashtagService
	ThreadSummary        threadsummarysvciface.IThreadSummaryService
	Thread               threadsvciface.IThreadService
	Event                eventsvciface.IEventService
	Explore              exploresvciface.IExploreService
	ThreadMedia          threadmediasvciface.IThreadMediaService
	Like                 likesvciface.ILikeService
	Bookmark             bookmarksvciface.IBookmarkService
	ThreadPin            threadpinsvciface.IThreadPinService
	Category             categorysvciface.ICategoryService
	Violation            violationsvciface.IViolationService
	Report               reportsvciface.IReportService
	Community            communitysvciface.ICommunityService
	CommunityMember      communitymembersvciface.ICommunityMemberService
	CommunityRequest     communityrequestsvciface.ICommunityRequestService
	CommunityRule        communityrulesvciface.ICommunityRuleService
	CommunityThreadPin   communitythreadpinsvciface.ICommunityThreadPinService
	CommunityPin         communitypinsvciface.ICommunityPinService
	EventPin             eventpinsvciface.IEventPinService
	EventParticipant     eventparticipantsvciface.IEventParticipantService
	Faculty              facultysvciface.IFacultyService
	Order                ordersvciface.IOrderService
	PremiumSubscription  premiumsubscriptionsvciface.IPremiumSubscriptionService
	ChatRoom             chatroomsvciface.IChatRoomService
	Chat                 chatsvciface.IChatService
	ChatMedia            chatmediasvciface.IChatMediaService
	ChatReaction         chatreactionsvciface.IChatReactionService
	ThreadPoll           threadpollsvciface.IThreadPollService
	EventThreadPin       eventthreadpinsvciface.IEventThreadPinService
}

func NewServices(repos *Repositories, ext *External, c *cache.Cache) *Services {
	profile := userprofileuc.NewUserProfileService(repos.User, repos.Follow, repos.UserProfile, repos.UserFaculty, repos.Cache, ext.Storage)
	notification := notificationuc.NewNotificationService(repos.Notification, repos.User, repos.UserProfile, repos.UserFaculty, ext.Storage, ext.SocketHub)
	follow := followuc.NewFollowService(repos.Follow, repos.User, repos.Cache, notification)
	hashtag := hashtaguc.NewHashtagService(repos.Hashtag, c)
	summary := threadsummaryuc.NewThreadSummaryService(repos.Thread, repos.Cache, ext.Ollama)
	thread := threaduc.NewThreadService(repos.Thread, repos.Hashtag, repos.Event, repos.EventParticipant, hashtag, repos.Community, repos.CommunityMember, repos.UserSettings, repos.Follow, notification, summary)
	like := likeuc.NewLikeService(repos.Like, repos.Thread, repos.Community, repos.CommunityMember, repos.UserSettings, repos.Follow, notification)
	bookmark := bookmarkuc.NewBookmarkService(repos.Bookmark, repos.Thread, repos.Community, repos.CommunityMember, repos.UserSettings, repos.Follow, notification)

	return &Services{
		Cache:                cacheuc.NewCacheService(repos.Cache),
		Auth:                 authuc.NewAuthService(repos.User, repos.Role, repos.Verification, repos.UserSettings, c),
		Role:                 roleuc.NewRoleService(repos.Role),
		User:                 useruc.NewUserService(repos.User, repos.Follow, repos.UserProfile),
		UserSettings:         usersettingsuc.NewUserSettingsService(repos.UserSettings, repos.Thread),
		UserProfile:          profile,
		Verification:         verificationuc.NewVerificationService(repos.Verification, repos.User, ext.Email, c),
		Notification:         notification,
		Follow:               follow,
		Hashtag:              hashtag,
		ThreadSummary:        summary,
		Thread:               thread,
		Event:                eventuc.NewEventService(repos.Event, repos.EventReview, repos.EventPin, repos.EventParticipant, ext.Storage, repos.UserSettings, repos.Follow, repos.Thread),
		Explore:              exploreuc.NewExploreService(repos.Explore, repos.Thread, repos.UserSettings, repos.Follow),
		ThreadMedia:          threadmediauc.NewThreadMediaService(repos.Thread, repos.ThreadMedia, ext.Storage),
		Like:                 like,
		Bookmark:             bookmark,
		ThreadPin:            threadpinuc.NewThreadPinService(repos.ThreadPin, repos.Thread),
		Category:             categoryuc.NewCategoryService(repos.Category),
		Violation:            violationuc.NewViolationService(repos.Violation),
		Report:               reportuc.NewReportService(repos.Report, repos.Violation, repos.User, repos.Thread, profile, repos.Cache),
		Community:            communityuc.NewCommunityService(repos.Community, repos.CommunityMember, repos.User, repos.UserProfile, repos.Follow, repos.Category, repos.CommunityRule, repos.CommunityPin, repos.Cache, ext.Storage),
		CommunityMember:      communityuc.NewCommunityMemberService(repos.Community, repos.CommunityMember, repos.Follow, repos.Cache),
		CommunityRequest:     communityrequestuc.NewCommunityRequestService(repos.Community, repos.CommunityMember, repos.CommunityRequest, repos.Cache),
		CommunityRule:        communityruleuc.NewCommunityRuleService(repos.Community, repos.CommunityRule),
		CommunityThreadPin:   communitythreadpinuc.NewCommunityThreadPinService(repos.CommunityThreadPin, repos.Community, repos.CommunityMember, repos.Thread),
		CommunityPin:         communitypinuc.NewCommunityPinService(repos.CommunityPin, repos.Community, repos.CommunityMember),
		EventPin:             eventpinuc.NewEventPinService(repos.EventPin, repos.Event),
		EventParticipant:     eventparticipantuc.NewEventParticipantService(repos.EventParticipant, repos.Event),
		Faculty:              facultyuc.NewFacultyService(repos.Faculty, repos.UserFaculty),
		Order:                orderuc.NewOrderService(repos.Order, repos.User, repos.PremiumSubscription, repos.Community, repos.Cache),
		PremiumSubscription:  premiumsubscriptionuc.NewPremiumSubscriptionService(repos.User, repos.PremiumSubscription, repos.Community, repos.Cache),
		ChatRoom:             chatroomuc.NewChatRoomService(repos.ChatRoom, repos.User, ext.Storage),
		Chat:                 chatuc.NewChatService(repos.Chat, repos.ChatRoom, repos.ChatRead, notification),
		ChatMedia:            chatmediauc.NewChatMediaService(repos.Chat, repos.ChatMedia, ext.Storage),
		ChatReaction:         chatreactionuc.NewChatReactionService(repos.Chat, repos.ChatReaction, repos.ChatRoom),
		ThreadPoll:           threadpolluc.NewThreadPollService(repos.Thread, repos.ThreadPoll, ext.Storage),
		EventThreadPin:       eventthreadpinuc.NewEventThreadPinService(repos.EventThreadPin, repos.Event, repos.Thread),
	}
}
