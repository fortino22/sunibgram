package controller

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/webapi/ctxutil"
)

func (c *ThreadController) Mention(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	query := strings.TrimSpace(ctx.Query("q"))
	query = strings.TrimPrefix(query, "@")

	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))
	users, err := c.threadService.GetMentionUsers(ctx.Context(), claims.UserID, query, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": c.toMentionUserResponses(users)})
}

func (c *ThreadController) toMentionUserResponses(users []model.ExploreUserResult) []dto.MentionUserResponse {
	responses := make([]dto.MentionUserResponse, len(users))
	for i, user := range users {
		avatarURL := ""
		if user.AvatarObjectName != "" {
			avatarURL = c.profileService.GetURL(user.AvatarBucket, user.AvatarObjectName)
		}

		responses[i] = dto.MentionUserResponse{
			UserID:      strconv.FormatUint(uint64(user.UserID), 10),
			Username:    user.Username,
			Name:        user.Name,
			IsPremium:   user.IsPremium,
			UserProfile: avatarURL,
			IsFollowing: user.IsFollowing,
		}
	}
	return responses
}
