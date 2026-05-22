package authsvciface

import (
	"context"

	usermodel "github.com/sunibgram/domain/model/user"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
)

type IAuthService interface {
	Register(user *usermodel.User, profile *userprofilemodel.UserProfile, confirmPassword string) error
	Login(email, password string) (usermodel.TokenPair, error)
	LoginWithOTP(email, password, otp string) (usermodel.TokenPair, error)
	LoginProd(email, password string) (*usermodel.User, usermodel.TokenPair, error)
	RefreshToken(ctx context.Context, sid string) (usermodel.TokenPair, error)
	Logout(ctx context.Context, sid string) error
	ChangePassword(userID uint, oldPassword, newPassword string) error
	ForgotPassword(email, newPassword string) error
}
