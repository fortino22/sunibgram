package dto

import "github.com/sunibgram/dto/utils"

type UserResponse struct {
	UserID    uint                `json:"user_id"`
	Name      string              `json:"name"`
	UserName  string              `json:"username"`
	Email     string              `json:"email"`
	RoleID    uint                `json:"role_id"`
	RoleName  string              `json:"role_name"`
	CreatedAt utils.FormattedTime `json:"created_at"`
}

type PaginatedUsersResponse struct {
	Data       []UserResponse `json:"data"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	TotalPages int            `json:"total_pages"`
}
