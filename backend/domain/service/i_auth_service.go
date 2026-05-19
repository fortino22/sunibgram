package service

import (
	"context"

	"github.com/sunibgram/domain/model"
)

type IAuthService interface {
	Register(user *model.User, profile *model.UserProfile, confirmPassword string) error
	Login(email, password string) (model.TokenPair, error)
	LoginWithOTP(email, password, otp string) (model.TokenPair, error)
	LoginProd(email, password string) (*model.User, model.TokenPair, error)
	RefreshToken(ctx context.Context, sid string) (model.TokenPair, error)
	Logout(ctx context.Context, sid string) error
	ChangePassword(userID uint, oldPassword, newPassword string) error
	ForgotPassword(email, newPassword string) error
}
