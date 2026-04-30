package repository

import "github.com/sunibgram/domain/model"

type IUserProfileRepository interface {
	Upsert(profile *model.UserProfile) error
	FindByUserID(userID uint) (*model.UserProfile, error)
	ExistsByUserID(userID uint) (bool, error)
}
