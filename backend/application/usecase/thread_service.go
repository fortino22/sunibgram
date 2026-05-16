package usecase

import (
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type ThreadService struct {
	threadRepo     repository.IThreadRepository
	hashtagRepo    repository.IHashtagRepository
	eventRepo      repository.IEventRepository
	hashtagService service.IHashtagService
	communityRepo  repository.ICommunityRepository
	memberRepo     repository.ICommunityMemberRepository
	settingsRepo   repository.IUserSettingsRepository
	followRepo     repository.IFollowRepository
	notification   service.INotificationService
	summaryService service.IThreadSummaryService
}

func NewThreadService(
	threadRepo repository.IThreadRepository,
	hashtagRepo repository.IHashtagRepository,
	eventRepo repository.IEventRepository,
	hashtagService service.IHashtagService,
	communityRepo repository.ICommunityRepository,
	memberRepo repository.ICommunityMemberRepository,
	settingsRepo repository.IUserSettingsRepository,
	followRepo repository.IFollowRepository,
	notification service.INotificationService,
	summaryService service.IThreadSummaryService,
) service.IThreadService {
	return &ThreadService{
		threadRepo:     threadRepo,
		hashtagRepo:    hashtagRepo,
		eventRepo:      eventRepo,
		hashtagService: hashtagService,
		communityRepo:  communityRepo,
		memberRepo:     memberRepo,
		settingsRepo:   settingsRepo,
		followRepo:     followRepo,
		notification:   notification,
		summaryService: summaryService,
	}
}
