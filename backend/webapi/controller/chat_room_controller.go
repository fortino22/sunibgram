package controller

import (
	"mime/multipart"

	"github.com/gofiber/fiber/v2"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	dtoutils "github.com/sunibgram/webapi/dto/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	chatdto "github.com/sunibgram/webapi/dto/chat"
	threaddto "github.com/sunibgram/webapi/dto/thread"
)

type ChatRoomController struct {
	chatRoomService service.IChatRoomService
	hub             service.IHub
}

func NewChatRoomController(chatRoomService service.IChatRoomService, hubs ...service.IHub) *ChatRoomController {
	var hub service.IHub
	if len(hubs) > 0 {
		hub = hubs[0]
	}
	return &ChatRoomController{chatRoomService: chatRoomService, hub: hub}
}

func (c *ChatRoomController) CreatePrivate(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	var req chatdto.CreatePrivateChatRoomRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid payload"})
	}

	room, err := c.chatRoomService.CreatePrivate(ctx.Context(), claims.UserID, req.TargetUserID)
	if err != nil {
		return HandleServiceError(ctx, err)
	}

	// Notify the target user so their conversation list updates without a refresh
	if c.hub != nil {
		if roomForTarget, err := c.chatRoomService.GetRoomForUser(room.RoomID, req.TargetUserID); err == nil {
			c.hub.EmitToUser(req.TargetUserID, "room:new", toChatRoomResponseWithParticipant(roomForTarget))
		}
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"room_id": room.RoomID})
}

func (c *ChatRoomController) CreateGroup(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	var req chatdto.CreateGroupChatRoomRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid payload"})
	}

	room, err := c.chatRoomService.CreateGroup(ctx.Context(), claims.UserID, req.Name, req.Description, req.MemberIDs)
	if err != nil {
		return HandleServiceError(ctx, err)
	}

	// Notify all non-creator members so their conversation list updates without a refresh
	if c.hub != nil {
		payload := toChatRoomResponse(room)
		for _, memberID := range req.MemberIDs {
			if memberID != claims.UserID {
				c.hub.EmitToUser(memberID, "room:new", payload)
			}
		}
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"room_id": room.RoomID})
}

func (c *ChatRoomController) AddMember(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ParseIDParam(ctx, "id", "chat room")
	if !ok {
		return nil
	}

	var req chatdto.AddGroupMemberRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid payload"})
	}

	room, alreadyMembers, err := c.chatRoomService.AddMemberToGroup(ctx.Context(), id, claims.UserID, req.UserIDs)
	if err != nil {
		return HandleServiceError(ctx, err)
	}
	if alreadyMembers == nil {
		alreadyMembers = []uint{}
	}

	addedUserIDs := addedMembers(req.UserIDs, alreadyMembers)
	if c.hub != nil && len(addedUserIDs) > 0 {
		c.hub.Broadcast(id, "room:member_added", fiber.Map{
			"room_id":         room.RoomID,
			"added_user_ids":  addedUserIDs,
			"added_by":        claims.UserID,
			"already_members": alreadyMembers,
			"room":            toChatRoomResponse(room),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":         "members added successfully",
		"already_members": alreadyMembers,
	})
}

// UpdateGroup accepts multipart/form-data so an avatar image can be uploaded alongside the name/description.
func (c *ChatRoomController) UpdateGroup(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ParseIDParam(ctx, "id", "chat room")
	if !ok {
		return nil
	}

	name := ctx.FormValue("name")
	descStr := ctx.FormValue("description")
	var description *string
	if descStr != "" {
		description = &descStr
	}

	var avatarFile *multipart.FileHeader
	if f, err := ctx.FormFile("avatar"); err == nil {
		avatarFile = f
	}

	if err := c.chatRoomService.UpdateGroup(ctx.Context(), id, claims.UserID, name, description, avatarFile); err != nil {
		return HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "group updated successfully"})
}

func (c *ChatRoomController) GetMembers(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ParseIDParam(ctx, "id", "chat room")
	if !ok {
		return nil
	}

	members, err := c.chatRoomService.GetGroupMembers(ctx.Context(), id, claims.UserID)
	if err != nil {
		return HandleServiceError(ctx, err)
	}

	data := make([]chatdto.GroupMemberResponse, len(members))
	for i, m := range members {
		data[i] = chatdto.GroupMemberResponse{
			UserID:      m.UserID,
			Username:    m.Username,
			Name:        m.Name,
			FacultyName: m.FacultyName,
			AvatarURL:   m.AvatarURL,
		}
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": data})
}

func (c *ChatRoomController) LeaveGroup(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ParseIDParam(ctx, "id", "chat room")
	if !ok {
		return nil
	}

	if err := c.chatRoomService.LeaveGroup(ctx.Context(), id, claims.UserID); err != nil {
		return HandleServiceError(ctx, err)
	}
	if c.hub != nil {
		c.hub.Broadcast(id, "room:member_left", fiber.Map{
			"room_id": id,
			"user_id": claims.UserID,
		})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "left group successfully"})
}

