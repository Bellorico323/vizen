package controllers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/usecases"
)

type RefreshTokenHandler struct {
	RefreshTokenUseCase *usecases.RefreshTokenUseCase
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken,omitempty"`
}

// Handle refreshes the user access token
// @Summary      Refresh Access Token
// @Description  Generates a new access token using a valid refresh token from Cookie (Web) or Body (Mobile)
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body controllers.RefreshTokenRequest false "Refresh Token (Optional if using Cookies)"
// @Success      200  {object}  controllers.RefreshTokenResponse "Token refreshed successfully"
// @Failure      400  {object}  common.ErrResponse             "Token not found"
// @Failure      401  {object}  common.ErrResponse             "Invalid or expired session"
// @Failure      500  {object}  common.ErrResponse             "Internal server error"
// @Router       /auth/refresh [post]
func (h *RefreshTokenHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var refreshToken string

	if cookie, err := r.Cookie("refresh_token"); err == nil {
		refreshToken = cookie.Value
	}

	if refreshToken == "" {
		if req, err := jsonutils.DecodeJson[RefreshTokenRequest](r); err == nil {
			refreshToken = req.RefreshToken
		}
	}

	if refreshToken == "" {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
			Message: "Refresh token not found in cookie or body",
		})
		return
	}

	tokens, err := h.RefreshTokenUseCase.Exec(r.Context(), refreshToken)
	if err != nil {
		slog.Warn("Failed to refresh token", "error", err)

		if errors.Is(err, usecases.ErrInvalidToken) {
			http.SetCookie(w, &http.Cookie{
				Name:   "refresh_token",
				Value:  "",
				Path:   "/api/v1/auth",
				MaxAge: -1,
			})

			jsonutils.EncodeJson(w, r, http.StatusUnauthorized, common.ErrResponse{
				Message: "Invalid or expired session. Please log in again.",
			})
			return
		}

		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{
			Message: "Internal server error",
		})
		return
	}

	if r.Header.Get("X-Client-Type") == "mobile" {
		jsonutils.EncodeJson(w, r, http.StatusOK, RefreshTokenResponse{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		HttpOnly: true,
		Secure:   false, // true em prod
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60, // 7 days
		SameSite: http.SameSiteStrictMode,
	})

	jsonutils.EncodeJson(w, r, http.StatusOK, RefreshTokenResponse{
		AccessToken: tokens.AccessToken,
	})
}
