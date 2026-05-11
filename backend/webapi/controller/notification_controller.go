package controller

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/webapi/ctxutil"
)

type NotificationController struct {
	notificationService service.INotificationService
}

func NewNotificationController(notificationService service.INotificationService) *NotificationController {
	return &NotificationController{notificationService: notificationService}
}

func (c *NotificationController) Send(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	var req dto.CreateNotificationRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid payload"})
	}
	if claims.RoleName != "admin" && req.SenderUserID != claims.UserID {
		return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: "forbidden"})
	}

	if err := c.notificationService.SendNotification(service.CreateNotificationInput{
		Type:            req.Type,
		Message:         req.Message,
		SenderUserID:    req.SenderUserID,
		RecipientUserID: req.RecipientUserID,
		ReferenceID:     req.ReferenceID,
	}); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "success"})
}

func (c *NotificationController) GetMine(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))
	offset, _ := strconv.Atoi(ctx.Query("offset", "0"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	recipientUserID := strconv.FormatUint(uint64(claims.UserID), 10)

	items, err := c.notificationService.GetNotificationsByRecipientUserID(recipientUserID, limit, offset)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	data := make([]dto.NotificationListResponse, len(items))
	for i, item := range items {
		data[i] = toNotificationListResponse(item)
	}

	return ctx.JSON(dto.PaginatedNotificationResponse{
		Data: data,
		Meta: dto.NotificationListMeta{Limit: limit, Offset: offset},
	})
}

func (c *NotificationController) Delete(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid notification id"})
	}

	err = c.notificationService.DeleteNotification(uint(id), claims.UserID, claims.RoleName == "admin")
	if err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.JSON(fiber.Map{"message": "notification deleted successfully"})
}

func toNotificationListResponse(item service.NotificationListItem) dto.NotificationListResponse {
	n := item.Notification
	return dto.NotificationListResponse{
		ID:        n.ID,
		Type:      n.Type,
		Message:   n.Message,
		CreatedAt: n.CreatedAt,
		Actor:     toNotificationSenderResponse(item.Actor),
	}
}

func toNotificationSenderResponse(sender *model.NotificationSender) dto.NotificationSenderResponse {
	if sender == nil {
		return dto.NotificationSenderResponse{}
	}
	var faculty *dto.NotificationFacultyResponse
	if sender.Faculty != nil {
		faculty = &dto.NotificationFacultyResponse{
			FacultyID:   sender.Faculty.FacultyID,
			FacultyName: sender.Faculty.FacultyName,
			TotalUsers:  sender.Faculty.TotalUsers,
		}
	}
	return dto.NotificationSenderResponse{
		UserID:    sender.UserID,
		Username:  sender.Username,
		Name:      sender.Name,
		AvatarURL: sender.AvatarURL,
		IsPremium: sender.IsPremium,
		Faculty:   faculty,
	}
}
