package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/sunibgram/application/usecase"
	"github.com/sunibgram/infrastructure/cache"
	"github.com/sunibgram/infrastructure/config"
	infraemail "github.com/sunibgram/infrastructure/email"
	"github.com/sunibgram/infrastructure/migration"
	"github.com/sunibgram/infrastructure/ollama"
	infrarepo "github.com/sunibgram/infrastructure/repository"
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
	cacheRepo := infrarepo.NewCacheRepository(redisCache)
	roleRepo := infrarepo.NewRoleRepository(writeDB, readDB)
	userRepo := infrarepo.NewUserRepository(writeDB, readDB, redisCache)
	userSettingsRepo := infrarepo.NewUserSettingsRepository(writeDB, readDB, redisCache)
	profileRepo := infrarepo.NewUserProfileRepository(writeDB, readDB, redisCache)
	verificationRepo := infrarepo.NewVerificationRepository(writeDB, readDB)
	followRepo := infrarepo.NewFollowRepository(writeDB, readDB, redisCache)
	threadRepo, patchCount := infrarepo.NewThreadRepository(writeDB, readDB, redisCache)
	eventRepo := infrarepo.NewEventRepository(writeDB, readDB, redisCache)
	eventReviewRepo := infrarepo.NewEventReviewRepository(writeDB, readDB)
	hashtagRepo := infrarepo.NewHashtagRepository(writeDB, readDB)
	exploreRepo := infrarepo.NewExploreRepository(readDB, redisCache)
	threadMediaRepo := infrarepo.NewThreadMediaRepository(writeDB, readDB, redisCache)
	likeRepo := infrarepo.NewLikeRepository(writeDB, readDB, redisCache, patchCount)
	bookmarkRepo := infrarepo.NewBookmarkRepository(writeDB, readDB, redisCache, patchCount)
	threadPinRepo := infrarepo.NewThreadPinRepository(writeDB, readDB, redisCache)
	categoryRepo := infrarepo.NewCategoryRepository(writeDB, readDB, redisCache)
	violationRepo := infrarepo.NewViolationRepository(writeDB, readDB)
	reportRepo := infrarepo.NewReportRepository(writeDB, readDB)
	communityRepo := infrarepo.NewCommunityRepository(writeDB, readDB, redisCache)
	communityMemberRepo := infrarepo.NewCommunityMemberRepository(writeDB, readDB, redisCache)
	communityRequestRepo := infrarepo.NewCommunityRequestRepository(writeDB, readDB, redisCache)
	communityRuleRepo := infrarepo.NewCommunityRuleRepository(writeDB, readDB)
	communityThreadPinRepo := infrarepo.NewCommunityThreadPinRepository(writeDB, readDB)
	communityPinRepo := infrarepo.NewCommunityPinRepository(writeDB, readDB, redisCache)
	eventPinRepo := infrarepo.NewEventPinRepository(writeDB, readDB, redisCache)
	eventParticipantRepo := infrarepo.NewEventParticipantRepository(writeDB, readDB, redisCache)
	facultyRepo := infrarepo.NewFacultyRepository(writeDB, readDB, redisCache)
	userFacultyRepo := infrarepo.NewUserFacultyRepository(writeDB, readDB, redisCache)
	orderRepo := infrarepo.NewOrderRepository(writeDB, readDB)
	premiumSubscriptionRepo := infrarepo.NewPremiumSubscriptionRepository(writeDB, readDB)
	chatRoomRepo := infrarepo.NewChatRoomRepository(writeDB, readDB, redisCache)
	chatRepo := infrarepo.NewChatRepository(writeDB, readDB)
	chatMediaRepo := infrarepo.NewChatMediaRepository(writeDB, readDB)
	chatReadRepo := infrarepo.NewChatReadRepository(writeDB, readDB)
	chatReactionRepo := infrarepo.NewChatReactionRepository(writeDB, readDB)
	notificationRepo := infrarepo.NewNotificationRepository(writeDB, readDB, redisCache)
	threadPollRepo := infrarepo.NewThreadPollRepository(writeDB, readDB, redisCache)
	eventThreadPinRepo := infrarepo.NewEventThreadPinRepository(writeDB, readDB, redisCache)

	// --- External Services ---
	storageUsecase := usecase.NewStorageUsecase(storage.NewMinioClient())
	emailService := usecase.NewEmailService(infraemail.NewSMTPClient(), redisCache)
	go emailService.StartWorker(context.Background())
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
	likeCtrl := controller.NewLikeController(likeService, bookmarkService, threadService, threadMediaService, threadPollService)
	bookmarkCtrl := controller.NewBookmarkController(bookmarkService, likeService, threadService, threadMediaService, threadPollService)
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

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := premiumSubscriptionService.ExpireSubscriptions(); err != nil {
				log.Printf("ExpireSubscriptions error: %v", err)
			}
		}
	}()

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
