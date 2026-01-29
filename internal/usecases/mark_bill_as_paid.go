package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type MarkBillAsPaidUC interface {
	Exec(ctx context.Context, req MarkBillAsPaidReq) error
}

type MarkBillAsPaidReq struct {
	BillID        uuid.UUID
	CondominiumID uuid.UUID
	UserID        uuid.UUID
}

type MarkBillAsPaidUseCase struct {
	querier pgstore.Querier
}

func NewMarkBillASPaidUseCase(q pgstore.Querier) *MarkBillAsPaidUseCase {
	return &MarkBillAsPaidUseCase{
		querier: q,
	}
}

var (
	ErrBillNotFound    = errors.New("bill not found")
	ErrBillAlreadyPaid = errors.New("bill is already paid")
	ErrBillCancelled   = errors.New("cannot pay a cancelled bill")
)

func (uc *MarkBillAsPaidUseCase) Exec(ctx context.Context, req MarkBillAsPaidReq) error {
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
		return ErrBillCancelled
	}

	if bill.Status == "paid" {
		return ErrBillAlreadyPaid
	}

	bill, err = uc.querier.UpdateBillStatus(ctx, pgstore.UpdateBillStatusParams{
		ID:            req.BillID,
		CondominiumID: req.CondominiumID,
		Status:        "paid",
	})

	if err != nil {
		return fmt.Errorf("error updating bill status: %w", err)
	}

	return nil
}
