package controllers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type RevokeInviteHandler struct {
	RevokeInvite usecases.RevokeInviteUC
}

// Handle revokes an invite
// @Summary      Revoke Invite
// @Description  Cancels an existing invite. Only the user who created the invite can revoke it.
// @Tags         Invites
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Invite UUID"
// @Success      204  "No Content"
// @Failure      401  {object}  common.ErrResponse  "User not authenticated"
// @Failure      403  {object}  common.ErrResponse  "Permission denied"
// @Failure      404  {object}  common.ErrResponse  "Invite not found"
// @Failure      500  {object}  common.ErrResponse  "Internal server error"
// @Router       /invites/{id}/revoke [patch]
func (h *RevokeInviteHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusUnauthorized, common.ErrResponse{
			Message: "User not authenticated",
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	inviteID, err := uuid.Parse(idStr)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
			Message: "Invalid invite ID format",
		})
		return
	}

	err = h.RevokeInvite.Exec(r.Context(), usecases.RevokeInviteReq{
		InviteID: inviteID,
		UserID:   userID,
	})
	if err != nil {
		slog.Error("failed to validate invite",
			"error", err,
			"inviteId", inviteID,
			"userId", userID,
		)

		switch {
		case errors.Is(err, usecases.ErrInviteNotFound):
			jsonutils.EncodeJson(w, r, http.StatusNotFound, common.ErrResponse{
				Message: "Invite not found",
			})
		case errors.Is(err, usecases.ErrNoPermission):
			jsonutils.EncodeJson(w, r, http.StatusForbidden, common.ErrResponse{
				Message: "You do not have permission to revoke this invite",
			})
		case errors.Is(err, usecases.ErrInviteAlreadyRevoked):
			jsonutils.EncodeJson(w, r, http.StatusConflict, common.ErrResponse{
				Message: "Invite is already revoked",
			})
		default:
			jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{
				Message: "Failed to revoke invite",
			})
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
