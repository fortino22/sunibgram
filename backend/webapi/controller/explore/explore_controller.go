package explorectrl

import (
	"github.com/gofiber/fiber/v2"

	bookmarksvciface "github.com/sunibgram/domain/service/bookmark"
	communitysvciface "github.com/sunibgram/domain/service/community"
	exploresvciface "github.com/sunibgram/domain/service/explore"
	likesvciface "github.com/sunibgram/domain/service/like"
	threadsvciface "github.com/sunibgram/domain/service/thread"
	threadmediasvciface "github.com/sunibgram/domain/service/thread_media"
	threadpollsvciface "github.com/sunibgram/domain/service/thread_poll"
	userprofilesvciface "github.com/sunibgram/domain/service/user_profile"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	authdto "github.com/sunibgram/webapi/dto/auth"
	webapiutils "github.com/sunibgram/webapi/utils"
)

type ExploreController struct {
	exploreService   exploresvciface.IExploreService
	profileService   userprofilesvciface.IUserProfileService
	mediaService     threadmediasvciface.IThreadMediaService
	likeService      likesvciface.ILikeService
	bookmarkService  bookmarksvciface.IBookmarkService
	threadService    threadsvciface.IThreadService
	communityService communitysvciface.ICommunityService
	pollService      threadpollsvciface.IThreadPollService
}

func NewExploreController(
	exploreService exploresvciface.IExploreService,
	profileService userprofilesvciface.IUserProfileService,
	mediaService threadmediasvciface.IThreadMediaService,
	likeService likesvciface.ILikeService,
	bookmarkService bookmarksvciface.IBookmarkService,
	threadService threadsvciface.IThreadService,
	communityService communitysvciface.ICommunityService,
	pollService threadpollsvciface.IThreadPollService,
) *ExploreController {
	return &ExploreController{
		exploreService:   exploreService,
		profileService:   profileService,
		mediaService:     mediaService,
		likeService:      likeService,
		bookmarkService:  bookmarkService,
		threadService:    threadService,
		communityService: communityService,
		pollService:      pollService,
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
			return ctx.Status(fiber.StatusBadRequest).JSON(authdto.ErrorResponse{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(authdto.ErrorResponse{Message: "internal server error"})
	}
	return ctx.Status(fiber.StatusOK).JSON(c.toResponse(result))
}

func (c *ExploreController) GetTop(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	threads, parents, total, err := c.exploreService.GetTopPosts(ctx.Context(), ctx.Query("q"), claims.UserID)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(c.toFeedResponse(threads, parents, total, &claims.UserID))
}

func (c *ExploreController) GetLatest(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	threads, parents, total, err := c.exploreService.GetLatestPosts(ctx.Context(), ctx.Query("q"), claims.UserID)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(c.toFeedResponse(threads, parents, total, &claims.UserID))
}

func (c *ExploreController) GetPeople(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	users, err := c.exploreService.GetPeople(ctx.Context(), ctx.Query("q"), claims.UserID)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": c.toPeopleResponses(users)})
}

func (c *ExploreController) GetMedia(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	threads, total, err := c.exploreService.GetMediaPosts(ctx.Context(), ctx.Query("q"), claims.UserID)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(c.toMediaResponse(threads, total))
}

func (c *ExploreController) GetCommunities(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	communities, _, err := c.exploreService.GetCommunities(ctx.Context(), ctx.Query("q"), claims.UserID)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": c.toExploreCommunityResponses(communities)})
}

func (c *ExploreController) GetSuggestedPeople(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	_, limit := ctrlshared.ParsePagination(ctx)

	users, err := c.exploreService.GetSuggestedPeople(ctx.Context(), claims.UserID, limit)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": c.toPeopleResponses(users)})
}
