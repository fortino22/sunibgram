package main

import (
	"fmt"
	"net/http"

	"github.com/sunibgram/application/usecase"
	"github.com/sunibgram/infrastructure/config"
	"github.com/sunibgram/infrastructure/migration"
	infrarepo "github.com/sunibgram/infrastructure/repository"
	"github.com/sunibgram/webapi/controller"
)

func main() {
	writeDB := config.NewPostgresWrite()
	readDB := config.NewPostgresRead()

	migration.AutoMigrate(writeDB)

	userRepo := infrarepo.NewUserRepository(writeDB, readDB)
	authService := usecase.NewAuthService(userRepo)
	authController := controller.NewAuthController(authService)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/register", authController.Register)
	mux.HandleFunc("POST /api/auth/login", authController.Login)
	mux.HandleFunc("POST /api/auth/refresh", authController.Refresh)

	addr := ":8080"
	fmt.Printf("server running on %s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		panic(err)
	}
}
