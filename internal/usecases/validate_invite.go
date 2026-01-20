package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Bellorico323/vizen/internal/services"
	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/Bellorico323/vizen/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ValidateInviteUC interface {
	Exec(ctx context.Context, req ValidateInviteReq) (pgstore.GetInviteByTokenRow, error)
}

type ValidateInviteReq struct {
	CondominiumID uuid.UUID
	UserID        uuid.UUID
	Token         uuid.UUID
}

type ValidateInviteUseCase struct {
	pool     *pgxpool.Pool
	notifier services.NotificationService
}

func NewValidateInviteUseCase(pool *pgxpool.Pool, n services.NotificationService) *ValidateInviteUseCase {
	return &ValidateInviteUseCase{
		pool:     pool,
		notifier: n,
	}
}

var (
	ErrInviteNotFound       = errors.New("Invite not found")
	ErrInviteAlreadyRevoked = errors.New("Invite already revoked")
	ErrInviteNotStarted     = errors.New("Invite start time has not been reached yet")
	ErrInviteExpired        = errors.New("Invite has expired")
)

func (uc *ValidateInviteUseCase) Exec(ctx context.Context, req ValidateInviteReq) (pgstore.GetInviteByTokenRow, error) {
	tx, err := uc.pool.Begin(ctx)
	if err != nil {
		return pgstore.GetInviteByTokenRow{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := pgstore.New(tx)

	_, err = qtx.GetCondominiumMemberRole(ctx, pgstore.GetCondominiumMemberRoleParams{
		CondominiumID: req.CondominiumID,
		UserID:        req.UserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgstore.GetInviteByTokenRow{}, ErrNoPermission
		}
		return pgstore.GetInviteByTokenRow{}, err
	}

	invite, err := qtx.GetInviteByToken(ctx, req.Token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgstore.GetInviteByTokenRow{}, ErrInviteNotFound
		}
		return pgstore.GetInviteByTokenRow{}, fmt.Errorf("Error while fetching invite: %w", err)
	}

	if req.CondominiumID != invite.CondominiumID {
		return pgstore.GetInviteByTokenRow{}, ErrNoPermission
	}

	if invite.RevokedAt != nil {
		return pgstore.GetInviteByTokenRow{}, ErrInviteAlreadyRevoked
	}

	now := time.Now()
	if now.Before(invite.StartsAt) {
		return pgstore.GetInviteByTokenRow{}, ErrInviteNotStarted
	}

	if now.After(invite.EndsAt) {
		return pgstore.GetInviteByTokenRow{}, ErrInviteExpired
	}

	_, err = qtx.LogAccessEntry(ctx, pgstore.LogAccessEntryParams{
		InviteID:      invite.ID,
		CondominiumID: invite.CondominiumID,
		AuthorizedBy:  utils.ToPtr(req.UserID),
	})
	if err != nil {
		return pgstore.GetInviteByTokenRow{}, fmt.Errorf("failed to create access entry: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return pgstore.GetInviteByTokenRow{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	go func() {
		bgCtx := context.Background()

		var title, body string
		title = "Visitante chegou!"
		body = fmt.Sprintf("%s acabou de passar pela portaria.", invite.GuestName)

		_ = uc.notifier.SendToUser(bgCtx, invite.IssuedBy, title, body)

	}()

	return invite, nil
}
