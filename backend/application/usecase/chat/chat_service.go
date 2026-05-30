package chatuc

import (
	"context"
	"errors"
	"log"
	"strings"
	"unicode/utf8"

	"github.com/sunibgram/application/helper"
	chatmodel "github.com/sunibgram/domain/model/chat"
	chatroommodel "github.com/sunibgram/domain/model/chat_room"
	chatrepoiface "github.com/sunibgram/domain/repository/chat"
	chatreadrepoiface "github.com/sunibgram/domain/repository/chat_read"
	chatroomrepoiface "github.com/sunibgram/domain/repository/chat_room"
	userblockrepoiface "github.com/sunibgram/domain/repository/user_block"
	chatsvciface "github.com/sunibgram/domain/service/chat"
	notificationsvciface "github.com/sunibgram/domain/service/notification"
)

type ChatService struct {
	chatRepo            chatrepoiface.IChatRepository
	roomRepo            chatroomrepoiface.IChatRoomRepository
	chatReadRepo        chatreadrepoiface.IChatReadRepository
	blockRepo           userblockrepoiface.IUserBlockRepository
	notificationService notificationsvciface.INotificationService
}

func NewChatService(
	chatRepo chatrepoiface.IChatRepository,
	roomRepo chatroomrepoiface.IChatRoomRepository,
	chatReadRepo chatreadrepoiface.IChatReadRepository,
	blockRepo userblockrepoiface.IUserBlockRepository,
	notificationServices ...notificationsvciface.INotificationService,
) chatsvciface.IChatService {
	var notificationService notificationsvciface.INotificationService
	if len(notificationServices) > 0 {
		notificationService = notificationServices[0]
	}
	return &ChatService{
		chatRepo:            chatRepo,
		roomRepo:            roomRepo,
		chatReadRepo:        chatReadRepo,
		blockRepo:           blockRepo,
		notificationService: notificationService,
	}
}

// privateRoomBlocked reports whether roomID is a private room whose other member
// is in a mutual block relationship with userID. Group rooms always return false
// (FindOtherMemberInfo only resolves private-room counterparts). Used to seal an
// existing A↔B DM after a block — IsMember alone still passes post-block.
func (s *ChatService) privateRoomBlocked(roomID, userID uint) bool {
	if s.blockRepo == nil {
		return false
	}
	memberMap, err := s.roomRepo.FindOtherMemberInfo([]uint{roomID}, userID)
	if err != nil {
		return false
	}
	info, ok := memberMap[roomID]
	if !ok {
		return false
	}
	blocked, err := s.blockRepo.IsBlocked(userID, info.UserID)
	return err == nil && blocked
}

func (s *ChatService) Send(ctx context.Context, roomID, senderID uint, message *string, gifURL *string, replyToID *uint, sharedThreadID *uint) (*chatmodel.Chat, error) {
	if !s.roomRepo.IsMember(roomID, senderID) {
		return nil, errors.New("forbidden")
	}
	// Seal an existing private DM once a block is in place (either direction).
	if s.privateRoomBlocked(roomID, senderID) {
		return nil, errors.New("blocked")
	}
	room, err := s.roomRepo.FindByID(roomID)
	if err != nil {
		return nil, errors.New("chat room not found")
	}
	if message != nil {
		trimmed := strings.TrimSpace(*message)
		if trimmed == "" {
			message = nil
		} else {
			if utf8.RuneCountInString(trimmed) > 5000 {
				return nil, errors.New("message must be 5000 characters or less")
			}
			message = &trimmed
		}
	}
	if gifURL != nil {
		trimmed := strings.TrimSpace(*gifURL)
		if utf8.RuneCountInString(trimmed) > 2000 {
			return nil, errors.New("gifURL must be 2000 characters or less")
		}
		*gifURL = trimmed
	}
	roomKey, err := s.roomKey(room)
	if err != nil {
		return nil, err
	}
	if message != nil {
		encrypted, err := helper.ChatEncrypt([]byte(*message), roomKey)
		if err != nil {
			return nil, err
		}
		message = &encrypted
	}
	chat := &chatmodel.Chat{RoomID: roomID, SenderID: senderID, Message: message, GifURL: gifURL, ReplyToID: replyToID, SharedThreadID: sharedThreadID}
	if err := s.chatRepo.Save(chat); err != nil {
		return nil, err
	}
	s.decryptChat(chat, roomKey)
	_ = s.roomRepo.TouchUpdatedAt(roomID)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("chat notif fanout panic: %v", r)
			}
		}()
		s.sendMessageNotifications(room, chat)
	}()
	return chat, nil
}

