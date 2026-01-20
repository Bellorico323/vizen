package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type RevokeInviteUC interface {
	Exec(ctx context.Context, req RevokeInviteReq) error
}

type RevokeInviteReq struct {
	InviteID uuid.UUID
	UserID   uuid.UUID
}

type RevokeInviteUseCase struct {
	querier pgstore.Querier
}

func NewRevokeInviteUseCase(q pgstore.Querier) *RevokeInviteUseCase {
	return &RevokeInviteUseCase{
		querier: q,
	}
}

func (uc *RevokeInviteUseCase) Exec(ctx context.Context, req RevokeInviteReq) error {
	invite, err := uc.querier.GetInviteById(ctx, req.InviteID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInviteNotFound
		}

		return fmt.Errorf("failed to fetch invite: %w", err)
	}

	if invite.IssuedBy != req.UserID {
		return ErrNoPermission
	}

	if invite.RevokedAt != nil {
		return ErrInviteAlreadyRevoked
	}

	err = uc.querier.RevokeInvite(ctx, pgstore.RevokeInviteParams{
		ID:       req.InviteID,
		IssuedBy: req.UserID,
	})
	if err != nil {
		return fmt.Errorf("failed to revoke invite: %w", err)
	}

	return nil
}
