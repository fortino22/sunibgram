package bootstrap

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/helmet/v2"

	premiumsubscriptionsvciface "github.com/sunibgram/domain/service/premium_subscription"
	verificationsvciface "github.com/sunibgram/domain/service/verification"
	"github.com/sunibgram/infrastructure/cache"
	"github.com/sunibgram/infrastructure/scheduler"
	"github.com/sunibgram/webapi/router"
	webapisocket "github.com/sunibgram/webapi/socket"
)

func NewServer(ctrls *Controllers, svcs *Services, ext *External, c *cache.Cache) *fiber.App {
	app := fiber.New(fiber.Config{
		BodyLimit:               25 * 1024 * 1024,
		EnableTrustedProxyCheck: true,
		TrustedProxies:          trustedProxies(),
		ProxyHeader:             fiber.HeaderXForwardedFor,
		ReadTimeout:             15 * time.Second,
		WriteTimeout:            30 * time.Second,
		IdleTimeout:             60 * time.Second,
	})
	app.Use(recover.New())
	app.Use(helmet.New(helmet.Config{
		ContentSecurityPolicy: "default-src 'self'",
		ReferrerPolicy:        "no-referrer",
		HSTSMaxAge:            31536000,
		HSTSPreloadEnabled:    true,
		XSSProtection:         "0",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
	}))
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
		// Never compress the LLM streams — gzip buffering would defeat the
		// incremental flush and make the response arrive in one blob.
		Next: func(c *fiber.Ctx) bool {
			p := c.Path()
			return strings.HasSuffix(p, "/summary/stream") || strings.HasSuffix(p, "/fox/ask")
		},
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     corsAllowOrigins(),
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		AllowCredentials: true,
	}))
	webapisocket.Mount(app, ext.SocketHub, svcs.ChatRoom, ctrls.NotificationHandler, c)
	registerRoutes(app, ctrls, c)
	return app
}


func corsAllowOrigins() string {
	raw := strings.TrimSpace(os.Getenv("CORS_ALLOW_ORIGINS"))
	if raw == "" {
		log.Fatal("CORS_ALLOW_ORIGINS must be set (comma-separated list of allowed origins)")
	}
	if raw == "*" {
		log.Fatal("CORS_ALLOW_ORIGINS wildcard '*' is incompatible with AllowCredentials")
	}
	return raw
}


func StartVerificationCodePurge(svc verificationsvciface.IVerificationService) {
	go scheduler.RunVerificationCodePurge(context.Background(), svc)
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
		ctrls.BookmarkCollection,
		ctrls.ThreadPin,
		ctrls.Repost,
		ctrls.ThreadSummary,
		ctrls.FoxChat,
		ctrls.Email,
		ctrls.Verification,
		ctrls.Follow,
		ctrls.UserBlock,
		ctrls.NotificationUserMute,
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
		ctrls.Feedback,
		ctrls.Order,
		ctrls.PremiumSubscription,
		ctrls.DevSubscription,
		ctrls.ChatRoom,
		ctrls.Chat,
		ctrls.Share,
		ctrls.Notification,
		ctrls.ThreadPoll,
		ctrls.TrendingUser,
		ctrls.TrendingThread,
		ctrls.TrendReport,
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
