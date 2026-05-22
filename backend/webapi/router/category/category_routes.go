package categoryroute

import (
	"github.com/gofiber/fiber/v2"

	categoryctrl "github.com/sunibgram/webapi/controller/category"
	"github.com/sunibgram/webapi/middleware"
)

func RegisterCategoryRoutes(api fiber.Router, categoryCtrl *categoryctrl.CategoryController) {
	api.Get("/categories", middleware.Auth(), categoryCtrl.GetAll)

	categories := api.Group("/categories", middleware.Auth(), middleware.AdminOnly())
	categories.Post("", categoryCtrl.Create)
	categories.Put("/:id", categoryCtrl.Update)
	categories.Delete("/:id", categoryCtrl.Delete)
}
