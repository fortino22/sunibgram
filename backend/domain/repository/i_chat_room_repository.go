package repository

import "github.com/sunibgram/domain/model"

type IChatRoomRepository interface {
	Save(room *model.ChatRoom) error
	FindByID(id uint) (*model.ChatRoom, error)
	FindByUserID(userID uint, page, limit int) ([]model.ChatRoom, int64, error)
	IsMember(roomID, userID uint) bool
	CountMembers(roomID uint) int64
	AddMember(roomID, userID uint) error
	FindPrivateRoom(userA, userB uint) (*model.ChatRoom, error)
}
