package repository

import (
	"time"

	"github.com/sunibgram/domain/model"
)

type IThreadPinRepository interface {
	Pin(pin *model.ThreadPin) error
	Unpin(userID, threadID uint) error
	ExistsPin(userID, threadID uint) (bool, error)
	FindByUserID(userID uint, page, limit int) ([]model.Thread, []time.Time, int64, error)
}
