package dto

import "time"

type AvatarResponse struct {
	AvatarID   uint      `json:"avatar_id"`
	UserID     uint      `json:"user_id"`
	URL        string    `json:"url"`
	ObjectName string    `json:"object_name"`
	Bucket     string    `json:"bucket"`
	UpdatedAt  time.Time `json:"updated_at"`
}
