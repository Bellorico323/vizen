package api

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	_ "github.com/Bellorico323/vizen/docs"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func (api *Api) BindRoutes() {
	api.Router.Use(middleware.RequestID, middleware.Recoverer, middleware.Logger)

	authMiddleware := auth.Auth(api.TokenService)

	api.Router.Route("/api", func(r chi.Router) {
		r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
			jsonutils.EncodeJson(w, r, http.StatusOK, map[string]any{
				"message": "Online",
			})
		})
		workDir, _ := os.Getwd()
		filesDir := http.Dir(filepath.Join(workDir, "docs"))
		common.FileServer(r, "/docs", filesDir)

		// 2. Rota do Scalar (A nova interface)
		r.Get("/reference", func(w http.ResponseWriter, r *http.Request) {
			htmlContent := `
						<!doctype html>
						<html>
						<head>
							<title>Vizen API Reference</title>
							<meta charset="utf-8" />
							<meta name="viewport" content="width=device-width, initial-scale=1" />
							<style>
								body { margin: 0; }
							</style>
						</head>
						<body>
							<script
								id="api-reference"
								data-url="/api/docs/swagger.json"
							></script>
							<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
						</body>
						</html>
					`
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(htmlContent))
		})

		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("doc.json"),
		))

		r.Route("/v1", func(r chi.Router) {
			r.Route("/users", func(r chi.Router) {
				r.Post("/", api.SignupController.Handle)
				r.Post("/signin", api.SigninController.Handle)

				r.Group(func(r chi.Router) {
					r.Use(authMiddleware)
					r.Get("/me", api.UsersController.Handle)
					r.Get("/condominiums", api.ListUserCondominiusController.Handle)
					r.Get("/apartments", api.ListUserApartmentsController.Handle)
					r.Post("/devices", api.RegisterDeviceController.Handle)
				})
			})

			r.Route("/auth", func(r chi.Router) {
				r.Post("/refresh", api.RefreshTokenController.Handle)
			})

			// Needs auth
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware)

				r.Route("/condominiums", func(r chi.Router) {
					r.Post("/", api.CreateCondominiumController.Handle)
				})
				r.Route("/apartments", func(r chi.Router) {
					r.Post("/", api.CreateApartmentController.Handle)
				})
				r.Route("/access_requests", func(r chi.Router) {
					r.Post("/", api.CreateAccessRequestController.Handle)
					r.Post("/approve", api.ApproveAccessRequestController.Handle)
					r.Post("/reject", api.RejectAccessRequestController.Handle)
					r.Get("/pending", api.ListPendingAccessRequestsController.Handle)
				})
				r.Route("/announcements", func(r chi.Router) {
					r.Post("/", api.CreateAnnouncementController.Handle)
					r.Get("/", api.ListAnnouncementsController.Handle)
					r.Delete("/{id}", api.DeleteAnnouncementController.Handle)
				})
				r.Route("/packages", func(r chi.Router) {
					r.Post("/", api.CreatePackageController.Handle)
					r.Get("/{id}", api.GetPackageController.Handle)
					r.Get("/", api.ListPackagesController.Handle)
					r.Patch("/{id}/withdraw", api.WithdrawPackageController.Handle)
				})
				r.Route("/invites", func(r chi.Router) {
					r.Post("/", api.CreateInviteController.Handle)
					r.Post("/validate", api.ValidateInviteController.Handle)
					r.Patch("/{id}/revoke", api.RevokeInviteController.Handle)
					r.Get("/", api.ListInvitesController.Handle)
				})
				r.Route("/common_areas", func(r chi.Router) {
					r.Post("/", api.CreateCommonAreaController.Handle)
					r.Get("/", api.ListCommonAreasController.Handle)
					r.Get("/{id}/availability", api.GetAreaAvailabilityController.Handle)
				})
				r.Route("/bookings", func(r chi.Router) {
					r.Post("/", api.CreateBookingController.Handle)
					r.Patch("/{id}/status", api.EditBookingController.Handle)
					r.Get("/", api.ListBookingsController.Handle)
				})
				r.Route("/bills", func(r chi.Router) {
					r.Post("/", api.CreateBillController.Handle)
					r.Patch("/{id}/pay", api.MarkBillAsPaidController.Handle)
					r.Get("/", api.ListBillsController.Handle)
				})
			})
		})
	})

}
