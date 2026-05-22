package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sunibgram/webapi/controller"
)

func registerChatRoutes(api fiber.Router, chatRoomCtrl *controller.ChatRoomController, chatCtrl *controller.ChatController, auth fiber.Handler) {
	chats := api.Group("/chats", auth)

	chats.Post("/rooms/private", chatRoomCtrl.CreatePrivate)
	chats.Post("/rooms/group", chatRoomCtrl.CreateGroup)
	chats.Get("/rooms", chatRoomCtrl.GetMyRooms)
	chats.Get("/rooms/:id", chatRoomCtrl.GetByID)
	chats.Get("/rooms/:id/members", chatRoomCtrl.GetMembers)
	chats.Post("/rooms/:id/members", chatRoomCtrl.AddMember)
	chats.Delete("/rooms/:id/leave", chatRoomCtrl.LeaveGroup)
	chats.Patch("/rooms/:id", chatRoomCtrl.UpdateGroup)
	chats.Delete("/rooms/:id", chatRoomCtrl.DeleteGroup)
	chats.Post("/rooms/:id/mute", chatRoomCtrl.ToggleMute)
	chats.Post("/rooms/:room_id/messages", chatCtrl.Send)
	chats.Get("/rooms/:room_id/messages", chatCtrl.GetMessages)
	chats.Patch("/messages/:id", chatCtrl.EditMessage)
	chats.Post("/messages/read", chatCtrl.MarkMessagesRead)
	chats.Delete("/messages/:id", chatCtrl.Unsend)
	chats.Post("/messages/:id/reactions", chatCtrl.AddReaction)
	chats.Delete("/messages/:id/reactions/:emoji", chatCtrl.RemoveReaction)
}
