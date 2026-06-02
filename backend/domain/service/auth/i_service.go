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
	User            *usermodel.User
	Profile         *userprofilemodel.UserProfile
	ConfirmPassword string
	Avatar          *multipart.FileHeader
	Banner          *multipart.FileHeader
	Bio             string
	FacultyID       *uint
}

type IAuthService interface {
	Register(ctx context.Context, input RegisterInput) error
	IsUsernameAvailable(username string) (bool, error)
	Login(email, password string) (usermodel.TokenPair, error)
	VerifyCredentials(email, password string) error
	LoginWithOTP(email, password, otp string) (usermodel.TokenPair, error)
	LoginProd(email, password string) (*usermodel.User, usermodel.TokenPair, error)
	RefreshToken(ctx context.Context, sid string) (usermodel.TokenPair, error)
	Logout(ctx context.Context, sid string) error
	ChangePassword(userID uint, email, newPassword string) error
	ForgotPassword(email, newPassword string) error
}
