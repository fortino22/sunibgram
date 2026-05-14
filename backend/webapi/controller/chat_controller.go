package controller

import (
	"context"
	"log"
	"mime/multipart"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	dtoutils "github.com/sunibgram/dto/utils"
)

type ChatController struct {
	chatService      service.IChatService
	chatMediaService service.IChatMediaService
	hub              service.IHub
}

func NewChatController(chatService service.IChatService, chatMediaService service.IChatMediaService, hubs ...service.IHub) *ChatController {
	var hub service.IHub
	if len(hubs) > 0 {
		hub = hubs[0]
	}
	return &ChatController{chatService: chatService, chatMediaService: chatMediaService, hub: hub}
}

func (c *ChatController) Send(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	roomID, ok := ParseIDParam(ctx, "room_id", "chat room")
	if !ok {
		return nil
	}

	var message *string
	if raw := strings.TrimSpace(ctx.FormValue("message")); raw != "" {
		message = &raw
	}

	var files []*multipart.FileHeader
	form, _ := ctx.MultipartForm()
	if form != nil {
		files = form.File["files"]
	}
	if message == nil && len(files) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "message or media is required"})
	}

	chat, err := c.chatService.Send(ctx.Context(), roomID, claims.UserID, message)
	if err != nil {
		return HandleServiceError(ctx, err)
	}

	var media []model.ChatMedia
	if len(files) > 0 {
		media, err = c.chatMediaService.Upload(ctx.Context(), chat.ChatID, claims.UserID, files)
		if err != nil {
			_ = c.chatService.Delete(ctx.Context(), chat.ChatID)
			return HandleServiceError(ctx, err)
		}
	}

	c.broadcastRoom(chat.RoomID, "message:new", c.toChatResponse(ctx.Context(), chat, media))

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "chat sent successfully"})
}

func (c *ChatController) GetMessages(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	roomID, ok := ParseIDParam(ctx, "room_id", "chat room")
	if !ok {
		return nil
	}
	page, limit := ParsePagination(ctx)

	markedIDs, err := c.chatService.MarkRead(roomID, claims.UserID)
	if err != nil {
		return HandleServiceError(ctx, err)
	}
	chats, total, err := c.chatService.GetMessages(roomID, claims.UserID, page, limit)
	if err != nil {
		return HandleServiceError(ctx, err)
	}

	mediaByChat := c.loadMediaByChat(chats)
	data := make([]dto.ChatResponse, len(chats))
	for i := range chats {
		data[i] = c.toChatResponse(ctx.Context(), &chats[i], mediaByChat[chats[i].ChatID])
	}

	if len(markedIDs) > 0 {
		marked := make(map[uint]bool, len(markedIDs))
		for _, id := range markedIDs {
			marked[id] = true
		}
		var results []service.ChatMarkResult
		for i := range chats {
			if marked[chats[i].ChatID] {
				results = append(results, service.ChatMarkResult{
					ChatID:   chats[i].ChatID,
					SenderID: chats[i].SenderID,
					Status:   "marked",
				})
			}
		}
		c.broadcastReadReceipts(roomID, claims.UserID, results)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": data,
		"meta": dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *ChatController) Unsend(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ParseIDParam(ctx, "id", "chat")
	if !ok {
		return nil
	}

	chat, err := c.chatService.GetByID(ctx.Context(), id)
	if err != nil {
		return HandleServiceError(ctx, err)
	}

	if err := c.chatService.Unsend(ctx.Context(), id, claims.UserID); err != nil {
		return HandleServiceError(ctx, err)
	}

	c.broadcastRoom(chat.RoomID, "message:deleted", fiber.Map{
		"chat_id":    chat.ChatID,
		"room_id":    chat.RoomID,
		"sender_id":  chat.SenderID,
		"deleted_by": claims.UserID,
	})

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "chat unsent successfully"})
}

