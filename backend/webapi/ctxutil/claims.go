package ctxutil

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sunibgram/webapi/middleware"
)

type Claims struct {
	UserID   uint
	Username string
	Email    string
	RoleID   uint
	RoleName string
}

func Get(ctx *fiber.Ctx) (Claims, bool) {
	raw, ok := ctx.Locals(middleware.LocalClaims).(jwt.MapClaims)
	if !ok {
		return Claims{}, false
	}

	return Claims{
		UserID:   uint(raw["user_id"].(float64)),
		Username: str(raw["username"]),
		Email:    str(raw["email"]),
		RoleID:   uint(raw["role_id"].(float64)),
		RoleName: str(raw["role_name"]),
	}, true
}

func str(v any) string {
	s, _ := v.(string)
	return s
}
