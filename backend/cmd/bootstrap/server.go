package bootstrap

import (
	"context"
	"fmt"
	"os"
	"strings"

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
		BodyLimit:               50 * 1024 * 1024,
		EnableTrustedProxyCheck: true,
		TrustedProxies:          trustedProxies(),
		ProxyHeader:             fiber.HeaderXForwardedFor,
	})
	app.Use(compress.New(compress.Config{Level: compress.LevelBestSpeed}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		AllowCredentials: true,
	}))
	webapisocket.Mount(app, ext.SocketHub, svcs.ChatRoom, ctrls.NotificationHandler, c)
	registerRoutes(app, ctrls, c)
	return app
}

func registerRoutes(app *fiber.App, ctrls *Controllers, c *cache.Cache) {
	router.Register(
		app,
		ctrls.Auth,
		ctrls.Role,
		ctrls.User,
		ctrls.Cache,
		ctrls.UserSettings,
		ctrls.UserProfile,
		ctrls.Explore,
		ctrls.Hashtag,
		ctrls.Thread,
		ctrls.Event,
		ctrls.Like,
		ctrls.Bookmark,
		ctrls.ThreadPin,
		ctrls.Repost,
		ctrls.ThreadSummary,
		ctrls.Email,
		ctrls.Verification,
		ctrls.Follow,
		ctrls.Community,
		ctrls.CommunityMember,
		ctrls.CommunityRequest,
		ctrls.CommunityThread,
		ctrls.CommunityRule,
		ctrls.CommunityThreadPin,
		ctrls.CommunityPin,
		ctrls.EventPin,
		ctrls.EventParticipant,
		ctrls.EventThreadPin,
		ctrls.Category,
		ctrls.Violation,
		ctrls.Report,
		ctrls.Faculty,
		ctrls.Order,
		ctrls.PremiumSubscription,
		ctrls.DevSubscription,
		ctrls.ChatRoom,
		ctrls.Chat,
		ctrls.Notification,
		ctrls.ThreadPoll,
		c,
	)
}

func StartSubscriptionExpiry(svc premiumsubscriptionsvciface.IPremiumSubscriptionService) {
	go scheduler.RunSubscriptionExpiry(context.Background(), svc)
}

func Listen(app *fiber.App) error {
	addr := fmt.Sprintf(":%s", getEnv("PORT", ""))
	fmt.Printf("server running on %s\n", addr)
	return app.Listen(addr)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func trustedProxies() []string {
	raw := strings.TrimSpace(os.Getenv("TRUSTED_PROXIES"))

	out := make([]string, 0, 4)
	for _, p := range strings.Split(raw, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
