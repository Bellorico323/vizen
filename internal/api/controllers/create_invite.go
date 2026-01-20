package controllers

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/Bellorico323/vizen/internal/validator"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type CreateInviteHandler struct {
	CreateInvite usecases.CreateInviteUC
}

type CreateInviteRequest struct {
	CondominiumID uuid.UUID `json:"condominiumId" validate:"required,uuid4"`
	ApartmentID   uuid.UUID `json:"apartmentId" validate:"required,uuid4"`
	GuestName     string    `json:"guestName" validate:"required,min=1"`
	GuestType     *string   `json:"guestType" validate:"omitempty,oneof=guest service delivery"`
	StartsAt      time.Time `json:"startsAt" validate:"required"`
	EndsAt        time.Time `json:"endsAt" validate:"required"`
}

type CreateInviteResponse struct {
	Message string         `json:"message"`
	Invite  pgstore.Invite `json:"invite"`
}

// Handle creates a new invite
// @Summary      Create Guest Invite
// @Description  Generates a new access invite (QR Code Token). Only residents of the specified apartment can create invites. Dates must be in ISO8601 format.
// @Tags         Invites
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body      controllers.CreateInviteRequest true "Invite Creation Data"
// @Success      201     {object}  controllers.CreateInviteResponse
// @Failure      400     {object}  common.ErrResponse            "Invalid Payload or Logic Error (e.g., End date before Start date)"
// @Failure      401     {object}  common.ErrResponse            "User not authenticated"
// @Failure      403     {object}  common.ErrResponse            "Permission denied (User does not reside in this apartment)"
// @Failure      409     {object}  common.ErrResponse            "Conflict (Condominium or Apartment not found)"
// @Failure      422     {object}  common.ValidationErrResponse  "Validation Failed"
// @Failure      500     {object}  common.ErrResponse            "Internal server error"
// @Router       /invites [post]
func (h *CreateInviteHandler) Handle(w http.ResponseWriter, r *http.Request) {
	data, err := jsonutils.DecodeJson[CreateInviteRequest](r)
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

	payload := usecases.CreateInviteReq{
		CondominiumID: data.CondominiumID,
		ApartmentID:   data.ApartmentID,
		IssuedBy:      userId,
		GuestName:     data.GuestName,
		GuestType:     data.GuestType,
		StartsAt:      data.StartsAt,
		EndsAt:        data.EndsAt,
	}

	invite, err := h.CreateInvite.Exec(r.Context(), payload)
	if err != nil {
		slog.Error("Error while creating invite",
			"error", err,
			"condominiumId", data.CondominiumID,
			"apartmentId", data.ApartmentID,
			"issuedBy", userId,
		)

		switch {
		case errors.Is(err, usecases.ErrNoPermission):
			jsonutils.EncodeJson(w, r, http.StatusForbidden, common.ErrResponse{
				Message: err.Error(),
			})

		case errors.Is(err, usecases.ErrGuestNameIsRequired):
			jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
				Message: err.Error(),
			})

		case errors.Is(err, usecases.ErrInviteInvalidTimeRange):
			jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
				Message: err.Error(),
			})

		case errors.Is(err, usecases.ErrInviteInThePast):
			jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
				Message: err.Error(),
			})

		case errors.Is(err, usecases.ErrApartmentNotFound):
			jsonutils.EncodeJson(w, r, http.StatusNotFound, common.ErrResponse{Message: "Apartment not found"})

		case errors.Is(err, usecases.ErrApartmentIsNotFromCondominium):
			jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{Message: "Apartment does not belong to this condominium"})

		default:
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				jsonutils.EncodeJson(w, r, http.StatusConflict, common.ErrResponse{
					Message: "Condominium or apartment not found",
				})
			}
			jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{
				Message: "An unexpected error occurred while creating apartment",
			})
		}

		return
	}

	jsonutils.EncodeJson(w, r, http.StatusCreated, CreateInviteResponse{
		Message: "Invite successfully created",
		Invite:  invite,
	})
}
