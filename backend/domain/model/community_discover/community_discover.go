package communitydiscovermodel

import (
	categorymodel "github.com/sunibgram/domain/model/category"
)

type DiscoverCommunityMember struct {
	UserID           uint
	AvatarObjectName string
}

type DiscoverCommunity struct {
	CommunityID      uint
	Name             string
	BannerObjectName string
	MemberCount      int
	Category         *categorymodel.Category
	TopMembers       []DiscoverCommunityMember
}
