package ctrlshared

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	authdto "github.com/sunibgram/webapi/dto/auth"
	webapiutils "github.com/sunibgram/webapi/utils"
)

func RequireAuth(ctx *fiber.Ctx) (*webapiutils.Claims, bool) {
	claims, ok := webapiutils.Get(ctx)
	if !ok {
		_ = ctx.Status(fiber.StatusUnauthorized).JSON(authdto.ErrorResponse{Message: "unauthorized"})
		return nil, false
	}
	return &claims, true
}

func ParseIDParam(ctx *fiber.Ctx, paramName string, entityName string) (uint, bool) {
	id, err := strconv.ParseUint(ctx.Params(paramName), 10, 64)
	if err != nil {
		_ = ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid " + entityName + " id"})
		return 0, false
	}
	return uint(id), true
}

func ParsePagination(ctx *fiber.Ctx) (page int, limit int) {
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

// unsafeErrorMarkers flag substrings that indicate the error string carries
// infra-layer detail (database driver, Redis, network, panic). Such errors
// must never reach the client — they leak schema, IPs, or stack info. The
// raw message is logged server-side; the client sees a generic message.
var unsafeErrorMarkers = []string{
	"sql:", "gorm", "pq:", "pgx", "redis:", "panic:",
	"connection refused", "no such host", "i/o timeout",
	"syntax error", "duplicate key", "violates", "constraint",
	"ERROR:", "deadlock", "dial tcp", "dial unix",
	"EOF", "broken pipe",
}

// safeErrorMessage returns msg if it looks like a deliberate user-facing
// error; otherwise logs the raw message and returns "internal error". The
// allowlist is by exclusion: business errors never contain SQL or driver
// keywords, so anything matching unsafeErrorMarkers is treated as a leak.
func safeErrorMessage(ctx *fiber.Ctx, msg string) string {
	lower := strings.ToLower(msg)
	for _, marker := range unsafeErrorMarkers {
		if strings.Contains(lower, strings.ToLower(marker)) {
			log.Printf("internal error on %s %s: %s", ctx.Method(), ctx.Path(), msg)
			return "internal error"
		}
	}
	return msg
}

func HandleServiceError(ctx *fiber.Ctx, err error) error {
	status := fiber.StatusBadRequest
	msg := err.Error()
	switch {
	case strings.Contains(msg, "forbidden"):
		status = fiber.StatusForbidden
	case strings.Contains(msg, "not found"):
		status = fiber.StatusNotFound
	}
	return ctx.Status(status).JSON(authdto.ErrorResponse{Message: safeErrorMessage(ctx, msg)})
}

func BuildMediaURL(objectName string, getURL func(string) string) string {
	if objectName == "" {
		return ""
	}
	return getURL(objectName)
}

func ParseDateField(raw string, fieldName string) (time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, errors.New(fieldName + " is required")
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, errors.New(fieldName + " must use YYYY-MM-DD format")
	}
	return parsed, nil
}

func FormatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}
