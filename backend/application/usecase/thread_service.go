package usecase

import (
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type ThreadService struct {
	threadRepo    repository.IThreadRepository
	hashtagRepo   repository.IHashtagRepository
	communityRepo repository.ICommunityRepository
	memberRepo    repository.ICommunityMemberRepository
	settingsRepo  repository.IUserSettingsRepository
	followRepo    repository.IFollowRepository
	notification  service.INotificationService
}

func NewThreadService(
	threadRepo repository.IThreadRepository,
	hashtagRepo repository.IHashtagRepository,
	communityRepo repository.ICommunityRepository,
	memberRepo repository.ICommunityMemberRepository,
	settingsRepo repository.IUserSettingsRepository,
	followRepo repository.IFollowRepository,
	notification service.INotificationService,
) service.IThreadService {
	return &ThreadService{
		threadRepo:    threadRepo,
		hashtagRepo:   hashtagRepo,
		communityRepo: communityRepo,
		memberRepo:    memberRepo,
		settingsRepo:  settingsRepo,
		followRepo:    followRepo,
		notification:  notification,
	}
}
