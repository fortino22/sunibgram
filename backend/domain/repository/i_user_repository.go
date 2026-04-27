package repository

import "github.com/sunibgram/domain/model"

type IUserRepository interface {
	Save(user *model.User) error
	FindByID(id uint) (*model.User, error)
	FindByEmail(email string) (*model.User, error)
	FindByUserName(username string) (*model.User, error)
	ExistsByEmail(email string) (bool, error)
	ExistsByUserName(username string) (bool, error)
}
