package controllers

import (
	"net/http"
	"time"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type GetAreaAvailabilityHandler struct {
	GetAreaAvailability usecases.GetAreaAvailabilityUC
}

type GetAreaAvailabilityResponse struct {
	Bookings []pgstore.GetAreaAvailabilityRow `json:"bookings"`
}

// Handle lists occupied slots for a common area
// @Summary      Get Area Availability
// @Description  Returns a list of occupied time slots for a specific common area within a date range. Used to populate calendars.
// @Tags         Common Areas
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string  true  "Common Area UUID"
// @Param        from  query     string  true  "Start Date (ISO 8601 e.g. 2026-01-01T00:00:00Z)"
// @Param        to    query     string  true  "End Date (ISO 8601)"
// @Success      200   {object}  controllers.GetAreaAvailabilityResponse
// @Failure      400   {object}  common.ErrResponse "Invalid ID or Date format"
// @Failure      401   {object}  common.ErrResponse "Unauthorized"
// @Failure      500   {object}  common.ErrResponse "Internal Server Error"
// @Router       /common-areas/{id}/availability [get]
func (h *GetAreaAvailabilityHandler) Handle(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	commonAreaID, err := uuid.Parse(idStr)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
			Message: "Invalid common area ID format",
		})
		return
	}

	fromStr := r.URL.Query().Get("from")
	if fromStr == "" {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{Message: "Missing required parameter: from"})
		return
	}
	fromDate, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
			Message: "Invalid from date",
		})
		return
	}

	toStr := r.URL.Query().Get("to")
	if toStr == "" {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{Message: "Missing required parameter: to"})
		return
	}
	toDate, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
			Message: "Invalid to date",
		})
		return
	}

	bookings, err := h.GetAreaAvailability.Exec(r.Context(), usecases.GetAreaAvailabilityReq{
		CommonAreaID: commonAreaID,
		FromDate:     fromDate,
		ToDate:       toDate,
	})
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{
			Message: "Failed to fetch bookings",
		})
		return
	}

	if bookings == nil {
		bookings = []pgstore.GetAreaAvailabilityRow{}
	}

	jsonutils.EncodeJson(w, r, http.StatusOK, GetAreaAvailabilityResponse{
		Bookings: bookings,
	})

}
