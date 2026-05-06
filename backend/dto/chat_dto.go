package dto

import "github.com/sunibgram/dto/utils"

type ChatRoomResponse struct {
	RoomID    uint                `json:"room_id"`
	Name      *string             `json:"name,omitempty"`
	Type      string              `json:"type"`
	CreatedAt utils.FormattedTime `json:"created_at"`
	UpdatedAt utils.FormattedTime `json:"updated_at"`
}

type ChatMediaResponse struct {
	MediaID   uint   `json:"media_id"`
	URL       string `json:"url"`
	MediaType string `json:"media_type"`
	FileSize  int64  `json:"file_size"`
}

type ChatSender struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

type ChatResponse struct {
	ChatID    uint                 `json:"chat_id"`
	RoomID    uint                 `json:"room_id"`
	Sender    ChatSender           `json:"sender"`
	Message   *string              `json:"message,omitempty"`
	Media     []ChatMediaResponse  `json:"media"`
	IsRead    bool                 `json:"is_read"`
	DeletedAt *utils.FormattedTime `json:"deleted_at,omitempty"`
	CreatedAt utils.FormattedTime  `json:"created_at"`
}

type CreatePrivateChatRoomRequest struct {
	TargetUserID uint `json:"target_user_id"`
}

type CreateGroupChatRoomRequest struct {
	Name      string `json:"name"`
	MemberIDs []uint `json:"member_ids"`
}

type SendMessageRequest struct {
	Message *string `json:"message"`
}
