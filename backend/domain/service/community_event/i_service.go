package communityeventsvciface

import (
	"context"
	"mime/multipart"
	"time"

	communityeventmodel "github.com/sunibgram/domain/model/community_event"
	communityeventrepoiface "github.com/sunibgram/domain/repository/community_event"
)

// CreateCommunityEventInput carries the fields needed to create or update a
// community event.
type CreateCommunityEventInput struct {
	Title          string
	Description    string
	Location       string
	StartTime      time.Time
	EndTime        time.Time
	Banner         *multipart.FileHeader
	ClearBanner    bool
	InviteeEmails  []string
	InviteeUserIDs []uint
}

// InviteeCandidate is a taggable user surfaced by the picker search (never
// includes email).
type InviteeCandidate struct {
	UserID           uint
	Username         string
	Name             string
	AvatarObjectName string
}

// CommunityEventDetail is the fully hydrated view of a single event.
type CommunityEventDetail struct {
	Event          *communityeventmodel.CommunityEvent
	Invitees       []communityeventmodel.CommunityEventInvitee
	ReactionCounts []communityeventrepoiface.ReactionCount
	MyReaction     string
	PinnedComment  *communityeventmodel.CommunityEventComment
	IsAdmin        bool
	CanComment     bool
}

// CommentList is a paginated comment listing plus viewer context.
type CommentList struct {
	Comments        []communityeventmodel.CommunityEventComment
	Total           int64
	PinnedCommentID *uint
	IsAdmin         bool
}

type ICommunityEventService interface {
	Create(ctx context.Context, communityID, actorID uint, input CreateCommunityEventInput) (*communityeventmodel.CommunityEvent, error)
	SearchInviteeCandidates(ctx context.Context, communityID, requesterID uint, query string) ([]InviteeCandidate, error)
	GetByID(ctx context.Context, eventID, requesterID uint) (*CommunityEventDetail, error)
	GetByCommunity(ctx context.Context, communityID, requesterID uint, page, limit int) ([]communityeventmodel.CommunityEvent, int64, error)
	Update(ctx context.Context, eventID, actorID uint, input CreateCommunityEventInput) (*communityeventmodel.CommunityEvent, error)
	Delete(ctx context.Context, eventID, actorID uint) error

	AddComment(ctx context.Context, eventID, actorID uint, content string) (*communityeventmodel.CommunityEventComment, error)
	UpdateComment(ctx context.Context, eventID, commentID, actorID uint, content string) (*communityeventmodel.CommunityEventComment, error)
	GetComments(ctx context.Context, eventID, requesterID uint, page, limit int) (*CommentList, error)
	DeleteComment(ctx context.Context, eventID, commentID, actorID uint) error
	PinComment(ctx context.Context, eventID, commentID, actorID uint) error
	UnpinComment(ctx context.Context, eventID, actorID uint) error

	// SetReaction sets the actor's single emoji reaction; picking the same emoji
	// again removes it. Returns the resulting emoji ("" when removed).
	SetReaction(ctx context.Context, eventID, actorID uint, emoji string) (string, error)
	RemoveReaction(ctx context.Context, eventID, actorID uint) error

	GetURL(objectName string) string
}
