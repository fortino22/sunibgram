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
	chats.Post("/rooms/:id/members", chatRoomCtrl.AddMember)
	chats.Patch("/rooms/:id", chatRoomCtrl.UpdateGroup)
	chats.Post("/rooms/:room_id/messages", chatCtrl.Send)
	chats.Get("/rooms/:room_id/messages", chatCtrl.GetMessages)
	chats.Patch("/messages/:id", chatCtrl.EditMessage)
	chats.Post("/messages/read", chatCtrl.MarkMessagesRead)
	chats.Delete("/messages/:id", chatCtrl.Unsend)
}
