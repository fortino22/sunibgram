package repository

import "github.com/sunibgram/domain/model"

type IThreadPinRepository interface {
	Pin(pin *model.ThreadPin) error
	Unpin(userID, threadID uint) error
	ExistsPin(userID, threadID uint) (bool, error)
}
