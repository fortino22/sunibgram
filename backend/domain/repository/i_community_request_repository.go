package repository

import communityrequestmodel "github.com/sunibgram/domain/model/community_request"

type ICommunityRequestRepository interface {
	Save(req *communityrequestmodel.CommunityRequest) error
	FindRequest(communityID, userID uint) (*communityrequestmodel.CommunityRequest, error)
	UpdateStatus(communityID, userID uint, status communityrequestmodel.CommunityRequestStatus) error
	Delete(communityID, userID uint) error
	FindPending(communityID uint, page, limit int) ([]communityrequestmodel.CommunityRequest, int64, error)
}
