package repository

import "github.com/sunibgram/domain/model"

type IThreadRepository interface {
	Save(thread *model.Thread) error
	FindByID(id uint) (*model.Thread, error)
	FindRoots(page, limit int) ([]model.Thread, int64, error)
	FindByRootID(rootID uint, page, limit int) ([]model.Thread, int64, error)
	Update(thread *model.Thread) error
	Delete(id uint) error
	IncrementCount(threadID uint, column string) error
	DecrementCount(threadID uint, column string) error
}
