package service

import (
	"time"

	"github.com/sunibgram/domain/model"
)

type IThreadPinService interface {
	Pin(userID, threadID uint) error
	Unpin(userID, threadID uint) error
	IsPinned(userID, threadID uint) bool
	GetMyPins(userID uint, page, limit int) ([]model.Thread, []time.Time, int64, error)
}
