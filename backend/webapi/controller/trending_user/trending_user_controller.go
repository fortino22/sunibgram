package trendinguserctrl

import (
	"time"

	"github.com/gofiber/fiber/v2"

	trendinguserrepoiface "github.com/sunibgram/domain/repository/trending_user"
	trendingusersvciface "github.com/sunibgram/domain/service/trending_user"
	userprofilesvciface "github.com/sunibgram/domain/service/user_profile"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	trendinguserdto "github.com/sunibgram/webapi/dto/trending_user"
)

type TrendingUserController struct {
	service        trendingusersvciface.ITrendingUserService
	profileService userprofilesvciface.IUserProfileService
}

func NewTrendingUserController(
	service trendingusersvciface.ITrendingUserService,
	profileService userprofilesvciface.IUserProfileService,
) *TrendingUserController {
	return &TrendingUserController{service: service, profileService: profileService}
}

func (c *TrendingUserController) GetWeeklyTrending(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	page, limit := ctrlshared.ParsePagination(ctx)

	result, err := c.service.GetWeeklyTrending(ctx.Context(), claims.UserID, page, limit)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(c.toResponse(result, page, limit))
}

func (c *TrendingUserController) toResponse(result trendinguserrepoiface.TrendingUsersPage, page, limit int) trendinguserdto.PaginatedTrendingUserResponse {
	data := make([]trendinguserdto.TrendingUserResponse, len(result.Users))
	for i, u := range result.Users {
		data[i] = trendinguserdto.TrendingUserResponse{
			UserID:      u.UserID,
			Username:    u.Username,
			Name:        u.Name,
			Profile:     ctrlshared.BuildMediaURL(u.AvatarObjectName, c.profileService.GetURL),
			IsPremium:   u.IsPremium,
			IsFollowing: u.IsFollowing,
			Score:       u.Score,
			Detail: trendinguserdto.TrendingUserScoreDetail{
				Received: trendinguserdto.TrendingUserReceivedDetail{
					Likes:       u.LikesIn,
					Comments:    u.CommentsIn,
					Bookmarks:   u.BookmarksIn,
					Reposts:     u.RepostsIn,
					ThreadCount: u.ThreadCount,
				},
				Given: trendinguserdto.TrendingUserGivenDetail{
					Likes:     u.LikesOut,
					Bookmarks: u.BookmarksOut,
					Comments:  u.CommentsOut,
					Posts:     u.PostsOut,
				},
			},
		}
	}
	refreshIn := int64(time.Until(result.RefreshAt).Seconds())
	if refreshIn < 0 {
		refreshIn = 0
	}
	return trendinguserdto.PaginatedTrendingUserResponse{
		Data: data,
		Meta: trendinguserdto.TrendingUserMeta{
			Total:            result.Total,
			Page:             page,
			Limit:            limit,
			CachedAt:         result.CachedAt,
			RefreshAt:        result.RefreshAt,
			RefreshInSeconds: refreshIn,
		},
	}
}
