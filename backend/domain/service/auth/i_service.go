package authsvciface

import (
	"context"
	"errors"

	usermodel "github.com/sunibgram/domain/model/user"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
)

var (
	ErrSessionRevoked     = errors.New("session expired or revoked")
	ErrSessionTransient   = errors.New("session lookup temporarily unavailable")
	ErrRefreshRateLimited = errors.New("refresh rate limit exceeded")
)

type IAuthService interface {
	Register(user *usermodel.User, profile *userprofilemodel.UserProfile, confirmPassword string) error
	Login(email, password string) (usermodel.TokenPair, error)
	LoginWithOTP(email, password, otp string) (usermodel.TokenPair, error)
	LoginProd(email, password string) (*usermodel.User, usermodel.TokenPair, error)
	RefreshToken(ctx context.Context, sid string) (usermodel.TokenPair, error)
	Logout(ctx context.Context, sid string) error
	ChangePassword(userID uint, email, newPassword string) error
	ForgotPassword(email, newPassword string) error
}
