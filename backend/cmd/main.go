package main

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/application/usecase"
	"github.com/sunibgram/infrastructure/cache"
	"github.com/sunibgram/infrastructure/config"
	infraemail "github.com/sunibgram/infrastructure/email"
	"github.com/sunibgram/infrastructure/migration"
	infrarepo "github.com/sunibgram/infrastructure/repository"
	"github.com/sunibgram/infrastructure/storage"
	"github.com/sunibgram/webapi/controller"
	"github.com/sunibgram/webapi/router"
)

func main() {
	// --- Database ---
	writeDB := config.NewPostgresWrite()
	readDB  := config.NewPostgresRead()
	migration.AutoMigrate(writeDB)

	// --- Cache ---
	redisCache := cache.New()

	// --- Repositories ---
	roleRepo             := infrarepo.NewRoleRepository(writeDB, readDB)
	userRepo             := infrarepo.NewUserRepository(writeDB, readDB)
	profileRepo          := infrarepo.NewUserProfileRepository(writeDB, readDB)
	verificationRepo     := infrarepo.NewVerificationRepository(writeDB, readDB)
	followRepo           := infrarepo.NewFollowRepository(writeDB, readDB, redisCache)
	threadRepo, patchCount := infrarepo.NewThreadRepository(writeDB, readDB, redisCache)
	threadMediaRepo      := infrarepo.NewThreadMediaRepository(writeDB, readDB)
	likeRepo             := infrarepo.NewLikeRepository(writeDB, readDB, redisCache, patchCount)
	bookmarkRepo         := infrarepo.NewBookmarkRepository(writeDB, readDB, redisCache, patchCount)
	communityRepo        := infrarepo.NewCommunityRepository(writeDB, readDB)
	communityMemberRepo  := infrarepo.NewCommunityMemberRepository(writeDB, readDB)
	communityRequestRepo := infrarepo.NewCommunityRequestRepository(writeDB, readDB)
	facultyRepo          := infrarepo.NewFacultyRepository(writeDB, readDB)
	userFacultyRepo      := infrarepo.NewUserFacultyRepository(writeDB, readDB)

	// --- External Services ---
	storageUsecase := usecase.NewStorageUsecase(storage.NewMinioClient())
	emailService   := usecase.NewEmailService(infraemail.NewSMTPClient())

	// --- Usecases ---
	authService            := usecase.NewAuthService(userRepo, roleRepo, verificationRepo, redisCache)
	roleService            := usecase.NewRoleService(roleRepo)
	userService            := usecase.NewUserService(userRepo, followRepo, profileRepo)
	profileService         := usecase.NewUserProfileService(userRepo, followRepo, profileRepo, userFacultyRepo, storageUsecase)
	verificationService    := usecase.NewVerificationService(verificationRepo, userRepo, emailService)
	followService          := usecase.NewFollowService(followRepo, userRepo)
	threadService          := usecase.NewThreadService(threadRepo, communityRepo, communityMemberRepo)
	threadMediaService     := usecase.NewThreadMediaService(threadRepo, threadMediaRepo, storageUsecase)
	likeService            := usecase.NewLikeService(likeRepo, threadRepo, communityMemberRepo)
	bookmarkService        := usecase.NewBookmarkService(bookmarkRepo, threadRepo, communityMemberRepo)
	communityService       := usecase.NewCommunityService(communityRepo, communityMemberRepo, storageUsecase)
	communityMemberService := usecase.NewCommunityMemberService(communityRepo, communityMemberRepo)
	communityRequestService := usecase.NewCommunityRequestService(communityRepo, communityMemberRepo, communityRequestRepo)
	facultyService         := usecase.NewFacultyService(facultyRepo, userFacultyRepo)

	// --- Controllers ---
	authCtrl            := controller.NewAuthController(authService, profileService, facultyService)
	roleCtrl            := controller.NewRoleController(roleService)
	userCtrl            := controller.NewUserController(userService)
	profileCtrl         := controller.NewUserProfileController(profileService, facultyService)
	emailCtrl           := controller.NewEmailController(emailService)
	verificationCtrl    := controller.NewVerificationController(verificationService)
	followCtrl          := controller.NewFollowController(followService)
	threadCtrl          := controller.NewThreadController(threadService, threadMediaService, communityMemberService, likeService, bookmarkService)
	likeCtrl            := controller.NewLikeController(likeService, threadMediaService)
	bookmarkCtrl        := controller.NewBookmarkController(bookmarkService, threadMediaService)
	repostCtrl          := controller.NewRepostController(threadService, threadMediaService)
	communityCtrl       := controller.NewCommunityController(communityService)
	communityMemberCtrl := controller.NewCommunityMemberController(communityMemberService)
	communityRequestCtrl := controller.NewCommunityRequestController(communityRequestService)
	communityThreadCtrl  := controller.NewCommunityThreadController(threadService, threadMediaService, likeService, bookmarkService)
	facultyCtrl          := controller.NewFacultyController(facultyService)

	// --- Server ---
	app := fiber.New()
	router.Register(
		app,
		authCtrl,
		roleCtrl,
		userCtrl,
		profileCtrl,
		threadCtrl,
		likeCtrl,
		bookmarkCtrl,
		repostCtrl,
		emailCtrl,
		verificationCtrl,
		followCtrl,
		communityCtrl,
		communityMemberCtrl,
		communityRequestCtrl,
		communityThreadCtrl,
		facultyCtrl,
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
