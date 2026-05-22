package repository

import userprofilemodel "github.com/sunibgram/domain/model/user_profile"

type IUserProfileRepository interface {
	Upsert(profile *userprofilemodel.UserProfile) error
	FindByUserID(userID uint) (*userprofilemodel.UserProfile, error)
	FindByUserIDs(ids []uint) ([]userprofilemodel.UserProfile, error)
	ExistsByUserID(userID uint) (bool, error)
}
