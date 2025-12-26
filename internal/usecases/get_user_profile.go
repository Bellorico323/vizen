package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type GetUserProfileUC interface {
	Exec(ctx context.Context, userID uuid.UUID) (*GetUserProfileRes, error)
}

type GetUserProfileUseCase struct {
	querier pgstore.Querier
}

type GetUserProfileRes struct {
	User        *pgstore.User
	Residences  *[]pgstore.GetResidencesByUserIdRow
	Memberships *[]pgstore.GetUserMembershipsRow
}

func NewGetUserProfileUseCase(querier pgstore.Querier) *GetUserProfileUseCase {
	return &GetUserProfileUseCase{
		querier: querier,
	}
}

func (uc *GetUserProfileUseCase) Exec(ctx context.Context, userID uuid.UUID) (*GetUserProfileRes, error) {
	user, err := uc.querier.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("User not found: %w", err)
		}
		return nil, fmt.Errorf("Unexpected error occured: %w", err)
	}

	residences, err := uc.querier.GetResidencesByUserId(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch residences: %w", err)
	}

	memberships, err := uc.querier.GetUserMemberships(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch memberships: %w", err)
	}

	return &GetUserProfileRes{
		User:        &user,
		Residences:  &residences,
		Memberships: &memberships,
	}, nil
}
