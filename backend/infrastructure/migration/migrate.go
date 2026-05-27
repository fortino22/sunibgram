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
	communitymembermodel "github.com/sunibgram/domain/model/community_member"
	communitypinmodel "github.com/sunibgram/domain/model/community_pin"
	communityrequestmodel "github.com/sunibgram/domain/model/community_request"
	communityrulemodel "github.com/sunibgram/domain/model/community_rule"
	communitythreadpinmodel "github.com/sunibgram/domain/model/community_thread_pin"
	eventparticipantmodel "github.com/sunibgram/domain/model/event_participant"
	eventpinmodel "github.com/sunibgram/domain/model/event_pin"
	eventreviewmodel "github.com/sunibgram/domain/model/event_review"
	eventthreadpinmodel "github.com/sunibgram/domain/model/event_thread_pin"
	facultymodel "github.com/sunibgram/domain/model/faculty"
	feedbackmodel "github.com/sunibgram/domain/model/feedback"
	hashtagmodel "github.com/sunibgram/domain/model/hashtag"
	likemodel "github.com/sunibgram/domain/model/like"
	notificationmodel "github.com/sunibgram/domain/model/notification"
	ordermodel "github.com/sunibgram/domain/model/order"
	premiumsubscriptionmodel "github.com/sunibgram/domain/model/premium_subscription"
	reportmodel "github.com/sunibgram/domain/model/report"
	rolemodel "github.com/sunibgram/domain/model/role"
	threadmodel "github.com/sunibgram/domain/model/thread"
	threadmediamodel "github.com/sunibgram/domain/model/thread_media"
	threadpinmodel "github.com/sunibgram/domain/model/thread_pin"
	threadpollmodel "github.com/sunibgram/domain/model/thread_poll"
	usermodel "github.com/sunibgram/domain/model/user"
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

	models := []any{
		&rolemodel.Role{},
		&usermodel.User{},
		&usersettingsmodel.UserSettings{},
		&userprofilemodel.UserProfile{},
		&categorymodel.Category{},
		&violationmodel.Violation{},
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
		&bookmarkmodel.Bookmark{},
		&bookmarkcollectionmodel.BookmarkCollection{},
		&bookmarkcollectionmodel.BookmarkCollectionItem{},
		&threadpinmodel.ThreadPin{},
		&verificationcodemodel.VerificationCode{},
		&userrelationshipmodel.UserRelationship{},
		&userstatmodel.UserStat{},
		&communitymembermodel.CommunityMember{},
		&communityrequestmodel.CommunityRequest{},
		&communityrulemodel.CommunityRule{},
		&communitythreadpinmodel.CommunityThreadPin{},
		&communitypinmodel.UserCommunityPin{},
		&eventpinmodel.UserEventPin{},
		&eventparticipantmodel.EventParticipant{},
		&facultymodel.Faculty{},
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
		&feedbackmodel.Feedback{},
		&feedbackmodel.FeedbackVote{},
		&premiumsubscriptionmodel.PremiumSubscription{},
	}

	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			log.Fatalf("AutoMigrate %T: %v", m, err)
		}
	}

	createIndexes(db)
	addCheckConstraints(db)
	addColumns(db)
	addForeignKeys(db)
	migrateChatEncryption(db)

	seedRoles(db)
	seeder.SeedCategory(db)
	seeder.SeedViolation(db)
	seeder.SeedFaculty(db)
	seeder.SeedAdmin(db)
	seeder.SeedDummy(db)
	seeder.SeedThreadWithReplies(db)
}

func createIndexes(db *gorm.DB) {
	stmts := []string{
		"CREATE INDEX IF NOT EXISTS idx_threads_feed_score ON threads(feed_score DESC)",
		"CREATE INDEX IF NOT EXISTS idx_threads_created_at ON threads(created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_threads_user_type ON threads(user_id, type)",
		"CREATE INDEX IF NOT EXISTS idx_threads_root_id_not_null ON threads(root_id) WHERE root_id IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_thread_media_thread_id ON thread_media(thread_id)",
		"CREATE INDEX IF NOT EXISTS idx_user_rel_follower ON user_relationships(follower_id)",
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
		"CREATE INDEX IF NOT EXISTS idx_notifications_recipient_created ON notifications(recipient_user_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_threads_parent_id ON threads(parent_id) WHERE parent_id IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_threads_original_thread_id ON threads(original_thread_id) WHERE original_thread_id IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_threads_root_like ON threads(root_id, like_count DESC, created_at DESC) WHERE root_id IS NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_user_rel_follower_created ON user_relationships(follower_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_user_rel_following_created ON user_relationships(following_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_verification_codes_expires_at ON verification_codes(expires_at)",

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
	}
	for _, s := range stmts {
		db.Exec(s)
	}
}

func addColumns(db *gorm.DB) {
	db.Exec(`ALTER TABLE events ADD COLUMN IF NOT EXISTS pdfs jsonb DEFAULT '[]'::jsonb`)
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
