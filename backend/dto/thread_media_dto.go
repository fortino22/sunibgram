package dto

type DeleteMediaRequest struct {
	MediaIDs []uint `json:"media_ids"`
}

type ThreadMediaResponse struct {
	MediaID uint   `json:"media_id"`
	URL     string `json:"url"`
}
