package model

type DiscoverCommunityMember struct {
	UserID           uint
	AvatarObjectName string
	AvatarBucket     string
}

type DiscoverCommunity struct {
	CommunityID      uint
	Name             string
	BannerObjectName string
	BannerBucket     string
	MemberCount      int
	Category         *Category
	TopMembers       []DiscoverCommunityMember
}
