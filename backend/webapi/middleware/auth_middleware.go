package middleware

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sunibgram/infrastructure/cache"
)

const (
	claimKindAccess = "access"
	LocalClaims     = "claims"
)

var sessionCache *cache.Cache

func Init(c *cache.Cache) {
	sessionCache = c
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
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "session expired or revoked"})
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
		if roleName != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "admin access required"})
		}

		return c.Next()
	}
}

func parseAccessToken(tokenStr string) (jwt.MapClaims, error) {
	secret := jwtSecret()
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
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
