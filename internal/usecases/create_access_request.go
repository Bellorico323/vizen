package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CreateAccessRequestUC interface {
	Exec(ctx context.Context, req CreateAccessRequestReq) (uuid.UUID, error)
}

type CreateAccessRequestReq struct {
	UserID      uuid.UUID
	ApartmentID uuid.UUID
	Type        string
}

type CreateAccessRequestUseCase struct {
	querier pgstore.Querier
}

func NewCreateAccessRequestUseCase(q pgstore.Querier) CreateAccessRequestUC {
	return &CreateAccessRequestUseCase{
		querier: q,
	}
}

var (
	ErrApartmentNotFound   = errors.New("Apartment not found")
	ErrInvalidResidentType = errors.New("Invalid resident type")
)

func (uc *CreateAccessRequestUseCase) Exec(ctx context.Context, req CreateAccessRequestReq) (uuid.UUID, error) {
	user, err := uc.querier.GetUserByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrUserNotFound
		}

		return uuid.Nil, fmt.Errorf("Error while searching for user: %w", err)
	}

	apartment, err := uc.querier.GetApartmentById(ctx, req.ApartmentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrApartmentNotFound
		}

		return uuid.Nil, fmt.Errorf("Error while searching for apartment: %w", err)
	}

	ok := validateResidentType(req.Type)
	if !ok {
		return uuid.Nil, ErrInvalidResidentType
	}

	params := pgstore.CreateAccessRequestParams{
		UserID:        user.ID,
		ApartmentID:   apartment.ID,
		CondominiumID: apartment.CondominiumID,
		Type:          req.Type,
	}

	id, err := uc.querier.CreateAccessRequest(ctx, params)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func validateResidentType(inp string) bool {
	switch inp {
	case "owner", "tenant", "dependent":
		return true
	default:
		return false
	}
}
