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
	SignupController                    *controllers.SignupHandler
	SigninController                    *controllers.SigninHandler
	UsersController                     *controllers.UsersController
	RefreshTokenController              *controllers.RefreshTokenHandler
	CreateCondominiumController         *controllers.CreateCondominiumHandler
	ListUserCondominiusController       *controllers.ListUserCondominiumsHandler
	CreateApartmentController           *controllers.CreateApartmentHandler
	ListUserApartmentsController        *controllers.ListUserApartmentsHandler
	CreateAccessRequestController       *controllers.CreateAccessRequestHandler
	ApproveAccessRequestController      *controllers.ApproveAccessRequestHandler
	RejectAccessRequestController       *controllers.RejectAccessRequestHandler
	ListPendingAccessRequestsController *controllers.ListPendingAccessRequestHandler
	RegisterDeviceController            *controllers.RegisterDeviceHandler
	CreateAnnouncementController        *controllers.CreateAnnouncementHandler
	ListAnnouncementsController         *controllers.ListAnnouncementsHandler
	DeleteAnnouncementController        *controllers.DeleteAnnouncementHandler
	CreatePackageController             *controllers.CreatePackageHandler
	GetPackageController                *controllers.GetPackageHandler
	ListPackagesController              *controllers.ListPackagesHandler
	WithdrawPackageController           *controllers.WithdrawPackageHandler
	CreateInviteController              *controllers.CreateInviteHandler
	ValidateInviteController            *controllers.ValidateInviteHandler
	RevokeInviteController              *controllers.RevokeInviteHandler
	ListInvitesController               *controllers.ListInvitesHandler
	CreateCommonAreaController          *controllers.CreateCommonAreaHandler
	ListCommonAreasController           *controllers.ListCommonAreasHandler
	CreateBookingController             *controllers.CreateBookingsHandler
	EditBookingController               *controllers.EditBookingHandler
	ListBookingsController              *controllers.ListBookingsHandler
	GetAreaAvailabilityController       *controllers.GetAreaAvailabilityHandler
}
