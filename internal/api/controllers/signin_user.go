package controllers

import (
	"errors"
	"log/slog"
	"net"
	"net/http"

	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/Bellorico323/vizen/internal/validator"
)

type SigninHandler struct {
	SigninUseCase *usecases.SigninUserWithCredentials
}

type SigninWithCredentialsRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

func (h *SigninHandler) Handle(w http.ResponseWriter, r *http.Request) {
	data, err := jsonutils.DecodeJson[SigninWithCredentialsRequest](r)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
			"message": "Invalid json",
		})
		return
	}

	if validationErrors := validator.ValidateStruct(data); len(validationErrors) > 0 {
		jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, map[string]any{
			"message": "Validation failed",
			"errors":  validationErrors,
		})
		return
	}

	host, _, _ := net.SplitHostPort(r.RemoteAddr)

	payload := usecases.SigninUserWithCredentialsReq{
		Email:     data.Email,
		Password:  data.Password,
		IpAddress: host,
		UserAgent: r.UserAgent(),
	}

	tokens, err := h.SigninUseCase.Exec(r.Context(), payload)
	if err != nil {
		slog.Error("Error while sign in user", "error", err)

		if errors.Is(err, usecases.ErrInvalidCredentials) {
			jsonutils.EncodeJson(w, r, http.StatusBadRequest, map[string]any{
				"message": err,
			})
			return
		}
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, map[string]any{
			"message": "An error occured while sign in user",
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		HttpOnly: true,
		Secure:   false, // TRUE if https
		Path:     "/api/v1/auth",
		MaxAge:   7 * 24 * 60 * 60, // 7 days
		SameSite: http.SameSiteStrictMode,
	})

	jsonutils.EncodeJson(w, r, http.StatusOK, map[string]any{
		"message":     "User successfully logged in",
		"accessToken": tokens.AccessToken,
	})
}
