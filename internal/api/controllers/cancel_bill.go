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
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CancelBillHandler struct {
	CancelBill usecases.CancelBillUC
}

type CancelBillRequest struct {
	CondominiumID uuid.UUID `json:"condominiumId" validate:"required"`
}

// Handle cancels a bill
// @Summary      Cancel Bill
// @Description  Cancels a pending or overdue bill. Cannot cancel paid bills. Only for Admins/Syndics.// @Security	   BearerAuth
// @Tags         Bills
// @Accept			 json
// @Param        id   path      string  true  "Bill UUID"
// @Param				 request body controllers.CancelBillRequest true "Condominium id"
// @Success      204  "No Content"
// @Failure      400  {object} common.ErrResponse "Invalid JSON or Logic Error"
// @Failure      422  {objct}  common.ValidationErrResponse "Validation failed"
// @Failure      409  {object} common.ErrResponse "Bill already paid or cancelled"
// @Failure      403  {object} common.ErrResponse "Permission denied"
// @Failure      500  {object} common.ErrResponse "Internal server error"
// @Router       /bills/{id}/cancel [patch]
func (h *CancelBillHandler) Handle(w http.ResponseWriter, r *http.Request) {
	data, err := jsonutils.DecodeJson[CancelBillRequest](r)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{Message: "Invalid JSON payload"})
		return
	}

	if errs := validator.ValidateStruct(data); len(errs) > 0 {
		jsonutils.EncodeJson(w, r, http.StatusUnprocessableEntity, common.ValidationErrResponse{Message: "Validation failed", Errors: errs})
		return
	}

	billIDStr := chi.URLParam(r, "id")
	billID, err := uuid.Parse(billIDStr)
	if err != nil {
		jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{
			Message: "Invalid UUID",
		})
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		jsonutils.EncodeJson(w, r, http.StatusUnauthorized, common.ErrResponse{Message: "Unauthorized"})
		return
	}

	params := usecases.CancelBillReq{
		BillID:        billID,
		CondominiumID: data.CondominiumID,
		UserID:        userID,
	}

	err = h.CancelBill.Exec(r.Context(), params)
	if err != nil {
		slog.Error("Error while cancelling bill",
			"error", err,
			"userId", userID,
			"condominiumId", data.CondominiumID,
			"billId", billID,
		)
		switch {
		case errors.Is(err, usecases.ErrBillNotFound):
			jsonutils.EncodeJson(w, r, http.StatusNotFound, common.ErrResponse{Message: "Bill not found"})
		case errors.Is(err, usecases.ErrNoPermission):
			jsonutils.EncodeJson(w, r, http.StatusForbidden, common.ErrResponse{Message: "Permission denied"})
		case errors.Is(err, usecases.ErrBillAlreadyCancelled), errors.Is(err, usecases.ErrCannotCancelPaidBill):
			jsonutils.EncodeJson(w, r, http.StatusConflict, common.ErrResponse{Message: err.Error()})

		default:
			jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{
				Message: "An unexpected error occured",
			})
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
