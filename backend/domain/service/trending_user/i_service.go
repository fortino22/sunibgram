package trendingusersvciface

import (
	"context"

	trendinguserrepoiface "github.com/sunibgram/domain/repository/trending_user"
)

type ITrendingUserService interface {
	GetWeeklyTrending(ctx context.Context, viewerID uint, page, limit int) (trendinguserrepoiface.TrendingUsersPage, error)
}
