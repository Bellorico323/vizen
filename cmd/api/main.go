package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/Bellorico323/vizen/internal/api"
	"github.com/Bellorico323/vizen/internal/api/controllers"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Env variables file not found")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s",
		os.Getenv("VIZEN_DATABASE_USER"),
		os.Getenv("VIZEN_DATABASE_PASSWORD"),
		os.Getenv("VIZEN_DATABASE_HOST"),
		os.Getenv("VIZEN_DATABASE_PORT"),
		os.Getenv("VIZEN_DATABASE_NAME"),
		os.Getenv("VIZEN_DATABASE_SSLMODE"),
	))

	if err != nil {
		panic(err)
	}

	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		panic(err)
	}

	tokenService, err := auth.NewTokenService()
	if err != nil {
		panic(err)
	}

	signupWithCredentials := usecases.NewSignupWithCredentialsUseCase(pool)
	signinWithCredentials := usecases.NewSigninUserWithCredentials(pool, tokenService)
	refreshToken := usecases.NewRefreshTokenUseCase(pool, tokenService)
	getUserProfile := usecases.NewGetUserProfile(pool)

	api := api.Api{
		Router:       chi.NewMux(),
		TokenService: tokenService,

		SignupController: &controllers.SignupHandler{
			SignUpUseCase: signupWithCredentials,
		},
		SigninController: &controllers.SigninHandler{
			SigninUseCase: signinWithCredentials,
		},
		RefreshTokenController: &controllers.RefreshTokenHandler{
			RefreshTokenUseCase: refreshToken,
		},
		UsersController: &controllers.UsersController{
			GetUserProfile: getUserProfile,
		},
	}

	api.BindRoutes()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fmt.Printf("Starting Server on port: %s\n", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, api.Router); err != nil {
		panic(err)
	}
}
