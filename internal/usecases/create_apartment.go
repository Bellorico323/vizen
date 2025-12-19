package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/Bellorico323/vizen/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CreateApartmentUC interface {
	Exec(ctx context.Context, req CreateApartmentReq) (uuid.UUID, error)
}

type CreateApartmentReq struct {
	CondominiumID uuid.UUID
	Block         string
	Number        string
	UserID        uuid.UUID
}

type CreateApartmentUseCase struct {
	querier pgstore.Querier
}

func NewCreateApartmentUseCase(q pgstore.Querier) *CreateApartmentUseCase {
	return &CreateApartmentUseCase{
		querier: q,
	}
}

var (
	ErrCondominiumNotFound       = errors.New("condominium not found")
	ErrApartmentNumberIsRequired = errors.New("apartment number is required")
)

func (uc *CreateApartmentUseCase) Exec(ctx context.Context, req CreateApartmentReq) (uuid.UUID, error) {
	user, err := uc.querier.GetUserByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, fmt.Errorf("user not found: %w", err)
		}

		return uuid.Nil, err
	}

	if user.Role != "admin" {
		return uuid.Nil, ErrNoPermission
	}

	condominium, err := uc.querier.GetCondominiumById(ctx, req.CondominiumID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrCondominiumNotFound
		}

		return uuid.Nil, err
	}

	if req.Number == "" {
		return uuid.Nil, ErrApartmentNumberIsRequired
	}

	args := pgstore.CreateApartmentParams{
		CondominiumID: condominium.ID,
		Block:         utils.ToNullString(req.Block),
		Number:        req.Number,
	}

	id, err := uc.querier.CreateApartment(ctx, args)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}
