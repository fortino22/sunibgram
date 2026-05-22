package userrepoiface

import (
	usermodel "github.com/sunibgram/domain/model/user"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
)

type IUserRepository interface {
	Save(user *usermodel.User) error
	SaveWithProfile(user *usermodel.User, profile *userprofilemodel.UserProfile) error
	FindByID(id uint) (*usermodel.User, error)
	FindByIDs(ids []uint) ([]usermodel.User, error)
	FindByEmail(email string) (*usermodel.User, error)
	FindByUserName(username string) (*usermodel.User, error)
	ExistsByEmail(email string) (bool, error)
	ExistsByUserName(username string) (bool, error)
	FindAllPaginated(page, limit int) ([]usermodel.User, int64, error)
	UpdatePassword(userID uint, hashedPassword string) error
	UpdateInfo(userID uint, name, username string) error
	UpdatePremium(userID uint, premium bool) error
}
