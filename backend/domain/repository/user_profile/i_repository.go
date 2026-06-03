package userprofilerepoiface

import userprofilemodel "github.com/sunibgram/domain/model/user_profile"

type IUserProfileRepository interface {
	Upsert(profile *userprofilemodel.UserProfile) error
	FindByUserID(userID uint) (*userprofilemodel.UserProfile, error)
	FindByUserIDs(ids []uint) ([]userprofilemodel.UserProfile, error)
	ExistsByUserID(userID uint) (bool, error)
	// UpdateFacultyLocation sets the user's home campus (Upsert's fixed column
	// list doesn't cover it) and invalidates the profile cache.
	UpdateFacultyLocation(userID uint, locationID *uint) error
}
