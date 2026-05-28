package seeder

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	bookmarkmodel "github.com/sunibgram/domain/model/bookmark"
	likemodel "github.com/sunibgram/domain/model/like"
	repostmodel "github.com/sunibgram/domain/model/repost"
	threadmodel "github.com/sunibgram/domain/model/thread"
	usermodel "github.com/sunibgram/domain/model/user"
)

// SeedTrendingActivity makes the live "/trending-users" leaderboard non-empty
// in dev: every non-admin user gets one fresh POST in the current weekly window
// plus a fan-out of likes/bookmarks/reposts/replies from peers. Idempotent:
// skips if any thread already exists with marker '[trending-seed]'.
func SeedTrendingActivity(db *gorm.DB) {
	const marker = "[trending-seed]"

	var existing int64
	if err := db.Model(&threadmodel.Thread{}).
		Where("content LIKE ?", "%"+marker+"%").
		Count(&existing).Error; err != nil {
		log.Printf("trending activity seeder: count failed: %v", err)
		return
	}
	if existing > 0 {
		return
	}

	var users []usermodel.User
	if err := db.
		Joins("JOIN roles ON roles.role_id = users.role_id").
		Where("roles.name <> ?", "admin").
		Select("users.user_id, users.user_name").
		Order("users.user_id ASC").
		Find(&users).Error; err != nil {
		log.Printf("trending activity seeder: list users failed: %v", err)
		return
	}
	n := len(users)
	if n < 2 {
		return
	}

	now := time.Now().UTC()

	// One root post per user, staggered hours in the past week so feed_score
	// decay differs across users.
	roots := make([]threadmodel.Thread, 0, n)
	for i, u := range users {
		createdAt := now.Add(-time.Duration(i%72) * time.Hour)
		roots = append(roots, threadmodel.Thread{
			UserID:    u.UserID,
			Type:      threadmodel.ThreadTypePost,
			Content:   fmt.Sprintf("%s seeded engagement post from @%s", marker, u.UserName),
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		})
	}
	if err := db.CreateInBatches(&roots, 100).Error; err != nil {
		log.Printf("trending activity seeder: thread insert failed: %v", err)
		return
	}

	likes := make([]likemodel.Like, 0, n*8)
	bookmarks := make([]bookmarkmodel.Bookmark, 0, n*5)
	reposts := make([]repostmodel.Repost, 0, n*3)
	replies := make([]threadmodel.Thread, 0, n*4)

	for i, root := range roots {
		// Per-thread fan-out sizes vary to spread out scores.
		nLikes := 3 + (i*7)%9
		nBookmarks := 1 + (i*3)%5
		nReposts := (i * 5) % 4
		nReplies := 1 + (i*11)%4

		for k := 1; k <= nLikes; k++ {
			j := (i + k) % n
			if users[j].UserID == root.UserID {
				continue
			}
			likes = append(likes, likemodel.Like{
				UserID:    users[j].UserID,
				ThreadID:  root.ThreadID,
				CreatedAt: now.Add(-time.Duration((i+k)%96) * time.Hour),
			})
		}
		for k := 1; k <= nBookmarks; k++ {
			j := (i + k*2) % n
			if users[j].UserID == root.UserID {
				continue
			}
			bookmarks = append(bookmarks, bookmarkmodel.Bookmark{
				UserID:    users[j].UserID,
				ThreadID:  root.ThreadID,
				CreatedAt: now.Add(-time.Duration((i*2+k)%96) * time.Hour),
			})
		}
		for k := 1; k <= nReposts; k++ {
			j := (i + k*3) % n
			if users[j].UserID == root.UserID {
				continue
			}
			reposts = append(reposts, repostmodel.Repost{
				UserID:    users[j].UserID,
				ThreadID:  root.ThreadID,
				CreatedAt: now.Add(-time.Duration((i*3+k)%96) * time.Hour),
			})
		}
		rootID := root.ThreadID
		for k := 1; k <= nReplies; k++ {
			j := (i + k*4) % n
			if users[j].UserID == root.UserID {
				continue
			}
			replyAt := now.Add(-time.Duration((i*4+k)%96) * time.Hour)
			replies = append(replies, threadmodel.Thread{
				UserID:    users[j].UserID,
				Type:      threadmodel.ThreadTypeReply,
				Content:   fmt.Sprintf("%s reply %d", marker, k),
				ParentID:  &rootID,
				RootID:    &rootID,
				CreatedAt: replyAt,
				UpdatedAt: replyAt,
			})
		}
	}

	if len(likes) > 0 {
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).
			CreateInBatches(&likes, 500).Error; err != nil {
			log.Printf("trending activity seeder: likes insert failed: %v", err)
		}
	}
	if len(bookmarks) > 0 {
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).
			CreateInBatches(&bookmarks, 500).Error; err != nil {
			log.Printf("trending activity seeder: bookmarks insert failed: %v", err)
		}
	}
	if len(reposts) > 0 {
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).
			CreateInBatches(&reposts, 500).Error; err != nil {
			log.Printf("trending activity seeder: reposts insert failed: %v", err)
		}
	}
	if len(replies) > 0 {
		if err := db.CreateInBatches(&replies, 500).Error; err != nil {
			log.Printf("trending activity seeder: replies insert failed: %v", err)
		}
	}

	// Recompute thread counters + feed_score for the threads we just touched.
	if err := db.Exec(`
		UPDATE threads t SET
			like_count     = (SELECT COUNT(*) FROM likes      WHERE thread_id = t.thread_id),
			bookmark_count = (SELECT COUNT(*) FROM bookmarks  WHERE thread_id = t.thread_id),
			repost_count   = (SELECT COUNT(*) FROM reposts    WHERE thread_id = t.thread_id),
			comment_count  = (SELECT COUNT(*) FROM threads r  WHERE r.parent_id = t.thread_id AND r.type = 'REPLY')
		WHERE t.content LIKE ?
	`, "%"+marker+"%").Error; err != nil {
		log.Printf("trending activity seeder: counter update failed: %v", err)
	}
	if err := db.Exec(`
		UPDATE threads SET feed_score = (
			like_count::float +
			3.0 * comment_count::float +
			5.0 * repost_count::float +
			4.0 * bookmark_count::float
		) / (
			1.0 + POWER(EXTRACT(EPOCH FROM (NOW() - created_at)) / 3600.0, 1.5)
		)
		WHERE content LIKE ?
	`, "%"+marker+"%").Error; err != nil {
		log.Printf("trending activity seeder: feed_score recalc failed: %v", err)
	}

	log.Printf("trending activity seeder: %d posts, %d likes, %d bookmarks, %d reposts, %d replies across %d users",
		len(roots), len(likes), len(bookmarks), len(reposts), len(replies), n)
}
