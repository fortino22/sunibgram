package shareroute

import (
	"github.com/gofiber/fiber/v2"

	sharectrl "github.com/sunibgram/webapi/controller/share"
)

func RegisterShareRoutes(api fiber.Router, shareCtrl *sharectrl.ShareController, auth fiber.Handler) {
	api.Post("/threads/:id/share/copy", auth, shareCtrl.RecordCopy)
	api.Post("/threads/:id/share/chat", auth, shareCtrl.ShareViaChat)
	api.Get("/users/search-share", auth, shareCtrl.SearchUsers)
}
