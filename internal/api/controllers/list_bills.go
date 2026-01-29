package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/google/uuid"
)

type ListBillsHandler struct {
	ListBills usecases.ListBillsUC
}

type ListBillsResponse struct {
	Data []struct {
		ID            uuid.UUID `json:"id"`
		CondominiumID uuid.UUID `json:"condominiumId"`
		ApartmentID   uuid.UUID `json:"apartmentId"`
		Status        string    `json:"status"`
		BillType      string    `json:"billType"`
		ValueInCents  int32     `json:"valueInCents"`
		DueDate       time.Time `json:"dueDate"`
		PaidAt        *string   `json:"paidAt,omitempty"`
		CreatedAt     string    `json:"createdAt"`
		UpdatedAt     *string   `json:"updatedAt,omitempty"`
		DigitableLine *string   `json:"digitableLine,omitempty"`
		PixCode       *string   `json:"pixCode,omitempty"`
	} `json:"data"`
}

// Handle lists bills based on filters
// @Summary      List Bills
// @Description  List bills. Admins see all (or filter). Residents only see their own apartment's bills.
// @Security     BearerAuth
// @Tags         Bills
// @Produce      json
// @Param        condominiumId query string true  "Condominium UUID"
// @Param        apartmentId   query string false "Apartment UUID (Required for Residents)"
// @Param        status        query string false "Filter by status (pending, paid, overdue, cancelled)"
// @Param        page          query int    false "Page number (default 1)"
// @Param        limit         query int    false "Items per page (default 10)"
// @Success      200 {object} controllers.ListBillsResponse
// @Failure      403 {object} common.ErrResponse "Permission denied"
// @Router       /bills [get]
func (h *ListBillsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusUnauthorized, common.ErrResponse{
			Message: "Unauthorized",
		})
		return
	}

	q := r.URL.Query()

	condoID, err := uuid.Parse(q.Get("condominiumId"))
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
			Message: "Invalid condominiumId",
		})
		return
	}

	var aptID *uuid.UUID
	if idStr := q.Get("apartmentId"); idStr != "" {
		id, err := uuid.Parse(idStr)
		if err == nil {
			aptID = &id
		}
	}

	var status *string
	if statusStr := q.Get("status"); statusStr != "" {
		status = &statusStr
	}

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit < 1 {
		limit = 20
	}

	offset := (page - 1) * limit

	params := usecases.ListBillsReq{
		CondominiumID: condoID,
		UserID:        userID,
		ApartmentID:   aptID,
		Status:        status,
		Limit:         int32(limit),
		Offset:        int32(offset),
	}

	bills, err := h.ListBills.Exec(r.Context(), params)
	if err != nil {
		if err == usecases.ErrNoPermission {
			jsonutils.EncodeJson(w, r, http.StatusForbidden, common.ErrResponse{
				Message: "Permission denied",
			})
			return
		}
		jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{
			Message: "Internal server error",
		})
		return
	}

	resp := ListBillsResponse{}
	resp.Data = make([]struct {
		ID            uuid.UUID `json:"id"`
		CondominiumID uuid.UUID `json:"condominiumId"`
		ApartmentID   uuid.UUID `json:"apartmentId"`
		Status        string    `json:"status"`
		BillType      string    `json:"billType"`
		ValueInCents  int32     `json:"valueInCents"`
		DueDate       time.Time `json:"dueDate"`
		PaidAt        *string   `json:"paidAt,omitempty"`
		CreatedAt     string    `json:"createdAt"`
		UpdatedAt     *string   `json:"updatedAt,omitempty"`
		DigitableLine *string   `json:"digitableLine,omitempty"`
		PixCode       *string   `json:"pixCode,omitempty"`
	}, len(bills))

	for i, bill := range bills {
		var (
			paidAt    *string
			updatedAt *string
		)
		if bill.PaidAt != nil {
			p := bill.PaidAt.Format(time.RFC3339)
			paidAt = &p
		}

		if bill.UpdatedAt != nil {
			u := bill.UpdatedAt.Format(time.RFC3339)
			updatedAt = &u
		}

		resp.Data[i].ID = bill.ID
		resp.Data[i].CondominiumID = bill.CondominiumID
		resp.Data[i].ApartmentID = bill.ApartmentID
		resp.Data[i].Status = bill.Status
		resp.Data[i].BillType = bill.BillType
		resp.Data[i].ValueInCents = int32(bill.ValueInCents)
		resp.Data[i].DueDate = bill.DueDate
		resp.Data[i].PaidAt = paidAt
		resp.Data[i].CreatedAt = bill.CreatedAt.Format(time.RFC3339)
		resp.Data[i].UpdatedAt = updatedAt
		resp.Data[i].DigitableLine = bill.DigitableLine
		resp.Data[i].PixCode = bill.PixCode
	}

	jsonutils.EncodeJson(w, r, http.StatusOK, resp)
}
