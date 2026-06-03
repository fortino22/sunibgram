package migration

import (
	"log"

	"gorm.io/gorm"

	bookmarkmodel "github.com/sunibgram/domain/model/bookmark"
	bookmarkcollectionmodel "github.com/sunibgram/domain/model/bookmark_collection"
	categorymodel "github.com/sunibgram/domain/model/category"
	chatmodel "github.com/sunibgram/domain/model/chat"
	chatmediamodel "github.com/sunibgram/domain/model/chat_media"
	chatreactionmodel "github.com/sunibgram/domain/model/chat_reaction"
	chatreadmodel "github.com/sunibgram/domain/model/chat_read"
	chatroommodel "github.com/sunibgram/domain/model/chat_room"
	chatroommembermodel "github.com/sunibgram/domain/model/chat_room_member"
	communitymodel "github.com/sunibgram/domain/model/community"
	communityactionconfirmationmodel "github.com/sunibgram/domain/model/community_action_confirmation"
	communityeventmodel "github.com/sunibgram/domain/model/community_event"
	communitymembermodel "github.com/sunibgram/domain/model/community_member"
	communitypinmodel "github.com/sunibgram/domain/model/community_pin"
	communityrequestmodel "github.com/sunibgram/domain/model/community_request"
	communityrulemodel "github.com/sunibgram/domain/model/community_rule"
	communitythreadpinmodel "github.com/sunibgram/domain/model/community_thread_pin"
	eventactionconfirmationmodel "github.com/sunibgram/domain/model/event_action_confirmation"
	eventfacultylocationmodel "github.com/sunibgram/domain/model/event_faculty_location"
	eventparticipantmodel "github.com/sunibgram/domain/model/event_participant"
	eventpinmodel "github.com/sunibgram/domain/model/event_pin"
	eventreviewmodel "github.com/sunibgram/domain/model/event_review"
	eventthreadpinmodel "github.com/sunibgram/domain/model/event_thread_pin"
	facultymodel "github.com/sunibgram/domain/model/faculty"
	facultylocationmodel "github.com/sunibgram/domain/model/faculty_location"
	feedbackmodel "github.com/sunibgram/domain/model/feedback"
	feedbackmediamodel "github.com/sunibgram/domain/model/feedback_media"
	foxhistorymodel "github.com/sunibgram/domain/model/fox_history"
	foxsessionmodel "github.com/sunibgram/domain/model/fox_session"
	hashtagmodel "github.com/sunibgram/domain/model/hashtag"
	likemodel "github.com/sunibgram/domain/model/like"
	notificationmodel "github.com/sunibgram/domain/model/notification"
	notificationusermutemodel "github.com/sunibgram/domain/model/notification_user_mute"
	ordermodel "github.com/sunibgram/domain/model/order"
	premiumsubscriptionmodel "github.com/sunibgram/domain/model/premium_subscription"
	profilefieldchangemodel "github.com/sunibgram/domain/model/profile_field_change"
	reportmodel "github.com/sunibgram/domain/model/report"
	repostmodel "github.com/sunibgram/domain/model/repost"
	rolemodel "github.com/sunibgram/domain/model/role"
	threadmodel "github.com/sunibgram/domain/model/thread"
	threadmediamodel "github.com/sunibgram/domain/model/thread_media"
	threadpinmodel "github.com/sunibgram/domain/model/thread_pin"
	threadpollmodel "github.com/sunibgram/domain/model/thread_poll"
	threadsummarymodel "github.com/sunibgram/domain/model/thread_summary"
	trendreportmodel "github.com/sunibgram/domain/model/trend_report"
	trendingusersnapshotmodel "github.com/sunibgram/domain/model/trending_user_snapshot"
	usermodel "github.com/sunibgram/domain/model/user"
	userblockmodel "github.com/sunibgram/domain/model/user_block"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
	userrelationshipmodel "github.com/sunibgram/domain/model/user_relationship"
	usersettingsmodel "github.com/sunibgram/domain/model/user_settings"
	userstatmodel "github.com/sunibgram/domain/model/user_stat"
	verificationcodemodel "github.com/sunibgram/domain/model/verification_code"
	violationmodel "github.com/sunibgram/domain/model/violation"
	"github.com/sunibgram/infrastructure/seeder"
)

