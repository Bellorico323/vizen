package controllers

import (
	"log/slog"
	"net/http"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/usecases"
)

type ListUserCondominiumsHandler struct {
	ListUserCondominiums usecases.ListUserCondominiumsUC
}

// Handle lists user condominiums
// @Summary 		List User Condominiums
// @Description	Returns all condominiums the user is part of, sorted by role (Syndic > Admin > Resident).
// @Security		BearerAuth
// @Tags			Users
// @Produce			json
// @Success			200 {object} []usecases.UserCondominiumDTO
// @Failure			401 {object} common.ErrResponse "User not authenticated"
// @Failure			500 {object} common.ErrResponse	"Internal server error"
// @Router			/users/condominiums [get]
func (h *ListUserCondominiumsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusUnauthorized, common.ErrResponse{
			Message: "User not authenticated",
		})
		return
	}

	condos, err := h.ListUserCondominiums.Exec(r.Context(), userID)
	if err != nil {
		slog.Error("Failed to list user condominiums", "error", err)
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{
			Message: "Failed to fetch condominiums",
		})
		return
	}

	jsonutils.EncodeJson(w, r, http.StatusOK, condos)
}
