package categorydto

import utils "github.com/sunibgram/webapi/dto/_shared"

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
