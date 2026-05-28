package trendingusermodel

type TrendingUser struct {
	UserID           uint
	Username         string
	Name             string
	AvatarObjectName string
	IsPremium        bool
	IsFollowing      bool
	Score            float64

	ThreadCount  int
	LikesIn      int
	CommentsIn   int
	BookmarksIn  int
	RepostsIn    int

	LikesOut     int
	CommentsOut  int
	BookmarksOut int
	PostsOut     int
}
