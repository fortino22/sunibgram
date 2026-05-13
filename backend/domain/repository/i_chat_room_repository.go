package repository

import "github.com/sunibgram/domain/model"

// ChatRoomMemberInfo holds basic public info about another room participant.
type ChatRoomMemberInfo struct {
	UserID      uint
	Username    string
	Name        string
	FacultyName string
}

type IChatRoomRepository interface {
	Save(room *model.ChatRoom) error
	FindByID(id uint) (*model.ChatRoom, error)
	FindByUserID(userID uint, page, limit int) ([]model.ChatRoom, int64, error)
	// FindOtherMemberInfo returns the other participant's info for private rooms,
	// keyed by room_id. Rooms without a counterpart are omitted from the map.
	FindOtherMemberInfo(roomIDs []uint, excludeUserID uint) (map[uint]ChatRoomMemberInfo, error)
	IsMember(roomID, userID uint) bool
	CountMembers(roomID uint) int64
	AddMember(roomID, userID uint) error
	UpdateGroup(roomID uint, name string, description *string) error
	DeleteGroup(roomID uint) error
	FindPrivateRoom(userA, userB uint) (*model.ChatRoom, error)
}
