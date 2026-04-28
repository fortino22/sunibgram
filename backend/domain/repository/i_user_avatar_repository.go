package repository

import "github.com/sunibgram/domain/model"

type IUserAvatarRepository interface {
	Save(avatar *model.UserAvatar) error
	Update(avatar *model.UserAvatar) error
	FindByUserID(userID uint) (*model.UserAvatar, error)
	ExistsByUserID(userID uint) (bool, error)
	DeleteByUserID(userID uint) error
}
