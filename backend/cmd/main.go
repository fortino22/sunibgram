package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"github.com/sunibgram/application/usecase"
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
	"github.com/sunibgram/webapi/controller"
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
	controller.SetFeedBuilderCache(redisCache)
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
	storageUsecase := usecase.NewStorageUsecase(storage.NewMinioClient())
	emailService := usecase.NewEmailService(infraemail.NewSMTPClient(), redisCache)
	const emailWorkerCount = 4
	for i := 0; i < emailWorkerCount; i++ {
		go emailService.StartWorker(context.Background())
	}
	ollamaClient := ollama.NewClient()
	socketHub := infrasocket.NewHub()

	// --- Usecases ---
	cacheService := usecase.NewCacheService(cacheRepo)
	authService := usecase.NewAuthService(userRepo, roleRepo, verificationRepo, userSettingsRepo, redisCache)
	roleService := usecase.NewRoleService(roleRepo)
	userService := usecase.NewUserService(userRepo, followRepo, profileRepo)
	userSettingsService := usecase.NewUserSettingsService(userSettingsRepo, threadRepo)
	profileService := usecase.NewUserProfileService(userRepo, followRepo, profileRepo, userFacultyRepo, cacheRepo, storageUsecase)
	verificationService := usecase.NewVerificationService(verificationRepo, userRepo, emailService)
	notificationService := usecase.NewNotificationService(notificationRepo, userRepo, profileRepo, userFacultyRepo, storageUsecase, socketHub)
	followService := usecase.NewFollowService(followRepo, userRepo, cacheRepo, notificationService)
	hashtagService := usecase.NewHashtagService(hashtagRepo, redisCache)
	summaryService := usecase.NewThreadSummaryService(threadRepo, cacheRepo, ollamaClient)
	threadService := usecase.NewThreadService(threadRepo, hashtagRepo, eventRepo, eventParticipantRepo, hashtagService, communityRepo, communityMemberRepo, userSettingsRepo, followRepo, notificationService, summaryService)
	eventService := usecase.NewEventService(eventRepo, eventReviewRepo, eventPinRepo, eventParticipantRepo, storageUsecase, userSettingsRepo, followRepo, threadRepo)
	exploreService := usecase.NewExploreService(exploreRepo, threadRepo, userSettingsRepo, followRepo)
	threadMediaService := usecase.NewThreadMediaService(threadRepo, threadMediaRepo, storageUsecase)
	likeService := usecase.NewLikeService(likeRepo, threadRepo, communityRepo, communityMemberRepo, userSettingsRepo, followRepo, notificationService)
	bookmarkService := usecase.NewBookmarkService(bookmarkRepo, threadRepo, communityRepo, communityMemberRepo, userSettingsRepo, followRepo, notificationService)
	threadPinService := usecase.NewThreadPinService(threadPinRepo, threadRepo)
	categoryService := usecase.NewCategoryService(categoryRepo)
	violationService := usecase.NewViolationService(violationRepo)
	reportService := usecase.NewReportService(reportRepo, violationRepo, userRepo, threadRepo, profileService, cacheRepo)
	communityService := usecase.NewCommunityService(communityRepo, communityMemberRepo, userRepo, profileRepo, followRepo, categoryRepo, communityRuleRepo, communityPinRepo, cacheRepo, storageUsecase)
	communityMemberService := usecase.NewCommunityMemberService(communityRepo, communityMemberRepo, followRepo, cacheRepo)
	communityRequestService := usecase.NewCommunityRequestService(communityRepo, communityMemberRepo, communityRequestRepo, cacheRepo)
	communityRuleService := usecase.NewCommunityRuleService(communityRepo, communityRuleRepo)
	communityThreadPinService := usecase.NewCommunityThreadPinService(communityThreadPinRepo, communityRepo, communityMemberRepo, threadRepo)
	communityPinService := usecase.NewCommunityPinService(communityPinRepo, communityRepo, communityMemberRepo)
	eventPinService := usecase.NewEventPinService(eventPinRepo, eventRepo)
	eventParticipantService := usecase.NewEventParticipantService(eventParticipantRepo, eventRepo)
	facultyService := usecase.NewFacultyService(facultyRepo, userFacultyRepo)
	orderService := usecase.NewOrderService(orderRepo, userRepo, premiumSubscriptionRepo, communityRepo, cacheRepo)
	premiumSubscriptionService := usecase.NewPremiumSubscriptionService(userRepo, premiumSubscriptionRepo, communityRepo, cacheRepo)
	chatRoomService := usecase.NewChatRoomService(chatRoomRepo, userRepo, storageUsecase)
	chatService := usecase.NewChatService(chatRepo, chatRoomRepo, chatReadRepo, notificationService)
	chatMediaService := usecase.NewChatMediaService(chatRepo, chatMediaRepo, storageUsecase)
	chatReactionService := usecase.NewChatReactionService(chatRepo, chatReactionRepo, chatRoomRepo)
	threadPollService := usecase.NewThreadPollService(threadRepo, threadPollRepo, storageUsecase)
	eventThreadPinService := usecase.NewEventThreadPinService(eventThreadPinRepo, eventRepo, threadRepo)

	// --- Controllers ---
	cacheCtrl := controller.NewCacheController(cacheService)
	authCtrl := controller.NewAuthController(authService, profileService, facultyService)
	roleCtrl := controller.NewRoleController(roleService)
	userCtrl := controller.NewUserController(userService, profileService)
	userSettingsCtrl := controller.NewUserSettingsController(userSettingsService)
	profileCtrl := controller.NewUserProfileController(profileService, facultyService, socketHub, chatRoomService)
	emailCtrl := controller.NewEmailController(emailService)
	verificationCtrl := controller.NewVerificationController(verificationService)
	followCtrl := controller.NewFollowController(followService, profileService)
	threadCtrl := controller.NewThreadController(threadService, threadMediaService, communityMemberService, likeService, bookmarkService, threadPinService, profileService, threadPollService)
	eventCtrl := controller.NewEventController(eventService, threadService, threadMediaService, likeService, bookmarkService, threadPinService, profileService, threadPollService, eventThreadPinService)
	eventThreadPinCtrl := controller.NewEventThreadPinController(eventThreadPinService)
	hashtagCtrl := controller.NewHashtagController(hashtagService)
	exploreCtrl := controller.NewExploreController(exploreService, profileService, threadMediaService, likeService, bookmarkService, threadService, communityService, threadPollService)
	likeCtrl := controller.NewLikeController(likeService, bookmarkService, threadService, threadMediaService, threadPollService, profileService)
	bookmarkCtrl := controller.NewBookmarkController(bookmarkService, likeService, threadService, threadMediaService, threadPollService, profileService)
	threadPinCtrl := controller.NewThreadPinController(threadPinService)
	repostCtrl := controller.NewRepostController(threadService, threadMediaService)
	summaryCtrl := controller.NewThreadSummaryController(summaryService)
	categoryCtrl := controller.NewCategoryController(categoryService)
	violationCtrl := controller.NewViolationController(violationService)
	reportCtrl := controller.NewReportController(reportService)
	communityCtrl := controller.NewCommunityController(communityService, communityRuleService, profileService)
	communityMemberCtrl := controller.NewCommunityMemberController(communityMemberService, profileService)
	communityRequestCtrl := controller.NewCommunityRequestController(communityRequestService, profileService)
	communityRuleCtrl := controller.NewCommunityRuleController(communityRuleService)
	communityThreadPinCtrl := controller.NewCommunityThreadPinController(communityThreadPinService)
	communityPinCtrl := controller.NewCommunityPinController(communityPinService, communityService)
	eventPinCtrl := controller.NewEventPinController(eventPinService, eventService)
	eventParticipantCtrl := controller.NewEventParticipantController(eventParticipantService, eventService)
	communityThreadCtrl := controller.NewCommunityThreadController(
		threadService,
		threadMediaService,
		likeService,
		bookmarkService,
		threadPinService,
		profileService,
		threadPollService,
		communityThreadPinService,
	)
	facultyCtrl := controller.NewFacultyController(facultyService)
	orderCtrl := controller.NewOrderController(orderService)
	premiumSubscriptionCtrl := controller.NewPremiumSubscriptionController(premiumSubscriptionService)
	devSubscriptionCtrl := controller.NewDevSubscriptionController(premiumSubscriptionService)
	chatRoomCtrl := controller.NewChatRoomController(chatRoomService, socketHub)
	chatCtrl := controller.NewChatController(chatService, chatMediaService, chatReactionService, socketHub)
	notificationCtrl := controller.NewNotificationController(notificationService)
	threadPollCtrl := controller.NewThreadPollController(threadPollService)
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
