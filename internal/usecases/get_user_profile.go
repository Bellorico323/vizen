package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserProfileGetter interface {
	Exec(ctx context.Context, userID uuid.UUID) (*pgstore.User, error)
}

type GetUserProfile struct {
	querier pgstore.Querier
}

func NewGetUserProfile(querier pgstore.Querier) *GetUserProfile {
	return &GetUserProfile{
		querier: querier,
	}
}

func (uc *GetUserProfile) Exec(ctx context.Context, userID uuid.UUID) (*pgstore.User, error) {
	user, err := uc.querier.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &pgstore.User{}, fmt.Errorf("User not found: %w", err)
		}
		return &pgstore.User{}, fmt.Errorf("Unexpected error occured: %w", err)
	}

	return &user, nil
}
