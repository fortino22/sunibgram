package controller

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
)

type VerificationController struct {
	verificationService service.IVerificationService
}

func NewVerificationController(verificationService service.IVerificationService) *VerificationController {
	return &VerificationController{verificationService: verificationService}
}

func (c *VerificationController) SendCode(ctx *fiber.Ctx) error {
	var req dto.SendVerificationRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Type = strings.TrimSpace(strings.ToUpper(req.Type))

	if err := c.verificationService.SendCode(req.Email, model.VerificationType(req.Type)); err != nil {
		return ctx.Status(verificationErrorStatus(err)).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "verification code sent"})
}

func (c *VerificationController) VerifyCode(ctx *fiber.Ctx) error {
	var req dto.VerifyCodeRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Code = strings.TrimSpace(req.Code)
	req.Type = strings.TrimSpace(strings.ToUpper(req.Type))

	if err := c.verificationService.VerifyCode(req.Email, req.Code, model.VerificationType(req.Type)); err != nil {
		return ctx.Status(verificationErrorStatus(err)).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "email verified successfully"})
}

func verificationErrorStatus(err error) int {
	msg := err.Error()
	switch {
	case strings.HasPrefix(msg, "try again in"):
		return fiber.StatusTooManyRequests
	case msg == "email must use @binus.ac.id domain" ||
		msg == "email is required" ||
		msg == "email already registered" ||
		msg == "email not found" ||
		msg == "type must be REGISTER or FORGOT_PASSWORD" ||
		msg == "invalid code" ||
		msg == "code has expired" ||
		msg == "invalid or expired code":
		return fiber.StatusBadRequest
	default:
		return fiber.StatusInternalServerError
	}
}
