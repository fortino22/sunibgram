package dto

import "github.com/sunibgram/dto/utils"

type CreateViolationRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateViolationRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ViolationResponse struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	CreatedAt   utils.FormattedTime `json:"created_at"`
	UpdatedAt   utils.FormattedTime `json:"updated_at"`
}
