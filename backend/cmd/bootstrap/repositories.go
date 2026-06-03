package bootstrap

import (
	bookmarkrepoiface "github.com/sunibgram/domain/repository/bookmark"
	bookmarkcollectionrepoiface "github.com/sunibgram/domain/repository/bookmark_collection"
	cacherepoiface "github.com/sunibgram/domain/repository/cache"
	categoryrepoiface "github.com/sunibgram/domain/repository/category"
	chatrepoiface "github.com/sunibgram/domain/repository/chat"
	chatmediarepoiface "github.com/sunibgram/domain/repository/chat_media"
	chatreactionrepoiface "github.com/sunibgram/domain/repository/chat_reaction"
	chatreadrepoiface "github.com/sunibgram/domain/repository/chat_read"
	chatroomrepoiface "github.com/sunibgram/domain/repository/chat_room"
	communityrepoiface "github.com/sunibgram/domain/repository/community"
	communityactionconfirmationrepoiface "github.com/sunibgram/domain/repository/community_action_confirmation"
	communitymemberrepoiface "github.com/sunibgram/domain/repository/community_member"
	communitypinrepoiface "github.com/sunibgram/domain/repository/community_pin"
	communityrequestrepoiface "github.com/sunibgram/domain/repository/community_request"
	communityrulerepoiface "github.com/sunibgram/domain/repository/community_rule"
	communitythreadpinrepoiface "github.com/sunibgram/domain/repository/community_thread_pin"
	eventrepoiface "github.com/sunibgram/domain/repository/event"
	eventactionconfirmationrepoiface "github.com/sunibgram/domain/repository/event_action_confirmation"
	eventparticipantrepoiface "github.com/sunibgram/domain/repository/event_participant"
	eventpinrepoiface "github.com/sunibgram/domain/repository/event_pin"
	eventreviewrepoiface "github.com/sunibgram/domain/repository/event_review"
	eventthreadpinrepoiface "github.com/sunibgram/domain/repository/event_thread_pin"
	explorerepoiface "github.com/sunibgram/domain/repository/explore"
	facultyrepoiface "github.com/sunibgram/domain/repository/faculty"
	feedbackrepoiface "github.com/sunibgram/domain/repository/feedback"
	feedbackmediarepoiface "github.com/sunibgram/domain/repository/feedback_media"
	followrepoiface "github.com/sunibgram/domain/repository/follow"
	hashtagrepoiface "github.com/sunibgram/domain/repository/hashtag"
	likerepoiface "github.com/sunibgram/domain/repository/like"
	notificationrepoiface "github.com/sunibgram/domain/repository/notification"
	notificationusermuterepoiface "github.com/sunibgram/domain/repository/notification_user_mute"
	orderrepoiface "github.com/sunibgram/domain/repository/order"
	premiumsubscriptionrepoiface "github.com/sunibgram/domain/repository/premium_subscription"
	profilefieldchangerepoiface "github.com/sunibgram/domain/repository/profile_field_change"
	reportrepoiface "github.com/sunibgram/domain/repository/report"
	rolerepoiface "github.com/sunibgram/domain/repository/role"
	threadrepoiface "github.com/sunibgram/domain/repository/thread"
	threadmediarepoiface "github.com/sunibgram/domain/repository/thread_media"
	threadpinrepoiface "github.com/sunibgram/domain/repository/thread_pin"
	threadpollrepoiface "github.com/sunibgram/domain/repository/thread_poll"
	threadsummaryrepoiface "github.com/sunibgram/domain/repository/thread_summary"
	trendreportrepoiface "github.com/sunibgram/domain/repository/trend_report"
	trendingthreadrepoiface "github.com/sunibgram/domain/repository/trending_thread"
	trendinguserrepoiface "github.com/sunibgram/domain/repository/trending_user"
	trendingusersnapshotrepoiface "github.com/sunibgram/domain/repository/trending_user_snapshot"
	userrepoiface "github.com/sunibgram/domain/repository/user"
	userblockrepoiface "github.com/sunibgram/domain/repository/user_block"
	userfacultyrepoiface "github.com/sunibgram/domain/repository/user_faculty"
	userprofilerepoiface "github.com/sunibgram/domain/repository/user_profile"
	usersettingsrepoiface "github.com/sunibgram/domain/repository/user_settings"
	verificationrepoiface "github.com/sunibgram/domain/repository/verification"
	violationrepoiface "github.com/sunibgram/domain/repository/violation"
	"github.com/sunibgram/infrastructure/cache"
	bookmarkrepo "github.com/sunibgram/infrastructure/repository/bookmark"
	bookmarkcollectionrepo "github.com/sunibgram/infrastructure/repository/bookmark_collection"
	cacherepo "github.com/sunibgram/infrastructure/repository/cache"
	categoryrepo "github.com/sunibgram/infrastructure/repository/category"
	chatrepo "github.com/sunibgram/infrastructure/repository/chat"
	chatmediarepo "github.com/sunibgram/infrastructure/repository/chat_media"
	chatreactionrepo "github.com/sunibgram/infrastructure/repository/chat_reaction"
	chatreadrepo "github.com/sunibgram/infrastructure/repository/chat_read"
	chatroomrepo "github.com/sunibgram/infrastructure/repository/chat_room"
	communityrepo "github.com/sunibgram/infrastructure/repository/community"
	communityactionconfirmationrepo "github.com/sunibgram/infrastructure/repository/community_action_confirmation"
	communitymemberrepo "github.com/sunibgram/infrastructure/repository/community_member"
	communitypinrepo "github.com/sunibgram/infrastructure/repository/community_pin"
	communityrequestrepo "github.com/sunibgram/infrastructure/repository/community_request"
	communityrulerepo "github.com/sunibgram/infrastructure/repository/community_rule"
	communitythreadpinrepo "github.com/sunibgram/infrastructure/repository/community_thread_pin"
	eventrepo "github.com/sunibgram/infrastructure/repository/event"
	eventactionconfirmationrepo "github.com/sunibgram/infrastructure/repository/event_action_confirmation"
	eventparticipantrepo "github.com/sunibgram/infrastructure/repository/event_participant"
	eventpinrepo "github.com/sunibgram/infrastructure/repository/event_pin"
	eventreviewrepo "github.com/sunibgram/infrastructure/repository/event_review"
	eventthreadpinrepo "github.com/sunibgram/infrastructure/repository/event_thread_pin"
	explorerepo "github.com/sunibgram/infrastructure/repository/explore"
	facultyrepo "github.com/sunibgram/infrastructure/repository/faculty"
	feedbackrepo "github.com/sunibgram/infrastructure/repository/feedback"
	feedbackmediarepo "github.com/sunibgram/infrastructure/repository/feedback_media"
	followrepo "github.com/sunibgram/infrastructure/repository/follow"
	hashtagrepo "github.com/sunibgram/infrastructure/repository/hashtag"
	likerepo "github.com/sunibgram/infrastructure/repository/like"
	notificationrepo "github.com/sunibgram/infrastructure/repository/notification"
	notificationusermuterepo "github.com/sunibgram/infrastructure/repository/notification_user_mute"
	orderrepo "github.com/sunibgram/infrastructure/repository/order"
	premiumsubscriptionrepo "github.com/sunibgram/infrastructure/repository/premium_subscription"
	profilefieldchangerepo "github.com/sunibgram/infrastructure/repository/profile_field_change"
	reportrepo "github.com/sunibgram/infrastructure/repository/report"
	rolerepo "github.com/sunibgram/infrastructure/repository/role"
	threadrepo "github.com/sunibgram/infrastructure/repository/thread"
	threadmediarepo "github.com/sunibgram/infrastructure/repository/thread_media"
	threadpinrepo "github.com/sunibgram/infrastructure/repository/thread_pin"
	threadpollrepo "github.com/sunibgram/infrastructure/repository/thread_poll"
	threadsummaryrepo "github.com/sunibgram/infrastructure/repository/thread_summary"
	trendreportrepo "github.com/sunibgram/infrastructure/repository/trend_report"
	trendingthreadrepo "github.com/sunibgram/infrastructure/repository/trending_thread"
	trendinguserrepo "github.com/sunibgram/infrastructure/repository/trending_user"
	trendingusersnapshotrepo "github.com/sunibgram/infrastructure/repository/trending_user_snapshot"
	userrepo "github.com/sunibgram/infrastructure/repository/user"
	userblockrepo "github.com/sunibgram/infrastructure/repository/user_block"
	userfacultyrepo "github.com/sunibgram/infrastructure/repository/user_faculty"
	userprofilerepo "github.com/sunibgram/infrastructure/repository/user_profile"
	usersettingsrepo "github.com/sunibgram/infrastructure/repository/user_settings"
	verificationrepo "github.com/sunibgram/infrastructure/repository/verification"
	violationrepo "github.com/sunibgram/infrastructure/repository/violation"
)

