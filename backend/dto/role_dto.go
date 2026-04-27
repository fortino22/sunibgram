package dto

type CreateRoleRequest struct {
	Name string `json:"name"`
}

type UpdateRoleRequest struct {
	Name string `json:"name"`
}

type RoleResponse struct {
	RoleID uint   `json:"role_id"`
	Name   string `json:"name"`
}
