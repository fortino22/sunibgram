package repository

import (
	"time"

	chatroommodel "github.com/sunibgram/domain/model/chat_room"
)

// ChatRoomMemberInfo holds basic public info about a room participant.
type ChatRoomMemberInfo struct {
	UserID           uint
	Username         string
	Name             string
	FacultyName      string
	AvatarObjectName string
	AvatarBucket     string
}

// LastMessageInfo holds a lightweight preview of the most recent message in a room.
type LastMessageInfo struct {
	RoomID    uint
	ChatID    uint
	SenderID  uint
	Message   *string
	GifURL    *string
	HasMedia  bool
	CreatedAt time.Time
}

type IChatRoomRepository interface {
	Save(room *chatroommodel.ChatRoom) error
	FindByID(id uint) (*chatroommodel.ChatRoom, error)
	FindByUserID(userID uint, page, limit int) ([]chatroommodel.ChatRoom, int64, error)
	// FindOtherMemberInfo returns the other participant's info for private rooms,
	// keyed by room_id. Rooms without a counterpart are omitted from the map.
	FindOtherMemberInfo(roomIDs []uint, excludeUserID uint) (map[uint]ChatRoomMemberInfo, error)
	FindMembers(roomID uint) ([]ChatRoomMemberInfo, error)
	IsMember(roomID, userID uint) bool
	CountMembers(roomID uint) int64
	AddMember(roomID, userID uint) error
	UpdateEncryptedKey(roomID uint, encryptedKey string) error
	RemoveMember(roomID, userID uint) error
	UpdateGroup(roomID uint, name string, description *string, avatarURL *string) error
	DeleteGroup(roomID uint) error
	FindPrivateRoom(userA, userB uint) (*chatroommodel.ChatRoom, error)
	ToggleMute(roomID, userID uint) (bool, error)
	GetMuteStatuses(userID uint, roomIDs []uint) (map[uint]bool, error)
	TouchUpdatedAt(roomID uint) error
	// FindLastMessages returns the most recent non-deleted message for each given room.
	FindLastMessages(roomIDs []uint) (map[uint]LastMessageInfo, error)
	// FindUnreadCounts returns the number of unread messages per room for the given user.
	FindUnreadCounts(userID uint, roomIDs []uint) (map[uint]int, error)
	// GetUserRoomIDs returns all room IDs the user is a member of.
	GetUserRoomIDs(userID uint) ([]uint, error)
}
