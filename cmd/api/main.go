package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/Bellorico323/vizen/internal/api"
	"github.com/Bellorico323/vizen/internal/api/controllers"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/infra/notification"
	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
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

	credsBase64 := os.Getenv("FIREBASE_CREDENTIALS_BASE64")
	if credsBase64 == "" {
		slog.Error("FIREBASE_CREDENTIALS_BASE64 environment variable is not set")
	}

	credJSON, err := base64.StdEncoding.DecodeString(credsBase64)
	if err != nil {
		slog.Error("failed to decode firebase credentials", "error", err)
	}

	opt := option.WithCredentialsJSON(credJSON)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		slog.Error("Error initializing app", "error", err)
	}

	messagingClient, err := app.Messaging(context.Background())
	if err != nil {
		slog.Error("Error getting client", "error", err)
	}

	notiService := notification.NewFireBaseService(messagingClient, queries)

	signupWithCredentials := usecases.NewSignupWithCredentialsUseCase(pool)
	signinWithCredentials := usecases.NewSigninUserWithCredentials(queries, tokenService)
	refreshToken := usecases.NewRefreshTokenUseCase(queries, tokenService)
	getUserProfile := usecases.NewGetUserProfileUseCase(queries)
	listUserCondominiums := usecases.NewListUserCondominiumsUseCase(queries)
	createCondominium := usecases.NewCreateCondominiumUseCase(pool)
	listUserApartments := usecases.NewListUserApartmentsUseCase(queries)
	createApartment := usecases.NewCreateApartmentUseCase(queries)
	createAccessRequest := usecases.NewCreateAccessRequestUseCase(queries, notiService)
	approveAccessRequest := usecases.NewApproveAccessRequestUseCase(pool, notiService)
	rejectAccessRequest := usecases.NewRejectAccessRequestUseCase(pool, notiService)
	listPendingAccessRequests := usecases.NewListPendingAccessRequestsUseCase(queries)
	registerUserDevice := usecases.NewRegisterDeviceUseCase(queries)
	createAnnouncement := usecases.NewCreateAnnouncementUseCase(queries, notiService)
	listAnnouncements := usecases.NewListAnnouncementsUseCase(queries)
	deleteAnnouncement := usecases.NewDeleteAnnouncementUseCase(queries)
	createPackage := usecases.NewCreatePackageUseCase(queries, notiService)
	getPackage := usecases.NewGetPackageUseCase(queries)
	listPackages := usecases.NewListPackagesUseCase(queries)
	withdrawPackage := usecases.NewWithdrawPackageUseCase(queries)
	createInvite := usecases.NewCreateInviteUseCase(queries)
	validateInvite := usecases.NewValidateInviteUseCase(pool, notiService)
	revokeInvite := usecases.NewRevokeInviteUseCase(queries)
	listInvites := usecases.NewListInvitesUseCase(queries)
	createCommonArea := usecases.NewCreateCommonAreaUseCase(queries)
	listCommonAreas := usecases.NewListCommonAreasUseCase(queries)
	createBooking := usecases.NewCreateBookingUseCase(pool, queries)
	editBooking := usecases.NewEditBookingUseCase(queries, notiService)
	listBookings := usecases.NewListBookingsUseCase(queries)
	getAreaAvailability := usecases.NewGetAreaAvailabilityUseCase(queries)

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
		ListUserCondominiusController: &controllers.ListUserCondominiumsHandler{
			ListUserCondominiums: listUserCondominiums,
		},
		CreateCondominiumController: &controllers.CreateCondominiumHandler{
			CreateCondominium: createCondominium,
		},
		ListUserApartmentsController: &controllers.ListUserApartmentsHandler{
			ListUserApartments: listUserApartments,
		},
		CreateApartmentController: &controllers.CreateApartmentHandler{
			CreateApartment: createApartment,
		},
		CreateAccessRequestController: &controllers.CreateAccessRequestHandler{
			CreateAccessRequest: createAccessRequest,
		},
		ApproveAccessRequestController: &controllers.ApproveAccessRequestHandler{
			ApproveAccessRequest: approveAccessRequest,
		},
		RejectAccessRequestController: &controllers.RejectAccessRequestHandler{
			RejectAccessRequest: rejectAccessRequest,
		},
		ListPendingAccessRequestsController: &controllers.ListPendingAccessRequestHandler{
			ListPendingAccessRequests: listPendingAccessRequests,
		},
		RegisterDeviceController: &controllers.RegisterDeviceHandler{
			RegisterDevice: *registerUserDevice,
		},
		CreateAnnouncementController: &controllers.CreateAnnouncementHandler{
			CreateAnnouncement: createAnnouncement,
		},
		ListAnnouncementsController: &controllers.ListAnnouncementsHandler{
			ListAnnouncements: listAnnouncements,
		},
		DeleteAnnouncementController: &controllers.DeleteAnnouncementHandler{
			DeleteAnnouncement: deleteAnnouncement,
		},
		CreatePackageController: &controllers.CreatePackageHandler{
			CreatePackage: createPackage,
		},
		GetPackageController: &controllers.GetPackageHandler{
			GetPackage: getPackage,
		},
		ListPackagesController: &controllers.ListPackagesHandler{
			ListPackages: listPackages,
		},
		WithdrawPackageController: &controllers.WithdrawPackageHandler{
			WithdrawPackage: withdrawPackage,
		},
		CreateInviteController: &controllers.CreateInviteHandler{
			CreateInvite: createInvite,
		},
		ValidateInviteController: &controllers.ValidateInviteHandler{
			ValidateInvite: validateInvite,
		},
		RevokeInviteController: &controllers.RevokeInviteHandler{
			RevokeInvite: revokeInvite,
		},
		ListInvitesController: &controllers.ListInvitesHandler{
			ListInvites: listInvites,
		},
		CreateCommonAreaController: &controllers.CreateCommonAreaHandler{
			CreateCommonArea: createCommonArea,
		},
		ListCommonAreasController: &controllers.ListCommonAreasHandler{
			ListCommonAreas: listCommonAreas,
		},
		CreateBookingController: &controllers.CreateBookingsHandler{
			CreateBooking: createBooking,
		},
		EditBookingController: &controllers.EditBookingHandler{
			EditBooking: editBooking,
		},
		ListBookingsController: &controllers.ListBookingsHandler{
			ListBookings: listBookings,
		},
		GetAreaAvailabilityController: &controllers.GetAreaAvailabilityHandler{
			GetAreaAvailability: getAreaAvailability,
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
