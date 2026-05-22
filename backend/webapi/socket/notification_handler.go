package socket

import (
	"encoding/json"
	"errors"
	"strconv"

	fibersocketio "github.com/gofiber/contrib/socketio"

	"github.com/sunibgram/domain/service"
)

type NotificationHandler struct {
	hub                 service.IHub
	notificationService service.INotificationService
}

type notificationEvent struct {
	Event  string          `json:"event"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
	ID     uint            `json:"id"`
	Data   json.RawMessage `json:"data"`
}

type notificationGetPayload struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type notificationDeletePayload struct {
	ID uint `json:"id"`
}

func NewNotificationHandler(hub service.IHub, notificationService service.INotificationService) *NotificationHandler {
	return &NotificationHandler{hub: hub, notificationService: notificationService}
}

func (h *NotificationHandler) HandleMessage(ep *fibersocketio.EventPayload) {
	userID, ok := userIDFromSocket(ep.Kws.GetAttribute(userIDAttribute))
	if !ok {
		emitSocketError(ep.Kws, "unauthorized")
		ep.Kws.Close()
		return
	}

	input, err := parseNotificationEvent(ep.Data)
	if err != nil {
		emitSocketError(ep.Kws, err.Error())
		return
	}

	switch input.Event {
	case "notification:get":
		h.getNotifications(ep.Kws, userID, input)
	case "notification:delete":
		h.deleteNotification(ep.Kws, userID, input)
	default:
		emitSocketError(ep.Kws, "unsupported event")
	}
}

func (h *NotificationHandler) getNotifications(kws *fibersocketio.Websocket, userID uint, input notificationEvent) {
	notifications, err := h.notificationService.GetNotificationsByRecipientUserID(strconv.Itoa(int(userID)), input.Limit, input.Offset)
	if err != nil {
		emitSocketError(kws, err.Error())
		return
	}

	emitSocketEvent(kws, "notification:list", notifications)
}

func (h *NotificationHandler) deleteNotification(kws *fibersocketio.Websocket, userID uint, input notificationEvent) {
	if input.ID == 0 {
		emitSocketError(kws, "id is required")
		return
	}
	if err := h.notificationService.DeleteNotification(input.ID, userID, false); err != nil {
		emitSocketError(kws, err.Error())
		return
	}

	emitSocketEvent(kws, "notification:deleted", map[string]uint{"id": input.ID})
}

func parseNotificationEvent(data []byte) (notificationEvent, error) {
	var input notificationEvent
	if err := json.Unmarshal(data, &input); err != nil {
		return notificationEvent{}, errors.New("invalid socket payload")
	}
	if input.Event == "" {
		return notificationEvent{}, errors.New("event is required")
	}

	if len(input.Data) == 0 {
		return input, nil
	}

	switch input.Event {
	case "notification:get":
		var payload notificationGetPayload
		if err := json.Unmarshal(input.Data, &payload); err == nil {
			if input.Limit == 0 {
				input.Limit = payload.Limit
			}
			if input.Offset == 0 {
				input.Offset = payload.Offset
			}
		}
	case "notification:delete":
		var payload notificationDeletePayload
		if err := json.Unmarshal(input.Data, &payload); err == nil && input.ID == 0 {
			input.ID = payload.ID
		}
	}

	return input, nil
}
