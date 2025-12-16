package controllers

import (
	"net/http"

	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/usecases"
)

type UsersController struct {
	GetUserProfile *usecases.GetUserProfile
}

func (uc *UsersController) Handle(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusUnauthorized, map[string]any{
			"message": "User not authenticated in context",
		})
		return
	}

	user, err := uc.GetUserProfile.Exec(r.Context(), userID)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusNotFound, map[string]any{
			"message": err.Error(),
		})
		return
	}

	jsonutils.EncodeJson(w, r, http.StatusOK, map[string]any{
		"user": user,
	})
}
