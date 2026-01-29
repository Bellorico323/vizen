package controllers

import (
	"log/slog"
	"net/http"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/usecases"
)

type LogoutHandler struct {
	Logout usecases.LogoutUC
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// Handle logs out a user
// @Summary      Logout User
// @Description  Revokes the user session.
// @Description  It checks for the 'refresh_token' in HttpOnly Cookies (Web) OR in the JSON Body (Mobile).
// @Description  If found, the session is deleted from the database and the cookie is cleared.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body controllers.LogoutRequest false "Refresh Token (Optional if using Cookies)"
// @Success      204 "No Content - Successfully logged out"
// @Failure      400 {object} common.ErrResponse "Refresh token not found in Cookie or Body"
// @Failure      500 {object} common.ErrResponse "Internal Server Error"
// @Router       /auth/logout [post]
func (h *LogoutHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var refreshToken string

	if cookie, err := r.Cookie("refresh_token"); err == nil {
		refreshToken = cookie.Value
	}

	if refreshToken == "" {
		data, err := jsonutils.DecodeJson[LogoutRequest](r)
		if err == nil && data.RefreshToken != "" {
			refreshToken = data.RefreshToken
		}
	}

	if refreshToken == "" {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
			Message: "Refresh token not found",
		})
		return
	}

	err := h.Logout.Exec(r.Context(), usecases.LogoutReq{
		RefreshToken: refreshToken,
	})
	if err != nil {
		slog.Error("failed to logout",
			"error", err,
		)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Secure:   false, // true em prod
		Path:     "/",
		MaxAge:   -1, // 7 days
		SameSite: http.SameSiteStrictMode,
	})

	w.WriteHeader(http.StatusNoContent)
}
