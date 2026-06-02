package authctrl

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	usermodel "github.com/sunibgram/domain/model/user"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
	authsvciface "github.com/sunibgram/domain/service/auth"
	facultysvciface "github.com/sunibgram/domain/service/faculty"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
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
	// Birthday is optional. Parse only when provided; a malformed value still errors.
	var birthdayDate time.Time
	if req.BirthdayDate != "" {
		parsed, err := ctrlshared.ParseDateField(req.BirthdayDate, "birthday_date")
		if err != nil {
			return ctrlshared.HandleServiceError(ctx, err)
		}
		birthdayDate = parsed
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
		Name:          req.Name,
		UserName:      req.UserName,
		Email:         req.Email,
		Password:      req.Password,
		TermsAccepted: ctx.FormValue("terms_accepted") == "true",
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

	meta := authsvciface.LoginMeta{
		UserAgent: ctx.Get("User-Agent"),
		IP:        ctx.IP(),
		BaseURL:   ctx.BaseURL(),
	}

	pair, err := c.authService.Login(req.Email, req.Password, meta)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(authdto.LoginResponse{
		Message:     "login successful",
		AccessToken: pair.AccessToken,
		SID:         pair.SID,
	})
}

// ConfirmLogin handles the yes/no links from the post-login alert email. It is a
// public endpoint — the one-time token in the query string is the only credential,
// so it works from any device. It always renders a friendly HTML page (even for an
// expired/unknown token) rather than a JSON error.
func (c *AuthController) ConfirmLogin(ctx *fiber.Ctx) error {
	token := ctx.Query("token")
	action := ctx.Query("action")

	status, err := c.authService.ConfirmLogin(ctx.Context(), token, action)
	if err != nil {
		return ctx.Type("html").SendString(loginConfirmPage(
			"Link expired",
			"This confirmation link is no longer valid. If you're worried about your account, change your password.",
		))
	}

	if status == "denied" {
		return ctx.Type("html").SendString(loginConfirmPage(
			"Signed out",
			"That device has been logged out of your account. We recommend changing your password now.",
		))
	}
	return ctx.Type("html").SendString(loginConfirmPage(
		"Thanks!",
		"Glad it was you. No further action is needed — you can close this tab.",
	))
}

// logoBoltSVG is the Sunibgram bolt mark (from frontend/components/ui/Logo.tsx),
// flattened to a plain inline SVG so it can be embedded in raw HTML/email without
// the React component. fill="currentColor" lets it follow the surrounding text
// color (auto-flips in dark mode).
const logoBoltSVG = `<svg width="26" height="26" viewBox="0 0 384 512" fill="currentColor" aria-hidden="true" style="display:block"><path d="M236 0 L40 300 L190 300 L150 512 L344 224 L190 224 Z"/></svg>`

func loginConfirmPage(title, message string) string {
	return fmt.Sprintf(`<!doctype html><html lang="en"><head><meta charset="utf-8">`+
		`<meta name="viewport" content="width=device-width,initial-scale=1"><title>%[1]s · Sunibgram</title>`+
		`<style>`+
		`:root{color-scheme:light dark}`+
		`*{box-sizing:border-box}`+
		`body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;`+
		`background:#ffffff;color:#0f1419;margin:0;min-height:100vh;display:flex;align-items:center;justify-content:center;padding:24px}`+
		`.card{width:100%%;max-width:400px;background:#ffffff;border:1px solid #eff3f4;border-radius:16px;padding:40px 32px;text-align:center}`+
		`.brand{display:flex;align-items:center;justify-content:center;gap:8px;font-size:20px;font-weight:800;letter-spacing:-0.02em;margin-bottom:28px}`+
		`h1{font-size:23px;font-weight:800;letter-spacing:-0.02em;margin:0 0 10px 0}`+
		`p{font-size:15px;color:#536471;line-height:1.5;margin:0 0 28px 0}`+
		`.btn{display:inline-block;padding:12px 28px;background:#0f1419;color:#ffffff;text-decoration:none;`+
		`border-radius:9999px;font-size:15px;font-weight:700;transition:background .2s ease}`+
		`.btn:hover{background:#272c30}`+
		`@media(prefers-color-scheme:dark){`+
		`body{background:#000000;color:#e7e9ea}`+
		`.card{background:#000000;border-color:#2f3336}`+
		`p{color:#71767b}`+
		`.btn{background:#eff3f4;color:#0f1419}`+
		`.btn:hover{background:#d7dbdc}}`+
		`</style></head>`+
		`<body><div class="card"><div class="brand">%[3]s<span>Sunibgram</span></div>`+
		`<h1>%[1]s</h1><p>%[2]s</p>`+
		`<a class="btn" href="/">Back to Sunibgram</a></div></body></html>`, title, message, logoBoltSVG)
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
