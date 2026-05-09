package model

type ExploreUserResult struct {
	UserID           uint
	Name             string
	Username         string
	IsPremium        bool
	AvatarObjectName string
	AvatarBucket     string
	IsFollowing      bool
}

type ExploreThreadResult struct {
	UserID  uint
	Content string
}

type ExploreResult struct {
	Users   []ExploreUserResult
	Threads []ExploreThreadResult
}
