package schemactrl

import (
	"github.com/gofiber/fiber/v2"

	schemasvciface "github.com/sunibgram/domain/service/schema"
	authdto "github.com/sunibgram/webapi/dto/auth"
	schemadto "github.com/sunibgram/webapi/dto/schema"
)

type SchemaController struct {
	schemaService schemasvciface.ISchemaService
}

func NewSchemaController(schemaService schemasvciface.ISchemaService) *SchemaController {
	return &SchemaController{schemaService: schemaService}
}

func (c *SchemaController) Get(ctx *fiber.Ctx) error {
	tables, err := c.schemaService.GetSchema()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}

	return ctx.Status(fiber.StatusOK).JSON(schemadto.ToSchemaResponse(tables))
}
