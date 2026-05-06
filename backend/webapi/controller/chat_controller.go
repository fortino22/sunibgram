package controller

import (
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	dtoutils "github.com/sunibgram/dto/utils"
	"github.com/sunibgram/webapi/ctxutil"
)

type ChatController struct {
	chatService      service.IChatService
	chatMediaService service.IChatMediaService
}

func NewChatController(chatService service.IChatService, chatMediaService service.IChatMediaService) *ChatController {
	return &ChatController{chatService: chatService, chatMediaService: chatMediaService}
}

func (c *ChatController) Send(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	roomID, err := strconv.ParseUint(ctx.Params("room_id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid chat room id"})
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

	chat, err := c.chatService.Send(ctx.Context(), uint(roomID), claims.UserID, message)
	if err != nil {
		return chatError(ctx, err)
	}

	if len(files) > 0 {
		_, err = c.chatMediaService.Upload(ctx.Context(), chat.ChatID, claims.UserID, files)
		if err != nil {
			return chatError(ctx, err)
		}
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "chat sent successfully"})
}

func (c *ChatController) GetMessages(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	roomID, err := strconv.ParseUint(ctx.Params("room_id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid chat room id"})
	}
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))

	if err := c.chatService.MarkRead(uint(roomID), claims.UserID); err != nil {
		return chatError(ctx, err)
	}
	chats, total, err := c.chatService.GetMessages(uint(roomID), claims.UserID, page, limit)
	if err != nil {
		return chatError(ctx, err)
	}

	mediaByChat := c.loadMediaByChat(chats)
	data := make([]dto.ChatResponse, len(chats))
	for i := range chats {
		data[i] = c.toChatResponse(&chats[i], mediaByChat[chats[i].ChatID])
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": data,
		"meta": dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *ChatController) Unsend(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid chat id"})
	}

	if err := c.chatService.Unsend(ctx.Context(), uint(id), claims.UserID); err != nil {
		return chatError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "chat unsent successfully"})
}

func (c *ChatController) EditMessage(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid chat id"})
	}

	var req dto.EditMessageRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}

	if err := c.chatService.UpdateMessage(ctx.Context(), uint(id), claims.UserID, req.Message); err != nil {
		return chatError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "chat updated successfully"})
}

func (c *ChatController) MarkMessagesRead(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
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
		return chatError(ctx, err)
	}
	data := make([]fiber.Map, len(results))
	for i, r := range results {
		data[i] = fiber.Map{"chat_id": r.ChatID, "status": r.Status}
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"results": data})
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

func (c *ChatController) toChatResponse(chat *model.Chat, media []model.ChatMedia) dto.ChatResponse {
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
		Media:     c.toChatMediaResponses(media),
		IsRead:    chat.IsRead,
		DeletedAt: deletedAt,
		CreatedAt: dtoutils.Wrap(chat.CreatedAt),
	}
}

func (c *ChatController) toChatMediaResponses(media []model.ChatMedia) []dto.ChatMediaResponse {
	out := make([]dto.ChatMediaResponse, len(media))
	for i := range media {
		out[i] = dto.ChatMediaResponse{
			MediaID:   media[i].MediaID,
			URL:       c.chatMediaService.GetURL(media[i].Bucket, media[i].ObjectName),
			MediaType: media[i].MediaType,
			FileSize:  media[i].FileSize,
		}
	}
	return out
}

func chatError(ctx *fiber.Ctx, err error) error {
	switch err.Error() {
	case "forbidden":
		return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
	case "chat room not found", "chat not found", "target user not found", "member user not found":
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
	default:
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}
}
