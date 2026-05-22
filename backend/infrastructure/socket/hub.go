package socket

import (
	"encoding/json"
	"sync"

	fibersocketio "github.com/gofiber/contrib/socketio"

	sockethubsvciface "github.com/sunibgram/domain/service/socket_hub"
)

type Hub struct {
	mu          sync.RWMutex
	rooms       map[uint]map[string]struct{}
	connections map[string]map[uint]struct{}
	users       map[uint]map[string]struct{}
	connUsers   map[string]uint
}

type envelope struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

func NewHub() sockethubsvciface.IHub {
	return &Hub{
		rooms:       make(map[uint]map[string]struct{}),
		connections: make(map[string]map[uint]struct{}),
		users:       make(map[uint]map[string]struct{}),
		connUsers:   make(map[string]uint),
	}
}

func (h *Hub) Register(connID string, userID uint) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.connUsers[connID] = userID
	if h.users[userID] == nil {
		h.users[userID] = make(map[string]struct{})
	}
	h.users[userID][connID] = struct{}{}
}

func (h *Hub) Unregister(connID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if roomIDs, ok := h.connections[connID]; ok {
		for roomID := range roomIDs {
			delete(h.rooms[roomID], connID)
			if len(h.rooms[roomID]) == 0 {
				delete(h.rooms, roomID)
			}
		}
		delete(h.connections, connID)
	}

	if userID, ok := h.connUsers[connID]; ok {
		delete(h.users[userID], connID)
		if len(h.users[userID]) == 0 {
			delete(h.users, userID)
		}
		delete(h.connUsers, connID)
	}
}

func (h *Hub) Join(roomID uint, connID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[roomID] == nil {
		h.rooms[roomID] = make(map[string]struct{})
	}
	h.rooms[roomID][connID] = struct{}{}

	if h.connections[connID] == nil {
		h.connections[connID] = make(map[uint]struct{})
	}
	h.connections[connID][roomID] = struct{}{}
}

func (h *Hub) Leave(roomID uint, connID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if connIDs, ok := h.rooms[roomID]; ok {
		delete(connIDs, connID)
		if len(connIDs) == 0 {
			delete(h.rooms, roomID)
		}
	}
	if roomIDs, ok := h.connections[connID]; ok {
		delete(roomIDs, roomID)
		if len(roomIDs) == 0 {
			delete(h.connections, connID)
		}
	}
}

func (h *Hub) IsInRoom(roomID uint, connID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	_, ok := h.rooms[roomID][connID]
	return ok
}

func (h *Hub) Broadcast(roomID uint, event string, payload any) {
	connIDs := h.roomConnIDs(roomID)
	if len(connIDs) == 0 {
		return
	}

	message, err := encode(event, payload)
	if err != nil {
		return
	}
	fibersocketio.EmitToList(connIDs, message, fibersocketio.TextMessage)
}

func (h *Hub) EmitToUser(userID uint, event string, payload any) {
	connIDs := h.userConnIDs(userID)
	if len(connIDs) == 0 {
		return
	}

	message, err := encode(event, payload)
	if err != nil {
		return
	}
	fibersocketio.EmitToList(connIDs, message, fibersocketio.TextMessage)
}

func (h *Hub) roomConnIDs(roomID uint) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	connIDs := make([]string, 0, len(h.rooms[roomID]))
	for connID := range h.rooms[roomID] {
		connIDs = append(connIDs, connID)
	}
	return connIDs
}

func (h *Hub) userConnIDs(userID uint) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	connIDs := make([]string, 0, len(h.users[userID]))
	for connID := range h.users[userID] {
		connIDs = append(connIDs, connID)
	}
	return connIDs
}

func encode(event string, payload any) ([]byte, error) {
	return json.Marshal(envelope{Event: event, Data: payload})
}
