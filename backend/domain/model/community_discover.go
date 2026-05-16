package model

type DiscoverCommunityMember struct {
	AvatarObjectName string
	AvatarBucket     string
}

type DiscoverCommunity struct {
	CommunityID uint
	Name        string
	MemberCount int
	Category    string
	TopMembers  []DiscoverCommunityMember
}
