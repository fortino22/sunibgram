package foxsessionrepo

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	foxhistorymodel "github.com/sunibgram/domain/model/fox_history"
	foxsessionmodel "github.com/sunibgram/domain/model/fox_session"
	foxsessionrepoiface "github.com/sunibgram/domain/repository/fox_session"
	"github.com/sunibgram/infrastructure/cache"
	infrarepoShared "github.com/sunibgram/infrastructure/repository/_shared"
)

const (
	foxSessionTTL = 5 * time.Minute
	maxTitleLen   = 120
)

type FoxSessionRepository struct {
	writeDB *gorm.DB
	readDB  *gorm.DB
	cache   *cache.Cache
}

func NewFoxSessionRepository(writeDB, readDB *gorm.DB, c *cache.Cache) foxsessionrepoiface.IFoxSessionRepository {
	return &FoxSessionRepository{writeDB: writeDB, readDB: readDB, cache: c}
}

func foxSessionListKey(userID uint, offset, limit int, q string) string {
	return fmt.Sprintf("fox_session:user:%d:q:%s:offset:%d:limit:%d", userID, q, offset, limit)
}

// foxSessionTurnsKey is keyed by user+session so a non-owner can never read an
// owner's cached turns (their key never exists for that session).
func foxSessionTurnsKey(userID, sessionID uint) string {
	return fmt.Sprintf("fox_session:turns:user:%d:session:%d", userID, sessionID)
}

type listCachePayload struct {
	Items []foxsessionrepoiface.SessionPreview `json:"items"`
	Total int64                                `json:"total"`
}

// truncateTitle trims a session title to a sane length for the list label.
func truncateTitle(s string) string {
	runes := []rune(s)
	if len(runes) > maxTitleLen {
		return string(runes[:maxTitleLen]) + "…"
	}
	return s
}

func (r *FoxSessionRepository) CreateChatSession(userID uint, title string) (uint, error) {
	session := foxsessionmodel.FoxSession{
		UserID: userID,
		Type:   foxsessionmodel.TypeChat,
		Title:  truncateTitle(title),
	}
	if err := r.writeDB.Create(&session).Error; err != nil {
		return 0, err
	}
	r.invalidateUser(userID)
	return session.ID, nil
}

func (r *FoxSessionRepository) AppendChatTurn(sessionID uint, question, answer string) error {
	var ownerID uint
	err := r.writeDB.Transaction(func(tx *gorm.DB) error {
		var session foxsessionmodel.FoxSession
		if err := tx.First(&session, sessionID).Error; err != nil {
			return err
		}
		ownerID = session.UserID
		entry := foxhistorymodel.FoxHistory{
			SessionID: sessionID,
			UserID:    session.UserID,
			Type:      foxhistorymodel.TypeChat,
			Question:  question,
			Body:      answer,
		}
		if err := tx.Create(&entry).Error; err != nil {
			return err
		}
		// Touch the session so it sorts to the top of the history list.
		return tx.Model(&foxsessionmodel.FoxSession{}).
			Where("id = ?", sessionID).
			Update("updated_at", time.Now()).Error
	})
	if err != nil {
		return err
	}
	// The per-user tag covers both the user's list pages and their session-turn
	// caches, so this single call invalidates this session's turns too.
	r.invalidateUser(ownerID)
	return nil
}

