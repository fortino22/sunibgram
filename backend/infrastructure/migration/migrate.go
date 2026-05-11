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

	db.Exec("UPDATE chat_rooms SET member_count = (SELECT COUNT(*) FROM chat_room_members WHERE chat_room_members.room_id = chat_rooms.room_id)")

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
