package reportdto

import (
	utils "github.com/sunibgram/webapi/dto/_shared"
	threaddto "github.com/sunibgram/webapi/dto/thread"
	userprofiledto "github.com/sunibgram/webapi/dto/user_profile"
	violationdto "github.com/sunibgram/webapi/dto/violation"
)

type CreateReportRequest struct {
	ReportedUserID *uint  `json:"reported_user_id"`
	ReportedPostID *uint  `json:"reported_post_id"`
	ViolationID    string `json:"violation_id"`
	Explanation    string `json:"explanation"`
}

type UpdateReportStatusRequest struct {
	Status string `json:"status"`
}

type ReportResponse struct {
	ID             string                          `json:"id"`
	ReporterID     uint                            `json:"reporter_id"`
	ReportedUserID *uint                           `json:"reported_user_id"`
	ReportedPostID *uint                           `json:"reported_post_id"`
	ViolationID    string                          `json:"violation_id"`
	Violation      *violationdto.ViolationResponse `json:"violation,omitempty"`
	ReporterUser   *ReportUserResponse             `json:"reporter_user,omitempty"`
	ReportedUser   *ReportUserResponse             `json:"reported_user,omitempty"`
	Explanation    string                          `json:"explanation"`
	Status         string                          `json:"status"`
	CreatedAt      utils.FormattedTime             `json:"created_at"`
	UpdatedAt      utils.FormattedTime             `json:"updated_at"`
}

type ReportUserResponse struct {
	UserID     uint                   `json:"user_id"`
	Username   string                 `json:"username"`
	UserAvatar string                 `json:"user_avatar"`
	IsPremium  bool                   `json:"is_premium"`
	Faculty    *ReportFacultyResponse `json:"faculty,omitempty"`
}

type ReportFacultyResponse struct {
	FacultyID   uint   `json:"faculty_id"`
	FacultyName string `json:"faculty_name"`
}

// ReportedThreadGroupResponse: one reported thread (full data) + its reports.
type ReportedThreadGroupResponse struct {
	Thread      threaddto.FeedItemResponse `json:"thread"`
	ReportCount int                        `json:"report_count"`
	Reports     []ReportResponse           `json:"reports"`
}

// ReportedUserGroupResponse: one reported user (full profile) + its reports.
type ReportedUserGroupResponse struct {
	User        userprofiledto.UserProfileResponse `json:"user"`
	ReportCount int                                `json:"report_count"`
	Reports     []ReportResponse                   `json:"reports"`
}

// ReportedEntitiesResponse: combined moderation payload.
type ReportedEntitiesResponse struct {
	Threads []ReportedThreadGroupResponse `json:"threads"`
	Users   []ReportedUserGroupResponse   `json:"users"`
}
