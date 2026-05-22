package controller

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	verificationcodemodel "github.com/sunibgram/domain/model/verification_code"
	verificationsvciface "github.com/sunibgram/domain/service/verification"
	authdto "github.com/sunibgram/webapi/dto/auth"
	verificationdto "github.com/sunibgram/webapi/dto/verification"
)

type VerificationController struct {
	verificationService verificationsvciface.IVerificationService
}

func NewVerificationController(verificationService verificationsvciface.IVerificationService) *VerificationController {
	return &VerificationController{verificationService: verificationService}
}

func (c *VerificationController) SendLoginOTP(ctx *fiber.Ctx) error {
	var req verificationdto.SendVerificationRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}

	req.Email = strings.TrimSpace(req.Email)

	if err := c.verificationService.SendCode(req.Email, verificationcodemodel.VerificationTypeLogin); err != nil {
		return HandleServiceError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "login otp sent"})
}

func (c *VerificationController) SendCode(ctx *fiber.Ctx) error {
	var req verificationdto.SendVerificationRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Type = strings.TrimSpace(strings.ToUpper(req.Type))

	if err := c.verificationService.SendCode(req.Email, verificationcodemodel.VerificationType(req.Type)); err != nil {
		return HandleServiceError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "verification code sent"})
}

func (c *VerificationController) VerifyCode(ctx *fiber.Ctx) error {
	var req verificationdto.VerifyCodeRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Code = strings.TrimSpace(req.Code)
	req.Type = strings.TrimSpace(strings.ToUpper(req.Type))

	if err := c.verificationService.VerifyCode(req.Email, req.Code, verificationcodemodel.VerificationType(req.Type)); err != nil {
		return HandleServiceError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "email verified successfully"})
}
