package repository

import threadmediamodel "github.com/sunibgram/domain/model/thread_media"

type IThreadMediaRepository interface {
	SaveAll(media []threadmediamodel.ThreadMedia) error
	FindByThreadID(threadID uint) ([]threadmediamodel.ThreadMedia, error)
	FindByThreadIDs(threadIDs []uint) ([]threadmediamodel.ThreadMedia, error)
	DeleteByThreadID(threadID uint) error
	DeleteByIDs(ids []uint) ([]threadmediamodel.ThreadMedia, error)
}
