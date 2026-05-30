package userblockrepoiface

type IUserBlockRepository interface {
	Block(blockerID, blockedID uint) error
	Unblock(blockerID, blockedID uint) error
	// IsBlocked reports whether a block exists in either direction between the two users.
	IsBlocked(viewerID, targetID uint) (bool, error)
	// BlocksUser reports whether blockerID has blocked blockedID (directional).
	BlocksUser(blockerID, blockedID uint) (bool, error)
	// GetBlockedIDs returns all user ids the given user blocks.
	GetBlockedIDs(userID uint) ([]uint, error)
}
