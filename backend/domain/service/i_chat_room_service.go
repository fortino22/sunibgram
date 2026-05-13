package service

import (
	"context"

	"github.com/sunibgram/domain/model"
)

// RoomParticipant is the other user in a private chat room.
type RoomParticipant struct {
	UserID      uint
	Username    string
	Name        string
	FacultyName string
}

// ChatRoomWithParticipant pairs a room with its other participant (private rooms only).
type ChatRoomWithParticipant struct {
	model.ChatRoom
	OtherUser *RoomParticipant
}

type IChatRoomService interface {
	CreatePrivate(ctx context.Context, initiatorID, targetUserID uint) (*model.ChatRoom, error)
	CreateGroup(ctx context.Context, creatorID uint, name string, description *string, memberIDs []uint) (*model.ChatRoom, error)
	AddMemberToGroup(ctx context.Context, roomID, requesterID uint, newUserIDs []uint) (*model.ChatRoom, []uint, error)
	UpdateGroup(ctx context.Context, roomID, requesterID uint, name string, description *string) error
	DeleteGroup(ctx context.Context, roomID, requesterID uint) error
	GetUserRooms(userID uint, page, limit int) ([]model.ChatRoom, int64, error)
	GetUserRoomsWithParticipants(userID uint, page, limit int) ([]ChatRoomWithParticipant, int64, error)
	GetByID(roomID, userID uint) (*model.ChatRoom, error)
}
