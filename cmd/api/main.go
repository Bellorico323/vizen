package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Bellorico323/vizen/internal/api"
	"github.com/Bellorico323/vizen/internal/api/controllers"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// @title           Vizen API
// @version         1.0
// @description     Backend server for Vizen Application.
// @termsOfService  http://swagger.io/terms/

// @contact.name    Suporte Vizen
// @contact.email   support@vizen.app

// @host            localhost:3000
// @BasePath        /api/v1
// @schemes         http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

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

	connected := false
	for i := range 10 {
		if err := pool.Ping(ctx); err == nil {
			connected = true
			break
		}
		fmt.Printf("Banco ainda indisponível (tentativa %d/10). Aguardando...\n", i+1)
		time.Sleep(2 * time.Second)
	}

	if !connected {
		panic("Não foi possível conectar ao banco de dados após várias tentativas")
	}

	fmt.Println("Conectado ao banco de dados com sucesso!")

	tokenService, err := auth.NewTokenService()
	if err != nil {
		panic(err)
	}

	queries := pgstore.New(pool)

	signupWithCredentials := usecases.NewSignupWithCredentialsUseCase(pool)
	signinWithCredentials := usecases.NewSigninUserWithCredentials(queries, tokenService)
	refreshToken := usecases.NewRefreshTokenUseCase(queries, tokenService)
	getUserProfile := usecases.NewGetUserProfile(queries)
	createCondominium := usecases.NewCreateCondominiumUseCase(pool)
	createApartment := usecases.NewCreateApartmentUseCase(queries)

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
		CreateCondominiumController: &controllers.CreateCondominiumHandler{
			CreateCondominium: createCondominium,
		},
		CreateApartmentController: &controllers.CreateApartmentHandler{
			CreateApartment: createApartment,
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
