package utils

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	authdto "github.com/sunibgram/webapi/dto/auth"
	"github.com/sunibgram/webapi/middleware"
)

type Claims struct {
	UserID   uint
	SID      string
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
		SID:      str(raw["sid"]),
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

func RequireAuth(ctx *fiber.Ctx) (Claims, bool) {
	claims, ok := Get(ctx)
	if !ok {
		_ = ctx.Status(401).JSON(authdto.ErrorResponse{Message: "unauthorized"})
		return Claims{}, false
	}
	return claims, true
}

func ParseIDParam(ctx *fiber.Ctx, paramName, entityName string) (uint, bool) {
	id, err := strconv.ParseUint(ctx.Params(paramName), 10, 64)
	if err != nil {
		_ = ctx.Status(400).JSON(authdto.ErrorResponse{Message: entityName + " id is invalid"})
		return 0, false
	}
	return uint(id), true
}

func ParsePagination(ctx *fiber.Ctx) (page, limit int) {
	page, _ = strconv.Atoi(ctx.Query("page", "1"))
	limit, _ = strconv.Atoi(ctx.Query("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 1
	}
	if limit > 100 {
		limit = 100
	}

	return page, limit
}

func HandleServiceError(ctx *fiber.Ctx, err error) error {
	status := fiber.StatusBadRequest
	if err.Error() == "forbidden" {
		status = fiber.StatusForbidden
	} else if strings.Contains(err.Error(), "not found") {
		status = fiber.StatusNotFound
	}
	return ctx.Status(status).JSON(authdto.ErrorResponse{Message: err.Error()})
}

func BuildMediaURL(objectName, bucket string, getURL func(string, string) string) string {
	if objectName == "" {
		return ""
	}
	return getURL(bucket, objectName)
}
