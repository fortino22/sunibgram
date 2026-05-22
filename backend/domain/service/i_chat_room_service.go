package service

import (
	"context"
	"mime/multipart"
	"time"

	chatroommodel "github.com/sunibgram/domain/model/chat_room"
)

// RoomParticipant is the other user in a private chat room, or a group member.
type RoomParticipant struct {
	UserID      uint
	Username    string
	Name        string
	FacultyName string
	AvatarURL   string
}

// LastMessage is a lightweight preview of the most recent message in a room.
type LastMessage struct {
	ChatID    uint
	SenderID  uint
	Message   *string
	GifURL    *string
	HasMedia  bool
	CreatedAt time.Time
}

// ChatRoomWithParticipant pairs a room with its other participant (private rooms only).
type ChatRoomWithParticipant struct {
	chatroommodel.ChatRoom
	OtherUser   *RoomParticipant
	IsMuted     bool
	LastMessage *LastMessage
	UnreadCount int
}

type IChatRoomService interface {
	CreatePrivate(ctx context.Context, initiatorID, targetUserID uint) (*chatroommodel.ChatRoom, error)
	CreateGroup(ctx context.Context, creatorID uint, name string, description *string, memberIDs []uint) (*chatroommodel.ChatRoom, error)
	AddMemberToGroup(ctx context.Context, roomID, requesterID uint, newUserIDs []uint) (*chatroommodel.ChatRoom, []uint, error)
	UpdateGroup(ctx context.Context, roomID, requesterID uint, name string, description *string, avatarFile *multipart.FileHeader) error
	DeleteGroup(ctx context.Context, roomID, requesterID uint) error
	LeaveGroup(ctx context.Context, roomID, userID uint) error
	GetGroupMembers(ctx context.Context, roomID, requesterID uint) ([]RoomParticipant, error)
	GetUserRooms(userID uint, page, limit int) ([]chatroommodel.ChatRoom, int64, error)
	GetUserRoomsWithParticipants(userID uint, page, limit int) ([]ChatRoomWithParticipant, int64, error)
	// GetRoomForUser returns a room with the other participant's info as seen by viewingUserID.
	GetRoomForUser(roomID, viewingUserID uint) (*ChatRoomWithParticipant, error)
	// GetUserRoomIDs returns all room IDs the user is a member of.
	GetUserRoomIDs(userID uint) ([]uint, error)
	GetByID(roomID, userID uint) (*chatroommodel.ChatRoom, error)
	ToggleMute(ctx context.Context, roomID, userID uint) (bool, error)
}
