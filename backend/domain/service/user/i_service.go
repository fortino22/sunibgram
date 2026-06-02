package usersvciface

import (
	usermodel "github.com/sunibgram/domain/model/user"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"

	userstatmodel "github.com/sunibgram/domain/model/user_stat"
)

type IUserService interface {
	GetAllPaginated(page, limit int) ([]usermodel.User, int64, error)
	GetByID(id uint) (*usermodel.User, error)
	GetProfile(userID uint) (*usermodel.User, *userstatmodel.UserStat, *userprofilemodel.UserProfile, error)
	CountUsersByRole(roleName string) (int64, error)
}
