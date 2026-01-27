package controllers

import (
	"net/http"
	"time"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/google/uuid"
)

type ListBookingsHandler struct {
	ListBookings usecases.ListBookingsUC
}

type ListBookingsResponse struct {
	Bookings []pgstore.ListBookingsRow `json:"bookings"`
}

// Handle lists bookings based on filters
// @Summary      List Bookings
// @Description  Get a list of bookings. Admin can filter by user. Residents are forced to see their own history unless they use public availability endpoint.
// @Tags         Bookings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        condominiumId query     string  true   "Condominium UUID"
// @Param        userId        query     string  false  "Filter by User UUID (Optional)"
// @Param        commonAreaId  query     string  false  "Filter by Common Area UUID (Optional)"
// @Param        from          query     string  false  "Start Date (ISO 8601 e.g. 2026-01-01T00:00:00Z)"
// @Param        to            query     string  false  "End Date (ISO 8601)"
// @Success      200     {object}  controllers.ListBookingsResponse
// @Failure      400     {object}  common.ErrResponse "Invalid UUID or Date format"
// @Failure      401     {object}  common.ErrResponse "Unauthorized"
// @Failure      500     {object}  common.ErrResponse "Internal Server Error"
// @Router       /bookings [get]
func (h *ListBookingsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusUnauthorized, common.ErrResponse{
			Message: "Unauthorized",
		})
		return
	}

	q := r.URL.Query()

	condoIdStr := q.Get("condominiumId")
	condoID, err := uuid.Parse(condoIdStr)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
			Message: "Invalid condominiumId",
		})
		return
	}

	var targetUserID *uuid.UUID
	if val := q.Get("userId"); val != "" {
		id, err := uuid.Parse(val)
		if err != nil {
			jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{Message: "Invalid userId format"})
			return
		}
		targetUserID = &id
	}

	var commonAreaID *uuid.UUID
	if val := q.Get("commonAreaId"); val != "" {
		id, err := uuid.Parse(val)
		if err != nil {
			jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{Message: "Invalid commonAreaId format"})
			return
		}
		commonAreaID = &id
	}

	var fromDate *time.Time
	if val := q.Get("from"); val != "" {
		t, err := time.Parse(time.RFC3339, val)
		if err != nil {
			jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{Message: "Invalid 'from' date format. Use ISO8601"})
			return
		}
		fromDate = &t
	}

	var toDate *time.Time
	if val := q.Get("to"); val != "" {
		t, err := time.Parse(time.RFC3339, val)
		if err != nil {
			jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{Message: "Invalid 'to' date format. Use ISO8601"})
			return
		}
		toDate = &t
	}

	bookings, err := h.ListBookings.Exec(r.Context(), usecases.ListBookingsReq{
		CondominiumID:    condoID,
		RequestingUserID: userID,
		TargetUserID:     targetUserID,
		CommonAreaID:     commonAreaID,
		FromDate:         fromDate,
		ToDate:           toDate,
	})
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{
			Message: "Failed to fetch bookings",
		})
		return
	}

	if bookings == nil {
		bookings = []pgstore.ListBookingsRow{}
	}

	jsonutils.EncodeJson(w, r, http.StatusOK, ListBookingsResponse{
		Bookings: bookings,
	})
}
