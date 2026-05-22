package facultyroute

import (
	"github.com/gofiber/fiber/v2"

	facultyctrl "github.com/sunibgram/webapi/controller/faculty"
	userprofilectrl "github.com/sunibgram/webapi/controller/user_profile"
	"github.com/sunibgram/webapi/middleware"
)

func RegisterFacultyRoutes(api fiber.Router, facultyCtrl *facultyctrl.FacultyController, profileCtrl *userprofilectrl.UserProfileController, auth fiber.Handler) {
	faculties := api.Group("/faculties")

	faculties.Get("", facultyCtrl.GetAll)
	faculties.Get("/me", auth, profileCtrl.GetMyFaculty)
	faculties.Get("/:id", facultyCtrl.GetByID)
	faculties.Post("", auth, middleware.AdminOnly(), facultyCtrl.Create)
	faculties.Put("/:id", auth, middleware.AdminOnly(), facultyCtrl.Update)
	faculties.Delete("/:id", auth, middleware.AdminOnly(), facultyCtrl.Delete)

}
