package dto

import "github.com/sunibgram/dto/utils"

type CreateCategoryRequest struct {
	Name string `json:"name"`
}

type UpdateCategoryRequest struct {
	Name string `json:"name"`
}

type CategoryResponse struct {
	ID        string              `json:"id"`
	Name      string              `json:"name"`
	CreatedAt utils.FormattedTime `json:"created_at"`
}
