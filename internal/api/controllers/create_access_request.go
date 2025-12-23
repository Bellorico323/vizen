package controllers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/Bellorico323/vizen/internal/validator"
	"github.com/google/uuid"
)

type CreateAccessRequestHandler struct {
	CreateAccessRequest usecases.CreateAccessRequestUC
}

type CreateAccessRequestReq struct {
	UserID      uuid.UUID `json:"userId" validate:"required"`
	ApartmentID uuid.UUID `json:"apartmentId" validate:"required"`
	Type        string    `json:"type" validate:"required,oneof=owner tenant dependent"`
}

type CreateAccessRequestResponse struct {
	Message   string    `json:"message"`
	RequestID uuid.UUID `json:"requestId"`
}

// Handle executes the creation of an access request
// @Summary 		Create Access Request
// @Description	Submits a request for a user to join an apartment (as owner, tenant, or dependent).
// @Tags			Access Requests
// @Accept			json
// @Produce			json
// @Security		BearerAuth
// @Param			request body controllers.CreateAccessRequestReq true "Access Request payload"
// @Success			201 {object} controllers.CreateAccessRequestResponse "Request created successfully"
// @Failure 		400	{object} common.ErrResponse "Invalid JSON or Resident Type"
// @Failure			401 {object} common.ErrResponse "User not authenticated"
// @Failure			404 {object} common.ErrResponse "User or Apartment not found"
// @Failure			422 {object} common.ValidationErrResponse "Validation failed"
// @Failure			500 {object} common.ErrResponse	"Internal server error"
// @Router			/access-requests [post]
func (h *CreateAccessRequestHandler) Handle(w http.ResponseWriter, r *http.Request) {
	data, err := jsonutils.DecodeJson[CreateAccessRequestReq](r)
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

	payload := usecases.CreateAccessRequestReq{
		UserID:      data.UserID,
		ApartmentID: data.ApartmentID,
		Type:        data.Type,
	}

	id, err := h.CreateAccessRequest.Exec(r.Context(), payload)
	if err != nil {
		slog.Error("Error while creating access request", "error", err)

		switch {
		case errors.Is(err, usecases.ErrApartmentNotFound):
			jsonutils.EncodeJson(w, r, http.StatusNotFound, common.ErrResponse{
				Message: "Apartment not found",
			})
			return

		case errors.Is(err, usecases.ErrInvalidResidentType):
			jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
				Message: "Invalid resident type",
			})
			return

		case errors.Is(err, usecases.ErrUserNotFound):
			jsonutils.EncodeJson(w, r, http.StatusNotFound, common.ErrResponse{
				Message: "User not found",
			})
			return

		default:
			jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{
				Message: "An unexpected error occurred while creating apartment",
			})
			return
		}

	}

	jsonutils.EncodeJson(w, r, http.StatusCreated, CreateAccessRequestResponse{
		Message:   "Access request successfully created",
		RequestID: id,
	})
}
