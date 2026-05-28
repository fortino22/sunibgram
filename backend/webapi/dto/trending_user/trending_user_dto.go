package trendinguserdto

import "time"

type TrendingUserReceivedDetail struct {
	Likes       int `json:"likes"`
	Comments    int `json:"comments"`
	Bookmarks   int `json:"bookmarks"`
	Reposts     int `json:"reposts"`
	ThreadCount int `json:"thread_count"`
}

type TrendingUserGivenDetail struct {
	Likes     int `json:"likes"`
	Bookmarks int `json:"bookmarks"`
	Comments  int `json:"comments"`
	Posts     int `json:"posts"`
}

type TrendingUserScoreDetail struct {
	Received TrendingUserReceivedDetail `json:"received"`
	Given    TrendingUserGivenDetail    `json:"given"`
}

type TrendingUserResponse struct {
	UserID      uint                    `json:"userId"`
	Username    string                  `json:"username"`
	Name        string                  `json:"name"`
	Profile     string                  `json:"profile"`
	IsPremium   bool                    `json:"is_premium"`
	IsFollowing bool                    `json:"is_following"`
	Score       float64                 `json:"score"`
	Detail      TrendingUserScoreDetail `json:"detail"`
}

type TrendingUserMeta struct {
	Total            int64     `json:"total"`
	Page             int       `json:"page"`
	Limit            int       `json:"limit"`
	CachedAt         time.Time `json:"cached_at"`
	RefreshAt        time.Time `json:"refresh_at"`
	RefreshInSeconds int64     `json:"refresh_in_seconds"`
}

type PaginatedTrendingUserResponse struct {
	Data []TrendingUserResponse `json:"data"`
	Meta TrendingUserMeta       `json:"meta"`
}
