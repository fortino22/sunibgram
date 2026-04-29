package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/webapi/controller"
	"github.com/sunibgram/webapi/middleware"
)

func Register(
	app *fiber.App,
	authCtrl *controller.AuthController,
	roleCtrl *controller.RoleController,
	userCtrl *controller.UserController,
	avatarCtrl *controller.AvatarController,
	threadCtrl *controller.ThreadController,
	threadMediaCtrl *controller.ThreadMediaController,
	likeCtrl *controller.LikeController,
	bookmarkCtrl *controller.BookmarkController,
	repostCtrl *controller.RepostController,
	emailCtrl *controller.EmailController,
	verificationCtrl *controller.VerificationController,
) {
	api := app.Group("/api")
	auth := middleware.Auth()

	registerAuthRoutes(api, authCtrl)
	registerVerificationRoutes(api, verificationCtrl)
	registerAvatarRoutes(api, avatarCtrl, auth)
	registerThreadRoutes(api, threadCtrl, threadMediaCtrl, likeCtrl, bookmarkCtrl, repostCtrl, auth)
	registerEmailRoutes(api, emailCtrl, auth)
	registerAdminRoutes(api, roleCtrl, userCtrl)
}
