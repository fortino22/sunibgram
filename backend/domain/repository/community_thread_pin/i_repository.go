package communitythreadpinrepoiface

import communitythreadpinmodel "github.com/sunibgram/domain/model/community_thread_pin"

type ICommunityThreadPinRepository interface {
	Pin(pin *communitythreadpinmodel.CommunityThreadPin) error
	Unpin(communityID, threadID uint) error
	Exists(communityID, threadID uint) (bool, error)
	FindByCommunityID(communityID uint) (*communitythreadpinmodel.CommunityThreadPin, error)
}
