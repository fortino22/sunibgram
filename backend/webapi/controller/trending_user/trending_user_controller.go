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

func (c *TrendingUserController) GetMyCurrent(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	result, err := c.service.GetMyCurrent(ctx.Context(), claims.UserID)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}

	resp := trendinguserdto.MyTrendingCurrentResponse{
		Rank: result.Rank,
		Meta: buildMeta(result.Page, 1, 0),
	}
	if result.User != nil {
		u := result.User
		resp.UserID = u.UserID
		resp.Username = u.Username
		resp.Name = u.Name
		resp.Profile = ctrlshared.BuildMediaURL(u.AvatarObjectName, c.profileService.GetURL)
		resp.IsPremium = u.IsPremium
		resp.Score = u.Score
		resp.Detail = trendinguserdto.TrendingUserScoreDetail{
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
		}
	} else {
		resp.UserID = claims.UserID
	}

	return ctx.Status(fiber.StatusOK).JSON(resp)
}

func (c *TrendingUserController) GetMyHistory(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}

	year := ctx.QueryInt("year", 0)
	month := ctx.QueryInt("month", 0)

	result, err := c.service.GetMyHistory(ctx.Context(), claims.UserID, year, month)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}

	rows := result.Rows
	data := make([]trendinguserdto.MyTrendingHistoryItem, len(rows))
	for i, r := range rows {
		data[i] = trendinguserdto.MyTrendingHistoryItem{
			SnapshotID: r.SnapshotID,
			WeekStart:  r.WeekStartDate.Format("2006-01-02"),
			Rank:       r.Rank,
			Score:      r.Score,
			Detail: trendinguserdto.TrendingUserScoreDetail{
				Received: trendinguserdto.TrendingUserReceivedDetail{
					Likes:       r.LikesIn,
					Comments:    r.CommentsIn,
					Bookmarks:   r.BookmarksIn,
					Reposts:     r.RepostsIn,
					ThreadCount: r.ThreadCount,
				},
				Given: trendinguserdto.TrendingUserGivenDetail{
					Likes:     r.LikesOut,
					Bookmarks: r.BookmarksOut,
					Comments:  r.CommentsOut,
					Posts:     r.PostsOut,
				},
			},
		}
	}
	months := make([]trendinguserdto.MyTrendingHistoryMonth, len(result.AvailableMonths))
	for i, m := range result.AvailableMonths {
		months[i] = trendinguserdto.MyTrendingHistoryMonth{Year: m.Year, Month: m.Month}
	}
	respYear, respMonth := year, month
	if (respYear == 0 || respMonth == 0) && len(months) > 0 {
		respYear, respMonth = months[0].Year, months[0].Month
	}
	return ctx.Status(fiber.StatusOK).JSON(trendinguserdto.MyTrendingHistoryResponse{
		Data: data,
		Meta: trendinguserdto.MyTrendingHistoryMeta{
			Year:            respYear,
			Month:           respMonth,
			AvailableMonths: months,
		},
	})
}

func (c *TrendingUserController) GetMySnapshot(ctx *fiber.Ctx) error {
	claims, ok := ctrlshared.RequireAuth(ctx)
	if !ok {
		return nil
	}
	id, ok2 := ctrlshared.ParseIDParam(ctx, "snapshotId", "snapshot")
	if !ok2 {
		return nil
	}

	r, err := c.service.GetSnapshotByID(ctx.Context(), claims.UserID, id)
	if err != nil {
		return ctrlshared.HandleServiceError(ctx, err)
	}

	resp := trendinguserdto.MyTrendingSnapshotResponse{
		SnapshotID: r.SnapshotID,
		WeekStart:  r.WeekStartDate.Format("2006-01-02"),
		Rank:       r.Rank,
		Score:      r.Score,
		SnapshotAt: r.SnapshotAt,
		Detail: trendinguserdto.TrendingUserScoreDetail{
			Received: trendinguserdto.TrendingUserReceivedDetail{
				Likes:       r.LikesIn,
				Comments:    r.CommentsIn,
				Bookmarks:   r.BookmarksIn,
				Reposts:     r.RepostsIn,
				ThreadCount: r.ThreadCount,
			},
			Given: trendinguserdto.TrendingUserGivenDetail{
				Likes:     r.LikesOut,
				Bookmarks: r.BookmarksOut,
				Comments:  r.CommentsOut,
				Posts:     r.PostsOut,
			},
		},
	}
	return ctx.Status(fiber.StatusOK).JSON(resp)
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
	return trendinguserdto.PaginatedTrendingUserResponse{
		Data: data,
		Meta: buildMeta(result, page, limit),
	}
}

func buildMeta(result trendinguserrepoiface.TrendingUsersPage, page, limit int) trendinguserdto.TrendingUserMeta {
	refreshIn := int64(time.Until(result.RefreshAt).Seconds())
	if refreshIn < 0 {
		refreshIn = 0
	}
	return trendinguserdto.TrendingUserMeta{
		Total:            result.Total,
		Page:             page,
		Limit:            limit,
		CachedAt:         result.CachedAt,
		RefreshAt:        result.RefreshAt,
		RefreshInSeconds: refreshIn,
	}
}
