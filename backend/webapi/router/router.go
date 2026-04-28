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
) {
	api := app.Group("/api")

	auth := api.Group("/auth")
	auth.Post("/register", authCtrl.Register)
	auth.Post("/login", authCtrl.Login)
	auth.Post("/login/prod", authCtrl.LoginProd)
	auth.Post("/refresh", authCtrl.Refresh)

	protected := api.Group("", middleware.Auth())

	protected.Post("/avatar", avatarCtrl.UploadAvatar)
	protected.Put("/avatar", avatarCtrl.UpdateAvatar)
	protected.Get("/avatar", avatarCtrl.GetAvatar)
	protected.Delete("/avatar", avatarCtrl.DeleteAvatar)

	threads := protected.Group("/threads")
	threads.Post("/", threadCtrl.Create)
	threads.Get("/", threadCtrl.GetRoots)
	threads.Get("/:id", threadCtrl.GetByID)
	threads.Get("/:id/chain", threadCtrl.GetChain)
	threads.Put("/:id", threadCtrl.Update)
	threads.Delete("/:id", threadCtrl.Delete)
	threads.Post("/:id/media", threadMediaCtrl.Upload)
	threads.Put("/:id/media", threadMediaCtrl.Update)
	threads.Delete("/:id/media", threadMediaCtrl.Delete)

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
