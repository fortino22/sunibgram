package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"

	authuc "github.com/sunibgram/application/usecase/auth"
	bookmarkuc "github.com/sunibgram/application/usecase/bookmark"
	cacheuc "github.com/sunibgram/application/usecase/cache"
	categoryuc "github.com/sunibgram/application/usecase/category"
	chatuc "github.com/sunibgram/application/usecase/chat"
	chatmediauc "github.com/sunibgram/application/usecase/chat_media"
	chatreactionuc "github.com/sunibgram/application/usecase/chat_reaction"
	chatroomuc "github.com/sunibgram/application/usecase/chat_room"
	communityuc "github.com/sunibgram/application/usecase/community"
	communitypinuc "github.com/sunibgram/application/usecase/community_pin"
	communityrequestuc "github.com/sunibgram/application/usecase/community_request"
	communityruleuc "github.com/sunibgram/application/usecase/community_rule"
	communitythreadpinuc "github.com/sunibgram/application/usecase/community_thread_pin"
	emailuc "github.com/sunibgram/application/usecase/email"
	eventuc "github.com/sunibgram/application/usecase/event"
	eventparticipantuc "github.com/sunibgram/application/usecase/event_participant"
	eventpinuc "github.com/sunibgram/application/usecase/event_pin"
	eventthreadpinuc "github.com/sunibgram/application/usecase/event_thread_pin"
	exploreuc "github.com/sunibgram/application/usecase/explore"
	facultyuc "github.com/sunibgram/application/usecase/faculty"
	followuc "github.com/sunibgram/application/usecase/follow"
	hashtaguc "github.com/sunibgram/application/usecase/hashtag"
	likeuc "github.com/sunibgram/application/usecase/like"
	notificationuc "github.com/sunibgram/application/usecase/notification"
	orderuc "github.com/sunibgram/application/usecase/order"
	premiumsubscriptionuc "github.com/sunibgram/application/usecase/premium_subscription"
	reportuc "github.com/sunibgram/application/usecase/report"
	roleuc "github.com/sunibgram/application/usecase/role"
	storageuc "github.com/sunibgram/application/usecase/storage"
	threaduc "github.com/sunibgram/application/usecase/thread"
	threadmediauc "github.com/sunibgram/application/usecase/thread_media"
	threadpinuc "github.com/sunibgram/application/usecase/thread_pin"
	threadpolluc "github.com/sunibgram/application/usecase/thread_poll"
	threadsummaryuc "github.com/sunibgram/application/usecase/thread_summary"
	useruc "github.com/sunibgram/application/usecase/user"
	userprofileuc "github.com/sunibgram/application/usecase/user_profile"
	usersettingsuc "github.com/sunibgram/application/usecase/user_settings"
	verificationuc "github.com/sunibgram/application/usecase/verification"
	violationuc "github.com/sunibgram/application/usecase/violation"
	"github.com/sunibgram/infrastructure/cache"
	"github.com/sunibgram/infrastructure/config"
	infraemail "github.com/sunibgram/infrastructure/email"
	"github.com/sunibgram/infrastructure/migration"
	"github.com/sunibgram/infrastructure/ollama"
	bookmarkrepo "github.com/sunibgram/infrastructure/repository/bookmark"
	cacherepo "github.com/sunibgram/infrastructure/repository/cache"
	categoryrepo "github.com/sunibgram/infrastructure/repository/category"
	chatrepo "github.com/sunibgram/infrastructure/repository/chat"
	chatmediarepo "github.com/sunibgram/infrastructure/repository/chat_media"
	chatreactionrepo "github.com/sunibgram/infrastructure/repository/chat_reaction"
	chatreadrepo "github.com/sunibgram/infrastructure/repository/chat_read"
	chatroomrepo "github.com/sunibgram/infrastructure/repository/chat_room"
	communityrepo "github.com/sunibgram/infrastructure/repository/community"
	communitymemberrepo "github.com/sunibgram/infrastructure/repository/community_member"
	communitypinrepo "github.com/sunibgram/infrastructure/repository/community_pin"
	communityrequestrepo "github.com/sunibgram/infrastructure/repository/community_request"
	communityrulerepo "github.com/sunibgram/infrastructure/repository/community_rule"
	communitythreadpinrepo "github.com/sunibgram/infrastructure/repository/community_thread_pin"
	eventrepo "github.com/sunibgram/infrastructure/repository/event"
	eventparticipantrepo "github.com/sunibgram/infrastructure/repository/event_participant"
	eventpinrepo "github.com/sunibgram/infrastructure/repository/event_pin"
	eventreviewrepo "github.com/sunibgram/infrastructure/repository/event_review"
	eventthreadpinrepo "github.com/sunibgram/infrastructure/repository/event_thread_pin"
	explorerepo "github.com/sunibgram/infrastructure/repository/explore"
	facultyrepo "github.com/sunibgram/infrastructure/repository/faculty"
	followrepo "github.com/sunibgram/infrastructure/repository/follow"
	hashtagrepo "github.com/sunibgram/infrastructure/repository/hashtag"
	likerepo "github.com/sunibgram/infrastructure/repository/like"
	notificationrepo "github.com/sunibgram/infrastructure/repository/notification"
	orderrepo "github.com/sunibgram/infrastructure/repository/order"
	premiumsubscriptionrepo "github.com/sunibgram/infrastructure/repository/premium_subscription"
	reportrepo "github.com/sunibgram/infrastructure/repository/report"
	rolerepo "github.com/sunibgram/infrastructure/repository/role"
	threadrepo "github.com/sunibgram/infrastructure/repository/thread"
	threadmediarepo "github.com/sunibgram/infrastructure/repository/thread_media"
	threadpinrepo "github.com/sunibgram/infrastructure/repository/thread_pin"
	threadpollrepo "github.com/sunibgram/infrastructure/repository/thread_poll"
	userrepo "github.com/sunibgram/infrastructure/repository/user"
	userfacultyrepo "github.com/sunibgram/infrastructure/repository/user_faculty"
	userprofilerepo "github.com/sunibgram/infrastructure/repository/user_profile"
	usersettingsrepo "github.com/sunibgram/infrastructure/repository/user_settings"
	verificationrepo "github.com/sunibgram/infrastructure/repository/verification"
	violationrepo "github.com/sunibgram/infrastructure/repository/violation"
	"github.com/sunibgram/infrastructure/scheduler"
	infrasocket "github.com/sunibgram/infrastructure/socket"
	"github.com/sunibgram/infrastructure/storage"
	ctrlshared "github.com/sunibgram/webapi/controller/_shared"
	authctrl "github.com/sunibgram/webapi/controller/auth"
	bookmarkctrl "github.com/sunibgram/webapi/controller/bookmark"
	cachectrl "github.com/sunibgram/webapi/controller/cache"
	categoryctrl "github.com/sunibgram/webapi/controller/category"
	chatctrl "github.com/sunibgram/webapi/controller/chat"
	chatroomctrl "github.com/sunibgram/webapi/controller/chat_room"
	communityctrl "github.com/sunibgram/webapi/controller/community"
	communitypinctrl "github.com/sunibgram/webapi/controller/community_pin"
	communityrequestctrl "github.com/sunibgram/webapi/controller/community_request"
	communityrulectrl "github.com/sunibgram/webapi/controller/community_rule"
	communitythreadpinctrl "github.com/sunibgram/webapi/controller/community_thread_pin"
	emailctrl "github.com/sunibgram/webapi/controller/email"
	eventctrl "github.com/sunibgram/webapi/controller/event"
	eventparticipantctrl "github.com/sunibgram/webapi/controller/event_participant"
	eventpinctrl "github.com/sunibgram/webapi/controller/event_pin"
	eventthreadpinctrl "github.com/sunibgram/webapi/controller/event_thread_pin"
	explorectrl "github.com/sunibgram/webapi/controller/explore"
	facultyctrl "github.com/sunibgram/webapi/controller/faculty"
	followctrl "github.com/sunibgram/webapi/controller/follow"
	hashtagctrl "github.com/sunibgram/webapi/controller/hashtag"
	likectrl "github.com/sunibgram/webapi/controller/like"
	notificationctrl "github.com/sunibgram/webapi/controller/notification"
	orderctrl "github.com/sunibgram/webapi/controller/order"
	premiumsubscriptionctrl "github.com/sunibgram/webapi/controller/premium_subscription"
	reportctrl "github.com/sunibgram/webapi/controller/report"
	repostctrl "github.com/sunibgram/webapi/controller/repost"
	rolectrl "github.com/sunibgram/webapi/controller/role"
	threadctrl "github.com/sunibgram/webapi/controller/thread"
	threadpinctrl "github.com/sunibgram/webapi/controller/thread_pin"
	threadpollctrl "github.com/sunibgram/webapi/controller/thread_poll"
	threadsummaryctrl "github.com/sunibgram/webapi/controller/thread_summary"
	userctrl "github.com/sunibgram/webapi/controller/user"
	userprofilectrl "github.com/sunibgram/webapi/controller/user_profile"
	usersettingsctrl "github.com/sunibgram/webapi/controller/user_settings"
	verificationctrl "github.com/sunibgram/webapi/controller/verification"
	violationctrl "github.com/sunibgram/webapi/controller/violation"
	"github.com/sunibgram/webapi/middleware"
	"github.com/sunibgram/webapi/router"
	webapisocket "github.com/sunibgram/webapi/socket"
)

