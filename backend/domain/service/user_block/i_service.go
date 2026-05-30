package userblocksvciface

type IUserBlockService interface {
	Block(blockerID, blockedID uint) error
	Unblock(blockerID, blockedID uint) error
	IsBlocked(viewerID, targetID uint) (bool, error)
}
