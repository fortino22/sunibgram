package controller

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/dto/utils"
)

type AuthController struct {
	authService    service.IAuthService
	profileService service.IUserProfileService
	facultyService service.IFacultyService
}

func NewAuthController(authService service.IAuthService, profileService service.IUserProfileService, facultyService service.IFacultyService) *AuthController {
	return &AuthController{authService: authService, profileService: profileService, facultyService: facultyService}
}

func (c *AuthController) Register(ctx *fiber.Ctx) error {
	var facultyID *uint
	if raw := ctx.FormValue("faculty_id"); raw != "" {
		fid, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid faculty_id"})
		}
		v := uint(fid)
		if _, err := c.facultyService.GetByID(v); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "faculty not found"})
		}
		facultyID = &v
	}

	user := &model.User{
		Name:     ctx.FormValue("name"),
		UserName: ctx.FormValue("username"),
		Email:    ctx.FormValue("email"),
		Password: ctx.FormValue("password"),
	}
	confirmPassword := ctx.FormValue("confirm_password")

	if err := c.authService.Register(user, confirmPassword); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	input := service.UpsertProfileInput{}
	if avatar, err := ctx.FormFile("avatar"); err == nil {
		input.Avatar = avatar
	}
	if banner, err := ctx.FormFile("banner"); err == nil {
		input.Banner = banner
	}
	if bio := ctx.FormValue("bio"); bio != "" {
		input.Bio = &bio
	}
	if input.Avatar != nil || input.Banner != nil || input.Bio != nil {
		if err := c.profileService.UpsertProfile(ctx.Context(), user.UserID, input); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
		}
	}

	if facultyID != nil {
		if err := c.facultyService.SetUserFaculty(user.UserID, facultyID); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
		}
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "user registered successfully"})
}

func (c *AuthController) ChangePassword(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	var req dto.ChangePasswordRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}

	if req.NewPassword != req.ConfirmPassword {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "new password and confirm password do not match"})
	}

	if err := c.authService.ChangePassword(claims.UserID, req.OldPassword, req.NewPassword); err != nil {
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
		SID:         pair.SID,
	})
}

func (c *AuthController) LoginWithOTP(ctx *fiber.Ctx) error {
	var req dto.LoginWithOTPRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid request body"})
	}

	pair, err := c.authService.LoginWithOTP(req.Email, req.Password, req.OTP)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.LoginResponse{
		Message:     "login successful",
		AccessToken: pair.AccessToken,
		SID:         pair.SID,
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
		Message:              "login successful",
		AccessToken:          pair.AccessToken,
		SID:                  pair.SID,
		AccessTokenExpiresAt: utils.Wrap(pair.AccessTokenExpiresAt),
	})
}

func (c *AuthController) Refresh(ctx *fiber.Ctx) error {
	var req dto.RefreshRequest
	if err := ctx.BodyParser(&req); err != nil || req.SID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "sid is required"})
	}

	pair, err := c.authService.RefreshToken(ctx.Context(), req.SID)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.RefreshResponse{
		AccessToken: pair.AccessToken,
		SID:         pair.SID,
	})
}

func (c *AuthController) Logout(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	if err := c.authService.Logout(ctx.Context(), claims.SID); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "logged out successfully"})
}
