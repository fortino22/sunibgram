package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/dto"
	dtoutils "github.com/sunibgram/dto/utils"
)

const defaultExploreResponseLimit = 20

func exploreError(ctx *fiber.Ctx, err error) error {
	if err.Error() == "query is required" {
		return ctx.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Message: err.Error()})
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

func (c *ExploreController) toPeopleResponses(users []model.ExploreUserResult) []dto.ExplorePeopleResponse {
	responses := make([]dto.ExplorePeopleResponse, len(users))
	for i, item := range users {
		avatarURL := ""
		if item.AvatarObjectName != "" {
			avatarURL = c.profileService.GetURL(item.AvatarBucket, item.AvatarObjectName)
		}

		responses[i] = dto.ExplorePeopleResponse{
			UserID:         item.UserID,
			UserProfilePic: avatarURL,
			Username:       item.Username,
			Name:           item.Name,
			IsPremium:      item.IsPremium,
			IsFollowing:    item.IsFollowing,
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

func (c *ExploreController) toFeedResponse(threads []model.Thread, total int64, viewerID *uint) dto.PaginatedFeedResponse {
	mediaByThread := loadThreadMediaByThread(c.mediaService, threads)
	fb := newFeedBuilder(c.mediaService, c.likeService, c.bookmarkService, c.threadService, nil, viewerID)
	data := buildFeedItemsFromThreads(threads, mediaByThread, fb)
	ensureFalseFeedViewerFlags(data)
	return dto.PaginatedFeedResponse{
		Data: data,
		Meta: dto.FeedMeta{Total: total, Page: 1, Limit: defaultExploreResponseLimit},
	}
}

func (c *ExploreController) toCommunityResponses(communities []model.Community) []dto.CommunityResponse {
	responses := make([]dto.CommunityResponse, len(communities))
	for i := range communities {
		responses[i] = c.toCommunityResponse(&communities[i])
	}
	return responses
}

func (c *ExploreController) toExploreCommunityResponses(communities []model.ExploreCommunityResult) []dto.ExploreCommunityResponse {
	responses := make([]dto.ExploreCommunityResponse, len(communities))
	for i, comm := range communities {
		avatarURL := ""
		if comm.AvatarObjectName != "" {
			avatarURL = c.communityService.GetURL(comm.AvatarBucket, comm.AvatarObjectName)
		}

		responses[i] = dto.ExploreCommunityResponse{
			CommunityID:      comm.CommunityID,
			CommunityProfile: avatarURL,
			CommunityName:    comm.Name,
			MemberCount:      comm.MemberCount,
			IsJoin:           comm.IsJoin,
		}
	}
	return responses
}

func (c *ExploreController) toCommunityResponse(comm *model.Community) dto.CommunityResponse {
	avatarURL := ""
	if comm.AvatarObjectName != "" {
		avatarURL = c.communityService.GetURL(comm.AvatarBucket, comm.AvatarObjectName)
	}
	bannerURL := ""
	if comm.BannerObjectName != "" {
		bannerURL = c.communityService.GetURL(comm.BannerBucket, comm.BannerObjectName)
	}

	return dto.CommunityResponse{
		CommunityID: comm.CommunityID,
		UserID:      comm.UserID,
		Username:    comm.User.UserName,
		Name:        comm.Name,
		Description: comm.Description,
		Rules:       comm.Rules,
		Category:    comm.Category,
		Visibility:  string(comm.Visibility),
		AvatarURL:   avatarURL,
		BannerURL:   bannerURL,
		MemberCount: comm.MemberCount,
		CreatedAt:   dtoutils.Wrap(comm.CreatedAt),
		UpdatedAt:   dtoutils.Wrap(comm.UpdatedAt),
	}
}
