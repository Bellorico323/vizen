package controllers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/Bellorico323/vizen/internal/validator"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type CreateApartmentHandler struct {
	CreateApartment usecases.CreateApartmentUC
}

type CreateApartmentRequest struct {
	CondominiumID uuid.UUID `json:"condominiumId" validate:"required"`
	Block         string    `json:"block,omitempty" validate:"omitempty"`
	Number        string    `json:"number" validate:"required,min=1"`
}

type CreateApartmentResponse struct {
	Message     string    `json:"message"`
	ApartmentID uuid.UUID `json:"apartmentId"`
}

// Handle executes the apartment creation
// @Summary 		Create Apartment
// @Description Creates a new apartment. The user must be an **admin** or **syndic** of the condominium.
// @Security 		BearerAuth
// @Tags			Apartments
// @Accept			json
// @Produce			json
// @Param			request body controllers.CreateApartmentRequest true "Apartment creation payload"
// @Success			201 {object} controllers.CreateApartmentResponse "Apartment successfully created"
// @Failure 		400 {object} common.ErrResponse "Invalid JSON payload"
// @Failure 		401 {object} common.ErrResponse	"User not authenticated"
// @Failure			403	{object} common.ErrResponse	"User does not have permission"
// @Failure			404	{object} common.ErrResponse	"Condominium not found"
// @Failure			409	{object} common.ErrResponse	"Apartment already exists in this block"
// @Failure			422 {object} common.ValidationErrResponse "Validation failed"
// @Failure			500 {object} common.ErrResponse "Internal server error"
// @Router			/apartments	[post]
func (h *CreateApartmentHandler) Handle(w http.ResponseWriter, r *http.Request) {
	data, err := jsonutils.DecodeJson[CreateApartmentRequest](r)
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
			Message: "User not found",
		})
		return
	}

	payload := usecases.CreateApartmentReq{
		CondominiumID: data.CondominiumID,
		Block:         data.Block,
		Number:        data.Number,
		UserID:        userId,
	}

	id, err := h.CreateApartment.Exec(r.Context(), payload)
	if err != nil {
		slog.Error("Error while creating apartment", "error", err)

		switch {
		case errors.Is(err, usecases.ErrNoPermission):
			jsonutils.EncodeJson(w, r, http.StatusForbidden, common.ErrResponse{
				Message: err.Error(),
			})

		case errors.Is(err, usecases.ErrCondominiumNotFound):
			jsonutils.EncodeJson(w, r, http.StatusNotFound, common.ErrResponse{
				Message: "Condominium not found",
			})

		default:
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				jsonutils.EncodeJson(w, r, http.StatusConflict, common.ErrResponse{
					Message: "This apartment number already exists in this block",
				})
				return
			}

			jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{
				Message: "An unexpected error occurred while creating apartment",
			})
		}
		return
	}

	jsonutils.EncodeJson(w, r, http.StatusCreated, CreateApartmentResponse{
		Message:     "Apartment successfully created",
		ApartmentID: id,
	})
}
