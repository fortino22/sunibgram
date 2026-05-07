package main

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/sunibgram/application/usecase"
	"github.com/sunibgram/infrastructure/cache"
	"github.com/sunibgram/infrastructure/config"
	infraemail "github.com/sunibgram/infrastructure/email"
	"github.com/sunibgram/infrastructure/migration"
	infrarepo "github.com/sunibgram/infrastructure/repository"
	"github.com/sunibgram/infrastructure/storage"
	"github.com/sunibgram/webapi/controller"
	"github.com/sunibgram/webapi/middleware"
	"github.com/sunibgram/webapi/router"
)

func main() {
	// --- Database ---
	writeDB := config.NewPostgresWrite()
	readDB := config.NewPostgresRead()
	migration.AutoMigrate(writeDB)

	// --- Cache ---
	redisCache := cache.New()
	middleware.Init(redisCache)

	// --- Repositories ---
	roleRepo := infrarepo.NewRoleRepository(writeDB, readDB)
	userRepo := infrarepo.NewUserRepository(writeDB, readDB)
	userSettingsRepo := infrarepo.NewUserSettingsRepository(writeDB, readDB)
	profileRepo := infrarepo.NewUserProfileRepository(writeDB, readDB)
	verificationRepo := infrarepo.NewVerificationRepository(writeDB, readDB)
	followRepo := infrarepo.NewFollowRepository(writeDB, readDB, redisCache)
	threadRepo, patchCount := infrarepo.NewThreadRepository(writeDB, readDB, redisCache)
	threadMediaRepo := infrarepo.NewThreadMediaRepository(writeDB, readDB)
	likeRepo := infrarepo.NewLikeRepository(writeDB, readDB, redisCache, patchCount)
	bookmarkRepo := infrarepo.NewBookmarkRepository(writeDB, readDB, redisCache, patchCount)
	threadPinRepo := infrarepo.NewThreadPinRepository(writeDB, readDB)
	communityRepo := infrarepo.NewCommunityRepository(writeDB, readDB)
	communityMemberRepo := infrarepo.NewCommunityMemberRepository(writeDB, readDB)
	communityRequestRepo := infrarepo.NewCommunityRequestRepository(writeDB, readDB)
	facultyRepo := infrarepo.NewFacultyRepository(writeDB, readDB)
	userFacultyRepo := infrarepo.NewUserFacultyRepository(writeDB, readDB)
	orderRepo := infrarepo.NewOrderRepository(writeDB, readDB)
	chatRoomRepo := infrarepo.NewChatRoomRepository(writeDB, readDB, redisCache)
	chatRepo := infrarepo.NewChatRepository(writeDB, readDB)
	chatMediaRepo := infrarepo.NewChatMediaRepository(writeDB, readDB)
	chatReadRepo := infrarepo.NewChatReadRepository(writeDB, readDB)

	// --- External Services ---
	storageUsecase := usecase.NewStorageUsecase(storage.NewMinioClient())
	emailService := usecase.NewEmailService(infraemail.NewSMTPClient())

	// --- Usecases ---
	authService := usecase.NewAuthService(userRepo, roleRepo, verificationRepo, redisCache)
	roleService := usecase.NewRoleService(roleRepo)
	userService := usecase.NewUserService(userRepo, followRepo, profileRepo)
	userSettingsService := usecase.NewUserSettingsService(userSettingsRepo)
	profileService := usecase.NewUserProfileService(userRepo, followRepo, profileRepo, userFacultyRepo, storageUsecase)
	verificationService := usecase.NewVerificationService(verificationRepo, userRepo, emailService)
	followService := usecase.NewFollowService(followRepo, userRepo)
	threadService := usecase.NewThreadService(threadRepo, communityRepo, communityMemberRepo)
	threadMediaService := usecase.NewThreadMediaService(threadRepo, threadMediaRepo, storageUsecase)
	likeService := usecase.NewLikeService(likeRepo, threadRepo, communityMemberRepo)
	bookmarkService := usecase.NewBookmarkService(bookmarkRepo, threadRepo, communityMemberRepo)
	threadPinService := usecase.NewThreadPinService(threadPinRepo, threadRepo)
	communityService := usecase.NewCommunityService(communityRepo, communityMemberRepo, userRepo, storageUsecase)
	communityMemberService := usecase.NewCommunityMemberService(communityRepo, communityMemberRepo)
	communityRequestService := usecase.NewCommunityRequestService(communityRepo, communityMemberRepo, communityRequestRepo)
	facultyService := usecase.NewFacultyService(facultyRepo, userFacultyRepo)
	orderService := usecase.NewOrderService(orderRepo, userRepo)
	chatRoomService := usecase.NewChatRoomService(chatRoomRepo, userRepo)
	chatService := usecase.NewChatService(chatRepo, chatRoomRepo, chatReadRepo)
	chatMediaService := usecase.NewChatMediaService(chatRepo, chatMediaRepo, storageUsecase)

	// --- Controllers ---
	authCtrl := controller.NewAuthController(authService, profileService, facultyService)
	roleCtrl := controller.NewRoleController(roleService)
	userCtrl := controller.NewUserController(userService)
	userSettingsCtrl := controller.NewUserSettingsController(userSettingsService)
	profileCtrl := controller.NewUserProfileController(profileService, facultyService)
	emailCtrl := controller.NewEmailController(emailService)
	verificationCtrl := controller.NewVerificationController(verificationService)
	followCtrl := controller.NewFollowController(followService, profileService)
	threadCtrl := controller.NewThreadController(threadService, threadMediaService, communityMemberService, likeService, bookmarkService, threadPinService)
	likeCtrl := controller.NewLikeController(likeService, bookmarkService, threadService, threadMediaService)
	bookmarkCtrl := controller.NewBookmarkController(bookmarkService, likeService, threadService, threadMediaService)
	threadPinCtrl := controller.NewThreadPinController(threadPinService)
	repostCtrl := controller.NewRepostController(threadService, threadMediaService)
	communityCtrl := controller.NewCommunityController(communityService)
	communityMemberCtrl := controller.NewCommunityMemberController(communityMemberService)
	communityRequestCtrl := controller.NewCommunityRequestController(communityRequestService)
	communityThreadCtrl := controller.NewCommunityThreadController(threadService, threadMediaService, likeService, bookmarkService, threadPinService)
	facultyCtrl := controller.NewFacultyController(facultyService)
	orderCtrl := controller.NewOrderController(orderService)
	chatRoomCtrl := controller.NewChatRoomController(chatRoomService)
	chatCtrl := controller.NewChatController(chatService, chatMediaService)

	// --- Server ---
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	router.Register(
		app,
		authCtrl,
		roleCtrl,
		userCtrl,
		userSettingsCtrl,
		profileCtrl,
		threadCtrl,
		likeCtrl,
		bookmarkCtrl,
		threadPinCtrl,
		repostCtrl,
		emailCtrl,
		verificationCtrl,
		followCtrl,
		communityCtrl,
		communityMemberCtrl,
		communityRequestCtrl,
		communityThreadCtrl,
		facultyCtrl,
		orderCtrl,
		chatRoomCtrl,
		chatCtrl,
	)

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