func (c *ChatRoomController) DeleteGroup(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ParseIDParam(ctx, "id", "chat room")
	if !ok {
		return nil
	}

	if err := c.chatRoomService.DeleteGroup(ctx.Context(), id, claims.UserID); err != nil {
		return HandleServiceError(ctx, err)
	}
	if c.hub != nil {
		c.hub.Broadcast(id, "group_chat_deleted", fiber.Map{
			"room_id":    id,
			"deleted_by": claims.UserID,
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "group deleted successfully"})
}

func (c *ChatRoomController) GetMyRooms(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	page, limit := ParsePagination(ctx)

	rooms, total, err := c.chatRoomService.GetUserRoomsWithParticipants(claims.UserID, page, limit)
	if err != nil {
		return HandleServiceError(ctx, err)
	}

	data := make([]chatdto.ChatRoomResponse, len(rooms))
	for i := range rooms {
		data[i] = toChatRoomResponseWithParticipant(&rooms[i])
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": data,
		"meta": threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *ChatRoomController) GetByID(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ParseIDParam(ctx, "id", "chat room")
	if !ok {
		return nil
	}

	room, err := c.chatRoomService.GetByID(id, claims.UserID)
	if err != nil {
		return HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(toChatRoomResponse(room))
}

func toChatRoomResponse(room *model.ChatRoom) chatdto.ChatRoomResponse {
	return chatdto.ChatRoomResponse{
		RoomID:      room.RoomID,
		Name:        room.Name,
		Description: room.Description,
		AvatarURL:   room.AvatarURL,
		Type:        string(room.Type),
		MemberCount: room.MemberCount,
		CreatedAt:   dtoutils.Wrap(room.CreatedAt),
		UpdatedAt:   dtoutils.Wrap(room.UpdatedAt),
	}
}

func toChatRoomResponseWithParticipant(room *service.ChatRoomWithParticipant) chatdto.ChatRoomResponse {
	resp := toChatRoomResponse(&room.ChatRoom)
	resp.IsMuted = room.IsMuted
	resp.UnreadCount = room.UnreadCount
	if room.OtherUser != nil {
		var avatarURL *string
		if room.OtherUser.AvatarURL != "" {
			avatarURL = &room.OtherUser.AvatarURL
		}
		resp.OtherUser = &chatdto.ChatRoomParticipant{
			UserID:      room.OtherUser.UserID,
			Username:    room.OtherUser.Username,
			Name:        room.OtherUser.Name,
			FacultyName: room.OtherUser.FacultyName,
			AvatarURL:   avatarURL,
		}
	}
	if room.LastMessage != nil {
		resp.LastMessage = &chatdto.LastMessagePreview{
			ChatID:    room.LastMessage.ChatID,
			SenderID:  room.LastMessage.SenderID,
			Message:   room.LastMessage.Message,
			GifURL:    room.LastMessage.GifURL,
			HasMedia:  room.LastMessage.HasMedia,
			CreatedAt: dtoutils.Wrap(room.LastMessage.CreatedAt),
		}
	}
	return resp
}

func (c *ChatRoomController) ToggleMute(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	id, ok := ParseIDParam(ctx, "id", "chat room")
	if !ok {
		return nil
	}
	isMuted, err := c.chatRoomService.ToggleMute(ctx.Context(), id, claims.UserID)
	if err != nil {
		return HandleServiceError(ctx, err)
	}
	return ctx.JSON(fiber.Map{"is_muted": isMuted})
}

func addedMembers(requested, alreadyMembers []uint) []uint {
	already := make(map[uint]struct{}, len(alreadyMembers))
	for _, id := range alreadyMembers {
		already[id] = struct{}{}
	}

	seen := make(map[uint]struct{}, len(requested))
	added := make([]uint, 0, len(requested))
	for _, id := range requested {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		if _, ok := already[id]; ok {
			continue
		}
		added = append(added, id)
	}
	return added
}