type Repositories struct {
	Cache                       cacherepoiface.ICacheRepository
	Role                        rolerepoiface.IRoleRepository
	User                        userrepoiface.IUserRepository
	UserSettings                usersettingsrepoiface.IUserSettingsRepository
	UserProfile                 userprofilerepoiface.IUserProfileRepository
	ProfileFieldChange          profilefieldchangerepoiface.IProfileFieldChangeRepository
	Verification                verificationrepoiface.IVerificationRepository
	Follow                      followrepoiface.IFollowRepository
	UserBlock                   userblockrepoiface.IUserBlockRepository
	NotificationUserMute        notificationusermuterepoiface.INotificationUserMuteRepository
	Thread                      threadrepoiface.IThreadRepository
	ThreadPatchCount            func(uint, string, int)
	Event                       eventrepoiface.IEventRepository
	EventReview                 eventreviewrepoiface.IEventReviewRepository
	Hashtag                     hashtagrepoiface.IHashtagRepository
	Explore                     explorerepoiface.IExploreRepository
	ThreadMedia                 threadmediarepoiface.IThreadMediaRepository
	Like                        likerepoiface.ILikeRepository
	Bookmark                    bookmarkrepoiface.IBookmarkRepository
	BookmarkCollection          bookmarkcollectionrepoiface.IBookmarkCollectionRepository
	ThreadPin                   threadpinrepoiface.IThreadPinRepository
	Category                    categoryrepoiface.ICategoryRepository
	Violation                   violationrepoiface.IViolationRepository
	Report                      reportrepoiface.IReportRepository
	Community                   communityrepoiface.ICommunityRepository
	CommunityActionConfirmation communityactionconfirmationrepoiface.ICommunityActionConfirmationRepository
	CommunityMember             communitymemberrepoiface.ICommunityMemberRepository
	CommunityRequest            communityrequestrepoiface.ICommunityRequestRepository
	CommunityRule               communityrulerepoiface.ICommunityRuleRepository
	CommunityThreadPin          communitythreadpinrepoiface.ICommunityThreadPinRepository
	CommunityPin                communitypinrepoiface.ICommunityPinRepository
	EventPin                    eventpinrepoiface.IEventPinRepository
	EventParticipant            eventparticipantrepoiface.IEventParticipantRepository
	EventActionConfirmation     eventactionconfirmationrepoiface.IEventActionConfirmationRepository
	Faculty                     facultyrepoiface.IFacultyRepository
	Feedback                    feedbackrepoiface.IFeedbackRepository
	FeedbackMedia               feedbackmediarepoiface.IFeedbackMediaRepository
	UserFaculty                 userfacultyrepoiface.IUserFacultyRepository
	Order                       orderrepoiface.IOrderRepository
	PremiumSubscription         premiumsubscriptionrepoiface.IPremiumSubscriptionRepository
	ChatRoom                    chatroomrepoiface.IChatRoomRepository
	Chat                        chatrepoiface.IChatRepository
	ChatMedia                   chatmediarepoiface.IChatMediaRepository
	ChatRead                    chatreadrepoiface.IChatReadRepository
	ChatReaction                chatreactionrepoiface.IChatReactionRepository
	Notification                notificationrepoiface.INotificationRepository
	ThreadPoll                  threadpollrepoiface.IThreadPollRepository
	ThreadSummary               threadsummaryrepoiface.IThreadSummaryRepository
	EventThreadPin              eventthreadpinrepoiface.IEventThreadPinRepository
	TrendingUser                trendinguserrepoiface.ITrendingUserRepository
	TrendingUserSnapshot        trendingusersnapshotrepoiface.ITrendingUserSnapshotRepository
	TrendingThread              trendingthreadrepoiface.ITrendingThreadRepository
	TrendReport                 trendreportrepoiface.ITrendReportRepository
}

