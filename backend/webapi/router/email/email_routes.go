package emailroute

import (
	"github.com/gofiber/fiber/v2"

	emailctrl "github.com/sunibgram/webapi/controller/email"
	"github.com/sunibgram/webapi/middleware"
)

func RegisterEmailRoutes(api fiber.Router, ctrl *emailctrl.EmailController, auth fiber.Handler) {
	api.Post("/email/send", auth, middleware.AdminOnly(), ctrl.Send)
}
