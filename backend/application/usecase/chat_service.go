package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/sunibgram/application/helper"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/repository"
	"github.com/sunibgram/domain/service"
)

type ChatService struct {
	chatRepo            repository.IChatRepository
	roomRepo            repository.IChatRoomRepository
	chatReadRepo        repository.IChatReadRepository
	notificationService service.INotificationService
}

func NewChatService(
	chatRepo repository.IChatRepository,
	roomRepo repository.IChatRoomRepository,
	chatReadRepo repository.IChatReadRepository,
	notificationServices ...service.INotificationService,
) service.IChatService {
	var notificationService service.INotificationService
	if len(notificationServices) > 0 {
		notificationService = notificationServices[0]
	}
	return &ChatService{
		chatRepo:            chatRepo,
		roomRepo:            roomRepo,
		chatReadRepo:        chatReadRepo,
		notificationService: notificationService,
	}
}

func (s *ChatService) Send(ctx context.Context, roomID, senderID uint, message *string, gifURL *string, replyToID *uint) (*model.Chat, error) {
	if !s.roomRepo.IsMember(roomID, senderID) {
		return nil, errors.New("forbidden")
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
			message = &trimmed
		}
	}
	chat := &model.Chat{RoomID: roomID, SenderID: senderID, Message: message, GifURL: gifURL, ReplyToID: replyToID}
	if err := s.chatRepo.Save(chat); err != nil {
		return nil, err
	}
	_ = s.roomRepo.TouchUpdatedAt(roomID)
	s.sendMessageNotifications(room, chat)
	return chat, nil
}

func (s *ChatService) sendMessageNotifications(room *model.ChatRoom, chat *model.Chat) {
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

		_ = s.notificationService.SendNotification(service.CreateNotificationInput{
			Type:            "chat_message",
			Message:         preview,
			SenderUserID:    chat.SenderID,
			RecipientUserID: member.UserID,
			ReferenceID:     &referenceID,
			Data:            payload,
		})
	}
}

func chatMessagePreview(chat *model.Chat) string {
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

func (s *ChatService) GetByID(ctx context.Context, chatID uint) (*model.Chat, error) {
	chat, err := s.chatRepo.FindByID(chatID)
	if err != nil {
		return nil, errors.New("chat not found")
	}
	return chat, nil
}

func (s *ChatService) GetMessages(roomID, userID uint, page, limit int) ([]model.Chat, int64, error) {
	if !s.roomRepo.IsMember(roomID, userID) {
		return nil, 0, errors.New("forbidden")
	}
	page, limit = helper.NormalizePageLimit(page, limit, 20)
	return s.chatRepo.FindByRoomID(roomID, page, limit)
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

	chat, err := s.chatRepo.FindByID(chatID)
	if err != nil {
		return errors.New("chat not found")
	}
	if chat.SenderID != senderID {
		return errors.New("forbidden")
	}
	return s.chatRepo.UpdateMessage(chatID, trimmed)
}

func (s *ChatService) MarkRead(roomID, userID uint) ([]uint, error) {
	if !s.roomRepo.IsMember(roomID, userID) {
		return nil, errors.New("forbidden")
	}

	room, err := s.roomRepo.FindByID(roomID)
	if err != nil {
		return nil, errors.New("chat room not found")
	}
	if room.Type == model.ChatRoomTypePrivate {
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
	for _, chatID := range inserted {
		if readCounts[chatID] >= memberCount-1 {
			_ = s.chatRepo.UpdateIsRead(chatID)
		}
	}
	return inserted, nil
}

func (s *ChatService) MarkMessagesRead(ctx context.Context, roomID uint, chatIDs []uint, userID uint) ([]service.ChatMarkResult, error) {
	if !s.roomRepo.IsMember(roomID, userID) {
		return nil, errors.New("forbidden")
	}
	room, err := s.roomRepo.FindByID(roomID)
	if err != nil {
		return nil, errors.New("chat room not found")
	}

	results := make([]service.ChatMarkResult, 0, len(chatIDs))
	chats, err := s.chatRepo.FindByIDs(chatIDs)
	if err != nil {
		for _, chatID := range chatIDs {
			results = append(results, service.ChatMarkResult{ChatID: chatID, Status: "not_found"})
		}
		return results, nil
	}
	chatByID := make(map[uint]*model.Chat, len(chats))
	for i := range chats {
		chatByID[chats[i].ChatID] = &chats[i]
	}
	for _, chatID := range chatIDs {
		chat, ok := chatByID[chatID]
		if !ok {
			results = append(results, service.ChatMarkResult{ChatID: chatID, Status: "not_found"})
			continue
		}
		if chat.RoomID != roomID || chat.SenderID == userID {
			results = append(results, service.ChatMarkResult{ChatID: chatID, SenderID: chat.SenderID, Status: "skipped"})
			continue
		}
		if chat.IsRead {
			results = append(results, service.ChatMarkResult{ChatID: chatID, SenderID: chat.SenderID, Status: "already_read"})
			continue
		}
		if room.Type == model.ChatRoomTypePrivate {
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
		results = append(results, service.ChatMarkResult{ChatID: chatID, SenderID: chat.SenderID, Status: "marked"})
	}
	return results, nil
}
