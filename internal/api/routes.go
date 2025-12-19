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
			})
		})
	})

}