func AutoMigrate(db *gorm.DB) {
	db.DisableForeignKeyConstraintWhenMigrating = true

	resetLegacyFoxHistory(db)

	models := []any{
		&rolemodel.Role{},
		&usermodel.User{},
		&usersettingsmodel.UserSettings{},
		&userprofilemodel.UserProfile{},
		&profilefieldchangemodel.ProfileFieldChange{},
		&categorymodel.Category{},
		&violationmodel.Violation{},
		&trendreportmodel.TrendReportReason{},
		&communitymodel.Community{},
		&threadmodel.Thread{},
		&reportmodel.Report{},
		&threadmodel.Event{},
		&eventreviewmodel.EventReview{},
		&threadmodel.ThreadEvent{},
		&eventthreadpinmodel.EventThreadPin{},
		&threadmediamodel.ThreadMedia{},
		&threadpollmodel.ThreadPoll{},
		&threadpollmodel.ThreadPollOption{},
		&threadpollmodel.ThreadPollVote{},
		&likemodel.Like{},
		&repostmodel.Repost{},
		&bookmarkmodel.Bookmark{},
		&bookmarkcollectionmodel.BookmarkCollection{},
		&bookmarkcollectionmodel.BookmarkCollectionItem{},
		&threadpinmodel.ThreadPin{},
		&threadsummarymodel.ThreadSummary{},
		&foxsessionmodel.FoxSession{},
		&foxhistorymodel.FoxHistory{},
		&verificationcodemodel.VerificationCode{},
		&userrelationshipmodel.UserRelationship{},
		&userblockmodel.UserBlock{},
		&notificationusermutemodel.NotificationUserMute{},
		&userstatmodel.UserStat{},
		&communitymembermodel.CommunityMember{},
		&communityrequestmodel.CommunityRequest{},
		&communityrulemodel.CommunityRule{},
		&communitythreadpinmodel.CommunityThreadPin{},
		&communityeventmodel.CommunityEvent{},
		&communityeventmodel.CommunityEventInvitee{},
		&communityeventmodel.CommunityEventComment{},
		&communityeventmodel.CommunityEventReaction{},
		&communitypinmodel.UserCommunityPin{},
		&communityactionconfirmationmodel.CommunityActionConfirmation{},
		&eventpinmodel.UserEventPin{},
		&eventparticipantmodel.EventParticipant{},
		&eventactionconfirmationmodel.EventActionConfirmation{},
		&facultymodel.Faculty{},
		&facultylocationmodel.FacultyLocation{},
		&eventfacultylocationmodel.EventFacultyLocation{},
		&usermodel.UserFaculty{},
		&ordermodel.Order{},
		&chatroommodel.ChatRoom{},
		&chatroommembermodel.ChatRoomMember{},
		&chatmodel.Chat{},
		&chatmediamodel.ChatMedia{},
		&chatreadmodel.ChatRead{},
		&chatreactionmodel.ChatReaction{},
		&notificationmodel.Notification{},
		&hashtagmodel.Hashtag{},
		&hashtagmodel.ThreadHashtag{},
		&trendreportmodel.TrendReport{},
		&feedbackmodel.Feedback{},
		&feedbackmodel.FeedbackVote{},
		&feedbackmediamodel.FeedbackMedia{},
		&premiumsubscriptionmodel.PremiumSubscription{},
		&trendingusersnapshotmodel.TrendingUserSnapshot{},
	}

	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			log.Fatalf("AutoMigrate %T: %v", m, err)
		}
	}

	dropDeadColumns(db)
	createIndexes(db)
	addCheckConstraints(db)
	addColumns(db)
	addForeignKeys(db)
	migrateChatEncryption(db)
	backfillShareCounters(db)

	seedRoles(db)
	seeder.SeedCategory(db)
	seeder.SeedViolation(db)
	seeder.SeedTrendReportReason(db)
	seeder.SeedFaculty(db)
	seeder.SeedFacultyLocation(db)
	seeder.SeedAdmin(db)
	seeder.SeedAgent(db)
	// seeder.SeedDummy(db)
	// seeder.SeedThreadWithReplies(db)
	// seeder.SeedTrendingUserSnapshot(db)
	// seeder.SeedMutualFollows(db)
	// seeder.SeedTrendingActivity(db)
}

