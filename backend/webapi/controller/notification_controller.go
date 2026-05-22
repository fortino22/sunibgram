package controller

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	dtoutils "github.com/sunibgram/webapi/dto/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	notificationdto "github.com/sunibgram/webapi/dto/notification"
)

type NotificationController struct {
	notificationService service.INotificationService
}

func NewNotificationController(notificationService service.INotificationService) *NotificationController {
	return &NotificationController{notificationService: notificationService}
}

func (c *NotificationController) Send(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	var req notificationdto.CreateNotificationRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid payload"})
	}
	if claims.RoleName != "admin" && req.SenderUserID != claims.UserID {
		return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "forbidden"})
	}

	if err := c.notificationService.SendNotification(service.CreateNotificationInput{
		Type:            req.Type,
		Message:         req.Message,
		SenderUserID:    req.SenderUserID,
		RecipientUserID: req.RecipientUserID,
		ReferenceID:     req.ReferenceID,
	}); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "success"})
}

func (c *NotificationController) GetMine(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
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
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	data := make([]notificationdto.NotificationListResponse, len(items))
	for i, item := range items {
		data[i] = toNotificationListResponse(item)
	}

	return ctx.JSON(notificationdto.PaginatedNotificationResponse{
		Data: data,
		Meta: notificationdto.NotificationListMeta{Limit: limit, Offset: offset},
	})
}

func (c *NotificationController) Delete(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	id, ok := ParseIDParam(ctx, "id", "notification")
	if !ok {
		return nil
	}

	err := c.notificationService.DeleteNotification(id, claims.UserID, claims.RoleName == "admin")
	if err != nil {
		if err.Error() == "forbidden" {
			return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.JSON(fiber.Map{"message": "notification deleted successfully"})
}

func toNotificationListResponse(item service.NotificationListItem) notificationdto.NotificationListResponse {
	n := item.Notification
	return notificationdto.NotificationListResponse{
		ID:          n.ID,
		Type:        n.Type,
		Message:     n.Message,
		ReferenceID: n.ReferenceID,
		CreatedAt:   dtoutils.Wrap(n.CreatedAt),
		Actor:       toNotificationSenderResponse(item.Actor),
	}
}

func toNotificationSenderResponse(sender *model.NotificationSender) notificationdto.NotificationSenderResponse {
	if sender == nil {
		return notificationdto.NotificationSenderResponse{}
	}
	var faculty *notificationdto.NotificationFacultyResponse
	if sender.Faculty != nil {
		faculty = &notificationdto.NotificationFacultyResponse{
			FacultyID:   sender.Faculty.FacultyID,
			FacultyName: sender.Faculty.FacultyName,
			TotalUsers:  sender.Faculty.TotalUsers,
		}
	}
	return notificationdto.NotificationSenderResponse{
		UserID:    sender.UserID,
		Username:  sender.Username,
		Name:      sender.Name,
		AvatarURL: sender.AvatarURL,
		IsPremium: sender.IsPremium,
		Faculty:   faculty,
	}
}
