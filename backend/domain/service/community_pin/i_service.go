package communitypinsvciface

import communitymodel "github.com/sunibgram/domain/model/community"

type ICommunityPinService interface {
	Pin(userID, communityID uint) error
	Unpin(userID, communityID uint) error
	GetMyPins(userID uint, page, limit int) ([]communitymodel.Community, int64, error)
}
