package controller

import (
	"github.com/sunibgram/domain/model"
	"github.com/sunibgram/dto"
	dtoutils "github.com/sunibgram/dto/utils"
)

const defaultExploreResponseLimit = 20

func (c *ExploreController) toResponse(result *model.ExploreResult) dto.ExploreResponse {
	return dto.ExploreResponse{
		Users:   c.toUserResponses(result.Users),
		Threads: c.toThreadResponses(result.Threads),
	}
}

func (c *ExploreController) toUserResponses(users []model.ExploreUserResult) []dto.ExploreUserResponse {
	responses := make([]dto.ExploreUserResponse, len(users))
	for i, item := range users {
		avatarURL := BuildMediaURL(item.AvatarObjectName, item.AvatarBucket, c.profileService.GetURL)

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
		avatarURL := BuildMediaURL(item.AvatarObjectName, item.AvatarBucket, c.profileService.GetURL)

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
	fb := newFeedBuilder(c.mediaService, c.likeService, c.bookmarkService, c.threadService, nil, c.profileService, viewerID)
	fb.LoadAvatarsForUserIDs(threadUserIDs(threads))
	data := buildFeedItemsFromThreads(threads, mediaByThread, fb)
	ensureFalseFeedViewerFlags(data)
	return dto.PaginatedFeedResponse{
		Data: data,
		Meta: dto.FeedMeta{Total: total, Page: 1, Limit: defaultExploreResponseLimit},
	}
}

func (c *ExploreController) toMediaResponse(threads []model.Thread, total int64) dto.PaginatedExploreMediaResponse {
	mediaByThread := loadThreadMediaByThread(c.mediaService, threads)
	data := make([]dto.ExploreMediaResponse, len(threads))
	for i, thread := range threads {
		media := mediaByThread[thread.ThreadID]
		mediaResponses := make([]dto.ThreadMediaResponse, len(media))
		for j, item := range media {
			mediaResponses[j] = dto.ThreadMediaResponse{
				MediaID: item.MediaID,
				URL:     c.mediaService.GetURL(item.Bucket, item.ObjectName),
			}
		}
		data[i] = dto.ExploreMediaResponse{
			ThreadID: thread.ThreadID,
			Media:    mediaResponses,
		}
	}
	return dto.PaginatedExploreMediaResponse{
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
		avatarURL := BuildMediaURL(comm.AvatarObjectName, comm.AvatarBucket, c.communityService.GetURL)

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
	avatarURL := BuildMediaURL(comm.AvatarObjectName, comm.AvatarBucket, c.communityService.GetURL)
	bannerURL := BuildMediaURL(comm.BannerObjectName, comm.BannerBucket, c.communityService.GetURL)
	var categoryName string
	if comm.Category != nil {
		categoryName = comm.Category.Name
	}
	return dto.CommunityResponse{
		CommunityID:  comm.CommunityID,
		UserID:       comm.UserID,
		Username:     comm.User.UserName,
		Name:         comm.Name,
		Description:  comm.Description,
		Rules:        comm.Rules,
		CategoryName: categoryName,
		Visibility:   string(comm.Visibility),
		AvatarURL:    avatarURL,
		BannerURL:    bannerURL,
		MemberCount:  comm.MemberCount,
		CreatedAt:    dtoutils.Wrap(comm.CreatedAt),
		UpdatedAt:    dtoutils.Wrap(comm.UpdatedAt),
	}
}
