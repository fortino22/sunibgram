package foxsessionroute

import (
	"github.com/gofiber/fiber/v2"

	foxsessionctrl "github.com/sunibgram/webapi/controller/fox_session"
)

func RegisterFoxSessionRoutes(api fiber.Router, foxSessionCtrl *foxsessionctrl.FoxSessionController, auth fiber.Handler) {
	api.Get("/fox/sessions", auth, foxSessionCtrl.List)
	api.Get("/fox/sessions/:id", auth, foxSessionCtrl.Detail)
}
