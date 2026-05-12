package dto

type CreateFacultyRequest struct {
	FacultyName string `json:"faculty_name"`
}

type UpdateFacultyRequest struct {
	FacultyName string `json:"faculty_name"`
}

type AssignUserFacultyRequest struct {
	UserID    uint `json:"user_id"`
	FacultyID uint `json:"faculty_id"`
}

type FacultyResponse struct {
	FacultyID   uint   `json:"faculty_id"`
	FacultyName string `json:"faculty_name"`
	TotalUsers  int    `json:"total_users"`
}

type PaginatedFacultyResponse struct {
	Data []FacultyResponse `json:"data"`
	Meta FeedMeta          `json:"meta"`
}

type MyFacultyResponse struct {
	FacultyID   uint   `json:"faculty_id"`
	FacultyName string `json:"faculty_name"`
	TotalUsers  int    `json:"total_users"`
}
