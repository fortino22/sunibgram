package model

type DiscoverCommunityMember struct {
	BannerObjectName string
	BannerBucket     string
}

type DiscoverCommunity struct {
	CommunityID uint
	Name        string
	MemberCount int
	Category    *Category
	TopMembers  []DiscoverCommunityMember
}
