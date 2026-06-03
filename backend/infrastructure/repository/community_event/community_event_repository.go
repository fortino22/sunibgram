package communityeventrepo

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	communityeventmodel "github.com/sunibgram/domain/model/community_event"
	communityeventrepoiface "github.com/sunibgram/domain/repository/community_event"
)

type CommunityEventRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
}

func NewCommunityEventRepository(writeDB, readDB *gorm.DB) communityeventrepoiface.ICommunityEventRepository {
	return &CommunityEventRepository{writeDB: writeDB, readDB: readDB}
}

// ---- Event CRUD ----

func (r *CommunityEventRepository) Save(event *communityeventmodel.CommunityEvent) error {
	return r.writeDB.Create(event).Error
}

func (r *CommunityEventRepository) FindByID(id uint) (*communityeventmodel.CommunityEvent, error) {
	var event communityeventmodel.CommunityEvent
	if err := r.readDB.
		Preload("Creator").
		First(&event, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *CommunityEventRepository) FindVisibleByCommunity(communityID, requesterID uint, isAdmin bool, page, limit int) ([]communityeventmodel.CommunityEvent, int64, error) {
	q := r.readDB.Model(&communityeventmodel.CommunityEvent{}).
		Where("community_id = ?", communityID)
	if !isAdmin {
		q = q.Where(
			"EXISTS (SELECT 1 FROM community_event_invitees i WHERE i.community_event_id = community_events.id AND i.user_id = ?)",
			requesterID,
		)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var events []communityeventmodel.CommunityEvent
	err := q.
		Preload("Creator").
		Order("start_time DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&events).Error
	return events, total, err
}

func (r *CommunityEventRepository) Update(event *communityeventmodel.CommunityEvent) error {
	return r.writeDB.Save(event).Error
}

func (r *CommunityEventRepository) Delete(id uint) error {
	return r.writeDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("community_event_id = ?", id).Delete(&communityeventmodel.CommunityEventInvitee{}).Error; err != nil {
			return err
		}
		if err := tx.Where("community_event_id = ?", id).Delete(&communityeventmodel.CommunityEventComment{}).Error; err != nil {
			return err
		}
		if err := tx.Where("community_event_id = ?", id).Delete(&communityeventmodel.CommunityEventReaction{}).Error; err != nil {
			return err
		}
		return tx.Delete(&communityeventmodel.CommunityEvent{}, "id = ?", id).Error
	})
}

// ---- Invitees ----

func (r *CommunityEventRepository) ReplaceInvitees(eventID uint, userIDs []uint) error {
	return r.writeDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("community_event_id = ?", eventID).Delete(&communityeventmodel.CommunityEventInvitee{}).Error; err != nil {
			return err
		}
		if len(userIDs) == 0 {
			return nil
		}
		rows := make([]communityeventmodel.CommunityEventInvitee, 0, len(userIDs))
		for _, uid := range userIDs {
			rows = append(rows, communityeventmodel.CommunityEventInvitee{CommunityEventID: eventID, UserID: uid})
		}
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
	})
}

func (r *CommunityEventRepository) FindInvitees(eventID uint) ([]communityeventmodel.CommunityEventInvitee, error) {
	var invitees []communityeventmodel.CommunityEventInvitee
	err := r.readDB.
		Preload("User").
		Preload("Profile").
		Where("community_event_id = ?", eventID).
		Order("created_at ASC").
		Find(&invitees).Error
	return invitees, err
}

func (r *CommunityEventRepository) IsInvitee(eventID, userID uint) (bool, error) {
	var count int64
	err := r.readDB.Model(&communityeventmodel.CommunityEventInvitee{}).
		Where("community_event_id = ? AND user_id = ?", eventID, userID).
		Count(&count).Error
	return count > 0, err
}

// ---- Comments ----

func (r *CommunityEventRepository) SaveComment(comment *communityeventmodel.CommunityEventComment) error {
	return r.writeDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(comment).Error; err != nil {
			return err
		}
		return tx.Model(&communityeventmodel.CommunityEvent{}).
			Where("id = ?", comment.CommunityEventID).
			UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error
	})
}

