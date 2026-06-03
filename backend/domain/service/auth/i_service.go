package authsvciface

import (
	"context"
	"errors"
	"mime/multipart"

	usermodel "github.com/sunibgram/domain/model/user"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
)

var (
	ErrSessionRevoked     = errors.New("session expired or revoked")
	ErrSessionTransient   = errors.New("session lookup temporarily unavailable")
	ErrRefreshRateLimited = errors.New("refresh rate limit exceeded")
)

// RegisterInput bundles everything a single atomic registration needs. Media is
// uploaded and all rows (user, profile, settings, faculty) are persisted as one
// unit so a partial failure never leaves an orphaned account behind.
type RegisterInput struct {
	User              *usermodel.User
	Profile           *userprofilemodel.UserProfile
	ConfirmPassword   string
	Avatar            *multipart.FileHeader
	Banner            *multipart.FileHeader
	Bio               string
	FacultyID         *uint
	FacultyLocationID *uint
}

// LoginMeta carries request context for the post-login security-alert email:
// the device + IP shown to the user, and the backend base URL the confirm/deny
// links point at (the endpoint then redirects to the frontend).
type LoginMeta struct {
	UserAgent string
	IP        string
	BaseURL   string
}

type IAuthService interface {
	Register(ctx context.Context, input RegisterInput) error
	IsUsernameAvailable(username string) (bool, error)
	Login(email, password string, meta LoginMeta) (usermodel.TokenPair, error)
	ConfirmLogin(ctx context.Context, token, action string) (string, error)
	RefreshToken(ctx context.Context, sid string) (usermodel.TokenPair, error)
	Logout(ctx context.Context, sid string) error
	ChangePassword(userID uint, email, newPassword string) error
	ForgotPassword(email, newPassword string) error
}
