package api

import (
	"net/http"

	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (api *Api) BindRoutes() {
	api.Router.Use(middleware.RequestID, middleware.Recoverer, middleware.Logger)

	api.Router.Route("/api", func(r chi.Router) {
		r.Get("/healthyz", func(w http.ResponseWriter, r *http.Request) {
			jsonutils.EncodeJson(w, r, http.StatusOK, map[string]any{
				"message": "Online",
			})
		})

		r.Route("/v1", func(r chi.Router) {
			r.Route("/users", func(r chi.Router) {
				r.Post("/", api.SignUpController.Handle)
			})
		})
	})

}
