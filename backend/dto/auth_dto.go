package dto

import "time"

type RegisterRequest struct {
	Name     string `json:"name"`
	UserName string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Message     string `json:"message"`
	AccessToken string `json:"access_token"`
}

type LoginProdResponse struct {
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
}

type RefreshResponse struct {
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
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
