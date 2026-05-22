package chatdto

import utils "github.com/sunibgram/webapi/dto/_shared"

type ChatRoomParticipant struct {
	UserID      uint    `json:"user_id"`
	Username    string  `json:"username"`
	Name        string  `json:"name"`
	FacultyName string  `json:"faculty_name"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
}

type LastMessagePreview struct {
	ChatID    uint                `json:"chat_id"`
	SenderID  uint                `json:"sender_id"`
	Message   *string             `json:"message,omitempty"`
	GifURL    *string             `json:"gif_url,omitempty"`
	HasMedia  bool                `json:"has_media"`
	CreatedAt utils.FormattedTime `json:"created_at"`
}

type ChatRoomResponse struct {
	RoomID      uint                 `json:"room_id"`
	Name        *string              `json:"name,omitempty"`
	Description *string              `json:"description,omitempty"`
	AvatarURL   *string              `json:"avatar_url,omitempty"`
	Type        string               `json:"type"`
	MemberCount int                  `json:"member_count"`
	IsMuted     bool                 `json:"is_muted"`
	UnreadCount int                  `json:"unread_count"`
	OtherUser   *ChatRoomParticipant `json:"other_user,omitempty"`
	LastMessage *LastMessagePreview  `json:"last_message,omitempty"`
	CreatedAt   utils.FormattedTime  `json:"created_at"`
	UpdatedAt   utils.FormattedTime  `json:"updated_at"`
}

type GroupMemberResponse struct {
	UserID      uint   `json:"user_id"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	FacultyName string `json:"faculty_name"`
	AvatarURL   string `json:"avatar_url"`
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

type ChatReplyResponse struct {
	ChatID  uint       `json:"chat_id"`
	Sender  ChatSender `json:"sender"`
	Message *string    `json:"message,omitempty"`
	GifURL  *string    `json:"gif_url,omitempty"`
}

type ChatReactionSummary struct {
	Emoji   string `json:"emoji"`
	Count   int    `json:"count"`
	UserIDs []uint `json:"user_ids"`
}

type ChatResponse struct {
	ChatID    uint                  `json:"chat_id"`
	RoomID    uint                  `json:"room_id"`
	Sender    ChatSender            `json:"sender"`
	ReplyTo   *ChatReplyResponse    `json:"reply_to,omitempty"`
	Message   *string               `json:"message,omitempty"`
	GifURL    *string               `json:"gif_url,omitempty"`
	Media     []ChatMediaResponse   `json:"media"`
	Reactions []ChatReactionSummary `json:"reactions"`
	IsRead    bool                  `json:"is_read"`
	IsEdited  bool                  `json:"is_edited"`
	EditedAt  *utils.FormattedTime  `json:"edited_at,omitempty"`
	DeletedAt *utils.FormattedTime  `json:"deleted_at,omitempty"`
	CreatedAt utils.FormattedTime   `json:"created_at"`
}

type CreatePrivateChatRoomRequest struct {
	TargetUserID uint `json:"target_user_id"`
}

type CreateGroupChatRoomRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	MemberIDs   []uint  `json:"member_ids"`
}

type UpdateGroupChatRoomRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type AddGroupMemberRequest struct {
	UserIDs []uint `json:"user_ids"`
}

type SendMessageRequest struct {
	Message   *string `json:"message"`
	ReplyToID *uint   `json:"reply_to_id,omitempty"`
}

type AddReactionRequest struct {
	Emoji string `json:"emoji"`
}

type EditMessageRequest struct {
	Message string `json:"message"`
}

type MarkMessagesReadRequest struct {
	RoomID  uint   `json:"room_id"`
	ChatIDs []uint `json:"chat_ids"`
}
