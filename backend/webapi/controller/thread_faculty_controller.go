package controller

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	authdto "github.com/sunibgram/webapi/dto/auth"
	threaddto "github.com/sunibgram/webapi/dto/thread"
	"github.com/sunibgram/webapi/middleware"
)

func (c *ThreadController) GetFacultyThreads(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	requesterFacultyID, ok := c.authenticatedFacultyID(ctx, claims.UserID)
	if !ok {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "faculty not found"})
	}

	facultyID := requesterFacultyID
	if raw := ctx.Query("faculty_id"); raw != "" {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil || id == 0 {
			return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid faculty_id"})
		}
		facultyID = uint(id)
	} else if raw := ctx.Query("facultyId"); raw != "" {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil || id == 0 {
			return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: "invalid facultyId"})
		}
		facultyID = uint(id)
	}

	if facultyID != requesterFacultyID {
		return ctx.Status(fiber.StatusForbidden).JSON(authdto.ErrorResponse{Message: "forbidden"})
	}

	page, limit := ParsePagination(ctx)

	threads, total, err := c.threadService.GetFacultyThreads(facultyID, claims.UserID, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}

	return ctx.Status(fiber.StatusOK).JSON(threaddto.PaginatedFeedResponse{
		Data: c.buildFeedItemsNoPin(threads, &claims.UserID),
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *ThreadController) authenticatedFacultyID(ctx *fiber.Ctx, userID uint) (uint, bool) {
	rawClaims, _ := ctx.Locals(middleware.LocalClaims).(jwt.MapClaims)
	if id, ok := numericClaim(rawClaims["faculty_id"]); ok && id != 0 {
		return id, true
	}
	if id, ok := numericClaim(rawClaims["facultyId"]); ok && id != 0 {
		return id, true
	}

	profile, err := c.profileService.GetProfile(userID, &userID)
	if err != nil || profile.UserFaculty == nil {
		return 0, false
	}
	return profile.UserFaculty.FacultyID, true
}

func numericClaim(value any) (uint, bool) {
	switch v := value.(type) {
	case float64:
		return uint(v), v > 0
	case int:
		return uint(v), v > 0
	case uint:
		return v, v > 0
	case string:
		id, err := strconv.ParseUint(v, 10, 64)
		return uint(id), err == nil && id > 0
	default:
		return 0, false
	}
}
