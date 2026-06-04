package userdto

import (
	utils "github.com/sunibgram/webapi/dto/_shared"
	userprofiledto "github.com/sunibgram/webapi/dto/user_profile"
)

// UserFacultyMini is the faculty summary embedded in an admin user row.
type UserFacultyMini struct {
	FacultyID   uint   `json:"faculty_id"`
	FacultyName string `json:"faculty_name"`
}

type UserResponse struct {
	UserID    uint                `json:"user_id"`
	Name      string              `json:"name"`
	UserName  string              `json:"username"`
	Email     string              `json:"email"`
	RoleID    uint                `json:"role_id"`
	RoleName  string              `json:"role_name"`
	AvatarURL string              `json:"avatar_url,omitempty"`
	IsPremium bool                `json:"is_premium"`
	Faculty   *UserFacultyMini    `json:"faculty,omitempty"`
	CreatedAt utils.FormattedTime `json:"created_at"`
}

// AdminUserDetailResponse is the admin-only single-user view. It embeds the full
// profile response and adds the role fields the public profile DTO omits — this
// is why the admin user-detail modal was rendering an empty role.
type AdminUserDetailResponse struct {
	userprofiledto.UserProfileResponse
	RoleID   uint   `json:"role_id"`
	RoleName string `json:"role_name"`
}

type PaginatedUsersResponse struct {
	Data       []UserResponse `json:"data"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	TotalPages int            `json:"total_pages"`
}
