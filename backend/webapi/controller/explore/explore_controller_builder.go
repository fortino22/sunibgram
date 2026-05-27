package explorectrl

import (
	communitymodel "github.com/sunibgram/domain/model/community"
	exploremodel "github.com/sunibgram/domain/model/explore"
	threadmodel "github.com/sunibgram/domain/model/thread"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	threadctrl "github.com/sunibgram/webapi/controller/thread"
	dtoutils "github.com/sunibgram/webapi/dto/_shared"
	communitydto "github.com/sunibgram/webapi/dto/community"
	exploredto "github.com/sunibgram/webapi/dto/explore"
	threaddto "github.com/sunibgram/webapi/dto/thread"
	threadmediadto "github.com/sunibgram/webapi/dto/thread_media"
	webapiutils "github.com/sunibgram/webapi/utils"
)

const defaultExploreResponseLimit = 20

func (c *ExploreController) toResponse(result *exploremodel.ExploreResult) exploredto.ExploreResponse {
	return exploredto.ExploreResponse{
		Users:   c.toUserResponses(result.Users),
		Threads: c.toThreadResponses(result.Threads),
	}
}

func (c *ExploreController) toUserResponses(users []exploremodel.ExploreUserResult) []exploredto.ExploreUserResponse {
	responses := make([]exploredto.ExploreUserResponse, len(users))
	for i, item := range users {
		avatarURL := ctrlshared.BuildMediaURL(item.AvatarObjectName, c.profileService.GetURL)

		responses[i] = exploredto.ExploreUserResponse{
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

func (c *ExploreController) toPeopleResponses(users []exploremodel.ExploreUserResult) []exploredto.ExplorePeopleResponse {
	responses := make([]exploredto.ExplorePeopleResponse, len(users))
	for i, item := range users {
		avatarURL := ctrlshared.BuildMediaURL(item.AvatarObjectName, c.profileService.GetURL)

		responses[i] = exploredto.ExplorePeopleResponse{
			UserID:         item.UserID,
			UserProfilePic: avatarURL,
			Username:       item.Username,
			Name:           item.Name,
			Bio:            item.Bio,
			IsPremium:      item.IsPremium,
			IsFollowing:    item.IsFollowing,
		}
	}
	return responses
}

func (c *ExploreController) toThreadResponses(threads []exploremodel.ExploreThreadResult) []exploredto.ExploreThreadResponse {
	responses := make([]exploredto.ExploreThreadResponse, len(threads))
	for i, thread := range threads {
		responses[i] = exploredto.ExploreThreadResponse{Content: thread.Content}
	}
	return responses
}

func (c *ExploreController) toFeedResponse(threads []threadmodel.Thread, parents []*threadmodel.Thread, total int64, viewerID *uint) threaddto.PaginatedFeedResponse {
	allIDs := make([]uint, 0, len(threads)*2)
	for _, t := range threads {
		allIDs = append(allIDs, t.ThreadID)
		if t.OriginalThread != nil {
			allIDs = append(allIDs, t.OriginalThread.ThreadID)
		}
	}
	for _, p := range parents {
		if p != nil {
			allIDs = append(allIDs, p.ThreadID)
		}
	}
	mediaByThread := ctrlshared.LoadThreadMediaByID(c.mediaService, allIDs)

	authorIDs := threadctrl.ThreadUserIDs(threads)
	for _, p := range parents {
		if p != nil {
			authorIDs = append(authorIDs, p.User.UserID)
		}
	}

	fb := ctrlshared.NewFeedBuilder(c.mediaService, c.likeService, c.bookmarkService, c.threadService, nil, c.profileService, c.pollService, viewerID)
	fb.LoadAvatarsForUserIDs(authorIDs)
	fb.LoadViewerFlagsForThreads(threads)

	pollIDs := make([]uint, 0, len(threads)+len(parents))
	for i := range threads {
		t := &threads[i]
		if t.Type == threadmodel.ThreadTypeRepost && t.OriginalThread != nil {
			pollIDs = append(pollIDs, t.OriginalThread.ThreadID)
		} else {
			pollIDs = append(pollIDs, t.ThreadID)
		}
	}
	for _, p := range parents {
		if p != nil {
			pollIDs = append(pollIDs, p.ThreadID)
		}
	}
	fb.LoadPollsForThreadIDs(pollIDs)

	data := make([]threaddto.FeedItemResponse, len(threads))
	for i, t := range threads {
		mediaID := t.ThreadID
		if t.Type == threadmodel.ThreadTypeRepost && t.OriginalThread != nil {
			mediaID = t.OriginalThread.ThreadID
		}
		item := fb.ToFeedItem(&t, mediaByThread[mediaID])
		if i < len(parents) && parents[i] != nil {
			p := parents[i]
			eventID, eventName := webapiutils.EventFields(p)
			item.ParentContext = &threaddto.FeedParentContext{
				ID:              p.ThreadID,
				ParentID:        p.ParentID,
				RootID:          p.RootID,
				EventID:         eventID,
				EventName:       eventName,
				Community:       webapiutils.CommunityField(p),
				Author:          fb.AuthorField(p.User),
				Content:         p.Content,
				ReplyPermission: string(p.ReplyPermission),
				Media:           fb.ToMediaResponses(mediaByThread[p.ThreadID]),
				Poll:            fb.PollForParent(p),
				Stats:           threaddto.FeedStats{Likes: p.LikeCount, Reposts: p.RepostCount, Bookmarks: p.BookmarkCount, Comments: p.CommentCount},
				CreatedAt:       dtoutils.Wrap(p.CreatedAt),
				UpdatedAt:       dtoutils.Wrap(p.UpdatedAt),
			}
		}
		data[i] = item
	}
	ctrlshared.EnsureFalseFeedViewerFlags(data)
	return threaddto.PaginatedFeedResponse{
		Data: data,
		Meta: threaddto.FeedMeta{Total: total, Page: 1, Limit: defaultExploreResponseLimit},
	}
}

func (c *ExploreController) toMediaResponse(threads []threadmodel.Thread, total int64) exploredto.PaginatedExploreMediaResponse {
	mediaByThread := ctrlshared.LoadThreadMediaByThread(c.mediaService, threads)
	data := make([]exploredto.ExploreMediaResponse, len(threads))
	for i, thread := range threads {
		media := mediaByThread[thread.ThreadID]
		mediaResponses := make([]threadmediadto.ThreadMediaResponse, len(media))
		for j, item := range media {
			mediaResponses[j] = threadmediadto.ThreadMediaResponse{
				MediaID: item.MediaID,
				URL:     c.mediaService.GetURL(item.ObjectName),
			}
		}
		data[i] = exploredto.ExploreMediaResponse{
			ThreadID: thread.ThreadID,
			Media:    mediaResponses,
		}
	}
	return exploredto.PaginatedExploreMediaResponse{
		Data: data,
		Meta: threaddto.FeedMeta{Total: total, Page: 1, Limit: defaultExploreResponseLimit},
	}
}

func (c *ExploreController) toCommunityResponses(communities []communitymodel.Community) []communitydto.CommunityResponse {
	responses := make([]communitydto.CommunityResponse, len(communities))
	for i := range communities {
		responses[i] = c.toCommunityResponse(&communities[i])
	}
	return responses
}

func (c *ExploreController) toExploreCommunityResponses(communities []exploremodel.ExploreCommunityResult) []exploredto.ExploreCommunityResponse {
	responses := make([]exploredto.ExploreCommunityResponse, len(communities))
	for i, comm := range communities {
		responses[i] = exploredto.ExploreCommunityResponse{
			CommunityID:   comm.CommunityID,
			CommunityName: comm.Name,
			MemberCount:   comm.MemberCount,
			IsJoin:        comm.IsJoin,
		}
	}
	return responses
}

func (c *ExploreController) toCommunityResponse(comm *communitymodel.Community) communitydto.CommunityResponse {
	bannerURL := ctrlshared.BuildMediaURL(comm.BannerObjectName, c.communityService.GetURL)
	var categoryName string
	if comm.Category != nil {
		categoryName = comm.Category.Name
	}
	return communitydto.CommunityResponse{
		CommunityID:  comm.CommunityID,
		UserID:       comm.UserID,
		Username:     comm.User.UserName,
		Name:         comm.Name,
		Description:  comm.Description,
		Rules:        nil,
		CategoryName: categoryName,
		Visibility:   string(comm.Visibility),
		BannerURL:    bannerURL,
		MemberCount:  comm.MemberCount,
		CreatedAt:    dtoutils.Wrap(comm.CreatedAt),
		UpdatedAt:    dtoutils.Wrap(comm.UpdatedAt),
	}
}
