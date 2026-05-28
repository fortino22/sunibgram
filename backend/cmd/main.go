package main

import (
	"github.com/sunibgram/cmd/bootstrap"
)

func main() {
	db := bootstrap.NewDatabase()
	redisCache := bootstrap.NewCache()

	bootstrap.StartWeeklySchedulers(db.Write)
	bootstrap.StartFeedScoreScheduler(db.Write, redisCache)

	ext := bootstrap.NewExternal(redisCache)
	bootstrap.StartEmailWorkers(ext.Email)

	repos := bootstrap.NewRepositories(db, redisCache)
	bootstrap.StartTrendingUserSnapshotScheduler(repos)
	svcs := bootstrap.NewServices(repos, ext, redisCache)
	ctrls := bootstrap.NewControllers(svcs, ext, redisCache)

	app := bootstrap.NewServer(ctrls, svcs, ext, redisCache)
	bootstrap.StartSubscriptionExpiry(svcs.PremiumSubscription)
	bootstrap.StartVerificationCodePurge(svcs.Verification)

	if err := bootstrap.Listen(app); err != nil {
		panic(err)
	}
}
