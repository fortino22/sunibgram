package repository

import "github.com/sunibgram/domain/model"

type ICommunityRepository interface {
	Save(community *model.Community) error
	FindByID(id uint) (*model.Community, error)
	Update(community *model.Community) error
	Delete(id uint) error
	FindAll(page, limit int, category string) ([]model.Community, int64, error)
	FindByCreatorID(userID uint, page, limit int) ([]model.Community, int64, error)
	FindJoinedByUserID(userID uint, page, limit int) ([]model.Community, int64, error)
}
