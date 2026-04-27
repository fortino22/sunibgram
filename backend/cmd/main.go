package main

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/sunibgram/application/usecase"
	"github.com/sunibgram/infrastructure/config"
	"github.com/sunibgram/infrastructure/migration"
	infrarepo "github.com/sunibgram/infrastructure/repository"
	"github.com/sunibgram/webapi/controller"
	"github.com/sunibgram/webapi/router"
)

func main() {
	writeDB := config.NewPostgresWrite()
	readDB := config.NewPostgresRead()

	migration.AutoMigrate(writeDB)

	roleRepo := infrarepo.NewRoleRepository(writeDB, readDB)
	userRepo := infrarepo.NewUserRepository(writeDB, readDB)

	authService := usecase.NewAuthService(userRepo, roleRepo)
	roleService := usecase.NewRoleService(roleRepo)
	userService := usecase.NewUserService(userRepo)

	authController := controller.NewAuthController(authService)
	roleController := controller.NewRoleController(roleService)
	userController := controller.NewUserController(userService)

	app := fiber.New()
	router.Register(app, authController, roleController, userController)

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
