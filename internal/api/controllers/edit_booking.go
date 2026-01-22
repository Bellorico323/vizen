package controllers

import (
	"errors"
	"net/http"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/Bellorico323/vizen/internal/validator"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type EditBookingHandler struct {
	EditBooking usecases.EditBookingtUC
}

type EditBookingRequest struct {
	Status string `json:"status" validate:"required,oneof=confirmed denied"`
}

// Handle processes the booking approval or rejection
// @Summary      Approve or Reject Booking
// @Description  Allows an Admin or Syndic to approve/deny a pending booking.
// @Tags         Bookings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id           path      string  true  "Booking UUID"
// @Param        request body      controllers.EditBookingRequest true "New Status (confirmed/denied)"
// @Success      204     "No Content"
// @Failure      400     {object}  common.ErrResponse            "Invalid Payload or ID"
// @Failure      401     {object}  common.ErrResponse            "Unauthorized"
// @Failure      403     {object}  common.ErrResponse            "Permission Denied"
// @Failure      404     {object}  common.ErrResponse            "Booking Not Found"
// @Failure      409     {object}  common.ErrResponse            "Conflict (Booking not pending)"
// @Failure      500     {object}  common.ErrResponse            "Internal Server Error"
// @Router       /bookings/{id}/status [patch]
func (h *EditBookingHandler) Handle(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	bookingID, err := uuid.Parse(idStr)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
			Message: "Invalid booking ID format",
		})
		return
	}

	data, err := jsonutils.DecodeJson[EditBookingRequest](r)
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

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusUnauthorized, common.ErrResponse{
			Message: "User not authenticated",
		})
		return
	}

	err = h.EditBooking.Exec(r.Context(), usecases.EditBookingReq{
		BookingID: bookingID,
		UserID:    userID,
		Status:    data.Status,
	})

	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrBookingNotFound):
			jsonutils.EncodeJson(w, r, http.StatusNotFound, common.ErrResponse{
				Message: "Booking not found",
			})

		case errors.Is(err, usecases.ErrBookingNotPending):
			jsonutils.EncodeJson(w, r, http.StatusConflict, common.ErrResponse{
				Message: "Booking is not in pending state and cannot be modified",
			})

		case errors.Is(err, usecases.ErrNoPermission):
			jsonutils.EncodeJson(w, r, http.StatusForbidden, common.ErrResponse{
				Message: "You do not have permission to review bookings",
			})

		case errors.Is(err, usecases.ErrInvalidBookingStatus):
			jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
				Message: "Invalid status provided",
			})

		default:
			jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{
				Message: "An unexpected error occurred while updating booking status",
			})
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
