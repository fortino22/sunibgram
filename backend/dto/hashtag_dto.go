package dto

type HashtagResponse struct {
	Name string `json:"name"`
}

type TrendingHashtagResponse struct {
	Name        string  `json:"name"`
	ThreadCount int64   `json:"thread_count"`
	Score       float64 `json:"score"`
}

type ThreadsByHashtagResponse struct {
	Data       []FeedItemResponse `json:"data"`
	NextCursor *uint              `json:"next_cursor,omitempty"`
	Limit      int                `json:"limit"`
}
