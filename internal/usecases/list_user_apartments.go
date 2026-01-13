package usecases

import (
	"context"
	"errors"

	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ListUserApartmentsUC interface {
	Exec(ctx context.Context, req ListUserApartmentsReq) ([]pgstore.GetApartmentsByUserIdRow, error)
}

type ListUserApartmentsReq struct {
	UserID        uuid.UUID
	CondominiumID *uuid.UUID
}

type ListUserApartmentsUseCase struct {
	querier pgstore.Querier
}

func NewListUserApartmentsUseCase(q pgstore.Querier) *ListUserApartmentsUseCase {
	return &ListUserApartmentsUseCase{
		querier: q,
	}
}

func (uc *ListUserApartmentsUseCase) Exec(ctx context.Context, req ListUserApartmentsReq) ([]pgstore.GetApartmentsByUserIdRow, error) {
	user, err := uc.querier.GetUserByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	apartments, err := uc.querier.GetApartmentsByUserId(ctx, pgstore.GetApartmentsByUserIdParams{
		CondominiumID: req.CondominiumID,
		UserID:        user.ID,
	})
	if err != nil {
		return []pgstore.GetApartmentsByUserIdRow{}, err
	}

	return apartments, nil
}
