package repository

import threadpinmodel "github.com/sunibgram/domain/model/thread_pin"

type IThreadPinRepository interface {
	Pin(pin *threadpinmodel.ThreadPin) error
	Unpin(userID, threadID uint) error
	ExistsPin(userID, threadID uint) (bool, error)
	ArePinnedFor(userID uint, threadIDs []uint) (map[uint]bool, error)
	AllPinnedThreadIDs(userID uint) ([]uint, error)
	FindPinnedByUserID(userID uint) (*threadpinmodel.ThreadPin, error)
}
