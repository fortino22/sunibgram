package repository

import "github.com/sunibgram/domain/model"

type ICommunityRequestRepository interface {
	Save(req *model.CommunityRequest) error
	FindRequest(communityID, userID uint) (*model.CommunityRequest, error)
	UpdateStatus(communityID, userID uint, status model.CommunityRequestStatus) error
	Delete(communityID, userID uint) error
	FindPending(communityID uint, page, limit int) ([]model.CommunityRequest, int64, error)
}
