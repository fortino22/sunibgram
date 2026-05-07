package controller

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	dtoutils "github.com/sunibgram/dto/utils"
	"github.com/sunibgram/webapi/ctxutil"
)

type ChatRoomController struct {
	chatRoomService service.IChatRoomService
}

func NewChatRoomController(chatRoomService service.IChatRoomService) *ChatRoomController {
	return &ChatRoomController{chatRoomService: chatRoomService}
}

func (c *ChatRoomController) CreatePrivate(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	var req dto.CreatePrivateChatRoomRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid payload"})
	}

	room, err := c.chatRoomService.CreatePrivate(ctx.Context(), claims.UserID, req.TargetUserID)
	if err != nil {
		return chatError(ctx, err)
	}
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"room_id": room.RoomID})
}

func (c *ChatRoomController) CreateGroup(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	var req dto.CreateGroupChatRoomRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid payload"})
	}

	room, err := c.chatRoomService.CreateGroup(ctx.Context(), claims.UserID, req.Name, req.Description, req.MemberIDs)
	if err != nil {
		return chatError(ctx, err)
	}
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"room_id": room.RoomID})
}

func (c *ChatRoomController) AddMember(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid chat room id"})
	}

	var req dto.AddGroupMemberRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid payload"})
	}

	room, err := c.chatRoomService.AddMemberToGroup(ctx.Context(), uint(id), claims.UserID, req.UserID)
	if err != nil {
		return chatError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(toChatRoomResponse(room))
}

func (c *ChatRoomController) UpdateGroup(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid chat room id"})
	}

	var req dto.UpdateGroupChatRoomRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid payload"})
	}

	if err := c.chatRoomService.UpdateGroup(ctx.Context(), uint(id), claims.UserID, req.Name, req.Description); err != nil {
		return chatError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "group updated successfully"})
}

func (c *ChatRoomController) GetMyRooms(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))

	rooms, total, err := c.chatRoomService.GetUserRooms(claims.UserID, page, limit)
	if err != nil {
		return chatError(ctx, err)
	}

	data := make([]dto.ChatRoomResponse, len(rooms))
	for i := range rooms {
		data[i] = toChatRoomResponse(&rooms[i])
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": data,
		"meta": dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *ChatRoomController) GetByID(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid chat room id"})
	}

	room, err := c.chatRoomService.GetByID(uint(id), claims.UserID)
	if err != nil {
		return chatError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(toChatRoomResponse(room))
}

func toChatRoomResponse(room *model.ChatRoom) dto.ChatRoomResponse {
	return dto.ChatRoomResponse{
		RoomID:      room.RoomID,
		Name:        room.Name,
		Description: room.Description,
		Type:        string(room.Type),
		CreatedAt:   dtoutils.Wrap(room.CreatedAt),
		UpdatedAt:   dtoutils.Wrap(room.UpdatedAt),
	}
}
