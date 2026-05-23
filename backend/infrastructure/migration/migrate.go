package migration

import (
	"log"

	"gorm.io/gorm"

	bookmarkmodel "github.com/sunibgram/domain/model/bookmark"
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
		&threadpinmodel.ThreadPin{},
		&verificationcodemodel.VerificationCode{},
		&userrelationshipmodel.UserRelationship{},
		&userstatmodel.UserStat{},
		&communitymembermodel.CommunityMember{},
		&communityrequestmodel.CommunityRequest{},
		&communityrulemodel.CommunityRule{},
		&communitythreadpinmodel.CommunityThreadPin{},
		&eventthreadpinmodel.EventThreadPin{},
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
	}

	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			log.Fatalf("AutoMigrate %T: %v", m, err)
		}
	}

	db.Exec("CREATE INDEX IF NOT EXISTS idx_threads_feed_score ON threads(feed_score DESC) WHERE created_at > NOW() - INTERVAL '7 days'")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_user_rel_follower ON user_relationships(follower_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_threads_created_at ON threads(created_at DESC)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_threads_user_type ON threads(user_id, type)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_threads_root_id_not_null ON threads(root_id) WHERE root_id IS NOT NULL")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_thread_media_thread_id ON thread_media(thread_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_hashtags_name ON hashtags(name)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_thread_hashtags_hashtag_id ON thread_hashtags(hashtag_id)")
	db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'fk_community_rules_community'
			) THEN
				ALTER TABLE community_rules ADD CONSTRAINT fk_community_rules_community FOREIGN KEY (community_id) REFERENCES communities(community_id) ON DELETE CASCADE;
			END IF;
		END $$;
	`)
	db.Exec("ALTER TABLE communities DROP CONSTRAINT IF EXISTS fk_communities_" + "gen" + "re")
	db.Exec("ALTER TABLE communities DROP COLUMN IF EXISTS " + "gen" + "re_id")
	db.Exec("ALTER TABLE communities DROP COLUMN IF EXISTS category")
	db.Exec("DROP TABLE IF EXISTS " + "gen" + "res")
	db.Exec("ALTER TABLE communities ADD COLUMN IF NOT EXISTS category_id UUID")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_communities_category_id ON communities(category_id)")
	db.Exec("ALTER TABLE communities DROP COLUMN IF EXISTS avatar_object_name")
	db.Exec("ALTER TABLE communities DROP COLUMN IF EXISTS avatar_bucket")
	db.Exec("ALTER TABLE communities DROP COLUMN IF EXISTS rules")
	db.Exec("ALTER TABLE user_stats DROP COLUMN IF EXISTS thread_count")
	db.Exec("ALTER TABLE communities DROP COLUMN IF EXISTS thread_count")
	db.Exec("ALTER TABLE hashtags ADD COLUMN IF NOT EXISTS thread_count INTEGER NOT NULL DEFAULT 0")
	db.Exec("ALTER TABLE hashtags ADD COLUMN IF NOT EXISTS like_count INTEGER NOT NULL DEFAULT 0")
	db.Exec("ALTER TABLE hashtags ADD COLUMN IF NOT EXISTS repost_count INTEGER NOT NULL DEFAULT 0")
	db.Exec("ALTER TABLE hashtags ADD COLUMN IF NOT EXISTS week_thread_count INTEGER NOT NULL DEFAULT 0")
	db.Exec("ALTER TABLE hashtags ADD COLUMN IF NOT EXISTS week_like_count INTEGER NOT NULL DEFAULT 0")
	db.Exec("ALTER TABLE hashtags ADD COLUMN IF NOT EXISTS week_repost_count INTEGER NOT NULL DEFAULT 0")
	db.Exec("ALTER TABLE events ADD COLUMN IF NOT EXISTS banner_object_name VARCHAR(500)")
	db.Exec("ALTER TABLE events ADD COLUMN IF NOT EXISTS banner_bucket VARCHAR(100)")
	db.Exec("ALTER TABLE events ADD COLUMN IF NOT EXISTS user_id INTEGER NOT NULL DEFAULT 0")
	db.Exec("ALTER TABLE events ADD COLUMN IF NOT EXISTS type VARCHAR(20) NOT NULL DEFAULT 'unofficial'")
	db.Exec("ALTER TABLE events ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'pending'")
	db.Exec("ALTER TABLE events ALTER COLUMN status SET DEFAULT 'pending'")
	db.Exec("ALTER TABLE events ADD COLUMN IF NOT EXISTS attachment_object_name VARCHAR(500)")
	db.Exec("ALTER TABLE events ADD COLUMN IF NOT EXISTS attachment_bucket VARCHAR(100)")
	db.Exec("ALTER TABLE events ADD COLUMN IF NOT EXISTS pdfs JSONB NOT NULL DEFAULT '[]'::jsonb")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_events_status ON events(status)")
	db.Exec("ALTER TABLE events ADD COLUMN IF NOT EXISTS participant_count INTEGER NOT NULL DEFAULT 0")
	db.Exec("ALTER TABLE events ADD COLUMN IF NOT EXISTS thread_count INTEGER NOT NULL DEFAULT 0")
	db.Exec("ALTER TABLE events ADD COLUMN IF NOT EXISTS like_count INTEGER NOT NULL DEFAULT 0")
	db.Exec("ALTER TABLE events ADD COLUMN IF NOT EXISTS repost_count INTEGER NOT NULL DEFAULT 0")
	db.Exec("ALTER TABLE events ADD COLUMN IF NOT EXISTS week_thread_count INTEGER NOT NULL DEFAULT 0")
	db.Exec("ALTER TABLE events ADD COLUMN IF NOT EXISTS week_like_count INTEGER NOT NULL DEFAULT 0")
	db.Exec("ALTER TABLE events ADD COLUMN IF NOT EXISTS week_repost_count INTEGER NOT NULL DEFAULT 0")
	db.Exec("ALTER TABLE events DROP COLUMN IF EXISTS view_count")
	db.Exec("ALTER TABLE events DROP COLUMN IF EXISTS trending_score")
	migratePremiumSubscriptions(db)
	migrateUserProfiles(db)
	migrateViolations(db)
	migrateReportReporter(db)
	migrateReports(db)
	db.Exec("CREATE INDEX IF NOT EXISTS idx_events_start_time ON events(start_time)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_events_end_time ON events(end_time)")
	db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'fk_communities_category'
			) THEN
				ALTER TABLE communities ADD CONSTRAINT fk_communities_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;
			END IF;
		END $$;
	`)
	db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'fk_event_reviews_event'
			) THEN
				ALTER TABLE event_reviews ADD CONSTRAINT fk_event_reviews_event FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE;
			END IF;
		END $$;
	`)
	db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'fk_event_reviews_admin'
			) THEN
				ALTER TABLE event_reviews ADD CONSTRAINT fk_event_reviews_admin FOREIGN KEY (admin_id) REFERENCES users(user_id);
			END IF;
		END $$;
	`)
	db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'fk_thread_hashtags_thread'
			) THEN
				ALTER TABLE thread_hashtags ADD CONSTRAINT fk_thread_hashtags_thread FOREIGN KEY (thread_id) REFERENCES threads(thread_id) ON DELETE CASCADE;
			END IF;
		END $$;
	`)
	db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'fk_thread_hashtags_hashtag'
			) THEN
				ALTER TABLE thread_hashtags ADD CONSTRAINT fk_thread_hashtags_hashtag FOREIGN KEY (hashtag_id) REFERENCES hashtags(id) ON DELETE CASCADE;
			END IF;
		END $$;
	`)
	db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'fk_thread_events_thread'
			) THEN
				ALTER TABLE thread_events ADD CONSTRAINT fk_thread_events_thread FOREIGN KEY (thread_id) REFERENCES threads(thread_id) ON DELETE CASCADE;
			END IF;
		END $$;
	`)
	db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'fk_thread_events_event'
			) THEN
				ALTER TABLE thread_events ADD CONSTRAINT fk_thread_events_event FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE;
			END IF;
		END $$;
	`)

	db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'fk_thread_polls_thread'
			) THEN
				ALTER TABLE thread_polls ADD CONSTRAINT fk_thread_polls_thread FOREIGN KEY (thread_id) REFERENCES threads(thread_id) ON DELETE CASCADE;
			END IF;
		END $$;
	`)
	db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'fk_thread_poll_options_poll'
			) THEN
				ALTER TABLE thread_poll_options ADD CONSTRAINT fk_thread_poll_options_poll FOREIGN KEY (poll_id) REFERENCES thread_polls(poll_id) ON DELETE CASCADE;
			END IF;
		END $$;
	`)
	db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'fk_thread_poll_votes_poll'
			) THEN
				ALTER TABLE thread_poll_votes ADD CONSTRAINT fk_thread_poll_votes_poll FOREIGN KEY (poll_id) REFERENCES thread_polls(poll_id) ON DELETE CASCADE;
			END IF;
		END $$;
	`)
	db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'fk_thread_poll_votes_option'
			) THEN
				ALTER TABLE thread_poll_votes ADD CONSTRAINT fk_thread_poll_votes_option FOREIGN KEY (option_id) REFERENCES thread_poll_options(option_id) ON DELETE CASCADE;
			END IF;
		END $$;
	`)
	db.Exec(`
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'fk_thread_poll_votes_user'
			) THEN
				ALTER TABLE thread_poll_votes ADD CONSTRAINT fk_thread_poll_votes_user FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE;
			END IF;
		END $$;
	`)

	db.Exec("ALTER TABLE chat_rooms ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(500)")
	migrateChatEncryption(db)
	db.Exec("ALTER TABLE chat_room_members ADD COLUMN IF NOT EXISTS is_muted BOOLEAN NOT NULL DEFAULT false")
	db.Exec("ALTER TABLE chats ADD COLUMN IF NOT EXISTS reply_to_id INTEGER REFERENCES chats(chat_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_chats_reply_to_id ON chats(reply_to_id)")
	db.Exec("UPDATE chat_rooms SET member_count = (SELECT COUNT(*) FROM chat_room_members WHERE chat_room_members.room_id = chat_rooms.room_id)")
	db.Exec(`
		UPDATE hashtags
		SET thread_count = counts.thread_count
		FROM (
			SELECT hashtag_id, COUNT(*)::integer AS thread_count
			FROM thread_hashtags
			GROUP BY hashtag_id
		) AS counts
		WHERE hashtags.id = counts.hashtag_id
	`)
	db.Exec(`
		UPDATE events
		SET participant_count = (
			SELECT COUNT(DISTINCT threads.user_id)::integer
			FROM thread_events
			JOIN threads ON threads.thread_id = thread_events.thread_id
			WHERE thread_events.event_id = events.id AND threads.type <> 'REPOST'
		)
	`)

	seedRoles(db)
	seeder.SeedCategory(db)
	seeder.SeedViolation(db)
	seeder.SeedFaculty(db)
	seeder.SeedAdmin(db)
	seeder.SeedDummy(db)
	seeder.SeedThreadWithReplies(db)
}

func seedRoles(db *gorm.DB) {
	defaultRoles := []string{"user", "lecturer", "admin"}
	for _, name := range defaultRoles {
		db.Where(rolemodel.Role{Name: name}).FirstOrCreate(&rolemodel.Role{}, rolemodel.Role{Name: name})
	}
}
