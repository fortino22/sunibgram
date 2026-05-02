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
	writeDB := config.NewPostgresWrite()
	readDB  := config.NewPostgresRead()

	migration.AutoMigrate(writeDB)

	redisCache := cache.New()

	roleRepo        := infrarepo.NewRoleRepository(writeDB, readDB)
	userRepo        := infrarepo.NewUserRepository(writeDB, readDB)
	profileRepo     := infrarepo.NewUserProfileRepository(writeDB, readDB)
	threadRepo, patchCount := infrarepo.NewThreadRepository(writeDB, readDB, redisCache)
	threadMediaRepo := infrarepo.NewThreadMediaRepository(writeDB, readDB)
	likeRepo        := infrarepo.NewLikeRepository(writeDB, readDB, redisCache, patchCount)
	bookmarkRepo    := infrarepo.NewBookmarkRepository(writeDB, readDB, redisCache, patchCount)
	verificationRepo := infrarepo.NewVerificationRepository(writeDB, readDB)
	followRepo            := infrarepo.NewFollowRepository(writeDB, readDB, redisCache)
	communityRepo         := infrarepo.NewCommunityRepository(writeDB, readDB)
	communityMemberRepo   := infrarepo.NewCommunityMemberRepository(writeDB, readDB)
	communityRequestRepo  := infrarepo.NewCommunityRequestRepository(writeDB, readDB)

	minioClient    := storage.NewMinioClient()
	storageUsecase := usecase.NewStorageUsecase(minioClient)

	smtpClient  := infraemail.NewSMTPClient()
	emailService := usecase.NewEmailService(smtpClient)

	authService        := usecase.NewAuthService(userRepo, roleRepo, verificationRepo, redisCache)
	roleService        := usecase.NewRoleService(roleRepo)
	userService        := usecase.NewUserService(userRepo, followRepo, profileRepo)
	profileService     := usecase.NewUserProfileService(userRepo, followRepo, profileRepo, storageUsecase)
	threadService      := usecase.NewThreadService(threadRepo)
	threadMediaService := usecase.NewThreadMediaService(threadRepo, threadMediaRepo, storageUsecase)
	likeService        := usecase.NewLikeService(likeRepo, threadRepo)
	bookmarkService    := usecase.NewBookmarkService(bookmarkRepo, threadRepo)
	verificationService := usecase.NewVerificationService(verificationRepo, userRepo, emailService)
	followService            := usecase.NewFollowService(followRepo, userRepo)
	communityService         := usecase.NewCommunityService(communityRepo, communityMemberRepo, storageUsecase)
	communityMemberService   := usecase.NewCommunityMemberService(communityRepo, communityMemberRepo)
	communityRequestService  := usecase.NewCommunityRequestService(communityRepo, communityMemberRepo, communityRequestRepo)

	authController        := controller.NewAuthController(authService, profileService)
	roleController        := controller.NewRoleController(roleService)
	userController        := controller.NewUserController(userService)
	profileController     := controller.NewUserProfileController(profileService)
	threadController := controller.NewThreadController(threadService, threadMediaService)
	likeController   := controller.NewLikeController(likeService, threadMediaService)
	bookmarkController    := controller.NewBookmarkController(bookmarkService, threadMediaService)
	repostController      := controller.NewRepostController(threadService, threadMediaService)
	emailController        := controller.NewEmailController(emailService)
	verificationController := controller.NewVerificationController(verificationService)
	followController             := controller.NewFollowController(followService)
	communityController         := controller.NewCommunityController(communityService)
	communityMemberController   := controller.NewCommunityMemberController(communityMemberService)
	communityRequestController  := controller.NewCommunityRequestController(communityRequestService)

	app := fiber.New()
	router.Register(
		app,
		authController,
		roleController,
		userController,
		profileController,
		threadController,
		likeController,
		bookmarkController,
		repostController,
		emailController,
		verificationController,
		followController,
		communityController,
		communityMemberController,
		communityRequestController,
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
