package adminroute

import (
	"github.com/gofiber/fiber/v2"

	cachectrl "github.com/sunibgram/webapi/controller/cache"
	rolectrl "github.com/sunibgram/webapi/controller/role"
	schemactrl "github.com/sunibgram/webapi/controller/schema"
	userctrl "github.com/sunibgram/webapi/controller/user"
	"github.com/sunibgram/webapi/middleware"
)

func RegisterAdminRoutes(api fiber.Router, roleCtrl *rolectrl.RoleController, userCtrl *userctrl.UserController, cacheCtrl *cachectrl.CacheController, schemaCtrl *schemactrl.SchemaController) {
	admin := api.Group("/admin", middleware.Auth(), middleware.AdminOnly())

	roles := admin.Group("/roles")
	roles.Post("/", roleCtrl.Create)
	roles.Get("/", roleCtrl.GetAll)
	roles.Get("/:id", roleCtrl.GetByID)
	roles.Put("/:id", roleCtrl.Update)
	roles.Delete("/:id", roleCtrl.Delete)

	users := admin.Group("/users")
	users.Get("/", userCtrl.GetAll)
	users.Get("/:id", userCtrl.GetByID)

	cache := admin.Group("/cache")
	cache.Post("/clear", cacheCtrl.Clear)

	schema := admin.Group("/schema")
	schema.Get("/", schemaCtrl.Get)
}
