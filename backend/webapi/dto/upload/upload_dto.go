package uploaddto

type UploadResponse struct {
	URL        string `json:"url"`
	ObjectName string `json:"object_name"`
	Type       string `json:"type"`
}
