package threaduc

import (
	communityrepoiface "github.com/sunibgram/domain/repository/community"
	communitymemberrepoiface "github.com/sunibgram/domain/repository/community_member"
	eventrepoiface "github.com/sunibgram/domain/repository/event"
	eventparticipantrepoiface "github.com/sunibgram/domain/repository/event_participant"
	followrepoiface "github.com/sunibgram/domain/repository/follow"
	hashtagrepoiface "github.com/sunibgram/domain/repository/hashtag"
	threadrepoiface "github.com/sunibgram/domain/repository/thread"
	userrepoiface "github.com/sunibgram/domain/repository/user"
	usersettingsrepoiface "github.com/sunibgram/domain/repository/user_settings"
	hashtagsvciface "github.com/sunibgram/domain/service/hashtag"
	notificationsvciface "github.com/sunibgram/domain/service/notification"
	threadsvciface "github.com/sunibgram/domain/service/thread"
	threadsummarysvciface "github.com/sunibgram/domain/service/thread_summary"
)

type ThreadService struct {
	threadRepo     threadrepoiface.IThreadRepository
	hashtagRepo    hashtagrepoiface.IHashtagRepository
	eventRepo      eventrepoiface.IEventRepository
	eventPartRepo  eventparticipantrepoiface.IEventParticipantRepository
	hashtagService hashtagsvciface.IHashtagService
	communityRepo  communityrepoiface.ICommunityRepository
	memberRepo     communitymemberrepoiface.ICommunityMemberRepository
	settingsRepo   usersettingsrepoiface.IUserSettingsRepository
	followRepo     followrepoiface.IFollowRepository
	notification   notificationsvciface.INotificationService
	summaryService threadsummarysvciface.IThreadSummaryService
	userRepo       userrepoiface.IUserRepository
}

func NewThreadService(
	threadRepo threadrepoiface.IThreadRepository,
	hashtagRepo hashtagrepoiface.IHashtagRepository,
	eventRepo eventrepoiface.IEventRepository,
	eventPartRepo eventparticipantrepoiface.IEventParticipantRepository,
	hashtagService hashtagsvciface.IHashtagService,
	communityRepo communityrepoiface.ICommunityRepository,
	memberRepo communitymemberrepoiface.ICommunityMemberRepository,
	settingsRepo usersettingsrepoiface.IUserSettingsRepository,
	followRepo followrepoiface.IFollowRepository,
	notification notificationsvciface.INotificationService,
	summaryService threadsummarysvciface.IThreadSummaryService,
	userRepo userrepoiface.IUserRepository,
) threadsvciface.IThreadService {
	return &ThreadService{
		threadRepo:     threadRepo,
		hashtagRepo:    hashtagRepo,
		eventRepo:      eventRepo,
		eventPartRepo:  eventPartRepo,
		hashtagService: hashtagService,
		communityRepo:  communityRepo,
		memberRepo:     memberRepo,
		settingsRepo:   settingsRepo,
		followRepo:     followRepo,
		notification:   notification,
		summaryService: summaryService,
		userRepo:       userRepo,
	}
}
