package controller

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
)

type EmailController struct {
	emailService service.IEmailService
}

func NewEmailController(emailService service.IEmailService) *EmailController {
	return &EmailController{emailService: emailService}
}

func (c *EmailController) Send(ctx *fiber.Ctx) error {
	var req dto.SendEmailRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}

	req.To = strings.TrimSpace(req.To)
	req.Subject = strings.TrimSpace(req.Subject)

	if req.To == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "to is required"})
	}
	if req.Subject == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "subject is required"})
	}
	if strings.TrimSpace(req.Body) == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "body is required"})
	}

	if err := c.emailService.QueueMail(ctx.Context(), req.To, req.Subject, req.Body); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: "internal server error"})
	}

	return ctx.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "email queued"})
}
