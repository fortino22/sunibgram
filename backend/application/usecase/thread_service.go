package usecase

import (
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type ThreadService struct {
	threadRepo    repository.IThreadRepository
	communityRepo repository.ICommunityRepository
	memberRepo    repository.ICommunityMemberRepository
	settingsRepo  repository.IUserSettingsRepository
	followRepo    repository.IFollowRepository
}

func NewThreadService(
	threadRepo repository.IThreadRepository,
	communityRepo repository.ICommunityRepository,
	memberRepo repository.ICommunityMemberRepository,
	settingsRepo repository.IUserSettingsRepository,
	followRepo repository.IFollowRepository,
) service.IThreadService {
	return &ThreadService{
		threadRepo:    threadRepo,
		communityRepo: communityRepo,
		memberRepo:    memberRepo,
		settingsRepo:  settingsRepo,
		followRepo:    followRepo,
	}
}