// resetLegacyFoxHistory drops the pre-session Fox history tables so AutoMigrate
// recreates them session-based. The old fox_histories had no session_id; if we
// detect that legacy shape we drop both fox tables (dev reset — existing Fox
// history is discarded). Runs once: after the first session-based migrate the
// session_id column exists and this is a no-op.
func resetLegacyFoxHistory(db *gorm.DB) {
	m := db.Migrator()
	if m.HasTable("fox_histories") && !m.HasColumn(&foxhistorymodel.FoxHistory{}, "session_id") {
		if err := db.Exec("DROP TABLE IF EXISTS fox_histories, fox_sessions CASCADE").Error; err != nil {
			log.Printf("reset legacy fox_histories failed: %v", err)
			return
		}
		log.Println("migration: reset legacy fox_histories → session-based fox_sessions + fox_histories")
	}
}

func createIndexes(db *gorm.DB) {
	stmts := []string{
		"CREATE INDEX IF NOT EXISTS idx_threads_feed_score ON threads(feed_score DESC)",
		"CREATE INDEX IF NOT EXISTS idx_threads_created_at ON threads(created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_threads_user_type ON threads(user_id, type)",
		"CREATE INDEX IF NOT EXISTS idx_threads_root_id_not_null ON threads(root_id) WHERE root_id IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_thread_media_thread_id ON thread_media(thread_id)",
		"CREATE INDEX IF NOT EXISTS idx_user_rel_follower ON user_relationships(follower_id)",
		"CREATE INDEX IF NOT EXISTS idx_user_block_blocker ON user_blocks(blocker_id)",
		"CREATE INDEX IF NOT EXISTS idx_user_block_blocked ON user_blocks(blocked_id)",
		"CREATE INDEX IF NOT EXISTS idx_notification_user_mute_user ON notification_user_mutes(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_notification_user_mute_muted ON notification_user_mutes(muted_user_id)",
		"CREATE INDEX IF NOT EXISTS idx_hashtags_name ON hashtags(name)",
		"CREATE INDEX IF NOT EXISTS idx_thread_hashtags_hashtag_id ON thread_hashtags(hashtag_id)",
		"CREATE INDEX IF NOT EXISTS idx_events_status ON events(status)",
		"CREATE INDEX IF NOT EXISTS idx_events_start_time ON events(start_time)",
		"CREATE INDEX IF NOT EXISTS idx_events_end_time ON events(end_time)",
		"CREATE INDEX IF NOT EXISTS idx_communities_category_id ON communities(category_id)",
		"CREATE INDEX IF NOT EXISTS idx_chats_reply_to_id ON chats(reply_to_id)",
		"CREATE INDEX IF NOT EXISTS idx_violations_name ON violations(name)",
		"CREATE INDEX IF NOT EXISTS idx_reports_status ON reports(status)",
		"CREATE INDEX IF NOT EXISTS idx_reports_reporter_id ON reports(reporter_id)",
		"CREATE INDEX IF NOT EXISTS idx_reports_reported_user_id ON reports(reported_user_id)",
		"CREATE INDEX IF NOT EXISTS idx_reports_reported_post_id ON reports(reported_post_id)",
		"CREATE INDEX IF NOT EXISTS idx_reports_violation_id ON reports(violation_id)",
		"CREATE INDEX IF NOT EXISTS idx_trend_reports_hashtag_id ON trend_reports(hashtag_id)",
		"CREATE INDEX IF NOT EXISTS idx_trend_reports_reason_id ON trend_reports(reason_id)",
		"CREATE INDEX IF NOT EXISTS idx_trend_reports_reporter_id ON trend_reports(reporter_id)",
		"DROP INDEX IF EXISTS uniq_trend_report_user_hashtag_reason",
		"CREATE UNIQUE INDEX IF NOT EXISTS uniq_trend_report_user_hashtag ON trend_reports(reporter_id, hashtag_id)",
		"CREATE INDEX IF NOT EXISTS idx_notifications_recipient_created ON notifications(recipient_user_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_threads_parent_id ON threads(parent_id) WHERE parent_id IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_threads_original_thread_id ON threads(original_thread_id) WHERE original_thread_id IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_threads_root_like ON threads(root_id, like_count DESC, created_at DESC) WHERE root_id IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_user_rel_follower_created ON user_relationships(follower_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_user_rel_following_created ON user_relationships(following_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_verification_codes_expires_at ON verification_codes(expires_at)",

		// Fox sessions: one summary session per (user, thread); chat sessions are
		// free. The partial unique index is the conflict arbiter for
		// UpsertSummarySession. The legacy fox_histories index is dropped (the
		// constraint moved to fox_sessions).
		"DROP INDEX IF EXISTS uniq_fox_history_user_thread_summary",
		"CREATE UNIQUE INDEX IF NOT EXISTS uniq_fox_session_user_thread_summary ON fox_sessions(user_id, thread_id) WHERE type = 'summary'",

		// For You source-query indexes. Each source's WHERE + ORDER BY pattern
		// gets one composite. The fresh-first boolean in ORDER BY cannot be
		// indexed directly (parameter-dependent), but the secondary sort
		// columns are covered so Postgres only has to bucket-sort the LIMIT 100
		// in memory after using the index for the filter.
		//
		// findForYouOwn: WHERE user_id = ? AND created_at > cutoff ORDER BY ..., feed_score DESC
		"CREATE INDEX IF NOT EXISTS idx_threads_user_created ON threads(user_id, created_at DESC)",

		// findForYouCommunity / findForYouFollowing: JOIN community_members and
		// user_relationships; the JOIN itself is keyed on community_id /
		// follower_id. Composite on (community_id, created_at DESC) plus the
		// existing follower_created index covers the JOIN's outer side.
		"CREATE INDEX IF NOT EXISTS idx_threads_community_created ON threads(community_id, created_at DESC) WHERE community_id IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_community_members_user_community ON community_members(user_id, community_id)",

		// findForYouGlobal: ORDER BY feed_score DESC, created_at DESC over the
		// 7-day window. Partial WHERE feed_score > 0 keeps the index small —
		// only rows with engagement are indexed. Cold rows (zero engagement)
		// drop out of the index automatically as their feed_score stays 0.
		// Old rows past 7d that retain a non-zero feed_score remain indexed
		// but the source-query WHERE clause filters them out.
		"CREATE INDEX IF NOT EXISTS idx_threads_score_created ON threads(feed_score DESC, created_at DESC) WHERE feed_score > 0",

		// CountNewSinceID: WHERE thread_id > since_id with visibility filters.
		// The PRIMARY KEY already covers thread_id ordering, but adding type +
		// created_at on top helps when since_id is small (large scan) and the
		// visibility predicates filter most of the table.
		"CREATE INDEX IF NOT EXISTS idx_threads_type_created ON threads(type, created_at DESC)",

		// Sort-covering composites for paginated GetMy* list endpoints. The
		// composite PKs already cover the WHERE user_id filter, but the
		// ORDER BY sort column is absent, forcing an in-memory sort after the
		// range scan. Adding the sort column lets Postgres read index-ordered
		// (no Sort node, skip-to-offset). Read-only optimization.
		"CREATE INDEX IF NOT EXISTS idx_likes_user_created ON likes(user_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_bookmarks_user_created ON bookmarks(user_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_reposts_user_created ON reposts(user_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_event_participants_user_joined ON event_participants(user_id, joined_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_user_event_pins_user_created ON user_event_pins(user_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_user_community_pins_user_created ON user_community_pins(user_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_event_faculty_locations_location ON event_faculty_locations(faculty_location_id)",
		"CREATE INDEX IF NOT EXISTS idx_user_profiles_faculty_location ON user_profiles(faculty_location_id) WHERE faculty_location_id IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_community_events_community_start ON community_events(community_id, start_time DESC)",
		"CREATE INDEX IF NOT EXISTS idx_community_event_invitees_user ON community_event_invitees(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_community_event_comments_event_created ON community_event_comments(community_event_id, created_at ASC)",
	}
	for _, s := range stmts {
		db.Exec(s)
	}
}

