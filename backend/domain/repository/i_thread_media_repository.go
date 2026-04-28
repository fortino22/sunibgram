package repository

import "github.com/sunibgram/domain/model"

type IThreadMediaRepository interface {
	SaveAll(media []model.ThreadMedia) error
	UpdateOne(media *model.ThreadMedia) error
	FindByThreadID(threadID uint) ([]model.ThreadMedia, error)
	FindByThreadIDs(threadIDs []uint) ([]model.ThreadMedia, error)
	DeleteByThreadID(threadID uint) error
	DeleteByIDs(ids []uint) ([]model.ThreadMedia, error)
}
