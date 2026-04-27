package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
)

type AuthController struct {
	authService service.IAuthService
}

func NewAuthController(authService service.IAuthService) *AuthController {
	return &AuthController{authService: authService}
}

func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Message: "invalid request body"})
		return
	}

	user := &model.User{
		Name:     req.Name,
		UserName: req.UserName,
		Email:    req.Email,
		Password: req.Password,
	}

	if err := c.authService.Register(user); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"message": "user registered successfully"})
}

func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, dto.ErrorResponse{Message: "invalid request body"})
		return
	}

	user, pair, err := c.authService.Login(req.Email, req.Password)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, dto.ErrorResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, dto.LoginResponse{
		AccessToken:           pair.AccessToken,
		RefreshToken:          pair.RefreshToken,
		AccessTokenExpiresAt:  pair.AccessTokenExpiresAt,
		RefreshTokenExpiresAt: pair.RefreshTokenExpiresAt,
		User: dto.UserData{
			UserID:   user.UserID,
			Name:     user.Name,
			UserName: user.UserName,
			Email:    user.Email,
		},
	})
}

func (c *AuthController) Refresh(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		writeJSON(w, http.StatusUnauthorized, dto.ErrorResponse{Message: "missing or invalid Authorization header"})
		return
	}

	refreshToken := strings.TrimPrefix(authHeader, "Bearer ")

	pair, err := c.authService.RefreshToken(refreshToken)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, dto.ErrorResponse{Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, dto.RefreshResponse{
		AccessToken:           pair.AccessToken,
		RefreshToken:          pair.RefreshToken,
		AccessTokenExpiresAt:  pair.AccessTokenExpiresAt,
		RefreshTokenExpiresAt: pair.RefreshTokenExpiresAt,
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
