package middleware

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	rolemodel "github.com/sunibgram/domain/model/role"
	"github.com/sunibgram/infrastructure/cache"
)

const (
	claimKindAccess = "access"
	LocalClaims     = "claims"
	expectedIssuer  = "sunibgram-api"
)

var (
	sessionCache    *cache.Cache
	cachedJWTSecret []byte
)

func Init(c *cache.Cache) {
	sessionCache = c
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET is required")
	}
	cachedJWTSecret = []byte(secret)
}

// JWTSecret returns the cached JWT secret loaded at boot via Init.
// Exposed so the socket package can share the same secret without re-reading env.
func JWTSecret() []byte {
	return cachedJWTSecret
}

func Auth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "missing or invalid Authorization header"})
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := parseAccessToken(tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": err.Error()})
		}

		sid, _ := claims["sid"].(string)
		if sid != "" && sessionCache != nil {
			var uid uint
			if err := sessionCache.Get(context.Background(), fmt.Sprintf("session:%s", sid), &uid); err != nil {
				// Only a true cache miss means the session was revoked/expired. Any
				// other error (Redis timeout, failover) is transient — return 503 so
				// the client retries instead of logging the user out on a blip.
				if sessionCache.Miss(err) {
					return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "session expired or revoked"})
				}
				return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"message": "session store unavailable"})
			}
		}

		c.Locals(LocalClaims, claims)
		return c.Next()
	}
}

func OptionalAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			return c.Next()
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := parseAccessToken(tokenStr)
		if err != nil {
			return c.Next()
		}

		sid, _ := claims["sid"].(string)
		if sid != "" && sessionCache != nil {
			var uid uint
			if err := sessionCache.Get(context.Background(), fmt.Sprintf("session:%s", sid), &uid); err != nil {
				uidClaim, _ := claims["user_id"].(float64)
				log.Printf("auth: revoked-session optional read sid=%s uid=%d ip=%s path=%s", sid, uint(uidClaim), c.IP(), c.Path())
				return c.Next()
			}
		}

		c.Locals(LocalClaims, claims)
		return c.Next()
	}
}

func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := c.Locals(LocalClaims).(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "unauthorized"})
		}

		roleName, _ := claims["role_name"].(string)
		if roleName != rolemodel.RoleAdmin {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "admin access required"})
		}

		return c.Next()
	}
}

func parseAccessToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return cachedJWTSecret, nil
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
