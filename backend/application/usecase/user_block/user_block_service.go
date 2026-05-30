package userblockuc

import (
	"errors"

	followrepoiface "github.com/sunibgram/domain/repository/follow"
	userrepoiface "github.com/sunibgram/domain/repository/user"
	userblockrepoiface "github.com/sunibgram/domain/repository/user_block"
	userblocksvciface "github.com/sunibgram/domain/service/user_block"
)

type UserBlockService struct {
	blockRepo  userblockrepoiface.IUserBlockRepository
	userRepo   userrepoiface.IUserRepository
	followRepo followrepoiface.IFollowRepository
}

func NewUserBlockService(
	blockRepo userblockrepoiface.IUserBlockRepository,
	userRepo userrepoiface.IUserRepository,
	followRepo followrepoiface.IFollowRepository,
) userblocksvciface.IUserBlockService {
	return &UserBlockService{blockRepo: blockRepo, userRepo: userRepo, followRepo: followRepo}
}

func (s *UserBlockService) Block(blockerID, blockedID uint) error {
	if blockerID == blockedID {
		return errors.New("cannot block yourself")
	}
	if _, err := s.userRepo.FindByID(blockedID); err != nil {
		return errors.New("user not found")
	}
	if err := s.blockRepo.Block(blockerID, blockedID); err != nil {
		return err
	}
	// Auto-unfollow in both directions; ignore "not following" errors.
	_ = s.followRepo.Unfollow(blockerID, blockedID)
	_ = s.followRepo.Unfollow(blockedID, blockerID)
	return nil
}

func (s *UserBlockService) Unblock(blockerID, blockedID uint) error {
	return s.blockRepo.Unblock(blockerID, blockedID)
}

func (s *UserBlockService) IsBlocked(viewerID, targetID uint) (bool, error) {
	return s.blockRepo.IsBlocked(viewerID, targetID)
}
