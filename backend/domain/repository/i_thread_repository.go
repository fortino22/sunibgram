package repository

import "github.com/sunibgram/domain/model"

type IThreadRepository interface {
	Save(thread *model.Thread) error
	FindByID(id uint) (*model.Thread, error)
	FindRoots(page, limit int) ([]model.Thread, int64, error)
	FindByRootID(rootID uint) ([]model.Thread, error)
	Update(thread *model.Thread) error
	Delete(id uint) error
}
