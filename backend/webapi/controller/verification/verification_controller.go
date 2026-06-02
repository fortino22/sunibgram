package verificationctrl

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	verificationcodemodel "github.com/sunibgram/domain/model/verification_code"
	authsvciface "github.com/sunibgram/domain/service/auth"
	verificationsvciface "github.com/sunibgram/domain/service/verification"
	"github.com/sunibgram/infrastructure/cache"
	"github.com/sunibgram/infrastructure/ratelimit"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	verificationdto "github.com/sunibgram/webapi/dto/verification"
)

type VerificationController struct {
	verificationService verificationsvciface.IVerificationService
	authService         authsvciface.IAuthService
	ipSendLimiter       *ratelimit.RateLimiter
	ipVerifyLimiter     *ratelimit.RateLimiter
}

func NewVerificationController(verificationService verificationsvciface.IVerificationService, authService authsvciface.IAuthService, c *cache.Cache) *VerificationController {
	return &VerificationController{
		verificationService: verificationService,
		authService:         authService,
		ipSendLimiter:       ratelimit.New(c, "verification_send_ip", 30, time.Hour),
		ipVerifyLimiter:     ratelimit.New(c, "verification_verify_ip", 60, time.Hour),
	}
}

func (c *VerificationController) SendLoginOTP(ctx *fiber.Ctx) error {
	if err := c.ipSendLimiter.Allow(ctx.IP()); err != nil {
		return ctx.Status(fiber.StatusTooManyRequests).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
	var req verificationdto.SendVerificationRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}

	req.Email = strings.TrimSpace(req.Email)

	// Verify email + password before issuing an OTP. A wrong password must fail
	// here (so the UI shows "invalid email or password" and stays on the login
	// step) instead of emailing a code that only fails at the final /login step.
	if err := c.authService.VerifyCredentials(req.Email, req.Password); err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	if err := c.verificationService.SendCode(req.Email, verificationcodemodel.VerificationTypeLogin); err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "login otp sent"})
}

func (c *VerificationController) SendCode(ctx *fiber.Ctx) error {
	if err := c.ipSendLimiter.Allow(ctx.IP()); err != nil {
		return ctx.Status(fiber.StatusTooManyRequests).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
	var req verificationdto.SendVerificationRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Type = strings.TrimSpace(strings.ToUpper(req.Type))

	if err := c.verificationService.SendCode(req.Email, verificationcodemodel.VerificationType(req.Type)); err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "verification code sent"})
}

func (c *VerificationController) VerifyCode(ctx *fiber.Ctx) error {
	if err := c.ipVerifyLimiter.Allow(ctx.IP()); err != nil {
		return ctx.Status(fiber.StatusTooManyRequests).JSON(authdto.ErrorResponse{Message: err.Error()})
	}
	var req verificationdto.VerifyCodeRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Code = strings.TrimSpace(req.Code)
	req.Type = strings.TrimSpace(strings.ToUpper(req.Type))

	if err := c.verificationService.VerifyCode(req.Email, req.Code, verificationcodemodel.VerificationType(req.Type)); err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "email verified successfully"})
}
