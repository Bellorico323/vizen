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
)

type CreateBookingsHandler struct {
	CreateBooking usecases.CreateBookingUC
}

type CreateBookingRequest struct {
	CondominiumID uuid.UUID `json:"condominiumId" validate:"required"`
	ApartmentID   uuid.UUID `json:"apartmentId" validate:"required"`
	CommonAreaID  uuid.UUID `json:"commonAreaId" validate:"required"`
	StartsAt      time.Time `json:"startsAt" validate:"required"`
	EndsAt        time.Time `json:"endsAt" validate:"required,gtfield=StartsAt"`
}

type CreateBookingResponse struct {
	Message string          `json:"message"`
	Booking pgstore.Booking `json:"booking"`
}

// Handle creates a new booking
// @Summary      Book Common Area
// @Description  Create a reservation for a common area (e.g. Party Hall). Prevents double booking via locking.
// @Tags         Bookings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body      controllers.CreateBookingRequest true "Booking Data"
// @Success      201     {object}  controllers.CreateBookingResponse
// @Failure      409     {object}  common.ErrResponse "Time slot conflict"
// @Failure      403     {object}  common.ErrResponse "Permission denied"
// @Router       /bookings [post]
func (h *CreateBookingsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	data, err := jsonutils.DecodeJson[CreateBookingRequest](r)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{Message: "Invalid JSON payload"})
		return
	}

	if errs := validator.ValidateStruct(data); len(errs) > 0 {
		jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, common.ValidationErrResponse{Message: "Validation failed", Errors: errs})
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusUnauthorized, common.ErrResponse{Message: "Unauthorized"})
		return
	}

	booking, err := h.CreateBooking.Exec(r.Context(), usecases.CreateBookingReq{
		CondominiumID: data.CondominiumID,
		UserID:        userID,
		ApartmentID:   data.ApartmentID,
		CommonAreaID:  data.CommonAreaID,
		StartsAt:      data.StartsAt,
		EndsAt:        data.EndsAt,
	})

	if err != nil {
		slog.Error("Error while creating booking",
			"error", err,
			"condominiumId", data.CondominiumID,
			"apartmentId", data.ApartmentID,
			"commonAreaId", data.CommonAreaID,
			"userId", userID,
		)

		switch {
		case errors.Is(err, usecases.ErrTimeSlotTaken):
			jsonutils.EncodeJson(w, r, http.StatusConflict, common.ErrResponse{
				Message: "This time slot is already booked by another resident.",
			})

		case errors.Is(err, usecases.ErrNoPermission):
			jsonutils.EncodeJson(w, r, http.StatusForbidden, common.ErrResponse{
				Message: "You must be a resident of the apartment to create a booking.",
			})

		case errors.Is(err, usecases.ErrInvalidBookingDate):
			jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
				Message: err.Error(),
			})

		default:
			jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{
				Message: "Failed to create booking",
			})
		}
		return
	}

	message := "Booking confirmed successfully"
	if booking.Status == "pending" {
		message = "Booking request received. Waiting for syndic approval."
	}

	jsonutils.EncodeJson(w, r, http.StatusCreated, CreateBookingResponse{
		Message: message,
		Booking: booking,
	})
}