func (s *ChatService) sendMessageNotifications(room *chatroommodel.ChatRoom, chat *chatmodel.Chat) {
	if s.notificationService == nil || room == nil || chat == nil {
		return
	}

	members, err := s.roomRepo.FindMembers(chat.RoomID)
	if err != nil {
		return
	}

	preview := chatMessagePreview(chat)
	referenceID := chat.RoomID
	payload := map[string]any{
		"sender": map[string]any{
			"user_id": chat.SenderID,
		},
		"sender_user_id":  chat.SenderID,
		"chat_id":         chat.ChatID,
		"room_id":         chat.RoomID,
		"conversation_id": chat.RoomID,
		"message":         preview,
		"preview":         preview,
		"message_type":    string(room.Type),
	}

	seen := make(map[uint]struct{}, len(members))
	for _, member := range members {
		if member.UserID == chat.SenderID {
			continue
		}
		if _, ok := seen[member.UserID]; ok {
			continue
		}
		seen[member.UserID] = struct{}{}

		_ = s.notificationService.SendNotification(notificationsvciface.CreateNotificationInput{
			Type:            "chat_message",
			Message:         chatNotificationMessage(chat),
			SenderUserID:    chat.SenderID,
			RecipientUserID: member.UserID,
			ReferenceID:     &referenceID,
			Data:            payload,
		})
	}
}

func chatMessagePreview(chat *chatmodel.Chat) string {
	if chat == nil {
		return "sent a message"
	}
	if chat.Message != nil {
		trimmed := strings.TrimSpace(*chat.Message)
		if trimmed != "" {
			return trimmed
		}
	}
	if chat.GifURL != nil && strings.TrimSpace(*chat.GifURL) != "" {
		return "sent a GIF"
	}
	return "sent an attachment"
}

func chatNotificationMessage(chat *chatmodel.Chat) string {
	if chat == nil {
		return "sent a message"
	}
	if chat.GifURL != nil && strings.TrimSpace(*chat.GifURL) != "" {
		return "sent a GIF"
	}
	return "sent a message"
}

func (s *ChatService) GetByID(ctx context.Context, chatID uint) (*chatmodel.Chat, error) {
	chat, err := s.chatRepo.FindByID(chatID)
	if err != nil {
		return nil, errors.New("chat not found")
	}
	room, err := s.roomRepo.FindByID(chat.RoomID)
	if err != nil {
		return nil, errors.New("chat room not found")
	}
	roomKey, err := s.roomKey(room)
	if err != nil {
		return nil, err
	}
	s.decryptChat(chat, roomKey)
	return chat, nil
}

func (s *ChatService) GetMessages(roomID, userID uint, page, limit int) ([]chatmodel.Chat, int64, error) {
	if !s.roomRepo.IsMember(roomID, userID) {
		return nil, 0, errors.New("forbidden")
	}
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	chats, total, err := s.chatRepo.FindByRoomID(roomID, page, limit)
	if err != nil {
		return nil, 0, err
	}
	room, err := s.roomRepo.FindByID(roomID)
	if err != nil {
		return nil, 0, errors.New("chat room not found")
	}
	roomKey, err := s.roomKey(room)
	if err != nil {
		return nil, 0, err
	}
	for i := range chats {
		s.decryptChat(&chats[i], roomKey)
	}
	return chats, total, nil
}

func (s *ChatService) Unsend(ctx context.Context, chatID, senderID uint) error {
	chat, err := s.chatRepo.FindByID(chatID)
	if err != nil {
		return errors.New("chat not found")
	}
	if chat.SenderID != senderID {
		return errors.New("forbidden")
	}
	return s.chatRepo.SoftDelete(chatID)
}

func (s *ChatService) Delete(ctx context.Context, chatID uint) error {
	return s.chatRepo.Delete(chatID)
}

func (s *ChatService) UpdateMessage(ctx context.Context, chatID, senderID uint, message string) error {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return errors.New("message must not be empty")
	}
	if utf8.RuneCountInString(trimmed) > 5000 {
		return errors.New("message must be 5000 characters or less")
	}

	chat, err := s.chatRepo.FindByID(chatID)
	if err != nil {
		return errors.New("chat not found")
	}
	if chat.SenderID != senderID {
		return errors.New("forbidden")
	}
	room, err := s.roomRepo.FindByID(chat.RoomID)
	if err != nil {
		return errors.New("chat room not found")
	}
	roomKey, err := s.roomKey(room)
	if err != nil {
		return err
	}
	encrypted, err := helper.ChatEncrypt([]byte(trimmed), roomKey)
	if err != nil {
		return err
	}
	return s.chatRepo.UpdateMessage(chatID, encrypted)
}

