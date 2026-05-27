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
	page, limit := ctrlshared.ParsePagination(ctx)

	result, err := c.exploreService.ExploreSearch(ctx.Context(), ctx.Query("q"), requesterID, page, limit)
	if err != nil {
		if err.Error() == "query is required" {
			return ctrlshared.HandleServiceError(ctx, err)
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
	page, limit := ctrlshared.ParsePagination(ctx)

	threads, parents, total, err := c.exploreService.GetTopPosts(ctx.Context(), ctx.Query("q"), claims.UserID, page, limit)
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
	page, limit := ctrlshared.ParsePagination(ctx)

	threads, parents, total, err := c.exploreService.GetLatestPosts(ctx.Context(), ctx.Query("q"), claims.UserID, page, limit)
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
	page, limit := ctrlshared.ParsePagination(ctx)

	users, err := c.exploreService.GetPeople(ctx.Context(), ctx.Query("q"), claims.UserID, page, limit)
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
	page, limit := ctrlshared.ParsePagination(ctx)

	threads, total, err := c.exploreService.GetMediaPosts(ctx.Context(), ctx.Query("q"), claims.UserID, page, limit)
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
	page, limit := ctrlshared.ParsePagination(ctx)

	communities, _, err := c.exploreService.GetCommunities(ctx.Context(), ctx.Query("q"), claims.UserID, page, limit)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": c.toExploreCommunityResponses(communities)})
}

// GetFacultySuggestions returns people who share a faculty with the requester.
func (c *ExploreController) GetFacultySuggestions(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	page, limit := ctrlshared.ParsePagination(ctx)

	users, err := c.exploreService.GetFacultySuggestions(ctx.Context(), claims.UserID, page, limit)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": c.toPeopleResponses(users)})
}

// GetMutualSuggestions returns mutual-following (friends-of-friends) candidates.
func (c *ExploreController) GetMutualSuggestions(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	page, limit := ctrlshared.ParsePagination(ctx)

	users, err := c.exploreService.GetMutualSuggestions(ctx.Context(), claims.UserID, page, limit)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": c.toPeopleResponses(users)})
}

// GetSuggestedPeoplePreview returns up to 3 people (mutual → faculty → random).
func (c *ExploreController) GetSuggestedPeoplePreview(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	users, err := c.exploreService.GetSuggestedPeoplePreview(ctx.Context(), claims.UserID)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"data": c.toPeopleResponses(users)})
}
