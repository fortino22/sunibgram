package dto

type ExploreResponse struct {
	Threads []ExploreThreadResponse `json:"threads"`
	Users   []ExploreUserResponse   `json:"users"`
}

type ExploreUserResponse struct {
	UserID      uint   `json:"userId"`
	Profile     string `json:"profile"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	IsPremium   bool   `json:"is_premium"`
	IsFollowing bool   `json:"is_following"`
}

type ExploreThreadResponse struct {
	Content string `json:"content"`
}
