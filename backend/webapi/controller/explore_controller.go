package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/service"
	"github.com/sunibgram/dto"
	webapiutils "github.com/sunibgram/webapi/utils"
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
	claims, ok := webapiutils.Get(ctx)
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
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	threads, total, err := c.exploreService.GetTopPosts(ctx.Context(), ctx.Query("q"), claims.UserID)
	if err != nil {
		return HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(c.toFeedResponse(threads, total, &claims.UserID))
}

func (c *ExploreController) GetLatest(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	threads, total, err := c.exploreService.GetLatestPosts(ctx.Context(), ctx.Query("q"), claims.UserID)
	if err != nil {
		return HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(c.toFeedResponse(threads, total, &claims.UserID))
}

func (c *ExploreController) GetPeople(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	users, err := c.exploreService.GetPeople(ctx.Context(), ctx.Query("q"), claims.UserID)
	if err != nil {
		return HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": c.toPeopleResponses(users)})
}

func (c *ExploreController) GetMedia(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	threads, total, err := c.exploreService.GetMediaPosts(ctx.Context(), ctx.Query("q"), claims.UserID)
	if err != nil {
		return HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(c.toMediaResponse(threads, total))
}

func (c *ExploreController) GetCommunities(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	communities, _, err := c.exploreService.GetCommunities(ctx.Context(), ctx.Query("q"), claims.UserID)
	if err != nil {
		return HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": c.toExploreCommunityResponses(communities)})
}

func (c *ExploreController) GetSuggestedPeople(ctx *fiber.Ctx) error {
	claims, ok := RequireAuth(ctx)
	if !ok {
		return nil
	}

	_, limit := ParsePagination(ctx)

	users, err := c.exploreService.GetSuggestedPeople(ctx.Context(), claims.UserID, limit)
	if err != nil {
		return HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": c.toPeopleResponses(users)})
}
