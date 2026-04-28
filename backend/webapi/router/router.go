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
) {
	api := app.Group("/api")

	auth := api.Group("/auth")
	auth.Post("/register", authCtrl.Register)
	auth.Post("/login", authCtrl.Login)
	auth.Post("/login/prod", authCtrl.LoginProd)
	auth.Post("/refresh", authCtrl.Refresh)

	// any authenticated user
	protected := api.Group("", middleware.Auth())
	protected.Post("/avatar", avatarCtrl.UploadAvatar)
	protected.Put("/avatar", avatarCtrl.UpdateAvatar)
	protected.Get("/avatar/:userId", avatarCtrl.GetAvatarByUserID)
	protected.Delete("/avatar", avatarCtrl.DeleteAvatar)

	// admin only
	admin := api.Group("", middleware.Auth(), middleware.AdminOnly())

	roles := admin.Group("/roles")
	roles.Post("/", roleCtrl.Create)
	roles.Get("/", roleCtrl.GetAll)
	roles.Get("/:id", roleCtrl.GetByID)
	roles.Put("/:id", roleCtrl.Update)
	roles.Delete("/:id", roleCtrl.Delete)

	users := admin.Group("/users")
	users.Get("/", userCtrl.GetAll)
	users.Get("/:id", userCtrl.GetByID)
}
