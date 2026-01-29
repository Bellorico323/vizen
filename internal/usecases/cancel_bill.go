package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CancelBillUC interface {
	Exec(ctx context.Context, req CancelBillReq) error
}

type CancelBillReq struct {
	BillID        uuid.UUID
	CondominiumID uuid.UUID
	UserID        uuid.UUID
}

type CancelBillUseCase struct {
	querier pgstore.Querier
}

func NewCancelBillUseCase(q pgstore.Querier) *CancelBillUseCase {
	return &CancelBillUseCase{
		querier: q,
	}
}

var (
	ErrBillAlreadyCancelled = errors.New("bill is already cancelled")
	ErrCannotCancelPaidBill = errors.New("cannot cancel a paid bill")
)

func (uc *CancelBillUseCase) Exec(ctx context.Context, req CancelBillReq) error {
	role, err := uc.querier.GetCondominiumMemberRole(ctx, pgstore.GetCondominiumMemberRoleParams{
		CondominiumID: req.CondominiumID,
		UserID:        req.UserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoPermission
		}
		return fmt.Errorf("error checking role: %w", err)
	}

	if role != "admin" && role != "syndic" {
		return ErrNoPermission
	}

	bill, err := uc.querier.GetBillById(ctx, pgstore.GetBillByIdParams{
		ID:            req.BillID,
		CondominiumID: req.CondominiumID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrBillNotFound
		}
		return fmt.Errorf("error fetching bill: %w", err)
	}

	if bill.CondominiumID != req.CondominiumID {
		return ErrBillNotFound
	}

	if bill.Status == "cancelled" {
		return ErrBillAlreadyCancelled
	}

	if bill.Status == "paid" {
		return ErrCannotCancelPaidBill
	}

	bill, err = uc.querier.UpdateBillStatus(ctx, pgstore.UpdateBillStatusParams{
		ID:            req.BillID,
		CondominiumID: req.CondominiumID,
		Status:        "cancelled",
	})

	if err != nil {
		return fmt.Errorf("error updating bill status: %w", err)
	}

	return nil
}
