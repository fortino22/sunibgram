package service

import (
	"context"
	"mime/multipart"

	"github.com/google/uuid"
	"github.com/sunibgram/domain/model"
)

type ICommunityService interface {
	Create(ctx context.Context, userID uint, name, description string, rules []string, categoryID *uuid.UUID, visibility model.CommunityVisibility, banner *multipart.FileHeader) (*model.Community, []model.CommunityRule, error)
	Update(ctx context.Context, id, userID uint, name, description string, keepRuleIDs []uint, newRules []string, categoryID *uuid.UUID, visibility model.CommunityVisibility, banner *multipart.FileHeader, clearBanner bool) (*model.Community, []model.CommunityRule, error)
	Delete(ctx context.Context, id, userID uint) error
	Discover(ctx context.Context, categoryID *uuid.UUID) ([]model.DiscoverCommunity, error)
	SearchByName(ctx context.Context, name string) ([]model.DiscoverCommunity, error)
	GetByID(id uint) (*model.Community, error)
	GetAbout(ctx context.Context, communityID uint, viewerID *uint) (*CommunityAbout, error)
	GetMyCreated(userID uint, page, limit int) ([]model.Community, int64, error)
	GetMyJoined(userID uint, page, limit int) ([]model.Community, int64, error)
	GetURL(bucket, objectName string) string
}

type CommunityAboutUser struct {
	User        *model.User
	Profile     *model.UserProfile
	IsFollowing bool
}

type CommunityAbout struct {
	Community  *model.Community
	Creator    CommunityAboutUser
	Rules      []model.CommunityRule
	Admins     []CommunityAboutUser
	TopMembers []CommunityAboutUser
}
