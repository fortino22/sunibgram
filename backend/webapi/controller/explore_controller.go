package controller

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	"github.com/sunibgram/webapi/ctxutil"
)

type ExploreController struct {
	exploreService   service.IExploreService
	profileService   service.IUserProfileService
	mediaService     service.IThreadMediaService
	likeService      service.ILikeService
	bookmarkService  service.IBookmarkService
	threadService    service.IThreadService
	communityService service.ICommunityService
}

func NewExploreController(
	exploreService service.IExploreService,
	profileService service.IUserProfileService,
	mediaService service.IThreadMediaService,
	likeService service.ILikeService,
	bookmarkService service.IBookmarkService,
	threadService service.IThreadService,
	communityService service.ICommunityService,
) *ExploreController {
	return &ExploreController{
		exploreService:   exploreService,
		profileService:   profileService,
		mediaService:     mediaService,
		likeService:      likeService,
		bookmarkService:  bookmarkService,
		threadService:    threadService,
		communityService: communityService,
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

func (c *ExploreController) GetTop(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	threads, total, err := c.exploreService.GetTopPosts(ctx.Context(), ctx.Query("q"), claims.UserID)
	if err != nil {
		return exploreError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(c.toFeedResponse(threads, total, &claims.UserID))
}

func (c *ExploreController) GetLatest(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	threads, total, err := c.exploreService.GetLatestPosts(ctx.Context(), ctx.Query("q"), claims.UserID)
	if err != nil {
		return exploreError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(c.toFeedResponse(threads, total, &claims.UserID))
}

func (c *ExploreController) GetPeople(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	users, err := c.exploreService.GetPeople(ctx.Context(), ctx.Query("q"), claims.UserID)
	if err != nil {
		return exploreError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": c.toPeopleResponses(users)})
}

func (c *ExploreController) GetMedia(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	threads, total, err := c.exploreService.GetMediaPosts(ctx.Context(), ctx.Query("q"), claims.UserID)
	if err != nil {
		return exploreError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(c.toFeedResponse(threads, total, &claims.UserID))
}

func (c *ExploreController) GetCommunities(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	communities, _, err := c.exploreService.GetCommunities(ctx.Context(), ctx.Query("q"), claims.UserID)
	if err != nil {
		return exploreError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": c.toExploreCommunityResponses(communities)})
}

func (c *ExploreController) GetSuggestedPeople(ctx *fiber.Ctx) error {
	claims, ok := ctxutil.Get(ctx)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Message: "unauthorized"})
	}

	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	users, err := c.exploreService.GetSuggestedPeople(ctx.Context(), claims.UserID, limit)
	if err != nil {
		return exploreError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": c.toPeopleResponses(users)})
}