func main() {
	// --- Database ---
	writeDB := config.NewPostgresWrite()
	readDB := config.NewPostgresRead()
	migration.AutoMigrate(writeDB)
	go scheduler.ResetWeeklyEventCounters(writeDB)
	go scheduler.ResetWeeklyHashtagCounters(writeDB)

	// --- Cache ---
	redisCache := cache.New()
	if err := redisCache.DeleteByPattern(context.Background(), "community:members:*"); err != nil {
		log.Printf("community members cache startup flush error: %v", err)
	}
	middleware.Init(redisCache)
	ctrlshared.SetFeedBuilderCache(redisCache)
	scheduler.NewFeedScoreScheduler(writeDB, redisCache).Start()

	// --- Repositories ---
	cacheRepo := cacherepo.NewCacheRepository(redisCache)
	roleRepo := rolerepo.NewRoleRepository(writeDB, readDB)
	userRepo := userrepo.NewUserRepository(writeDB, readDB, redisCache)
	userSettingsRepo := usersettingsrepo.NewUserSettingsRepository(writeDB, readDB, redisCache)
	profileRepo := userprofilerepo.NewUserProfileRepository(writeDB, readDB, redisCache)
	verificationRepo := verificationrepo.NewVerificationRepository(writeDB, readDB)
	followRepo := followrepo.NewFollowRepository(writeDB, readDB, redisCache)
	threadRepo, patchCount := threadrepo.NewThreadRepository(writeDB, readDB, redisCache)
	eventRepo := eventrepo.NewEventRepository(writeDB, readDB, redisCache)
	eventReviewRepo := eventreviewrepo.NewEventReviewRepository(writeDB, readDB)
	hashtagRepo := hashtagrepo.NewHashtagRepository(writeDB, readDB)
	exploreRepo := explorerepo.NewExploreRepository(readDB, redisCache)
	threadMediaRepo := threadmediarepo.NewThreadMediaRepository(writeDB, readDB, redisCache)
	likeRepo := likerepo.NewLikeRepository(writeDB, readDB, redisCache, patchCount)
	bookmarkRepo := bookmarkrepo.NewBookmarkRepository(writeDB, readDB, redisCache, patchCount)
	threadPinRepo := threadpinrepo.NewThreadPinRepository(writeDB, readDB, redisCache)
	categoryRepo := categoryrepo.NewCategoryRepository(writeDB, readDB, redisCache)
	violationRepo := violationrepo.NewViolationRepository(writeDB, readDB)
	reportRepo := reportrepo.NewReportRepository(writeDB, readDB)
	communityRepo := communityrepo.NewCommunityRepository(writeDB, readDB, redisCache)
	communityMemberRepo := communitymemberrepo.NewCommunityMemberRepository(writeDB, readDB, redisCache)
	communityRequestRepo := communityrequestrepo.NewCommunityRequestRepository(writeDB, readDB, redisCache)
	communityRuleRepo := communityrulerepo.NewCommunityRuleRepository(writeDB, readDB)
	communityThreadPinRepo := communitythreadpinrepo.NewCommunityThreadPinRepository(writeDB, readDB)
	communityPinRepo := communitypinrepo.NewCommunityPinRepository(writeDB, readDB, redisCache)
	eventPinRepo := eventpinrepo.NewEventPinRepository(writeDB, readDB, redisCache)
	eventParticipantRepo := eventparticipantrepo.NewEventParticipantRepository(writeDB, readDB, redisCache)
	facultyRepo := facultyrepo.NewFacultyRepository(writeDB, readDB, redisCache)
	userFacultyRepo := userfacultyrepo.NewUserFacultyRepository(writeDB, readDB, redisCache)
	orderRepo := orderrepo.NewOrderRepository(writeDB, readDB)
	premiumSubscriptionRepo := premiumsubscriptionrepo.NewPremiumSubscriptionRepository(writeDB, readDB)
	chatRoomRepo := chatroomrepo.NewChatRoomRepository(writeDB, readDB, redisCache)
	chatRepo := chatrepo.NewChatRepository(writeDB, readDB)
	chatMediaRepo := chatmediarepo.NewChatMediaRepository(writeDB, readDB)
	chatReadRepo := chatreadrepo.NewChatReadRepository(writeDB, readDB)
	chatReactionRepo := chatreactionrepo.NewChatReactionRepository(writeDB, readDB)
	notificationRepo := notificationrepo.NewNotificationRepository(writeDB, readDB, redisCache)
	threadPollRepo := threadpollrepo.NewThreadPollRepository(writeDB, readDB, redisCache)
	eventThreadPinRepo := eventthreadpinrepo.NewEventThreadPinRepository(writeDB, readDB, redisCache)

	// --- External Services ---
	storageUsecase := storageuc.NewStorageUsecase(storage.NewMinioClient())
	emailService := emailuc.NewEmailService(infraemail.NewSMTPClient(), redisCache)
	const emailWorkerCount = 4
	for i := 0; i < emailWorkerCount; i++ {
		go emailService.StartWorker(context.Background())
	}
	ollamaClient := ollama.NewClient()
	socketHub := infrasocket.NewHub()

	// --- Usecases ---
	cacheService := cacheuc.NewCacheService(cacheRepo)
	authService := authuc.NewAuthService(userRepo, roleRepo, verificationRepo, userSettingsRepo, redisCache)
	roleService := roleuc.NewRoleService(roleRepo)
	userService := useruc.NewUserService(userRepo, followRepo, profileRepo)
	userSettingsService := usersettingsuc.NewUserSettingsService(userSettingsRepo, threadRepo)
	profileService := userprofileuc.NewUserProfileService(userRepo, followRepo, profileRepo, userFacultyRepo, cacheRepo, storageUsecase)
	verificationService := verificationuc.NewVerificationService(verificationRepo, userRepo, emailService)
	notificationService := notificationuc.NewNotificationService(notificationRepo, userRepo, profileRepo, userFacultyRepo, storageUsecase, socketHub)
	followService := followuc.NewFollowService(followRepo, userRepo, cacheRepo, notificationService)
	hashtagService := hashtaguc.NewHashtagService(hashtagRepo, redisCache)
	summaryService := threadsummaryuc.NewThreadSummaryService(threadRepo, cacheRepo, ollamaClient)
	threadService := threaduc.NewThreadService(threadRepo, hashtagRepo, eventRepo, eventParticipantRepo, hashtagService, communityRepo, communityMemberRepo, userSettingsRepo, followRepo, notificationService, summaryService)
	eventService := eventuc.NewEventService(eventRepo, eventReviewRepo, eventPinRepo, eventParticipantRepo, storageUsecase, userSettingsRepo, followRepo, threadRepo)
	exploreService := exploreuc.NewExploreService(exploreRepo, threadRepo, userSettingsRepo, followRepo)
	threadMediaService := threadmediauc.NewThreadMediaService(threadRepo, threadMediaRepo, storageUsecase)
	likeService := likeuc.NewLikeService(likeRepo, threadRepo, communityRepo, communityMemberRepo, userSettingsRepo, followRepo, notificationService)
	bookmarkService := bookmarkuc.NewBookmarkService(bookmarkRepo, threadRepo, communityRepo, communityMemberRepo, userSettingsRepo, followRepo, notificationService)
	threadPinService := threadpinuc.NewThreadPinService(threadPinRepo, threadRepo)
	categoryService := categoryuc.NewCategoryService(categoryRepo)
	violationService := violationuc.NewViolationService(violationRepo)
	reportService := reportuc.NewReportService(reportRepo, violationRepo, userRepo, threadRepo, profileService, cacheRepo)
	communityService := communityuc.NewCommunityService(communityRepo, communityMemberRepo, userRepo, profileRepo, followRepo, categoryRepo, communityRuleRepo, communityPinRepo, cacheRepo, storageUsecase)
	communityMemberService := communityuc.NewCommunityMemberService(communityRepo, communityMemberRepo, followRepo, cacheRepo)
	communityRequestService := communityrequestuc.NewCommunityRequestService(communityRepo, communityMemberRepo, communityRequestRepo, cacheRepo)
	communityRuleService := communityruleuc.NewCommunityRuleService(communityRepo, communityRuleRepo)
	communityThreadPinService := communitythreadpinuc.NewCommunityThreadPinService(communityThreadPinRepo, communityRepo, communityMemberRepo, threadRepo)
	communityPinService := communitypinuc.NewCommunityPinService(communityPinRepo, communityRepo, communityMemberRepo)
	eventPinService := eventpinuc.NewEventPinService(eventPinRepo, eventRepo)
	eventParticipantService := eventparticipantuc.NewEventParticipantService(eventParticipantRepo, eventRepo)
	facultyService := facultyuc.NewFacultyService(facultyRepo, userFacultyRepo)
	orderService := orderuc.NewOrderService(orderRepo, userRepo, premiumSubscriptionRepo, communityRepo, cacheRepo)
	premiumSubscriptionService := premiumsubscriptionuc.NewPremiumSubscriptionService(userRepo, premiumSubscriptionRepo, communityRepo, cacheRepo)
	chatRoomService := chatroomuc.NewChatRoomService(chatRoomRepo, userRepo, storageUsecase)
	chatService := chatuc.NewChatService(chatRepo, chatRoomRepo, chatReadRepo, notificationService)
	chatMediaService := chatmediauc.NewChatMediaService(chatRepo, chatMediaRepo, storageUsecase)
	chatReactionService := chatreactionuc.NewChatReactionService(chatRepo, chatReactionRepo, chatRoomRepo)
	threadPollService := threadpolluc.NewThreadPollService(threadRepo, threadPollRepo, storageUsecase)
	eventThreadPinService := eventthreadpinuc.NewEventThreadPinService(eventThreadPinRepo, eventRepo, threadRepo)

	// --- Controllers ---
	cacheCtrl := cachectrl.NewCacheController(cacheService)
	authCtrl := authctrl.NewAuthController(authService, profileService, facultyService)
	roleCtrl := rolectrl.NewRoleController(roleService)
	userCtrl := userctrl.NewUserController(userService, profileService)
	userSettingsCtrl := usersettingsctrl.NewUserSettingsController(userSettingsService)
	profileCtrl := userprofilectrl.NewUserProfileController(profileService, facultyService, socketHub, chatRoomService)
	emailCtrl := emailctrl.NewEmailController(emailService)
	verificationCtrl := verificationctrl.NewVerificationController(verificationService)
	followCtrl := followctrl.NewFollowController(followService, profileService)
	threadCtrl := threadctrl.NewThreadController(threadService, threadMediaService, communityMemberService, likeService, bookmarkService, threadPinService, profileService, threadPollService)
	eventCtrl := eventctrl.NewEventController(eventService, threadService, threadMediaService, likeService, bookmarkService, threadPinService, profileService, threadPollService, eventThreadPinService)
	eventThreadPinCtrl := eventthreadpinctrl.NewEventThreadPinController(eventThreadPinService)
	hashtagCtrl := hashtagctrl.NewHashtagController(hashtagService)
	exploreCtrl := explorectrl.NewExploreController(exploreService, profileService, threadMediaService, likeService, bookmarkService, threadService, communityService, threadPollService)
	likeCtrl := likectrl.NewLikeController(likeService, bookmarkService, threadService, threadMediaService, threadPollService, profileService)
	bookmarkCtrl := bookmarkctrl.NewBookmarkController(bookmarkService, likeService, threadService, threadMediaService, threadPollService, profileService)
	threadPinCtrl := threadpinctrl.NewThreadPinController(threadPinService)
	repostCtrl := repostctrl.NewRepostController(threadService, threadMediaService)
	summaryCtrl := threadsummaryctrl.NewThreadSummaryController(summaryService)
	categoryCtrl := categoryctrl.NewCategoryController(categoryService)
	violationCtrl := violationctrl.NewViolationController(violationService)
	reportCtrl := reportctrl.NewReportController(reportService)
	communityCtrl := communityctrl.NewCommunityController(communityService, communityRuleService, profileService)
	communityMemberCtrl := communityctrl.NewCommunityMemberController(communityMemberService, profileService)
	communityRequestCtrl := communityrequestctrl.NewCommunityRequestController(communityRequestService, profileService)
	communityRuleCtrl := communityrulectrl.NewCommunityRuleController(communityRuleService)
	communityThreadPinCtrl := communitythreadpinctrl.NewCommunityThreadPinController(communityThreadPinService)
	communityPinCtrl := communitypinctrl.NewCommunityPinController(communityPinService, communityService)
	eventPinCtrl := eventpinctrl.NewEventPinController(eventPinService, eventService)
	eventParticipantCtrl := eventparticipantctrl.NewEventParticipantController(eventParticipantService, eventService)
	communityThreadCtrl := communityctrl.NewCommunityThreadController(
		threadService,
		threadMediaService,
		likeService,
		bookmarkService,
		threadPinService,
		profileService,
		threadPollService,
		communityThreadPinService,
	)
	facultyCtrl := facultyctrl.NewFacultyController(facultyService)
	orderCtrl := orderctrl.NewOrderController(orderService)
	premiumSubscriptionCtrl := premiumsubscriptionctrl.NewPremiumSubscriptionController(premiumSubscriptionService)
	devSubscriptionCtrl := premiumsubscriptionctrl.NewDevSubscriptionController(premiumSubscriptionService)
	chatRoomCtrl := chatroomctrl.NewChatRoomController(chatRoomService, socketHub)
	chatCtrl := chatctrl.NewChatController(chatService, chatMediaService, chatReactionService, socketHub)
	notificationCtrl := notificationctrl.NewNotificationController(notificationService)
	threadPollCtrl := threadpollctrl.NewThreadPollController(threadPollService)
	notificationHandler := webapisocket.NewNotificationHandler(socketHub, notificationService)

	// --- Server ---
	app := fiber.New(fiber.Config{
		BodyLimit: 50 * 1024 * 1024,
	})
	app.Use(compress.New(compress.Config{Level: compress.LevelBestSpeed}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		AllowCredentials: true,
	}))
	webapisocket.Mount(app, socketHub, chatRoomService, notificationHandler, redisCache)

	router.Register(
		app,
		authCtrl,
		roleCtrl,
		userCtrl,
		cacheCtrl,
		userSettingsCtrl,
		profileCtrl,
		exploreCtrl,
		hashtagCtrl,
		threadCtrl,
		eventCtrl,
		likeCtrl,
		bookmarkCtrl,
		threadPinCtrl,
		repostCtrl,
		summaryCtrl,
		emailCtrl,
		verificationCtrl,
		followCtrl,
		communityCtrl,
		communityMemberCtrl,
		communityRequestCtrl,
		communityThreadCtrl,
		communityRuleCtrl,
		communityThreadPinCtrl,
		communityPinCtrl,
		eventPinCtrl,
		eventParticipantCtrl,
		eventThreadPinCtrl,
		categoryCtrl,
		violationCtrl,
		reportCtrl,
		facultyCtrl,
		orderCtrl,
		premiumSubscriptionCtrl,
		devSubscriptionCtrl,
		chatRoomCtrl,
		chatCtrl,
		notificationCtrl,
		threadPollCtrl,
	)

	go scheduler.RunSubscriptionExpiry(context.Background(), premiumSubscriptionService)

	addr := fmt.Sprintf(":%s", getEnv("PORT", "8080"))
	fmt.Printf("server running on %s\n", addr)
	if err := app.Listen(addr); err != nil {
		panic(err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
