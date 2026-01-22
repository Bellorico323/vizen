package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CreateBookingUC interface {
	Exec(ctx context.Context, req CreateBookingReq) (pgstore.Booking, error)
}

type CreateBookingReq struct {
	CondominiumID uuid.UUID
	UserID        uuid.UUID
	ApartmentID   uuid.UUID
	CommonAreaID  uuid.UUID
	StartsAt      time.Time
	EndsAt        time.Time
}

type CreateBookingUseCase struct {
	pool    *pgxpool.Pool
	querier pgstore.Querier
}

func NewCreateBookingUseCase(pool *pgxpool.Pool, q pgstore.Querier) *CreateBookingUseCase {
	return &CreateBookingUseCase{
		pool:    pool,
		querier: q,
	}
}

var (
	ErrTimeSlotTaken      = errors.New("the selected time slot is already booked")
	ErrInvalidBookingDate = errors.New("start date must be before end date and in the future")
)

func (uc *CreateBookingUseCase) Exec(ctx context.Context, req CreateBookingReq) (pgstore.Booking, error) {
	if req.EndsAt.Before(req.StartsAt) {
		return pgstore.Booking{}, ErrInvalidBookingDate
	}
	if req.StartsAt.Before(time.Now()) {
		return pgstore.Booking{}, errors.New("cannot book in the past")
	}

	isResident, err := uc.querier.CheckIsResident(ctx, pgstore.CheckIsResidentParams{
		UserID:      req.UserID,
		ApartmentID: req.ApartmentID,
	})
	if err != nil {
		return pgstore.Booking{}, fmt.Errorf("error checking residency: %w", err)
	}
	if !isResident {
		return pgstore.Booking{}, ErrNoPermission
	}

	tx, err := uc.pool.Begin(ctx)
	if err != nil {
		return pgstore.Booking{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := pgstore.New(tx)

	area, err := qtx.GetCommonAreaIdForUpdate(ctx, req.CommonAreaID)
	if err != nil {
		return pgstore.Booking{}, fmt.Errorf("failed to lock common area: %w", err)
	}

	hasConflict, err := qtx.CheckBookingConflict(ctx, pgstore.CheckBookingConflictParams{
		CommonAreaID: req.CommonAreaID,
		StartsAt:     req.StartsAt,
		EndsAt:       req.EndsAt,
	})
	if err != nil {
		return pgstore.Booking{}, fmt.Errorf("failed to check conflicts: %w", err)
	}
	if hasConflict {
		return pgstore.Booking{}, ErrTimeSlotTaken
	}

	initialStatus := "confirmed"
	if area.RequiresApproval {
		initialStatus = "pending"
	}

	booking, err := qtx.CreateBooking(ctx, pgstore.CreateBookingParams{
		CondominiumID: req.CondominiumID,
		ApartmentID:   req.ApartmentID,
		UserID:        req.UserID,
		CommonAreaID:  req.CommonAreaID,
		StartsAt:      req.StartsAt,
		EndsAt:        req.EndsAt,
		Status:        initialStatus,
	})
	if err != nil {
		return pgstore.Booking{}, fmt.Errorf("failed to create booking: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return pgstore.Booking{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return booking, nil
}