func (r *FoxSessionRepository) UpsertSummarySession(userID, threadID uint, summary, snapshot string) (uint, error) {
	tid := threadID
	title := truncateTitle(snapshot)
	var sessionID uint

	err := r.writeDB.Transaction(func(tx *gorm.DB) error {
		session := foxsessionmodel.FoxSession{
			UserID:   userID,
			Type:     foxsessionmodel.TypeSummary,
			ThreadID: &tid,
			Title:    title,
		}
		// Conflict target is the partial unique index on (user_id, thread_id)
		// WHERE type='summary' (created in migration).
		if err := tx.Clauses(clause.OnConflict{
			Columns:     []clause.Column{{Name: "user_id"}, {Name: "thread_id"}},
			TargetWhere: clause.Where{Exprs: []clause.Expression{clause.Eq{Column: "type", Value: foxsessionmodel.TypeSummary}}},
			DoUpdates: clause.Assignments(map[string]any{
				"title":      title,
				"is_stale":   false,
				"updated_at": time.Now(),
			}),
		}).Create(&session).Error; err != nil {
			return err
		}
		sessionID = session.ID
		if sessionID == 0 {
			// RETURNING did not populate the PK on this driver — fetch it.
			var row struct{ ID uint }
			if err := tx.Model(&foxsessionmodel.FoxSession{}).
				Where("user_id = ? AND thread_id = ? AND type = ?", userID, tid, foxsessionmodel.TypeSummary).
				Select("id").Scan(&row).Error; err != nil {
				return err
			}
			sessionID = row.ID
		}

		// Upsert the single summary entry for this session: update in place, or
		// insert if this is the session's first summarize.
		res := tx.Model(&foxhistorymodel.FoxHistory{}).
			Where("session_id = ? AND type = ?", sessionID, foxhistorymodel.TypeSummary).
			Updates(map[string]any{"body": summary, "snapshot": snapshot, "updated_at": time.Now()})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			entry := foxhistorymodel.FoxHistory{
				SessionID: sessionID,
				UserID:    userID,
				Type:      foxhistorymodel.TypeSummary,
				ThreadID:  &tid,
				Body:      summary,
				Snapshot:  snapshot,
			}
			if err := tx.Create(&entry).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	r.invalidateUser(userID)
	return sessionID, nil
}

func (r *FoxSessionRepository) OwnsSession(userID, sessionID uint) (bool, error) {
	var count int64
	if err := r.readDB.Model(&foxsessionmodel.FoxSession{}).
		Where("id = ? AND user_id = ?", sessionID, userID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *FoxSessionRepository) GetSessionTurns(userID, sessionID uint) ([]foxsessionrepoiface.ChatTurn, error) {
	ctx := context.Background()
	key := foxSessionTurnsKey(userID, sessionID)
	if r.cache != nil {
		var cached []foxsessionrepoiface.ChatTurn
		if err := r.cache.Get(ctx, key, &cached); err == nil {
			return cached, nil
		}
	}

	// Ownership gate: a non-owner gets an empty list and we never cache it, so
	// only the owner's key is ever populated.
	owns, err := r.OwnsSession(userID, sessionID)
	if err != nil {
		return nil, err
	}
	if !owns {
		return []foxsessionrepoiface.ChatTurn{}, nil
	}

	rows := []foxhistorymodel.FoxHistory{}
	if err := r.readDB.
		Where("session_id = ? AND type = ?", sessionID, foxhistorymodel.TypeChat).
		Order("id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	turns := make([]foxsessionrepoiface.ChatTurn, len(rows))
	for i, row := range rows {
		turns[i] = foxsessionrepoiface.ChatTurn{Question: row.Question, Answer: row.Body}
	}

	if r.cache != nil {
		// Registered under the same per-user tag as the list, so any write that
		// calls invalidateUser drops these turns too.
		_ = r.cache.Set(ctx, key, turns, foxSessionTTL)
		_ = r.cache.RegisterCacheKey(ctx, infrarepoShared.FoxSessionUserTag(userID), key)
	}
	return turns, nil
}

func (r *FoxSessionRepository) ListSessions(userID uint, offset, limit int, q string) ([]foxsessionrepoiface.SessionPreview, int64, error) {
	ctx := context.Background()
	key := foxSessionListKey(userID, offset, limit, q)
	if r.cache != nil {
		var payload listCachePayload
		if err := r.cache.Get(ctx, key, &payload); err == nil {
			return payload.Items, payload.Total, nil
		}
	}

	base := r.readDB.Model(&foxsessionmodel.FoxSession{}).Where("user_id = ?", userID)
	if q != "" {
		pattern := "%" + q + "%"
		base = base.Where(
			"title ILIKE ? OR EXISTS (SELECT 1 FROM fox_histories h WHERE h.session_id = fox_sessions.id AND (h.question ILIKE ? OR h.body ILIKE ? OR h.snapshot ILIKE ?))",
			pattern, pattern, pattern, pattern,
		)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sessions := []foxsessionmodel.FoxSession{}
	if total > 0 {
		if err := base.Order("updated_at DESC").Offset(offset).Limit(limit).Find(&sessions).Error; err != nil {
			return nil, 0, err
		}
	}

	previews, err := r.hydratePreviews(sessions)
	if err != nil {
		return nil, 0, err
	}

	if r.cache != nil {
		_ = r.cache.Set(ctx, key, listCachePayload{Items: previews, Total: total}, foxSessionTTL)
		_ = r.cache.RegisterCacheKey(ctx, infrarepoShared.FoxSessionUserTag(userID), key)
	}
	return previews, total, nil
}

// hydratePreviews attaches the representative entry + turn count to each session:
// summary sessions get their single summary entry; chat sessions get their
// latest turn and total turn count.
func (r *FoxSessionRepository) hydratePreviews(sessions []foxsessionmodel.FoxSession) ([]foxsessionrepoiface.SessionPreview, error) {
	previews := make([]foxsessionrepoiface.SessionPreview, len(sessions))
	if len(sessions) == 0 {
		return previews, nil
	}

	ids := make([]uint, len(sessions))
	for i, s := range sessions {
		previews[i] = foxsessionrepoiface.SessionPreview{Session: s}
		ids[i] = s.ID
	}

	// Latest entry per session (works for both kinds: summary sessions have one
	// entry, chat sessions get their newest turn for the preview).
	latest := []foxhistorymodel.FoxHistory{}
	if err := r.readDB.
		Raw("SELECT DISTINCT ON (session_id) * FROM fox_histories WHERE session_id IN ? ORDER BY session_id, id DESC", ids).
		Scan(&latest).Error; err != nil {
		return nil, err
	}
	entryBySession := make(map[uint]foxhistorymodel.FoxHistory, len(latest))
	for _, e := range latest {
		entryBySession[e.SessionID] = e
	}

	// Chat turn counts.
	type countRow struct {
		SessionID uint
		Cnt       int
	}
	counts := []countRow{}
	if err := r.readDB.Model(&foxhistorymodel.FoxHistory{}).
		Select("session_id, COUNT(*) AS cnt").
		Where("session_id IN ? AND type = ?", ids, foxhistorymodel.TypeChat).
		Group("session_id").
		Scan(&counts).Error; err != nil {
		return nil, err
	}
	countBySession := make(map[uint]int, len(counts))
	for _, c := range counts {
		countBySession[c.SessionID] = c.Cnt
	}

	for i := range previews {
		sid := previews[i].Session.ID
		if e, ok := entryBySession[sid]; ok {
			entry := e
			previews[i].Entry = &entry
		}
		previews[i].TurnCount = countBySession[sid]
	}
	return previews, nil
}

func (r *FoxSessionRepository) invalidateUser(userID uint) {
	if r.cache == nil {
		return
	}
	_ = r.cache.InvalidateTag(context.Background(), infrarepoShared.FoxSessionUserTag(userID))
}