func (s *ChatService) MarkRead(roomID, userID uint) ([]uint, error) {
	if !s.roomRepo.IsMember(roomID, userID) {
		return nil, errors.New("forbidden")
	}

	room, err := s.roomRepo.FindByID(roomID)
	if err != nil {
		return nil, errors.New("chat room not found")
	}
	if room.Type == chatroommodel.ChatRoomTypePrivate {
		ids, err := s.chatRepo.MarkRoomRead(roomID, userID)
		return ids, err
	}

	inserted, err := s.chatReadRepo.BatchMarkRead(roomID, userID)
	if err != nil {
		return nil, err
	}
	memberCount := int64(room.MemberCount)
	readCounts, err := s.chatReadRepo.CountReadsBatch(context.Background(), inserted)
	if err != nil {
		return nil, err
	}
	toMark := make([]uint, 0, len(inserted))
	for _, chatID := range inserted {
		if readCounts[chatID] >= memberCount-1 {
			toMark = append(toMark, chatID)
		}
	}
	if len(toMark) > 0 {
		_ = s.chatRepo.UpdateIsReadBulk(toMark)
	}
	return inserted, nil
}

func (s *ChatService) roomKey(room *chatroommodel.ChatRoom) ([]byte, error) {
	if room == nil {
		return nil, errors.New("chat room not found")
	}
	masterKey, err := helper.ChatMasterKeyFromEnv("CHAT_MASTER_KEY")
	if err != nil {
		return nil, err
	}
	if room.EncryptedKey != nil && strings.TrimSpace(*room.EncryptedKey) != "" {
		return helper.ChatDecrypt(*room.EncryptedKey, masterKey)
	}
	roomKey, err := helper.ChatGenerateKey()
	if err != nil {
		return nil, err
	}
	encryptedKey, err := helper.ChatEncrypt(roomKey, masterKey)
	if err != nil {
		return nil, err
	}
	if err := s.roomRepo.UpdateEncryptedKey(room.RoomID, encryptedKey); err != nil {
		return nil, err
	}
	room.EncryptedKey = &encryptedKey
	return roomKey, nil
}

func (s *ChatService) decryptChat(chat *chatmodel.Chat, roomKey []byte) {
	if chat == nil {
		return
	}
	decryptMessage := func(message *string) {
		if message == nil || strings.TrimSpace(*message) == "" {
			return
		}
		plaintext, err := helper.ChatDecrypt(*message, roomKey)
		if err != nil {
			return
		}
		value := string(plaintext)
		*message = value
	}
	decryptMessage(chat.Message)
	if chat.ReplyTo != nil {
		decryptMessage(chat.ReplyTo.Message)
	}
}

func (s *ChatService) MarkMessagesRead(ctx context.Context, roomID uint, chatIDs []uint, userID uint) ([]chatsvciface.ChatMarkResult, error) {
	if !s.roomRepo.IsMember(roomID, userID) {
		return nil, errors.New("forbidden")
	}
	// No read receipts to a blocked counterpart in an existing private DM.
	if s.privateRoomBlocked(roomID, userID) {
		return nil, errors.New("blocked")
	}
	room, err := s.roomRepo.FindByID(roomID)
	if err != nil {
		return nil, errors.New("chat room not found")
	}

	results := make([]chatsvciface.ChatMarkResult, 0, len(chatIDs))
	chats, err := s.chatRepo.FindByIDs(chatIDs)
	if err != nil {
		for _, chatID := range chatIDs {
			results = append(results, chatsvciface.ChatMarkResult{ChatID: chatID, Status: "not_found"})
		}
		return results, nil
	}
	chatByID := make(map[uint]*chatmodel.Chat, len(chats))
	for i := range chats {
		chatByID[chats[i].ChatID] = &chats[i]
	}
	for _, chatID := range chatIDs {
		chat, ok := chatByID[chatID]
		if !ok {
			results = append(results, chatsvciface.ChatMarkResult{ChatID: chatID, Status: "not_found"})
			continue
		}
		if chat.RoomID != roomID || chat.SenderID == userID {
			results = append(results, chatsvciface.ChatMarkResult{ChatID: chatID, SenderID: chat.SenderID, Status: "skipped"})
			continue
		}
		if chat.IsRead {
			results = append(results, chatsvciface.ChatMarkResult{ChatID: chatID, SenderID: chat.SenderID, Status: "already_read"})
			continue
		}
		if room.Type == chatroommodel.ChatRoomTypePrivate {
			if err := s.chatRepo.UpdateIsRead(chatID); err != nil {
				return nil, err
			}
		} else {
			if err := s.chatReadRepo.MarkRead(chatID, userID); err != nil {
				return nil, err
			}
			if s.chatReadRepo.CountReads(chatID) >= int64(room.MemberCount)-1 {
				_ = s.chatRepo.UpdateIsRead(chatID)
			}
		}
		results = append(results, chatsvciface.ChatMarkResult{ChatID: chatID, SenderID: chat.SenderID, Status: "marked"})
	}
	return results, nil
}
