package trendingthreadctrl

import (
	"time"

	"github.com/gofiber/fiber/v2"

	trendingthreadrepoiface "github.com/sunibgram/domain/repository/trending_thread"
	trendingthreadsvciface "github.com/sunibgram/domain/service/trending_thread"
	userprofilesvciface "github.com/sunibgram/domain/service/user_profile"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	trendingthreaddto "github.com/sunibgram/webapi/dto/trending_thread"
)

type TrendingThreadController struct {
	service        trendingthreadsvciface.ITrendingThreadService
	profileService userprofilesvciface.IUserProfileService
}

func NewTrendingThreadController(
	service trendingthreadsvciface.ITrendingThreadService,
	profileService userprofilesvciface.IUserProfileService,
) *TrendingThreadController {
	return &TrendingThreadController{service: service, profileService: profileService}
}

func (c *TrendingThreadController) GetTopWeekly(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	result, err := c.service.GetTopWeekly(ctx.Context(), claims.UserID)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(c.toResponse(result))
}

func (c *TrendingThreadController) toResponse(result trendingthreadrepoiface.TrendingThreadsResult) trendingthreaddto.TrendingThreadListResponse {
	data := make([]trendingthreaddto.TrendingThreadResponse, len(result.Threads))
	for i, t := range result.Threads {
		data[i] = trendingthreaddto.TrendingThreadResponse{
			ThreadID:  t.ThreadID,
			Content:   t.Content,
			CreatedAt: t.CreatedAt,
			Score:     t.Score,
			Author: trendingthreaddto.TrendingThreadAuthor{
				UserID:    t.AuthorID,
				Username:  t.Username,
				Name:      t.Name,
				Profile:   ctrlshared.BuildMediaURL(t.AvatarObjectName, c.profileService.GetURL),
				IsPremium: t.IsPremium,
			},
			Detail: trendingthreaddto.TrendingThreadScoreDetail{
				Received: trendingthreaddto.TrendingThreadReceivedDetail{
					Likes:     t.LikesIn,
					Comments:  t.CommentsIn,
					Bookmarks: t.BookmarksIn,
					Reposts:   t.RepostsIn,
				},
				Given: trendingthreaddto.TrendingThreadGivenDetail{
					Likes:     t.LikesOut,
					Bookmarks: t.BookmarksOut,
					Comments:  t.CommentsOut,
					Posts:     t.PostsOut,
				},
			},
		}
	}

	refreshIn := int64(time.Until(result.RefreshAt).Seconds())
	if refreshIn < 0 {
		refreshIn = 0
	}

	return trendingthreaddto.TrendingThreadListResponse{
		Data: data,
		Meta: trendingthreaddto.TrendingThreadMeta{
			Total:            int64(len(data)),
			CachedAt:         result.CachedAt,
			RefreshAt:        result.RefreshAt,
			RefreshInSeconds: refreshIn,
		},
	}
}
