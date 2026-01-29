package usecases

import (
	"context"
	"errors"

	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ListBillsUC interface {
	Exec(ctx context.Context, req ListBillsReq) ([]pgstore.Bill, error)
}

type ListBillsReq struct {
	CondominiumID uuid.UUID
	UserID        uuid.UUID
	ApartmentID   *uuid.UUID
	Status        *string
	Limit         int32
	Offset        int32
}

type ListBillsUseCase struct {
	querier pgstore.Querier
}

func NewListBillsUseCase(q pgstore.Querier) *ListBillsUseCase {
	return &ListBillsUseCase{querier: q}
}

func (uc *ListBillsUseCase) Exec(ctx context.Context, req ListBillsReq) ([]pgstore.Bill, error) {
	role, err := uc.querier.GetCondominiumMemberRole(ctx, pgstore.GetCondominiumMemberRoleParams{
		CondominiumID: req.CondominiumID,
		UserID:        req.UserID,
	})

	isStaff := false
	if err == nil && (role == "admin" || role == "syndic") {
		isStaff = true
	} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	if !isStaff {
		if req.ApartmentID == nil {
			return nil, ErrNoPermission
		}

		isResident, err := uc.querier.CheckIsResident(ctx, pgstore.CheckIsResidentParams{
			UserID:      req.UserID,
			ApartmentID: *req.ApartmentID,
		})
		if err != nil {
			return nil, err
		}

		if !isResident {
			return nil, ErrNoPermission
		}

	}

	bills, err := uc.querier.ListBills(ctx, pgstore.ListBillsParams{
		CondominiumID: req.CondominiumID,
		ApartmentID:   req.ApartmentID,
		Status:        req.Status,
		Limit:         req.Limit,
		Offset:        req.Offset,
	})
	if err != nil {
		return nil, err
	}

	return bills, nil
}
