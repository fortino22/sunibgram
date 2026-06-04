package facultylocationroute

import (
	"github.com/gofiber/fiber/v2"

	facultylocationctrl "github.com/sunibgram/webapi/controller/faculty_location"
)

// RegisterFacultyLocationRoutes wires the public campus reference endpoint.
// Public (token optional), like /faculties — the signup form reads it pre-auth.
func RegisterFacultyLocationRoutes(api fiber.Router, ctrl *facultylocationctrl.FacultyLocationController) {
	locations := api.Group("/faculty-locations")
	locations.Get("", ctrl.List)
}
