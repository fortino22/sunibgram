package authctrl

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	usermodel "github.com/sunibgram/domain/model/user"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
	authsvciface "github.com/sunibgram/domain/service/auth"
	facultysvciface "github.com/sunibgram/domain/service/faculty"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	utils "github.com/sunibgram/webapi/dto/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
)

type AuthController struct {
	authService    authsvciface.IAuthService
	facultyService facultysvciface.IFacultyService
}

func NewAuthController(authService authsvciface.IAuthService, facultyService facultysvciface.IFacultyService) *AuthController {
	return &AuthController{authService: authService, facultyService: facultyService}
}

// CheckUsername is a public availability probe for the signup form (pre-auth).
func (c *AuthController) CheckUsername(ctx *fiber.Ctx) error {
	available, err := c.authService.IsUsernameAvailable(ctx.Query("username"))
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"available": available})
}

func (c *AuthController) Register(ctx *fiber.Ctx) error {
	req := authdto.RegisterRequest{
		Name:            ctx.FormValue("name"),
		UserName:        ctx.FormValue("username"),
		Email:           ctx.FormValue("email"),
		Password:        ctx.FormValue("password"),
		ConfirmPassword: ctx.FormValue("confirm_password"),
		Gender:          ctx.FormValue("gender"),
		BirthdayDate:    ctx.FormValue("birthday_date"),
	}
	birthdayDate, err := ctrlshared.ParseDateField(req.BirthdayDate, "birthday_date")
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}

	raw := ctx.FormValue("faculty_id")
	if raw == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "faculty_id is required"})
	}
	fid, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid faculty_id"})
	}
	v := uint(fid)
	if _, err := c.facultyService.GetByID(v); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "faculty not found"})
	}
	facultyID := &v

	user := &usermodel.User{
		Name:     req.Name,
		UserName: req.UserName,
		Email:    req.Email,
		Password: req.Password,
	}
	profile := &userprofilemodel.UserProfile{
		Gender:       userprofilemodel.ParseGender(req.Gender),
		BirthdayDate: birthdayDate,
	}

	input := authsvciface.RegisterInput{
		User:            user,
		Profile:         profile,
		ConfirmPassword: req.ConfirmPassword,
		Bio:             ctx.FormValue("bio"),
		FacultyID:       facultyID,
	}
	if avatar, err := ctx.FormFile("avatar"); err == nil {
		input.Avatar = avatar
	}
	if banner, err := ctx.FormFile("banner"); err == nil {
		input.Banner = banner
	}

	if err := c.authService.Register(ctx.Context(), input); err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "user registered successfully"})
}

func (c *AuthController) ChangePassword(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	var req authdto.ChangePasswordRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}

	if req.NewPassword != req.ConfirmPassword {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "new password and confirm password do not match"})
	}

	if err := c.authService.ChangePassword(claims.UserID, claims.Email, req.NewPassword); err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "password changed successfully"})
}

func (c *AuthController) ForgotPassword(ctx *fiber.Ctx) error {
	var req authdto.ForgotPasswordRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}

	if req.NewPassword != req.ConfirmPassword {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "new password and confirm password do not match"})
	}

	if err := c.authService.ForgotPassword(req.Email, req.NewPassword); err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "password reset successfully"})
}

func (c *AuthController) Login(ctx *fiber.Ctx) error {
	var req authdto.LoginRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}

	pair, err := c.authService.Login(req.Email, req.Password)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(authdto.LoginResponse{
		Message:     "login successful",
		AccessToken: pair.AccessToken,
		SID:         pair.SID,
	})
}

func (c *AuthController) LoginWithOTP(ctx *fiber.Ctx) error {
	var req authdto.LoginWithOTPRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}

	pair, err := c.authService.LoginWithOTP(req.Email, req.Password, req.OTP)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(authdto.LoginResponse{
		Message:     "login successful",
		AccessToken: pair.AccessToken,
		SID:         pair.SID,
	})
}

func (c *AuthController) LoginProd(ctx *fiber.Ctx) error {
	var req authdto.LoginRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid request body"})
	}

	_, pair, err := c.authService.LoginProd(req.Email, req.Password)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(authdto.LoginProdResponse{
		Message:              "login successful",
		AccessToken:          pair.AccessToken,
		SID:                  pair.SID,
		AccessTokenExpiresAt: utils.Wrap(pair.AccessTokenExpiresAt),
	})
}

func (c *AuthController) Refresh(ctx *fiber.Ctx) error {
	var req authdto.RefreshRequest
	if err := ctx.BodyParser(&req); err != nil || req.SID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "sid is required"})
	}

	pair, err := c.authService.RefreshToken(ctx.Context(), req.SID)
	if err != nil {
		switch {
		case errors.Is(err, authsvciface.ErrSessionRevoked):
			return ctx.Status(fiber.StatusUnauthorized).JSON(authdto.ErrorResponse{Message: err.Error()})
		case errors.Is(err, authsvciface.ErrSessionTransient):
			return ctx.Status(fiber.StatusServiceUnavailable).JSON(authdto.ErrorResponse{Message: err.Error()})
		case errors.Is(err, authsvciface.ErrRefreshRateLimited):
			ctx.Set("Retry-After", "1")
			return ctx.Status(fiber.StatusTooManyRequests).JSON(authdto.ErrorResponse{Message: err.Error()})
		default:
			return ctx.Status(fiber.StatusUnauthorized).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(authdto.RefreshResponse{
		AccessToken: pair.AccessToken,
		SID:         pair.SID,
	})
}

func (c *AuthController) Logout(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	if err := c.authService.Logout(ctx.Context(), claims.SID); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "logged out successfully"})
}
