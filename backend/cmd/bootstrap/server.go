package bootstrap

import (
	"context"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"

	premiumsubscriptionsvciface "github.com/sunibgram/domain/service/premium_subscription"
	"github.com/sunibgram/infrastructure/cache"
	"github.com/sunibgram/infrastructure/scheduler"
	"github.com/sunibgram/webapi/router"
	webapisocket "github.com/sunibgram/webapi/socket"
)

func NewServer(ctrls *Controllers, svcs *Services, ext *External, c *cache.Cache) *fiber.App {
	app := fiber.New(fiber.Config{
		BodyLimit: 50 * 1024 * 1024,
	})
	app.Use(compress.New(compress.Config{Level: compress.LevelBestSpeed}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		AllowCredentials: true,
	}))
	webapisocket.Mount(app, ext.SocketHub, svcs.ChatRoom, ctrls.NotificationHandler, c)
	registerRoutes(app, ctrls)
	return app
}

func registerRoutes(app *fiber.App, c *Controllers) {
	router.Register(
		app,
		c.Auth,
		c.Role,
		c.User,
		c.Cache,
		c.UserSettings,
		c.UserProfile,
		c.Explore,
		c.Hashtag,
		c.Thread,
		c.Event,
		c.Like,
		c.Bookmark,
		c.ThreadPin,
		c.Repost,
		c.ThreadSummary,
		c.Email,
		c.Verification,
		c.Follow,
		c.Community,
		c.CommunityMember,
		c.CommunityRequest,
		c.CommunityThread,
		c.CommunityRule,
		c.CommunityThreadPin,
		c.CommunityPin,
		c.EventPin,
		c.EventParticipant,
		c.EventThreadPin,
		c.Category,
		c.Violation,
		c.Report,
		c.Faculty,
		c.Order,
		c.PremiumSubscription,
		c.DevSubscription,
		c.ChatRoom,
		c.Chat,
		c.Notification,
		c.ThreadPoll,
	)
}

func StartSubscriptionExpiry(svc premiumsubscriptionsvciface.IPremiumSubscriptionService) {
	go scheduler.RunSubscriptionExpiry(context.Background(), svc)
}

func Listen(app *fiber.App) error {
	addr := fmt.Sprintf(":%s", getEnv("PORT", "8080"))
	fmt.Printf("server running on %s\n", addr)
	return app.Listen(addr)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