func addColumns(db *gorm.DB) {
	db.Exec(`ALTER TABLE events ADD COLUMN IF NOT EXISTS docs jsonb DEFAULT '[]'::jsonb`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS onboarding_completed boolean NOT NULL DEFAULT false`)
	db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS terms_accepted boolean NOT NULL DEFAULT false`)
}

func backfillShareCounters(db *gorm.DB) {
	db.Exec(`
		UPDATE events e SET
			share_count = COALESCE(s.total, 0),
			week_share_count = COALESCE(s.week, 0)
		FROM (
			SELECT te.event_id,
				SUM(t.share_count) AS total,
				SUM(CASE WHEN t.created_at > NOW() - INTERVAL '7 days' THEN t.share_count ELSE 0 END) AS week
			FROM thread_events te
			JOIN threads t ON t.thread_id = te.thread_id
			GROUP BY te.event_id
		) s
		WHERE e.id = s.event_id
	`)
	db.Exec(`
		UPDATE hashtags h SET
			share_count = COALESCE(s.total, 0),
			week_share_count = COALESCE(s.week, 0)
		FROM (
			SELECT th.hashtag_id,
				SUM(t.share_count) AS total,
				SUM(CASE WHEN t.created_at > NOW() - INTERVAL '7 days' THEN t.share_count ELSE 0 END) AS week
			FROM thread_hashtags th
			JOIN threads t ON t.thread_id = th.thread_id
			GROUP BY th.hashtag_id
		) s
		WHERE h.id = s.hashtag_id
	`)
}

