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
	avatarRepo      := infrarepo.NewUserAvatarRepository(writeDB, readDB)
	threadRepo      := infrarepo.NewThreadRepository(writeDB, readDB, redisCache)
	threadMediaRepo := infrarepo.NewThreadMediaRepository(writeDB, readDB)
	likeRepo             := infrarepo.NewLikeRepository(writeDB, readDB)
	bookmarkRepo         := infrarepo.NewBookmarkRepository(writeDB, readDB)
	repostRepo           := infrarepo.NewRepostRepository(writeDB, readDB)
	verificationRepo     := infrarepo.NewVerificationRepository(writeDB, readDB)

	minioClient    := storage.NewMinioClient()
	storageUsecase := usecase.NewStorageUsecase(minioClient)

	smtpClient  := infraemail.NewSMTPClient()
	emailService := usecase.NewEmailService(smtpClient)

	authService        := usecase.NewAuthService(userRepo, roleRepo, verificationRepo)
	roleService        := usecase.NewRoleService(roleRepo)
	userService        := usecase.NewUserService(userRepo)
	avatarService      := usecase.NewAvatarService(avatarRepo, storageUsecase)
	threadService      := usecase.NewThreadService(threadRepo)
	threadMediaService := usecase.NewThreadMediaService(threadRepo, threadMediaRepo, storageUsecase)
	likeService        := usecase.NewLikeService(likeRepo, threadRepo)
	bookmarkService    := usecase.NewBookmarkService(bookmarkRepo, threadRepo)
	repostService      := usecase.NewRepostService(repostRepo, threadRepo)
	verificationService := usecase.NewVerificationService(verificationRepo, userRepo, emailService)

	authController        := controller.NewAuthController(authService)
	roleController        := controller.NewRoleController(roleService)
	userController        := controller.NewUserController(userService)
	avatarController      := controller.NewAvatarController(avatarService)
	threadController      := controller.NewThreadController(threadService, threadMediaService)
	threadMediaController := controller.NewThreadMediaController(threadMediaService)
	likeController        := controller.NewLikeController(likeService, threadMediaService)
	bookmarkController    := controller.NewBookmarkController(bookmarkService, threadMediaService)
	repostController      := controller.NewRepostController(repostService, threadMediaService)
	emailController        := controller.NewEmailController(emailService)
	verificationController := controller.NewVerificationController(verificationService)

	app := fiber.New()
	router.Register(
		app,
		authController,
		roleController,
		userController,
		avatarController,
		threadController,
		threadMediaController,
		likeController,
		bookmarkController,
		repostController,
		emailController,
		verificationController,
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
