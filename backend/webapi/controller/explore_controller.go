package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/webapi/ctxutil"
)

type ExploreController struct {
	exploreService service.IExploreService
	profileService service.IUserProfileService
}

func NewExploreController(exploreService service.IExploreService, profileService service.IUserProfileService) *ExploreController {
	return &ExploreController{
		exploreService: exploreService,
		profileService: profileService,
	}
}

func (c *ExploreController) ExploreSearch(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	var requesterID *uint
	if ok {
		requesterID = &claims.UserID
	}

	result, err := c.exploreService.ExploreSearch(ctx.Context(), ctx.Query("q"), requesterID)
	if err != nil {
		if err.Error() == "query is required" {
			return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(c.toResponse(result))
}

func (c *ExploreController) toResponse(result *model.ExploreResult) dto.ExploreResponse {
	return dto.ExploreResponse{
		Users:   c.toUserResponses(result.Users),
		Threads: c.toThreadResponses(result.Threads),
	}
}

func (c *ExploreController) toUserResponses(users []model.ExploreUserResult) []dto.ExploreUserResponse {
	responses := make([]dto.ExploreUserResponse, len(users))
	for i, item := range users {
		avatarURL := ""
		if item.AvatarObjectName != "" {
			avatarURL = c.profileService.GetURL(item.AvatarBucket, item.AvatarObjectName)
		}

		responses[i] = dto.ExploreUserResponse{
			UserID:      item.UserID,
			Profile:     avatarURL,
			Username:    item.Username,
			Name:        item.Name,
			IsPremium:   item.IsPremium,
			IsFollowing: item.IsFollowing,
		}
	}
	return responses
}

func (c *ExploreController) toThreadResponses(threads []model.ExploreThreadResult) []dto.ExploreThreadResponse {
	responses := make([]dto.ExploreThreadResponse, len(threads))
	for i, thread := range threads {
		responses[i] = dto.ExploreThreadResponse{Content: thread.Content}
	}
	return responses
}
