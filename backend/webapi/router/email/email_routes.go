package emailroute

import (
	"github.com/gofiber/fiber/v2"

	emailctrl "github.com/sunibgram/webapi/controller/email"
)

func RegisterEmailRoutes(api fiber.Router, ctrl *emailctrl.EmailController, auth fiber.Handler) {
	api.Post("/email/send", auth, ctrl.Send)
}
