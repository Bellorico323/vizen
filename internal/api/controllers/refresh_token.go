package controllers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/usecases"
)

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"` // Optional (mobile only)
}

type RefreshTokenHandler struct {
	RefreshTokenUseCase *usecases.RefreshTokenUseCase
}

func (h *RefreshTokenHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var refreshToken string

	cookie, err := r.Cookie("refresh_token")
	if err == nil {
		refreshToken = cookie.Value
	}

	if refreshToken == "" {
		if req, err := jsonutils.DecodeJson[RefreshTokenRequest](r); err == nil {
			refreshToken = req.RefreshToken
		}
	}

	if refreshToken == "" {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"message": "Refresh token not found in cookie or body",
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

			jsonutils.EncodeJson(w, r, http.StatusUnauthorized, map[string]any{
				"message": "Invalid or expired session. Please log in again.",
			})
			return
		}

		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"message": "Internal server error",
		})
		return
	}

	if r.Header.Get("X-Client-Type") == "mobile" {
		jsonutils.EncodeJson(w, r, http.StatusOK, map[string]any{
			"accessToken":  tokens.AccessToken,
			"refreshToken": tokens.RefreshToken,
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		HttpOnly: true,
		Secure:   false, // True em PRD
		Path:     "/api/v1/auth",
		MaxAge:   7 * 24 * 60 * 60, // 7 days
		SameSite: http.SameSiteStrictMode,
	})

	jsonutils.EncodeJson(w, r, http.StatusOK, map[string]any{
		"accessToken": tokens.AccessToken,
	})
}
