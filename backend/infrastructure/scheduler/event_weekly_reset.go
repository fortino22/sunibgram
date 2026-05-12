package scheduler

import (
	"log"
	"time"

	"gorm.io/gorm"
)

// ResetWeeklyEventCounters resets week_* counter columns every Monday 00:00 WIB (UTC+7).
func ResetWeeklyEventCounters(db *gorm.DB) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	for {
		now := time.Now().In(loc)
		daysUntilMonday := (time.Monday - now.Weekday() + 7) % 7
		if daysUntilMonday == 0 {
			daysUntilMonday = 7
		}
		next := time.Date(now.Year(), now.Month(), now.Day()+int(daysUntilMonday), 0, 0, 0, 0, loc)
		time.Sleep(time.Until(next))
		if err := db.Exec("UPDATE events SET week_thread_count = 0, week_like_count = 0, week_repost_count = 0").Error; err != nil {
			log.Printf("weekly event counter reset failed: %v", err)
		} else {
			log.Println("weekly event counters reset")
		}
	}
}