func (r *CommunityEventRepository) UpdateComment(comment *communityeventmodel.CommunityEventComment) error {
	return r.writeDB.Model(&communityeventmodel.CommunityEventComment{}).
		Where("id = ?", comment.ID).
		Update("content", comment.Content).Error
}

func (r *CommunityEventRepository) FindComments(eventID uint, page, limit int) ([]communityeventmodel.CommunityEventComment, int64, error) {
	q := r.readDB.Model(&communityeventmodel.CommunityEventComment{}).
		Where("community_event_id = ?", eventID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var comments []communityeventmodel.CommunityEventComment
	err := q.
		Preload("User").
		Preload("Profile").
		Order("created_at ASC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&comments).Error
	return comments, total, err
}

func (r *CommunityEventRepository) FindCommentByID(commentID uint) (*communityeventmodel.CommunityEventComment, error) {
	var comment communityeventmodel.CommunityEventComment
	if err := r.readDB.
		Preload("User").
		Preload("Profile").
		First(&comment, "id = ?", commentID).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *CommunityEventRepository) DeleteComment(commentID uint) error {
	return r.writeDB.Transaction(func(tx *gorm.DB) error {
		var comment communityeventmodel.CommunityEventComment
		if err := tx.First(&comment, "id = ?", commentID).Error; err != nil {
			return err
		}
		if err := tx.Delete(&communityeventmodel.CommunityEventComment{}, "id = ?", commentID).Error; err != nil {
			return err
		}
		// Clear the pin if this comment was the pinned one.
		if err := tx.Model(&communityeventmodel.CommunityEvent{}).
			Where("id = ? AND pinned_comment_id = ?", comment.CommunityEventID, commentID).
			UpdateColumn("pinned_comment_id", gorm.Expr("NULL")).Error; err != nil {
			return err
		}
		return tx.Model(&communityeventmodel.CommunityEvent{}).
			Where("id = ? AND comment_count > 0", comment.CommunityEventID).
			UpdateColumn("comment_count", gorm.Expr("comment_count - 1")).Error
	})
}

func (r *CommunityEventRepository) SetPinnedComment(eventID uint, commentID *uint) error {
	q := r.writeDB.Model(&communityeventmodel.CommunityEvent{}).Where("id = ?", eventID)
	if commentID == nil {
		return q.UpdateColumn("pinned_comment_id", gorm.Expr("NULL")).Error
	}
	return q.UpdateColumn("pinned_comment_id", *commentID).Error
}

// ---- Reactions ----

func (r *CommunityEventRepository) UpsertReaction(eventID, userID uint, emoji string) error {
	reaction := communityeventmodel.CommunityEventReaction{
		CommunityEventID: eventID,
		UserID:           userID,
		Emoji:            emoji,
	}
	return r.writeDB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "community_event_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"emoji"}),
	}).Create(&reaction).Error
}

func (r *CommunityEventRepository) RemoveReaction(eventID, userID uint) error {
	return r.writeDB.
		Where("community_event_id = ? AND user_id = ?", eventID, userID).
		Delete(&communityeventmodel.CommunityEventReaction{}).Error
}

func (r *CommunityEventRepository) FindUserReaction(eventID, userID uint) (string, error) {
	var reaction communityeventmodel.CommunityEventReaction
	err := r.readDB.
		Where("community_event_id = ? AND user_id = ?", eventID, userID).
		First(&reaction).Error
	if err == gorm.ErrRecordNotFound {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return reaction.Emoji, nil
}

func (r *CommunityEventRepository) FindReactionSummary(eventID uint) ([]communityeventrepoiface.ReactionCount, error) {
	var rows []communityeventrepoiface.ReactionCount
	err := r.readDB.Model(&communityeventmodel.CommunityEventReaction{}).
		Select("emoji, COUNT(*) as count").
		Where("community_event_id = ?", eventID).
		Group("emoji").
		Order("count DESC").
		Scan(&rows).Error
	return rows, err
}
