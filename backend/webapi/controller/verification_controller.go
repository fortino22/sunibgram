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

func (c *VerificationController) SendLoginOTP(ctx *fiber.Ctx) error {
	var req dto.SendVerificationRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}

	req.Email = strings.TrimSpace(req.Email)

	if err := c.verificationService.SendCode(req.Email, model.VerificationTypeLogin); err != nil {
		return HandleServiceError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "login otp sent"})
}

func (c *VerificationController) SendCode(ctx *fiber.Ctx) error {
	var req dto.SendVerificationRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Type = strings.TrimSpace(strings.ToUpper(req.Type))

	if err := c.verificationService.SendCode(req.Email, model.VerificationType(req.Type)); err != nil {
		return HandleServiceError(ctx, err)
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
		return HandleServiceError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "email verified successfully"})
}
