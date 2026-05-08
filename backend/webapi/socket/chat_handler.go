package socket

import (
	"encoding/json"
	"errors"

	fibersocketio "github.com/gofiber/contrib/socketio"
	"github.com/sunibgram/domain/service"
)

const userIDAttribute = "user_id"

type ChatHandler struct {
	hub         service.IHub
	roomService service.IChatRoomService
}

type clientEvent struct {
	Event  string          `json:"event"`
	RoomID uint            `json:"room_id"`
	Data   json.RawMessage `json:"data"`
}

type roomPayload struct {
	RoomID uint `json:"room_id"`
}

func NewChatHandler(hub service.IHub, roomService service.IChatRoomService) *ChatHandler {
	return &ChatHandler{hub: hub, roomService: roomService}
}

func (h *ChatHandler) HandleMessage(ep *fibersocketio.EventPayload) {
	userID, ok := userIDFromSocket(ep.Kws.GetAttribute(userIDAttribute))
	if !ok {
		emitSocketError(ep.Kws, "unauthorized")
		ep.Kws.Close()
		return
	}

	event, roomID, err := parseClientEvent(ep.Data)
	if err != nil {
		emitSocketError(ep.Kws, err.Error())
		return
	}

	switch event {
	case "join_room":
		h.joinRoom(ep, userID, roomID)
	case "leave_room":
		h.leaveRoom(ep, roomID)
	case "typing:start", "typing:stop":
		h.typing(ep, event, userID, roomID)
	default:
		emitSocketError(ep.Kws, "unsupported event")
	}
}

func (h *ChatHandler) HandleDisconnect(ep *fibersocketio.EventPayload) {
	h.hub.Unregister(ep.SocketUUID)
}

func (h *ChatHandler) joinRoom(ep *fibersocketio.EventPayload, userID, roomID uint) {
	if roomID == 0 {
		emitSocketError(ep.Kws, "room_id is required")
		return
	}
	if _, err := h.roomService.GetByID(roomID, userID); err != nil {
		emitSocketError(ep.Kws, err.Error())
		return
	}

	h.hub.Join(roomID, ep.SocketUUID)
	emitSocketEvent(ep.Kws, "room:joined", roomPayload{RoomID: roomID})
}

func (h *ChatHandler) leaveRoom(ep *fibersocketio.EventPayload, roomID uint) {
	if roomID == 0 {
		emitSocketError(ep.Kws, "room_id is required")
		return
	}

	h.hub.Leave(roomID, ep.SocketUUID)
	emitSocketEvent(ep.Kws, "room:left", roomPayload{RoomID: roomID})
}

func (h *ChatHandler) typing(ep *fibersocketio.EventPayload, event string, userID, roomID uint) {
	if roomID == 0 {
		emitSocketError(ep.Kws, "room_id is required")
		return
	}
	if !h.hub.IsInRoom(roomID, ep.SocketUUID) {
		emitSocketError(ep.Kws, "join_room required")
		return
	}

	h.hub.Broadcast(roomID, event, map[string]any{
		"room_id": roomID,
		"user_id": userID,
	})
}

func parseClientEvent(data []byte) (string, uint, error) {
	var input clientEvent
	if err := json.Unmarshal(data, &input); err != nil {
		return "", 0, errors.New("invalid socket payload")
	}

	roomID := input.RoomID
	if roomID == 0 && len(input.Data) > 0 {
		var payload roomPayload
		if err := json.Unmarshal(input.Data, &payload); err == nil {
			roomID = payload.RoomID
		}
	}
	if input.Event == "" {
		return "", 0, errors.New("event is required")
	}
	return input.Event, roomID, nil
}

func userIDFromSocket(value any) (uint, bool) {
	switch v := value.(type) {
	case uint:
		return v, v != 0
	case int:
		if v <= 0 {
			return 0, false
		}
		return uint(v), true
	case float64:
		if v <= 0 {
			return 0, false
		}
		return uint(v), true
	default:
		return 0, false
	}
}

func emitSocketError(kws *fibersocketio.Websocket, message string) {
	emitSocketEvent(kws, "socket:error", map[string]string{"message": message})
}

func emitSocketEvent(kws *fibersocketio.Websocket, event string, payload any) {
	message, err := json.Marshal(map[string]any{
		"event": event,
		"data":  payload,
	})
	if err != nil {
		return
	}
	kws.Emit(message, fibersocketio.TextMessage)
}
