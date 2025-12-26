package api

import (
	"github.com/Bellorico323/vizen/internal/api/controllers"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/go-chi/chi/v5"
)

type Api struct {
	Router       *chi.Mux
	TokenService *auth.TokenService

	// Controllers
	SignupController               *controllers.SignupHandler
	SigninController               *controllers.SigninHandler
	UsersController                *controllers.UsersController
	RefreshTokenController         *controllers.RefreshTokenHandler
	CreateCondominiumController    *controllers.CreateCondominiumHandler
	CreateApartmentController      *controllers.CreateApartmentHandler
	CreateAccessRequestController  *controllers.CreateAccessRequestHandler
	ApproveAccessRequestController *controllers.ApproveAccessRequestHandler
	RegisterDeviceController       *controllers.RegisterDeviceHandler
}
