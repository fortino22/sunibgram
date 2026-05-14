package migration

import (
	"log"

	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/infrastructure/seeder"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) {
	db.DisableForeignKeyConstraintWhenMigrating = true

	models := []any{
		&model.Role{},
		&model.User{},
		&model.UserSettings{},
		&model.UserProfile{},
		&model.Community{},
		&model.Thread{},
		&model.Event{},
		&model.EventReview{},
		&model.ThreadEvent{},
		&model.ThreadMedia{},
		&model.Like{},
		&model.Bookmark{},
		&model.ThreadPin{},
		&model.VerificationCode{},
		&model.UserRelationship{},
		&model.UserStat{},
		&model.CommunityMember{},
		&model.CommunityRequest{},
		&model.Faculty{},
		&model.UserFaculty{},
		&model.Order{},
		&model.ChatRoom{},
		&model.ChatRoomMember{},
		&model.Chat{},
		&model.ChatMedia{},
		&model.ChatRead{},
		&model.Notification{},
		&model.Hashtag{},
		&model.ThreadHashtag{},
	}

	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			log.Fatalf("AutoMigrate %T: %v", m, err)
		}
	}

	db.Exec("CREATE INDEX IF NOT EXISTS idx_user_rel_follower ON user_relationships(follower_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_threads_created_at ON threads(created_at DESC)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_threads_user_type ON threads(user_id, type)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_thread_media_thread_id ON thread_media(thread_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_hashtags_name ON hashtags(name)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_thread_hashtags_hashtag_id ON thread_hashtags(hashtag_id)")
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
	db.Exec("CREATE INDEX IF NOT EXISTS idx_events_start_time ON events(start_time)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_events_end_time ON events(end_time)")
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
	seeder.SeedFaculty(db)
	seeder.SeedAdmin(db)
	seeder.SeedDummy(db)
}

func seedRoles(db *gorm.DB) {
	defaultRoles := []string{"user", "lecturer", "admin"}
	for _, name := range defaultRoles {
		db.Where(model.Role{Name: name}).FirstOrCreate(&model.Role{}, model.Role{Name: name})
	}
}
