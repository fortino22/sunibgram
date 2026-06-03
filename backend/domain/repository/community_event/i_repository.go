package communityeventrepoiface

import (
	communityeventmodel "github.com/sunibgram/domain/model/community_event"
)

// ReactionCount is one emoji and how many users picked it for an event.
type ReactionCount struct {
	Emoji string
	Count int
}

type ICommunityEventRepository interface {
	// Event CRUD
	Save(event *communityeventmodel.CommunityEvent) error
	FindByID(id uint) (*communityeventmodel.CommunityEvent, error)
	// FindVisibleByCommunity lists events for a community. Admins see all; other
	// users see only events they were tagged in.
	FindVisibleByCommunity(communityID, requesterID uint, isAdmin bool, page, limit int) ([]communityeventmodel.CommunityEvent, int64, error)
	Update(event *communityeventmodel.CommunityEvent) error
	Delete(id uint) error

	// Invitees
	ReplaceInvitees(eventID uint, userIDs []uint) error
	FindInvitees(eventID uint) ([]communityeventmodel.CommunityEventInvitee, error)
	IsInvitee(eventID, userID uint) (bool, error)

	// Comments
	SaveComment(comment *communityeventmodel.CommunityEventComment) error
	UpdateComment(comment *communityeventmodel.CommunityEventComment) error
	FindComments(eventID uint, page, limit int) ([]communityeventmodel.CommunityEventComment, int64, error)
	FindCommentByID(commentID uint) (*communityeventmodel.CommunityEventComment, error)
	DeleteComment(commentID uint) error
	SetPinnedComment(eventID uint, commentID *uint) error

	// Reactions (single per user; upsert toggles/replaces the emoji)
	UpsertReaction(eventID, userID uint, emoji string) error
	RemoveReaction(eventID, userID uint) error
	FindUserReaction(eventID, userID uint) (string, error)
	FindReactionSummary(eventID uint) ([]ReactionCount, error)
}
