package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CreateInviteUC interface {
	Exec(ctx context.Context, req CreateInviteReq) (pgstore.Invite, error)
}

type CreateInviteReq struct {
	CondominiumID uuid.UUID
	ApartmentID   uuid.UUID
	IssuedBy      uuid.UUID
	GuestName     string
	GuestType     *string
	StartsAt      time.Time
	EndsAt        time.Time
}

type CreateInviteUseCase struct {
	querier pgstore.Querier
}

func NewCreateInviteUseCase(q pgstore.Querier) *CreateInviteUseCase {
	return &CreateInviteUseCase{
		querier: q,
	}
}

var (
	ErrGuestNameIsRequired    = errors.New("guest name is required")
	ErrInviteInvalidTimeRange = errors.New("start date is lower than end date")
	ErrInviteInThePast        = errors.New("invite expiration must be in the future")
)

func (uc *CreateInviteUseCase) Exec(ctx context.Context, req CreateInviteReq) (pgstore.Invite, error) {
	apt, err := uc.querier.GetApartmentById(ctx, req.ApartmentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgstore.Invite{}, ErrApartmentNotFound
		}
		return pgstore.Invite{}, fmt.Errorf("failed to fetch apartment: %w", err)
	}

	if apt.CondominiumID != req.CondominiumID {
		return pgstore.Invite{}, ErrApartmentIsNotFromCondominium
	}

	isResident, err := uc.querier.CheckIsResident(ctx, pgstore.CheckIsResidentParams{
		UserID:      req.IssuedBy,
		ApartmentID: req.ApartmentID,
	})
	if err != nil {
		return pgstore.Invite{}, fmt.Errorf("Error while checking if user is resident: %w", err)
	}

	if !isResident {
		return pgstore.Invite{}, ErrNoPermission
	}

	if req.GuestName == "" {
		return pgstore.Invite{}, ErrGuestNameIsRequired
	}

	if req.EndsAt.Before(req.StartsAt) {
		return pgstore.Invite{}, ErrInviteInvalidTimeRange
	}

	if req.EndsAt.Before(time.Now()) {
		return pgstore.Invite{}, ErrInviteInThePast
	}

	guestType := "guest"
	if req.GuestType != nil {
		guestType = *req.GuestType
	}

	invite, err := uc.querier.CreateInvite(ctx, pgstore.CreateInviteParams{
		CondominiumID: req.CondominiumID,
		ApartmentID:   req.ApartmentID,
		IssuedBy:      req.IssuedBy,
		GuestName:     req.GuestName,
		GuestType:     guestType,
		StartsAt:      req.StartsAt,
		EndsAt:        req.EndsAt,
	})
	if err != nil {
		return pgstore.Invite{}, err
	}

	return invite, nil
}