func dropDeadColumns(db *gorm.DB) {
	if db.Migrator().HasColumn(&communityactionconfirmationmodel.CommunityActionConfirmation{}, "note") {
		if err := db.Migrator().DropColumn(&communityactionconfirmationmodel.CommunityActionConfirmation{}, "note"); err != nil {
			log.Printf("drop note column: %v", err)
		}
	}
}

func addCheckConstraints(db *gorm.DB) {
	db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'chk_user_profiles_gender'
			) THEN
				ALTER TABLE user_profiles ADD CONSTRAINT chk_user_profiles_gender
				CHECK (gender IN ('male', 'female', 'other', 'prefer_not_to_say'));
			END IF;
		END $$;
	`)
	db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'chk_user_settings_theme'
			) THEN
				ALTER TABLE user_settings ADD CONSTRAINT chk_user_settings_theme
				CHECK (user_theme IN ('black', 'white'));
			END IF;
		END $$;
	`)
}

func addForeignKeys(db *gorm.DB) {
	fks := []struct {
		name string
		stmt string
	}{
		{"fk_community_rules_community", "ALTER TABLE community_rules ADD CONSTRAINT fk_community_rules_community FOREIGN KEY (community_id) REFERENCES communities(community_id) ON DELETE CASCADE"},
		{"fk_communities_category", "ALTER TABLE communities ADD CONSTRAINT fk_communities_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL"},
		{"fk_event_reviews_event", "ALTER TABLE event_reviews ADD CONSTRAINT fk_event_reviews_event FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE"},
		{"fk_event_reviews_admin", "ALTER TABLE event_reviews ADD CONSTRAINT fk_event_reviews_admin FOREIGN KEY (admin_id) REFERENCES users(user_id)"},
		{"fk_thread_hashtags_thread", "ALTER TABLE thread_hashtags ADD CONSTRAINT fk_thread_hashtags_thread FOREIGN KEY (thread_id) REFERENCES threads(thread_id) ON DELETE CASCADE"},
		{"fk_thread_hashtags_hashtag", "ALTER TABLE thread_hashtags ADD CONSTRAINT fk_thread_hashtags_hashtag FOREIGN KEY (hashtag_id) REFERENCES hashtags(id) ON DELETE CASCADE"},
		{"fk_thread_events_thread", "ALTER TABLE thread_events ADD CONSTRAINT fk_thread_events_thread FOREIGN KEY (thread_id) REFERENCES threads(thread_id) ON DELETE CASCADE"},
		{"fk_thread_events_event", "ALTER TABLE thread_events ADD CONSTRAINT fk_thread_events_event FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE"},
		{"fk_thread_polls_thread", "ALTER TABLE thread_polls ADD CONSTRAINT fk_thread_polls_thread FOREIGN KEY (thread_id) REFERENCES threads(thread_id) ON DELETE CASCADE"},
		{"fk_thread_poll_options_poll", "ALTER TABLE thread_poll_options ADD CONSTRAINT fk_thread_poll_options_poll FOREIGN KEY (poll_id) REFERENCES thread_polls(poll_id) ON DELETE CASCADE"},
		{"fk_thread_poll_votes_poll", "ALTER TABLE thread_poll_votes ADD CONSTRAINT fk_thread_poll_votes_poll FOREIGN KEY (poll_id) REFERENCES thread_polls(poll_id) ON DELETE CASCADE"},
		{"fk_thread_poll_votes_option", "ALTER TABLE thread_poll_votes ADD CONSTRAINT fk_thread_poll_votes_option FOREIGN KEY (option_id) REFERENCES thread_poll_options(option_id) ON DELETE CASCADE"},
		{"fk_thread_poll_votes_user", "ALTER TABLE thread_poll_votes ADD CONSTRAINT fk_thread_poll_votes_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE"},
		{"fk_reports_reporter", "ALTER TABLE reports ADD CONSTRAINT fk_reports_reporter FOREIGN KEY (reporter_id) REFERENCES users(user_id) ON DELETE CASCADE"},
		{"fk_reports_reported_user", "ALTER TABLE reports ADD CONSTRAINT fk_reports_reported_user FOREIGN KEY (reported_user_id) REFERENCES users(user_id) ON DELETE SET NULL"},
		{"fk_reports_reported_post", "ALTER TABLE reports ADD CONSTRAINT fk_reports_reported_post FOREIGN KEY (reported_post_id) REFERENCES threads(thread_id) ON DELETE SET NULL"},
		{"fk_reports_violation", "ALTER TABLE reports ADD CONSTRAINT fk_reports_violation FOREIGN KEY (violation_id) REFERENCES violations(id) ON DELETE RESTRICT"},
		{"fk_trend_reports_reporter", "ALTER TABLE trend_reports ADD CONSTRAINT fk_trend_reports_reporter FOREIGN KEY (reporter_id) REFERENCES users(user_id) ON DELETE CASCADE"},
		{"fk_trend_reports_hashtag", "ALTER TABLE trend_reports ADD CONSTRAINT fk_trend_reports_hashtag FOREIGN KEY (hashtag_id) REFERENCES hashtags(id) ON DELETE CASCADE"},
		{"fk_trend_reports_reason", "ALTER TABLE trend_reports ADD CONSTRAINT fk_trend_reports_reason FOREIGN KEY (reason_id) REFERENCES trend_report_reasons(id) ON DELETE RESTRICT"},
		{"fk_user_profiles_faculty_location", "ALTER TABLE user_profiles ADD CONSTRAINT fk_user_profiles_faculty_location FOREIGN KEY (faculty_location_id) REFERENCES faculty_locations(id) ON DELETE SET NULL"},
		{"fk_event_faculty_locations_event", "ALTER TABLE event_faculty_locations ADD CONSTRAINT fk_event_faculty_locations_event FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE"},
		{"fk_event_faculty_locations_location", "ALTER TABLE event_faculty_locations ADD CONSTRAINT fk_event_faculty_locations_location FOREIGN KEY (faculty_location_id) REFERENCES faculty_locations(id) ON DELETE CASCADE"},
		{"fk_community_events_community", "ALTER TABLE community_events ADD CONSTRAINT fk_community_events_community FOREIGN KEY (community_id) REFERENCES communities(community_id) ON DELETE CASCADE"},
		{"fk_community_events_creator", "ALTER TABLE community_events ADD CONSTRAINT fk_community_events_creator FOREIGN KEY (created_by_user_id) REFERENCES users(user_id) ON DELETE CASCADE"},
		{"fk_community_event_invitees_event", "ALTER TABLE community_event_invitees ADD CONSTRAINT fk_community_event_invitees_event FOREIGN KEY (community_event_id) REFERENCES community_events(id) ON DELETE CASCADE"},
		{"fk_community_event_invitees_user", "ALTER TABLE community_event_invitees ADD CONSTRAINT fk_community_event_invitees_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE"},
		{"fk_community_event_comments_event", "ALTER TABLE community_event_comments ADD CONSTRAINT fk_community_event_comments_event FOREIGN KEY (community_event_id) REFERENCES community_events(id) ON DELETE CASCADE"},
		{"fk_community_event_comments_user", "ALTER TABLE community_event_comments ADD CONSTRAINT fk_community_event_comments_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE"},
		{"fk_community_event_reactions_event", "ALTER TABLE community_event_reactions ADD CONSTRAINT fk_community_event_reactions_event FOREIGN KEY (community_event_id) REFERENCES community_events(id) ON DELETE CASCADE"},
		{"fk_community_event_reactions_user", "ALTER TABLE community_event_reactions ADD CONSTRAINT fk_community_event_reactions_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE"},
	}
	for _, fk := range fks {
		db.Exec(`
			DO $$
			BEGIN
				IF NOT EXISTS (
					SELECT 1 FROM pg_constraint WHERE conname = '` + fk.name + `'
				) THEN
					` + fk.stmt + `;
				END IF;
			END $$;
		`)
	}
}

func seedRoles(db *gorm.DB) {
	defaultRoles := []string{"user", "lecturer", "admin"}
	for _, name := range defaultRoles {
		db.Where(rolemodel.Role{Name: name}).FirstOrCreate(&rolemodel.Role{}, rolemodel.Role{Name: name})
	}
}
