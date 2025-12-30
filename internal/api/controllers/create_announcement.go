package controllers

import (
	"errors"
	"net/http"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/Bellorico323/vizen/internal/validator"
	"github.com/google/uuid"
)

type CreateAnnouncementHandler struct {
	CreateAnnouncement usecases.CreateAnnouncementUC
}

type CreateAnnouncementRequest struct {
	CondominiumID uuid.UUID `json:"condominiumId" validate:"required"`
	Title         string    `json:"title" validate:"required"`
	Content       string    `json:"content" validate:"required"`
}

// Handle creates a new announcement and notifies residents
// @Summary 		Create Announcement
// @Description	Creates a new announcement for a condominium and sends push notifications to all residents.
// @Security		BearerAuth
// @Tags			Announcements
// @Accept			json
// @Produce			json
// @Param			request body controllers.CreateAnnouncementRequest true "Announcement payload"
// @Success			201 {object} common.SuccessResponse "Announcement created successfully"
// @Failure 		400	{object} common.ErrResponse "Invalid JSON or Empty fields"
// @Failure			401 {object} common.ErrResponse "User not authenticated"
// @Failure			403 {object} common.ErrResponse "User does not have permission (Must be Admin/Syndic)"
// @Failure			422 {object} common.ValidationErrResponse "Validation failed"
// @Failure			500 {object} common.ErrResponse	"Internal server error"
// @Router			/announcements [post]
func (h *CreateAnnouncementHandler) Handle(w http.ResponseWriter, r *http.Request) {
	data, err := jsonutils.DecodeJson[CreateAnnouncementRequest](r)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
			Message: "Invalid JSON payload",
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

	userId, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusUnauthorized, common.ErrResponse{
			Message: "User not authenticated",
		})
		return
	}

	err = h.CreateAnnouncement.Exec(r.Context(), &usecases.CreateAnnouncementReq{
		AuthorID:      userId,
		CondominiumID: data.CondominiumID,
		Title:         data.Title,
		Content:       data.Content,
	})

	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrNoPermission):
			jsonutils.EncodeJson(w, r, http.StatusForbidden, common.ErrResponse{
				Message: err.Error(),
			})
		case errors.Is(err, usecases.ErrEmptyTitle):
			jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
				Message: err.Error(),
			})
		case errors.Is(err, usecases.ErrEmptyContent):
			jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
				Message: err.Error(),
			})
		default:
			jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{
				Message: "An unexpected error occurred while creating announcement",
			})
		}

		return
	}

	jsonutils.EncodeJson(w, r, http.StatusCreated, common.SuccessResponse{
		Message: "Announcement created successfully",
	})
}
