package communitysvciface

import (
	"context"
	"mime/multipart"

	"github.com/google/uuid"

	communitymodel "github.com/sunibgram/domain/model/community"
	communitydiscovermodel "github.com/sunibgram/domain/model/community_discover"
	communityrulemodel "github.com/sunibgram/domain/model/community_rule"
	usermodel "github.com/sunibgram/domain/model/user"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
)

type ICommunityService interface {
	Create(ctx context.Context, userID uint, name, description string, rules []string, categoryID *uuid.UUID, visibility communitymodel.CommunityVisibility, banner *multipart.FileHeader) (*communitymodel.Community, []communityrulemodel.CommunityRule, error)
	Update(ctx context.Context, id, userID uint, name, description string, keepRuleIDs []uint, newRules []string, categoryID *uuid.UUID, visibility communitymodel.CommunityVisibility, banner *multipart.FileHeader, clearBanner bool) (*communitymodel.Community, []communityrulemodel.CommunityRule, error)
	Delete(ctx context.Context, id, userID uint) error
	Discover(ctx context.Context, categoryID *uuid.UUID) ([]communitydiscovermodel.DiscoverCommunity, error)
	SearchByName(ctx context.Context, name string) ([]communitydiscovermodel.DiscoverCommunity, error)
	GetByID(id uint) (*communitymodel.Community, error)
	GetAbout(ctx context.Context, communityID uint, viewerID *uint) (*CommunityAbout, error)
	GetMyCreated(userID uint, page, limit int) ([]communitymodel.Community, int64, error)
	GetMyJoined(userID uint, page, limit int) ([]communitymodel.Community, int64, error)
	GetURL(objectName string) string
}

type CommunityAboutUser struct {
	User        *usermodel.User
	Profile     *userprofilemodel.UserProfile
	IsFollowing bool
}

type CommunityAbout struct {
	Community  *communitymodel.Community
	Creator    CommunityAboutUser
	Rules      []communityrulemodel.CommunityRule
	Admins     []CommunityAboutUser
	TopMembers []CommunityAboutUser
}
