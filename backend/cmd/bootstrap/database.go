package bootstrap

import (
	"context"
	"log"

	"gorm.io/gorm"

	"github.com/sunibgram/infrastructure/cache"
	"github.com/sunibgram/infrastructure/config"
	"github.com/sunibgram/infrastructure/migration"
	"github.com/sunibgram/infrastructure/scheduler"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	"github.com/sunibgram/webapi/middleware"
	webapisocket "github.com/sunibgram/webapi/socket"
)

type Database struct {
	Write *gorm.DB
	Read  *gorm.DB
}

func NewDatabase() *Database {
	write := config.NewPostgresWrite()
	read := config.NewPostgresRead()
	migration.AutoMigrate(write)
	return &Database{Write: write, Read: read}
}

func NewCache() *cache.Cache {
	c := cache.New()
	if err := c.DeleteByPattern(context.Background(), "community:members:*"); err != nil {
		log.Printf("community members cache startup flush error: %v", err)
	}
	middleware.Init(c)
	webapisocket.InitRateLimiter(c)
	ctrlshared.SetFeedBuilderCache(c)
	return c
}

func StartWeeklySchedulers(db *gorm.DB) {
	ctx := context.Background()
	go scheduler.ResetWeeklyEventCounters(ctx, db)
	go scheduler.ResetWeeklyHashtagCounters(ctx, db)
}

func StartFeedScoreScheduler(db *gorm.DB, c *cache.Cache) {
	scheduler.NewFeedScoreScheduler(db, c).Start(context.Background())
}