func NewRepositories(db *Database, c *cache.Cache) *Repositories {
	w, r := db.Write, db.Read
	threadRepo, patchCount := threadrepo.NewThreadRepository(w, r, c)
	return &Repositories{
		Cache:                       cacherepo.NewCacheRepository(c),
		Role:                        rolerepo.NewRoleRepository(w, r),
		User:                        userrepo.NewUserRepository(w, r, c),
		UserSettings:                usersettingsrepo.NewUserSettingsRepository(w, r, c),
		UserProfile:                 userprofilerepo.NewUserProfileRepository(w, r, c),
		ProfileFieldChange:          profilefieldchangerepo.NewProfileFieldChangeRepository(w, r, c),
		Verification:                verificationrepo.NewVerificationRepository(w, r),
		Follow:                      followrepo.NewFollowRepository(w, r, c),
		UserBlock:                   userblockrepo.NewUserBlockRepository(w, r, c),
		NotificationUserMute:        notificationusermuterepo.NewNotificationUserMuteRepository(w, r, c),
		Thread:                      threadRepo,
		ThreadPatchCount:            patchCount,
		Event:                       eventrepo.NewEventRepository(w, r, c),
		EventReview:                 eventreviewrepo.NewEventReviewRepository(w, r),
		Hashtag:                     hashtagrepo.NewHashtagRepository(w, r),
		Explore:                     explorerepo.NewExploreRepository(r, c),
		ThreadMedia:                 threadmediarepo.NewThreadMediaRepository(w, r, c),
		Like:                        likerepo.NewLikeRepository(w, r, c, patchCount),
		Bookmark:                    bookmarkrepo.NewBookmarkRepository(w, r, c, patchCount),
		BookmarkCollection:          bookmarkcollectionrepo.NewBookmarkCollectionRepository(w, r),
		ThreadPin:                   threadpinrepo.NewThreadPinRepository(w, r, c),
		Category:                    categoryrepo.NewCategoryRepository(w, r, c),
		Violation:                   violationrepo.NewViolationRepository(w, r),
		Report:                      reportrepo.NewReportRepository(w, r),
		Community:                   communityrepo.NewCommunityRepository(w, r, c),
		CommunityActionConfirmation: communityactionconfirmationrepo.NewCommunityActionConfirmationRepository(w, r),
		CommunityMember:             communitymemberrepo.NewCommunityMemberRepository(w, r, c),
		CommunityRequest:            communityrequestrepo.NewCommunityRequestRepository(w, r, c),
		CommunityRule:               communityrulerepo.NewCommunityRuleRepository(w, r),
		CommunityThreadPin:          communitythreadpinrepo.NewCommunityThreadPinRepository(w, r),
		CommunityPin:                communitypinrepo.NewCommunityPinRepository(w, r, c),
		EventPin:                    eventpinrepo.NewEventPinRepository(w, r, c),
		EventParticipant:            eventparticipantrepo.NewEventParticipantRepository(w, r, c),
		EventActionConfirmation:     eventactionconfirmationrepo.NewEventActionConfirmationRepository(w, r),
		Faculty:                     facultyrepo.NewFacultyRepository(w, r, c),
		Feedback:                    feedbackrepo.NewFeedbackRepository(w, r, c),
		FeedbackMedia:               feedbackmediarepo.NewFeedbackMediaRepository(w, r),
		UserFaculty:                 userfacultyrepo.NewUserFacultyRepository(w, r, c),
		Order:                       orderrepo.NewOrderRepository(w, r),
		PremiumSubscription:         premiumsubscriptionrepo.NewPremiumSubscriptionRepository(w, r),
		ChatRoom:                    chatroomrepo.NewChatRoomRepository(w, r, c),
		Chat:                        chatrepo.NewChatRepository(w, r),
		ChatMedia:                   chatmediarepo.NewChatMediaRepository(w, r),
		ChatRead:                    chatreadrepo.NewChatReadRepository(w, r),
		ChatReaction:                chatreactionrepo.NewChatReactionRepository(w, r),
		Notification:                notificationrepo.NewNotificationRepository(w, r, c),
		ThreadPoll:                  threadpollrepo.NewThreadPollRepository(w, r, c),
		ThreadSummary:               threadsummaryrepo.NewThreadSummaryRepository(w, r),
		EventThreadPin:              eventthreadpinrepo.NewEventThreadPinRepository(w, r, c),
		TrendingUser:                trendinguserrepo.NewTrendingUserRepository(r, c),
		TrendingUserSnapshot:        trendingusersnapshotrepo.NewTrendingUserSnapshotRepository(w, r, c),
		TrendingThread:              trendingthreadrepo.NewTrendingThreadRepository(r, c),
		TrendReport:                 trendreportrepo.NewTrendReportRepository(w, r),
	}
}
