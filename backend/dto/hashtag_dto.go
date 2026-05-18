package dto

type HashtagResponse struct {
	Name string `json:"name"`
}

type TrendingHashtagResponse struct {
	Name        string `json:"name"`
	ThreadCount int64  `json:"thread_count"`
}
