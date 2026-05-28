package seeder

import (
	"log"
	"sort"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	trendingusersnapshotmodel "github.com/sunibgram/domain/model/trending_user_snapshot"
	usermodel "github.com/sunibgram/domain/model/user"
)

// SeedTrendingUserSnapshot inserts a few weeks of historical trending snapshots
// per existing user. Dev-only convenience so the "Your Stats" chart has data
// before the weekly scheduler has run for real.
//
// Upserts on (week_start_date, user_id), so re-running is a no-op once rows
// exist. Skips entirely if any snapshot rows are already present (preserves
// real history written by the scheduler).
func SeedTrendingUserSnapshot(db *gorm.DB) {
	var existing int64
	if err := db.Model(&trendingusersnapshotmodel.TrendingUserSnapshot{}).Count(&existing).Error; err != nil {
		log.Printf("trending snapshot seeder: count failed: %v", err)
		return
	}
	if existing > 0 {
		return
	}

	var users []usermodel.User
	if err := db.Select("user_id").Order("user_id ASC").Find(&users).Error; err != nil {
		log.Printf("trending snapshot seeder: list users failed: %v", err)
		return
	}
	if len(users) == 0 {
		return
	}

	// Compose 4 weeks of history: weekStart_n = (last Monday) - n*7 days, for n=4..1.
	now := time.Now().UTC()
	daysSinceMonday := (int(now.Weekday()) - int(time.Monday) + 7) % 7
	lastMonday := time.Date(now.Year(), now.Month(), now.Day()-daysSinceMonday, 0, 0, 0, 0, time.UTC)

	const totalWeeks = 30
	rows := make([]trendingusersnapshotmodel.TrendingUserSnapshot, 0, len(users)*totalWeeks)
	snapshotAt := now

	// Wave pattern across the 30 weeks (oldest → newest, x = 0..29):
	//   x in [0, 10)  → up   (60 → 200)
	//   x in [10, 20) → down (200 → 60)
	//   x in [20, 30) → up   (60 → 230)
	waveScore := func(xPos int) float64 {
		switch {
		case xPos < 10:
			return 60.0 + float64(xPos)*14.0
		case xPos < 20:
			return 200.0 - float64(xPos-10)*14.0
		default:
			return 60.0 + float64(xPos-20)*17.0
		}
	}

	for weeksAgo := totalWeeks; weeksAgo >= 1; weeksAgo-- {
		weekStart := lastMonday.AddDate(0, 0, -7*weeksAgo)
		xPos := totalWeeks - weeksAgo // 0 = oldest, 29 = newest

		// Per-week rows for all users, then sort by score DESC to assign rank.
		weekRows := make([]trendingusersnapshotmodel.TrendingUserSnapshot, 0, len(users))
		for _, u := range users {
			// Per-user offset so lines vary but follow the same wave shape.
			userOffset := float64((int(u.UserID)*7)%30) - 15.0
			score := waveScore(xPos) + userOffset

			likesIn := (int(u.UserID)+weeksAgo)%9 + 1
			commentsIn := (int(u.UserID)*3+weeksAgo)%7 + 1
			bookmarksIn := (int(u.UserID)+weeksAgo*2)%5
			repostsIn := (int(u.UserID)+weeksAgo)%4
			threadCount := (int(u.UserID)+weeksAgo*2)%6 + 1
			likesOut := (int(u.UserID)*2+weeksAgo)%8
			commentsOut := (int(u.UserID)+weeksAgo*3)%6
			bookmarksOut := (int(u.UserID)+weeksAgo)%4
			postsOut := (int(u.UserID)+weeksAgo)%5 + 1

			weekRows = append(weekRows, trendingusersnapshotmodel.TrendingUserSnapshot{
				WeekStartDate: weekStart,
				UserID:        u.UserID,
				Score:         score,
				LikesIn:       likesIn,
				CommentsIn:    commentsIn,
				BookmarksIn:   bookmarksIn,
				RepostsIn:     repostsIn,
				ThreadCount:   threadCount,
				LikesOut:      likesOut,
				CommentsOut:   commentsOut,
				BookmarksOut:  bookmarksOut,
				PostsOut:      postsOut,
				SnapshotAt:    snapshotAt,
			})
		}
		sort.SliceStable(weekRows, func(i, j int) bool {
			return weekRows[i].Score > weekRows[j].Score
		})
		for i := range weekRows {
			weekRows[i].Rank = i + 1
		}
		rows = append(rows, weekRows...)
	}

	if err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "week_start_date"}, {Name: "user_id"}},
		DoNothing: true,
	}).CreateInBatches(rows, 200).Error; err != nil {
		log.Printf("trending snapshot seeder: insert failed: %v", err)
		return
	}
	log.Printf("trending snapshot seeder: seeded %d rows across %d weeks for %d users",
		len(rows), totalWeeks, len(users))
}
