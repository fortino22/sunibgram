package service

type IThreadPinService interface {
	Pin(userID, threadID uint) error
	Unpin(userID, threadID uint) error
	IsPinned(userID, threadID uint) bool
}
