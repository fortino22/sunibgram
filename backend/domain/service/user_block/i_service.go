package userblocksvciface

import (
	usermodel "github.com/sunibgram/domain/model/user"
)

type IUserBlockService interface {
	Block(blockerID, blockedID uint) error
	Unblock(blockerID, blockedID uint) error
	IsBlocked(viewerID, targetID uint) (bool, error)
	GetBlockedUsers(userID uint, page, limit int) ([]usermodel.User, int64, error)
}
