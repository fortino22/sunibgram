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

type ExploreCommunityResult struct {
	CommunityID      uint
	Name             string
	AvatarObjectName string
	AvatarBucket     string
	MemberCount      int
	IsJoin           bool
}

type ExploreResult struct {
	Users   []ExploreUserResult
	Threads []ExploreThreadResult
}
