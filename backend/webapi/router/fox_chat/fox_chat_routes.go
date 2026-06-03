package foxchatroute

import (
	"github.com/gofiber/fiber/v2"

	foxchatctrl "github.com/sunibgram/webapi/controller/fox_chat"
)

func RegisterFoxChatRoutes(api fiber.Router, foxChatCtrl *foxchatctrl.FoxChatController, auth fiber.Handler) {
	api.Post("/fox/ask", auth, foxChatCtrl.Ask)
}
