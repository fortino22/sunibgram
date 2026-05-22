package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sunibgram/webapi/controller"
	"github.com/sunibgram/webapi/middleware"
)

func registerCategoryRoutes(api fiber.Router, categoryCtrl *controller.CategoryController) {
	api.Get("/categories", middleware.Auth(), categoryCtrl.GetAll)

	categories := api.Group("/categories", middleware.Auth(), middleware.AdminOnly())
	categories.Post("", categoryCtrl.Create)
	categories.Put("/:id", categoryCtrl.Update)
	categories.Delete("/:id", categoryCtrl.Delete)
}
