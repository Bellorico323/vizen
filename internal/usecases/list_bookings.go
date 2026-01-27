package usecases

import (
	"context"
	"time"

	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/google/uuid"
)

type ListBookingsUC interface {
	Exec(ctx context.Context, req ListBookingsReq) ([]pgstore.ListBookingsRow, error)
}

type ListBookingsReq struct {
	CondominiumID    uuid.UUID
	RequestingUserID uuid.UUID
	TargetUserID     *uuid.UUID
	CommonAreaID     *uuid.UUID
	FromDate         *time.Time
	ToDate           *time.Time
}

type ListBookingsUseCase struct {
	querier pgstore.Querier
}

func NewListBookingsUseCase(q pgstore.Querier) *ListBookingsUseCase {
	return &ListBookingsUseCase{
		querier: q,
	}
}

func (uc *ListBookingsUseCase) Exec(ctx context.Context, req ListBookingsReq) ([]pgstore.ListBookingsRow, error) {
	role, err := uc.querier.GetCondominiumMemberRole(ctx, pgstore.GetCondominiumMemberRoleParams{
		CondominiumID: req.CondominiumID,
		UserID:        req.RequestingUserID,
	})

	hasPermission := err == nil && (role == "admin" || role == "syndic")

	filterUserID := req.TargetUserID

	if !hasPermission {
		filterUserID = &req.RequestingUserID
	}

	bookings, err := uc.querier.ListBookings(ctx, pgstore.ListBookingsParams{
		CondominiumID: req.CondominiumID,
		UserID:        filterUserID,
		CommonAreaID:  req.CommonAreaID,
		FromDate:      req.FromDate,
		ToDate:        req.ToDate,
	})
	if err != nil {
		return nil, err
	}

	return bookings, nil
}
