package dto

type UploadResponse struct {
	URL        string `json:"url"`
	ObjectName string `json:"object_name"`
	Bucket     string `json:"bucket"`
	Type       string `json:"type"`
}
