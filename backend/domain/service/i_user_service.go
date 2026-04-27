package service

import "github.com/sunibgram/domain/model"

type IUserService interface {
	GetAllPaginated(page, limit int) ([]model.User, int64, error)
	GetByID(id uint) (*model.User, error)
}
