package api

import (
	"github.com/Bellorico323/vizen/internal/api/controllers"
	"github.com/go-chi/chi/v5"
)

type Api struct {
	Router *chi.Mux

	// Controllers
	SignUpController *controllers.SignupHandler
}
