package service

type IHub interface {
	Register(connID string, userID uint)
	Unregister(connID string)
	Join(roomID uint, connID string)
	Leave(roomID uint, connID string)
	IsInRoom(roomID uint, connID string) bool
	Broadcast(roomID uint, event string, payload any)
	EmitToUser(userID uint, event string, payload any)
}
