package controller

import (
	"strings"

	"github.com/gofiber/fiber/v2"
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
	if req.Email == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "email is required"})
	}

	if err := c.verificationService.SendCode(req.Email); err != nil {
		return ctx.Status(errorStatus(err)).JSON(dto.ErrorResponse{Message: err.Error()})
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

	if req.Email == "" || req.Code == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "email and code are required"})
	}

	if err := c.verificationService.VerifyCode(req.Email, req.Code); err != nil {
		return ctx.Status(errorStatus(err)).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "email verified successfully"})
}

func errorStatus(err error) int {
	msg := err.Error()
	switch {
	case strings.HasPrefix(msg, "rate limit exceeded"):
		return fiber.StatusTooManyRequests
	case msg == "email must use @binus.ac.id domain" || msg == "email is required" ||
		msg == "invalid code" || msg == "code has expired" || msg == "invalid or expired code" ||
		msg == "email and code are required":
		return fiber.StatusBadRequest
	default:
		return fiber.StatusInternalServerError
	}
}