func (c *ChatController) EditMessage(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ParseIDParam(ctx, "id", "chat")
	if !ok {
		return nil
	}

	var req dto.EditMessageRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}

	chat, err := c.chatService.GetByID(ctx.Context(), id)
	if err != nil {
		return HandleServiceError(ctx, err)
	}

	if err := c.chatService.UpdateMessage(ctx.Context(), id, claims.UserID, req.Message); err != nil {
		return HandleServiceError(ctx, err)
	}

	c.broadcastRoom(chat.RoomID, "message:edited", fiber.Map{
		"chat_id":   chat.ChatID,
		"room_id":   chat.RoomID,
		"sender_id": chat.SenderID,
		"message":   strings.TrimSpace(req.Message),
	})

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "chat updated successfully"})
}

func (c *ChatController) MarkMessagesRead(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	var req dto.MarkMessagesReadRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}
	if req.RoomID == 0 || len(req.ChatIDs) == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}

	results, err := c.chatService.MarkMessagesRead(ctx.Context(), req.RoomID, req.ChatIDs, claims.UserID)
	if err != nil {
		return HandleServiceError(ctx, err)
	}
	data := make([]fiber.Map, len(results))
	for i, r := range results {
		data[i] = fiber.Map{"chat_id": r.ChatID, "status": r.Status}
	}

	c.broadcastReadReceipts(req.RoomID, claims.UserID, results)

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"results": data})
}

func (c *ChatController) broadcastRoom(roomID uint, event string, payload any) {
	if c.hub == nil {
		return
	}
	c.hub.Broadcast(roomID, event, payload)
}

func (c *ChatController) broadcastReadReceipts(roomID, readBy uint, results []service.ChatMarkResult) {
	if c.hub == nil {
		return
	}

	chatIDsBySender := make(map[uint][]uint)
	for _, result := range results {
		if result.Status != "marked" || result.SenderID == 0 {
			continue
		}
		chatIDsBySender[result.SenderID] = append(chatIDsBySender[result.SenderID], result.ChatID)
	}

	for senderID, chatIDs := range chatIDsBySender {
		c.hub.EmitToUser(senderID, "message:read", fiber.Map{
			"room_id":  roomID,
			"chat_ids": chatIDs,
			"read_by":  readBy,
		})
	}
}

func (c *ChatController) loadMediaByChat(chats []model.Chat) map[uint][]model.ChatMedia {
	ids := make([]uint, len(chats))
	for i := range chats {
		ids[i] = chats[i].ChatID
	}
	allMedia, _ := c.chatMediaService.GetByChatIDs(ids)
	byChat := make(map[uint][]model.ChatMedia, len(ids))
	for _, m := range allMedia {
		byChat[m.ChatID] = append(byChat[m.ChatID], m)
	}
	return byChat
}

func (c *ChatController) toChatResponse(ctx context.Context, chat *model.Chat, media []model.ChatMedia) dto.ChatResponse {
	var deletedAt *dtoutils.FormattedTime
	if chat.DeletedAt != nil {
		wrapped := dtoutils.Wrap(*chat.DeletedAt)
		deletedAt = &wrapped
	}
	return dto.ChatResponse{
		ChatID: chat.ChatID,
		RoomID: chat.RoomID,
		Sender: dto.ChatSender{
			ID:       chat.SenderID,
			Username: chat.Sender.UserName,
		},
		Message:   chat.Message,
		Media:     c.toChatMediaResponses(ctx, media),
		IsRead:    chat.IsRead,
		DeletedAt: deletedAt,
		CreatedAt: dtoutils.Wrap(chat.CreatedAt),
	}
}

func (c *ChatController) toChatMediaResponses(ctx context.Context, media []model.ChatMedia) []dto.ChatMediaResponse {
	out := make([]dto.ChatMediaResponse, len(media))
	for i := range media {
		url, err := c.chatMediaService.GeneratePresignedURL(ctx, media[i].Bucket, media[i].ObjectName)
		if err != nil {
			log.Printf("presign error bucket=%s object=%s: %v", media[i].Bucket, media[i].ObjectName, err)
		}
		out[i] = dto.ChatMediaResponse{
			MediaID:   media[i].MediaID,
			URL:       url,
			MediaType: media[i].MediaType,
			FileSize:  media[i].FileSize,
		}
	}
	return out
}
