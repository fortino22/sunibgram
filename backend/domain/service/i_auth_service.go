package service

import "github.com/sunibgram/domain/model"

type IAuthService interface {
	Register(user *model.User, confirmPassword string) error
	Login(email, password string) (model.TokenPair, error)
	LoginProd(email, password string) (*model.User, model.TokenPair, error)
	RefreshToken(refreshToken string) (model.TokenPair, error)
	ChangePassword(userID uint, oldPassword, newPassword string) error
	ForgotPassword(email, newPassword string) error
}
