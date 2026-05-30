package userblockuc

import (
	"errors"

	"github.com/sunibgram/application/helper"
	usermodel "github.com/sunibgram/domain/model/user"
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

// GetBlockedUsers returns the users that userID has blocked, paginated.
// GetBlockedIDs returns all blocked IDs (small per-user set), so the page is
// sliced in memory; total is the full count for accurate pagination meta.
func (s *UserBlockService) GetBlockedUsers(userID uint, page, limit int) ([]usermodel.User, int64, error) {
	ids, err := s.blockRepo.GetBlockedIDs(userID)
	if err != nil {
		return nil, 0, err
	}
	total := int64(len(ids))
	if total == 0 {
		return []usermodel.User{}, 0, nil
	}

	page, limit = helper.NormalizePageLimit(page, limit, 20)
	start := (page - 1) * limit
	if start >= len(ids) {
		return []usermodel.User{}, total, nil
	}
	end := start + limit
	if end > len(ids) {
		end = len(ids)
	}

	users, err := s.userRepo.FindByIDs(ids[start:end])
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}
