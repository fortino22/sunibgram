package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/sunibgram/webapi/controller"
	"github.com/sunibgram/webapi/middleware"
)

func registerFacultyRoutes(api fiber.Router, facultyCtrl *controller.FacultyController, profileCtrl *controller.UserProfileController, auth fiber.Handler) {
	faculties := api.Group("/faculties")

	faculties.Get("", facultyCtrl.GetAll)
	faculties.Get("/me", auth, profileCtrl.GetMyFaculty)
	faculties.Get("/:id", facultyCtrl.GetByID)
	faculties.Post("", auth, middleware.AdminOnly(), facultyCtrl.Create)
	faculties.Put("/:id", auth, middleware.AdminOnly(), facultyCtrl.Update)
	faculties.Delete("/:id", auth, middleware.AdminOnly(), facultyCtrl.Delete)

}
