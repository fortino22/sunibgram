package exploredto

import (
	threaddto "github.com/sunibgram/webapi/dto/thread"
	threadmediadto "github.com/sunibgram/webapi/dto/thread_media"
)

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

type ExplorePeopleResponse struct {
	UserID         uint   `json:"userId"`
	UserProfilePic string `json:"user_profile_pic"`
	Username       string `json:"username"`
	Name           string `json:"name"`
	Bio            string `json:"bio"`
	IsPremium      bool   `json:"is_premium"`
	IsFollowing    bool   `json:"is_following"`
}

type ExploreThreadResponse struct {
	Content string `json:"content"`
}

type ExploreMediaResponse struct {
	ThreadID uint                                 `json:"thread_id"`
	Media    []threadmediadto.ThreadMediaResponse `json:"media"`
}

type PaginatedExploreMediaResponse struct {
	Data []ExploreMediaResponse `json:"data"`
	Meta threaddto.FeedMeta     `json:"meta"`
}

type ExploreCommunityResponse struct {
	CommunityID   uint   `json:"community_id"`
	CommunityName string `json:"community_name"`
	MemberCount   int    `json:"member_count"`
	IsJoin        bool   `json:"is_join"`
}
