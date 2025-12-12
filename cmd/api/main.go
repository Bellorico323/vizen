package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/Bellorico323/vizen/internal/api"
	"github.com/Bellorico323/vizen/internal/api/controllers"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s",
		os.Getenv("VIZEN_DATABASE_USER"),
		os.Getenv("VIZEN_DATABASE_PASSWORD"),
		os.Getenv("VIZEN_DATABASE_HOST"),
		os.Getenv("VIZEN_DATABASE_PORT"),
		os.Getenv("VIZEN_DATABASE_NAME"),
	))

	if err != nil {
		panic(err)
	}

	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		panic(err)
	}

	signupWithCredentials := usecases.NewSignupWithCredentialsUseCase(pool)

	api := api.Api{
		Router: chi.NewMux(),
		SignUpController: &controllers.SignupHandler{
			SignUpUseCase: &signupWithCredentials,
		},
	}

	api.BindRoutes()

	fmt.Println("Starting Server on port :3000")
	if err := http.ListenAndServe("127.0.0.1:3000", api.Router); err != nil {
		panic(err)
	}
}
