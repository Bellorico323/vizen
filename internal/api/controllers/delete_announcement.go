package controllers

import (
	"errors"
	"net/http"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type DeleteAnnouncementHandler struct {
	DeleteAnnouncement usecases.DeleteAnnouncementUC
}

// Handle deletes an announcement
// @Summary 		Delete Announcement
// @Description	Removes an announcement. Requires Admin or Syndic role.
// @Security		BearerAuth
// @Tags			Announcements
// @Param			id path string true "Announcement UUID"
// @Success			204 "No Content"
// @Failure 		400	{object} common.ErrResponse "Invalid ID"
// @Failure			401 {object} common.ErrResponse "User not authenticated"
// @Failure			403 {object} common.ErrResponse "Permission denied"
// @Failure			404 {object} common.ErrResponse "Announcement not found"
// @Failure			500 {object} common.ErrResponse	"Internal server error"
// @Router			/announcements/{id} [delete]
func (h *DeleteAnnouncementHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusUnauthorized, common.ErrResponse{
			Message: "User not authenticated",
		})
		return
	}

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
			Message: "Announcement ID is required",
		})
		return
	}

	announcementID, err := uuid.Parse(idStr)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
			Message: "Invalid UUID format",
		})
		return
	}

	err = h.DeleteAnnouncement.Exec(r.Context(), usecases.DeleteAnnouncementReq{
		AnnouncementID: announcementID,
		UserID:         userID,
	})

	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrAnnouncementNotFound):
			jsonutils.EncodeJson(w, r, http.StatusNotFound, common.ErrResponse{
				Message: "Announcement not found",
			})
		case errors.Is(err, usecases.ErrNoPermission):
			jsonutils.EncodeJson(w, r, http.StatusForbidden, common.ErrResponse{
				Message: "You do not have permission to delete this announcement",
			})
		default:
			jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{
				Message: "An unexpected error occured",
			})
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
