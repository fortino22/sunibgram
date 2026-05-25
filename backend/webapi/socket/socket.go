package socket

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	fibersocketio "github.com/gofiber/contrib/socketio"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	chatroomsvciface "github.com/sunibgram/domain/service/chat_room"
	sockethubsvciface "github.com/sunibgram/domain/service/socket_hub"
	"github.com/sunibgram/infrastructure/cache"
	"github.com/sunibgram/infrastructure/ratelimit"
	"github.com/sunibgram/webapi/middleware"
)

const expectedIssuer = "sunibgram-api"

var wsConnectLimiter *ratelimit.RateLimiter

// InitRateLimiter wires the per-IP WS connect limiter at boot. Called from bootstrap.
func InitRateLimiter(c *cache.Cache) {
	wsConnectLimiter = ratelimit.New(c, "ws_connect", 30, time.Minute)
}

const claimKindAccess = "access"

func Mount(app *fiber.App, hub sockethubsvciface.IHub, roomService chatroomsvciface.IChatRoomService, notificationHandler *NotificationHandler, sessionCache *cache.Cache) {
	handler := NewChatHandler(hub, roomService)

	fibersocketio.On(fibersocketio.EventMessage, func(ep *fibersocketio.EventPayload) {
		event, _, err := parseClientEvent(ep.Data)
		if err != nil {
			emitSocketError(ep.Kws, err.Error())
			return
		}

		switch event {
		case "notification:get", "notification:delete":
			if notificationHandler == nil {
				emitSocketError(ep.Kws, "unsupported event")
				return
			}
			notificationHandler.HandleMessage(ep)
		default:
			handler.HandleMessage(ep)
		}
	})
	fibersocketio.On(fibersocketio.EventDisconnect, handler.HandleDisconnect)
	fibersocketio.On(fibersocketio.EventClose, handler.HandleDisconnect)

	app.Use("/ws", func(c *fiber.Ctx) error {
		if !websocket.IsWebSocketUpgrade(c) {
			return fiber.ErrUpgradeRequired
		}

		if wsConnectLimiter != nil {
			if err := wsConnectLimiter.Allow(c.IP()); err != nil {
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"message": err.Error()})
			}
		}

		if _, err := authenticateSocket(c, sessionCache); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": err.Error()})
		}
		// Echo the chosen subprotocol back so the browser handshake succeeds
		// when the client passed the JWT via Sec-WebSocket-Protocol.
		if proto, _ := subprotocolToken(c.Get("Sec-WebSocket-Protocol")); proto != "" {
			c.Set("Sec-WebSocket-Protocol", subprotocolName)
		}
		return c.Next()
	})

	app.Get("/ws", fibersocketio.New(func(kws *fibersocketio.Websocket) {
		// Re-authenticate from connection-scoped state; kws.Locals may not
		// carry Fiber context locals set by middleware across the WebSocket
		// upgrade boundary. Prefer Sec-WebSocket-Protocol; fall back to query
		// (legacy). Query path remains for backward compatibility but should
		// be retired once clients migrate.
		var tokenStr string
		if _, token := subprotocolToken(kws.Conn.Headers("Sec-WebSocket-Protocol")); token != "" {
			tokenStr = token
		}
		userID, err := validateToken(tokenStr, sessionCache)
		if err != nil {
			kws.Close()
			return
		}

		hub.Register(kws.UUID, userID)
		kws.SetAttribute(userIDAttribute, userID)
	}))
}

// subprotocolName is the WebSocket subprotocol the server announces to
// clients sending the access token via Sec-WebSocket-Protocol. The browser
// API only allows the subprotocol slot for arbitrary handshake-time strings,
// so clients send two values: ["sunibgram.access.v1", "<jwt>"]. Server picks
// the recognized one back, satisfying the RFC 6455 handshake.
const subprotocolName = "sunibgram.access.v1"

// subprotocolToken parses a Sec-WebSocket-Protocol header (comma-separated)
// and returns (selectedProtocol, token). Returns empty strings if the header
// does not advertise our subprotocol or no token follows it.
func subprotocolToken(header string) (string, string) {
	if header == "" {
		return "", ""
	}
	parts := strings.Split(header, ",")
	var sawProto bool
	for _, raw := range parts {
		p := strings.TrimSpace(raw)
		if p == subprotocolName {
			sawProto = true
			continue
		}
		if sawProto && p != "" {
			return subprotocolName, p
		}
	}
	return "", ""
}

func authenticateSocket(c *fiber.Ctx, sessionCache *cache.Cache) (uint, error) {
	tokenStr := bearerToken(c.Get("Authorization"))
	if tokenStr == "" {
		if _, token := subprotocolToken(c.Get("Sec-WebSocket-Protocol")); token != "" {
			tokenStr = token
		}
	}
	return validateToken(tokenStr, sessionCache)
}

func validateToken(tokenStr string, sessionCache *cache.Cache) (uint, error) {
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
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return middleware.JWTSecret(), nil
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
	if iss, _ := claims["iss"].(string); iss != "" && iss != expectedIssuer {
		return nil, errors.New("invalid token issuer")
	}
	return claims, nil
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
