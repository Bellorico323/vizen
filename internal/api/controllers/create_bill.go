package controllers

import (
	"errors"
	"net/http"
	"time"

	"github.com/Bellorico323/vizen/internal/api/common"
	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/Bellorico323/vizen/internal/usecases"
	"github.com/Bellorico323/vizen/internal/validator"
	"github.com/google/uuid"
)

type CreateBillHandler struct {
	CreateBill usecases.CreateBillUC
}

type CreateBillRequest struct {
	CondominiumID uuid.UUID `json:"condominiumId" validate:"required"`
	ApartmentID   uuid.UUID `json:"apartmentId" validate:"required"`
	BillType      string    `json:"billType" validate:"required,oneof=rent condominium_fee water electricity gas fine"`
	ValueInCents  int64     `json:"valueInCents" validate:"required,gte=0"`
	DueDate       time.Time `json:"dueDate" validate:"required"`
	DigitableLine *string   `json:"digitableLine"`
	PixCode       *string   `json:"pixCode"`
}

type CreateBillResponse struct {
	ID            uuid.UUID  `json:"id"`
	CondominiumID uuid.UUID  `json:"condominiumId"`
	ApartmentID   uuid.UUID  `json:"apartmentId"`
	BillType      string     `json:"billType"`
	Status        string     `json:"status"`
	ValueInCents  int64      `json:"valueInCents"`
	DueDate       time.Time  `json:"dueDate"`
	DigitableLine *string    `json:"digitableLine,omitempty"`
	PixCode       *string    `json:"pixCode,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	PaidAt        *time.Time `json:"paidAt,omitempty"`
}

// Handle creates a new bill for an apartment
// @Summary      Create Bill
// @Description  Creates a new financial bill (boleto/pix) for a specific apartment. Only Admins/Syndics.
// @Security     BearerAuth
// @Tags         Bills
// @Accept       json
// @Produce      json
// @Param        request body controllers.CreateBillRequest true "Bill Information"
// @Success      201 {object} controllers.CreateBillResponse
// @Failure      400 {object} common.ErrResponse "Invalid JSON or Logic Error"
// @Failure      401 {object} common.ErrResponse "Unauthorized"
// @Failure      403 {object} common.ErrResponse "Permission denied"
// @Failure      404 {object} common.ErrResponse "Apartment not found"
// @Failure      422 {object} common.ValidationErrResponse "Validation failed"
// @Failure      500 {object} common.ErrResponse "Internal server error"
// @Router       /bills [post]
func (h *CreateBillHandler) Handle(w http.ResponseWriter, r *http.Request) {
	data, err := jsonutils.DecodeJson[CreateBillRequest](r)
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

	req := usecases.CreateBillReq{
		UserID:        userID,
		CondominiumID: data.CondominiumID,
		ApartmentID:   data.ApartmentID,
		BillType:      data.BillType,
		ValueInCents:  data.ValueInCents,
		DueDate:       data.DueDate,
		DigitableLine: data.DigitableLine,
		PixCode:       data.PixCode,
	}

	bill, err := h.CreateBill.Exec(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrNoPermission):
			jsonutils.EncodeJson(w, r, http.StatusForbidden, common.ErrResponse{Message: "You do not have permission to create bills"})
		case errors.Is(err, usecases.ErrApartmentNotFound):
			jsonutils.EncodeJson(w, r, http.StatusNotFound, common.ErrResponse{Message: "Apartment not found"})
		case errors.Is(err, usecases.ErrApartmentIsNotFromCondominium):
			jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{Message: "Apartment does not belong to this condominium"})
		case errors.Is(err, usecases.ErrInvalidBillType), errors.Is(err, usecases.ErrInvalidBillDueDate), errors.Is(err, usecases.ErrInvalidAmount):
			jsonutils.EncodeJson(w, r, http.StatusBadRequest, common.ErrResponse{Message: err.Error()})
		default:
			jsonutils.EncodeJson(w, r, http.StatusInternalServerError, common.ErrResponse{Message: "Internal server error"})
		}
		return
	}

	resp := CreateBillResponse{
		ID:            bill.ID,
		CondominiumID: bill.CondominiumID,
		ApartmentID:   bill.ApartmentID,
		BillType:      bill.BillType,
		Status:        bill.Status,
		ValueInCents:  bill.ValueInCents,
		DueDate:       bill.DueDate,
		DigitableLine: bill.DigitableLine,
		PixCode:       bill.PixCode,
		CreatedAt:     bill.CreatedAt,
		PaidAt:        bill.PaidAt,
	}

	jsonutils.EncodeJson(w, r, http.StatusCreated, resp)
}
