package socket

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	fibersocketio "github.com/gofiber/contrib/socketio"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/infrastructure/cache"
)

const (
	claimKindAccess = "access"
	localSocketUser = "socket_user_id"
)

func Mount(app *fiber.App, hub service.IHub, roomService service.IChatRoomService, sessionCache *cache.Cache) {
	handler := NewChatHandler(hub, roomService)

	fibersocketio.On(fibersocketio.EventMessage, handler.HandleMessage)
	fibersocketio.On(fibersocketio.EventDisconnect, handler.HandleDisconnect)
	fibersocketio.On(fibersocketio.EventClose, handler.HandleDisconnect)

	app.Use("/ws", func(c *fiber.Ctx) error {
		if !websocket.IsWebSocketUpgrade(c) {
			return fiber.ErrUpgradeRequired
		}

		userID, err := authenticateSocket(c, sessionCache)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": err.Error()})
		}
		c.Locals(localSocketUser, userID)
		return c.Next()
	})

	app.Get("/ws", fibersocketio.New(func(kws *fibersocketio.Websocket) {
		userID, ok := kws.Locals(localSocketUser).(uint)
		if !ok || userID == 0 {
			kws.Close()
			return
		}

		hub.Register(kws.UUID, userID)
		kws.SetAttribute(userIDAttribute, userID)
	}))
}

func authenticateSocket(c *fiber.Ctx, sessionCache *cache.Cache) (uint, error) {
	tokenStr := bearerToken(c.Get("Authorization"))
	if tokenStr == "" {
		tokenStr = c.Query("token")
	}
	if tokenStr == "" {
		tokenStr = c.Query("access_token")
	}
	if tokenStr == "" {
		return 0, errors.New("missing token")
	}

	claims, err := parseAccessToken(tokenStr)
	if err != nil {
		return 0, err
	}

	sid, _ := claims["sid"].(string)
	if sid != "" && sessionCache != nil {
		var uid uint
		if err := sessionCache.Get(context.Background(), fmt.Sprintf("session:%s", sid), &uid); err != nil {
			return 0, errors.New("session expired or revoked")
		}
	}

	userID, ok := numericClaim(claims["user_id"])
	if !ok || userID == 0 {
		return 0, errors.New("invalid token claims")
	}
	return userID, nil
}

func bearerToken(header string) string {
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
}

func parseAccessToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtSecret()), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}
	if kind, _ := claims["kind"].(string); kind != claimKindAccess {
		return nil, errors.New("wrong token type")
	}
	return claims, nil
}

func jwtSecret() string {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return s
	}
	return "sunibgram-secret"
}

func numericClaim(value any) (uint, bool) {
	switch v := value.(type) {
	case float64:
		if v <= 0 {
			return 0, false
		}
		return uint(v), true
	case int:
		if v <= 0 {
			return 0, false
		}
		return uint(v), true
	case uint:
		return v, v != 0
	default:
		return 0, false
	}
}
