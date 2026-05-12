package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
)

type FollowController struct {
	followService  service.IFollowService
	profileService service.IUserProfileService
}

func NewFollowController(followService service.IFollowService, profileService service.IUserProfileService) *FollowController {
	return &FollowController{followService: followService, profileService: profileService}
}

func (c *FollowController) Follow(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	targetID, ok := ParseIDParam(ctx, "id", "user")
	if !ok {
		return nil
	}
	if err := c.followService.Follow(claims.UserID, targetID); err != nil {
		if err.Error() == "user not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusConflict).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "followed successfully"})
}

func (c *FollowController) Unfollow(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}
	targetID, ok := ParseIDParam(ctx, "id", "user")
	if !ok {
		return nil
	}
	if err := c.followService.Unfollow(claims.UserID, targetID); err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "unfollowed successfully"})
}

func (c *FollowController) GetFollowing(ctx *fiber.Ctx) error {
	if _, ok := RequireAuth(ctx); !ok {
		return nil
	}

	userID, ok := ParseIDParam(ctx, "user_id", "user")
	if !ok {
		return nil
	}

	page, limit := ParsePagination(ctx)

	users, total, err := c.followService.GetFollowing(userID, page, limit)
	if err != nil {
		return HandleServiceError(ctx, err)
	}
	if len(users) == 0 {
		return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedFollowUsersResponse{
			Data: []dto.FollowUserResponse{},
			Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
		})
	}

	profileMap, err := c.followUserProfileMap(users)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedFollowUsersResponse{
		Data: c.toFollowUserResponses(users, profileMap),
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *FollowController) GetFollowers(ctx *fiber.Ctx) error {
	userID, ok := ParseIDParam(ctx, "user_id", "user")
	if !ok {
		return nil
	}

	page, limit := ParsePagination(ctx)

	users, total, err := c.followService.GetFollowers(userID, page, limit)
	if err != nil {
		return HandleServiceError(ctx, err)
	}
	if len(users) == 0 {
		return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedFollowUsersResponse{
			Data: []dto.FollowUserResponse{},
			Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
		})
	}

	profileMap, err := c.followUserProfileMap(users)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedFollowUsersResponse{
		Data: c.toFollowUserResponses(users, profileMap),
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *FollowController) GetMyFollowing(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	page, limit := ParsePagination(ctx)

	users, total, err := c.followService.GetFollowing(claims.UserID, page, limit)
	if err != nil {
		return HandleServiceError(ctx, err)
	}
	if len(users) == 0 {
		return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedFollowUsersResponse{
			Data: []dto.FollowUserResponse{},
			Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
		})
	}

	profileMap, err := c.followUserProfileMap(users)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedFollowUsersResponse{
		Data: c.toFollowUserResponses(users, profileMap),
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *FollowController) GetMyFollowers(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	page, limit := ParsePagination(ctx)

	users, total, err := c.followService.GetFollowers(claims.UserID, page, limit)
	if err != nil {
		return HandleServiceError(ctx, err)
	}
	if len(users) == 0 {
		return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedFollowUsersResponse{
			Data: []dto.FollowUserResponse{},
			Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
		})
	}

	profileMap, err := c.followUserProfileMap(users)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(dto.PaginatedFollowUsersResponse{
		Data: c.toFollowUserResponses(users, profileMap),
		Meta: dto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *FollowController) toFollowUserResponses(users []model.User, profileMap map[uint]model.UserProfile) []dto.FollowUserResponse {
	data := make([]dto.FollowUserResponse, len(users))
	for i, user := range users {
		avatarURL := ""
		profile, ok := profileMap[user.UserID]
		if ok {
			avatarURL = BuildMediaURL(profile.AvatarObjectName, profile.AvatarBucket, c.profileService.GetURL)
		}
		data[i] = dto.FollowUserResponse{
			UserID:    user.UserID,
			Name:      user.Name,
			UserName:  user.UserName,
			IsPremium: user.IsPremium,
			AvatarURL: avatarURL,
		}
	}
	return data
}

func (c *FollowController) followUserProfileMap(users []model.User) (map[uint]model.UserProfile, error) {
	userIDs := make([]uint, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.UserID)
	}

	profiles, err := c.profileService.GetProfiles(userIDs)
	if err != nil {
		return nil, err
	}

	profileMap := make(map[uint]model.UserProfile, len(profiles))
	for _, profile := range profiles {
		profileMap[profile.UserID] = profile
	}
	return profileMap, nil
}
