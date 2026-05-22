package controller

import (
	"github.com/gofiber/fiber/v2"

	usermodel "github.com/sunibgram/domain/model/user"
	userprofilemodel "github.com/sunibgram/domain/model/user_profile"
	followsvciface "github.com/sunibgram/domain/service/follow"
	userprofilesvciface "github.com/sunibgram/domain/service/user_profile"
	authdto "github.com/sunibgram/webapi/dto/auth"
	followdto "github.com/sunibgram/webapi/dto/follow"
	threaddto "github.com/sunibgram/webapi/dto/thread"
)

type FollowController struct {
	followService  followsvciface.IFollowService
	profileService userprofilesvciface.IUserProfileService
}

func NewFollowController(followService followsvciface.IFollowService, profileService userprofilesvciface.IUserProfileService) *FollowController {
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
			return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusConflict).JSON(authdto.ErrorResponse{Message: err.Error()})
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
		return ctx.Status(fiber.StatusNotFound).JSON(authdto.ErrorResponse{Message: err.Error()})
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
		return ctx.Status(fiber.StatusOK).JSON(followdto.PaginatedFollowUsersResponse{
			Data: []followdto.FollowUserResponse{},
			Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
		})
	}

	profileMap, err := c.followUserProfileMap(users)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(followdto.PaginatedFollowUsersResponse{
		Data: c.toFollowUserResponses(users, profileMap),
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
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
		return ctx.Status(fiber.StatusOK).JSON(followdto.PaginatedFollowUsersResponse{
			Data: []followdto.FollowUserResponse{},
			Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
		})
	}

	profileMap, err := c.followUserProfileMap(users)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(followdto.PaginatedFollowUsersResponse{
		Data: c.toFollowUserResponses(users, profileMap),
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
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
		return ctx.Status(fiber.StatusOK).JSON(followdto.PaginatedFollowUsersResponse{
			Data: []followdto.FollowUserResponse{},
			Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
		})
	}

	profileMap, err := c.followUserProfileMap(users)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(followdto.PaginatedFollowUsersResponse{
		Data: c.toFollowUserResponses(users, profileMap),
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
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
		return ctx.Status(fiber.StatusOK).JSON(followdto.PaginatedFollowUsersResponse{
			Data: []followdto.FollowUserResponse{},
			Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
		})
	}

	profileMap, err := c.followUserProfileMap(users)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(followdto.PaginatedFollowUsersResponse{
		Data: c.toFollowUserResponses(users, profileMap),
		Meta: threaddto.FeedMeta{Total: total, Page: page, Limit: limit},
	})
}

func (c *FollowController) toFollowUserResponses(users []usermodel.User, profileMap map[uint]userprofilemodel.UserProfile) []followdto.FollowUserResponse {
	data := make([]followdto.FollowUserResponse, len(users))
	for i, user := range users {
		avatarURL := ""
		profile, ok := profileMap[user.UserID]
		if ok {
			avatarURL = BuildMediaURL(profile.AvatarObjectName, profile.AvatarBucket, c.profileService.GetURL)
		}
		data[i] = followdto.FollowUserResponse{
			UserID:    user.UserID,
			Name:      user.Name,
			UserName:  user.UserName,
			IsPremium: user.IsPremium,
			AvatarURL: avatarURL,
		}
	}
	return data
}

func (c *FollowController) followUserProfileMap(users []usermodel.User) (map[uint]userprofilemodel.UserProfile, error) {
	userIDs := make([]uint, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.UserID)
	}

	profiles, err := c.profileService.GetProfiles(userIDs)
	if err != nil {
		return nil, err
	}

	profileMap := make(map[uint]userprofilemodel.UserProfile, len(profiles))
	for _, profile := range profiles {
		profileMap[profile.UserID] = profile
	}
	return profileMap, nil
}
