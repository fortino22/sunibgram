package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/webapi/ctxutil"
)

type AuthController struct {
	authService service.IAuthService
}

func NewAuthController(authService service.IAuthService) *AuthController {
	return &AuthController{authService: authService}
}

func (c *AuthController) Register(ctx *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}

	user := &model.User{
		Name:     req.Name,
		UserName: req.UserName,
		Email:    req.Email,
		Password: req.Password,
	}

	if err := c.authService.Register(user, req.ConfirmPassword); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "user registered successfully"})
}

func (c *AuthController) ChangePassword(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	var req dto.ChangePasswordRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}

	if req.NewPassword != req.ConfirmPassword {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "new password and confirm password do not match"})
	}

	if err := c.authService.ChangePassword(claims.UserID, req.OldPassword, req.NewPassword); err != nil {
		if err.Error() == "old password is incorrect" {
			return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "password changed successfully"})
}

func (c *AuthController) ForgotPassword(ctx *fiber.Ctx) error {
	var req dto.ForgotPasswordRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}

	if req.NewPassword != req.ConfirmPassword {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "new password and confirm password do not match"})
	}

	if err := c.authService.ForgotPassword(req.Email, req.NewPassword); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "password reset successfully"})
}

func (c *AuthController) Login(ctx *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}

	pair, err := c.authService.Login(req.Email, req.Password)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.LoginResponse{
		Message:     "login successful",
		AccessToken: pair.AccessToken,
	})
}

func (c *AuthController) LoginProd(ctx *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}

	_, pair, err := c.authService.LoginProd(req.Email, req.Password)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.LoginProdResponse{
		AccessToken:           pair.AccessToken,
		RefreshToken:          pair.RefreshToken,
		AccessTokenExpiresAt:  pair.AccessTokenExpiresAt,
		RefreshTokenExpiresAt: pair.RefreshTokenExpiresAt,
	})
}

func (c *AuthController) Refresh(ctx *fiber.Ctx) error {
	header := ctx.Get("Authorization")
	if len(header) < 8 || header[:7] != "Bearer " {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "missing or invalid Authorization header"})
	}

	refreshToken := header[7:]

	pair, err := c.authService.RefreshToken(refreshToken)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.RefreshResponse{
		AccessToken:           pair.AccessToken,
		RefreshToken:          pair.RefreshToken,
		AccessTokenExpiresAt:  pair.AccessTokenExpiresAt,
		RefreshTokenExpiresAt: pair.RefreshTokenExpiresAt,
	})
}
