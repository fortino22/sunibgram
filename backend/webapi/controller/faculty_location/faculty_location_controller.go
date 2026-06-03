package facultylocationctrl

import (
	"github.com/gofiber/fiber/v2"

	facultylocationsvciface "github.com/sunibgram/domain/service/faculty_location"
	authdto "github.com/sunibgram/webapi/dto/auth"
	facultylocationdto "github.com/sunibgram/webapi/dto/faculty_location"
)

type FacultyLocationController struct {
	service facultylocationsvciface.IFacultyLocationService
}

func NewFacultyLocationController(service facultylocationsvciface.IFacultyLocationService) *FacultyLocationController {
	return &FacultyLocationController{service: service}
}

// List returns campuses. `?include_all=true` adds the "All Regions" sentinel for
// the event-audience picker; the default (register/profile) omits it.
func (c *FacultyLocationController) List(ctx *fiber.Ctx) error {
	includeAll := ctx.Query("include_all") == "true"
	items, err := c.service.List(includeAll)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}
	return ctx.JSON(facultylocationdto.FacultyLocationListResponse{Data: facultylocationdto.ToResponses(items)})
}
