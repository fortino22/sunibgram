package verificationroute

import (
	"github.com/gofiber/fiber/v2"

	verificationctrl "github.com/sunibgram/webapi/controller/verification"
)

func RegisterVerificationRoutes(api fiber.Router, ctrl *verificationctrl.VerificationController) {
	v := api.Group("/verification")
	v.Post("/send", ctrl.SendCode)
	v.Post("/verify", ctrl.VerifyCode)
}
