package dto

type ExploreResponse struct {
	Users   []ExploreUserResponse   `json:"users"`
	Threads []ExploreThreadResponse `json:"threads"`
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
