package trendingthreaddto

import "time"

type TrendingThreadReceivedDetail struct {
	Likes     int `json:"likes"`
	Comments  int `json:"comments"`
	Bookmarks int `json:"bookmarks"`
	Reposts   int `json:"reposts"`
}

type TrendingThreadGivenDetail struct {
	Likes     int `json:"likes"`
	Bookmarks int `json:"bookmarks"`
	Comments  int `json:"comments"`
	Posts     int `json:"posts"`
}

type TrendingThreadScoreDetail struct {
	Received TrendingThreadReceivedDetail `json:"received"`
	Given    TrendingThreadGivenDetail    `json:"given"`
}

type TrendingThreadAuthor struct {
	UserID    uint   `json:"userId"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	Profile   string `json:"profile"`
	IsPremium bool   `json:"is_premium"`
}

type TrendingThreadResponse struct {
	ThreadID  uint                      `json:"threadId"`
	Content   string                    `json:"content"`
	CreatedAt time.Time                 `json:"created_at"`
	Score     float64                   `json:"score"`
	Author    TrendingThreadAuthor      `json:"author"`
	Detail    TrendingThreadScoreDetail `json:"detail"`
}

type TrendingThreadMeta struct {
	Total            int64     `json:"total"`
	CachedAt         time.Time `json:"cached_at"`
	RefreshAt        time.Time `json:"refresh_at"`
	RefreshInSeconds int64     `json:"refresh_in_seconds"`
}

type TrendingThreadListResponse struct {
	Data []TrendingThreadResponse `json:"data"`
	Meta TrendingThreadMeta       `json:"meta"`
}
