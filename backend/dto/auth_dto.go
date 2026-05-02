package dto

import "github.com/sunibgram/dto/utils"

type RegisterRequest struct {
	Name            string `json:"name"`
	UserName        string `json:"username"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

type ForgotPasswordRequest struct {
	Email           string `json:"email"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Message     string `json:"message"`
	AccessToken string `json:"access_token"`
	SID         string `json:"sid"`
}

type LoginProdResponse struct {
	Message              string              `json:"message"`
	AccessToken          string              `json:"access_token"`
	SID                  string              `json:"sid"`
	AccessTokenExpiresAt utils.FormattedTime `json:"access_token_expires_at"`
}

type RefreshRequest struct {
	SID string `json:"sid"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	SID         string `json:"sid"`
}

type UserData struct {
	UserID   uint   `json:"user_id"`
	Name     string `json:"name"`
	UserName string `json:"username"`
	Email    string `json:"email"`
	RoleID   uint   `json:"role_id"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}
