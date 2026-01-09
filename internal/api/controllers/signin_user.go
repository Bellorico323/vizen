package controllers

import (
	"errors"
	"log/slog"
	"net"
	"net/http"

	"github.com/Bellorico323/vizen/internal/api/common"
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

// SigninResponse defines the success response structure
// The RefreshToken is omitted from JSON if empty (standard for Web clients using cookies)
type SigninResponse struct {
	Message      string `json:"message"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken,omitempty"`
}

// Handle executes the user login
// @Summary      User Login
// @Description  Authenticates the user and returns the Access Token. Refresh Token is returned in the Body (Mobile) or HttpOnly Cookie (Web).
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body controllers.SigninWithCredentialsRequest true "Login Credentials"
// @Success      200  {object}  controllers.SigninResponse     "Login successful"
// @Failure      400  {object}  common.ErrResponse        "Invalid JSON payload"
// @Failure      401  {object}  common.ErrResponse        "Invalid credentials"
// @Failure      422  {object}  common.ValidationErrResponse "Validation failed"
// @Failure      500  {object}  common.ErrResponse        "Internal server error"
// @Router       /users/signin [post]
func (h *SigninHandler) Handle(w http.ResponseWriter, r *http.Request) {
	data, err := jsonutils.DecodeJson[SigninWithCredentialsRequest](r)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
			Message: "Invalid json body",
		})
		return
	}

	if validationErrors := validator.ValidateStruct(data); len(validationErrors) > 0 {
		jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, common.ValidationErrResponse{
			Message: "Validation failed",
			Errors:  validationErrors,
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
			jsonutils.EncodeJson(w, r, http.StatusUnauthorized, common.ErrResponse{
				Message: err.Error(),
			})
			return
		}
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{
			Message: "Internal server error",
		})
		return
	}

	isMobile := r.Header.Get("X-Client-Type") == "mobile"

	if isMobile {
		jsonutils.EncodeJson(w, r, http.StatusOK, SigninResponse{
			Message:      "User successfully logged in",
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
		})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		HttpOnly: true,
		Secure:   false, // TRUE if https
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60, // 7 days
		SameSite: http.SameSiteStrictMode,
	})

	jsonutils.EncodeJson(w, r, http.StatusOK, SigninResponse{
		Message:     "User successfully logged in",
		AccessToken: tokens.AccessToken,
	})
}
