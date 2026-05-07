package controller

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/webapi/ctxutil"
)

type FollowController struct {
	followService  service.IFollowService
	profileService service.IUserProfileService
}

func NewFollowController(followService service.IFollowService, profileService service.IUserProfileService) *FollowController {
	return &FollowController{followService: followService, profileService: profileService}
}

func (c *FollowController) Follow(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}
	targetID, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid user id"})
	}
	if err := c.followService.Follow(claims.UserID, uint(targetID)); err != nil {
		if err.Error() == "user not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusConflict).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "followed successfully"})
}

func (c *FollowController) Unfollow(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}
	targetID, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid user id"})
	}
	if err := c.followService.Unfollow(claims.UserID, uint(targetID)); err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"message": "unfollowed successfully"})
}

func (c *FollowController) GetFollowing(ctx *fiber.Ctx) error {
	if _, ok := ctxutil.Get(ctx); !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	userID, err := strconv.ParseUint(ctx.Params("user_id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid user id"})
	}

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	users, total, err := c.followService.GetFollowing(uint(userID), page, limit)
	if err != nil {
		return followListError(ctx, err)
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
	userID, err := strconv.ParseUint(ctx.Params("user_id"), 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: "invalid user id"})
	}

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	users, total, err := c.followService.GetFollowers(uint(userID), page, limit)
	if err != nil {
		return followListError(ctx, err)
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
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	users, total, err := c.followService.GetFollowing(claims.UserID, page, limit)
	if err != nil {
		return followListError(ctx, err)
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
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	users, total, err := c.followService.GetFollowers(claims.UserID, page, limit)
	if err != nil {
		return followListError(ctx, err)
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
		if ok && profile.AvatarBucket != "" && profile.AvatarObjectName != "" {
			avatarURL = c.profileService.GetURL(profile.AvatarBucket, profile.AvatarObjectName)
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

func followListError(ctx *fiber.Ctx, err error) error {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "forbidden"):
		return ctx.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Message: msg})
	case strings.Contains(msg, "not found"):
		return ctx.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Message: msg})
	default:
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: msg})
	}
}
