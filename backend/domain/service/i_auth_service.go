package service

import "github.com/sunibgram/domain/model"

type IAuthService interface {
	Register(user *model.User) error
	Login(email, password string) (*model.User, model.TokenPair, error)
	RefreshToken(refreshToken string) (model.TokenPair, error)
}
