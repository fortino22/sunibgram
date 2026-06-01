package trendingthreadmodel

import "time"

// TrendingThread is a read-model: a top thread of the rolling weekly window
// together with the engagement it received and the activity its author gave in
// the same window. Not migrated — assembled from aggregate queries over
// threads/likes/bookmarks/reposts.
type TrendingThread struct {
	ThreadID  uint
	Content   string
	CreatedAt time.Time
	Score     float64

	// Author — the "they" whose given activity we attach.
	AuthorID         uint
	Username         string
	Name             string
	AvatarObjectName string
	IsPremium        bool

	// Received by this thread, within the window.
	LikesIn     int
	CommentsIn  int
	BookmarksIn int
	RepostsIn   int

	// Given by the author, within the window.
	LikesOut     int
	CommentsOut  int
	BookmarksOut int
	PostsOut     int
}
