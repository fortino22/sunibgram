package exploremodel

type ExploreUserResult struct {
	UserID           uint
	Name             string
	Username         string
	Bio              string
	IsPremium        bool
	AvatarObjectName string
	IsFollowing      bool
}

type ExploreThreadResult struct {
	UserID  uint
	Content string
}

type ExploreCommunityResult struct {
	CommunityID      uint
	Name             string
	MemberCount      int
	IsJoin           bool
	BannerObjectName string
}

type ExploreResult struct {
	Users   []ExploreUserResult
	Threads []ExploreThreadResult
}
