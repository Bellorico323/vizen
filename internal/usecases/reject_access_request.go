package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bellorico323/vizen/internal/services"
	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/Bellorico323/vizen/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RejectAccessRequestUC interface {
	Exec(ctx context.Context, req RejectAccessRequestReq) error
}

type RejectAccessRequestReq struct {
	ReviewerID      uuid.UUID
	AccessRequestID uuid.UUID
}

type RejectAccessRequestUseCase struct {
	pool     *pgxpool.Pool
	notifier services.NotificationService
}

func NewRejectAccessRequestUseCase(pool *pgxpool.Pool, n services.NotificationService) RejectAccessRequestUC {
	return &RejectAccessRequestUseCase{
		pool:     pool,
		notifier: n,
	}
}

func (uc *RejectAccessRequestUseCase) Exec(ctx context.Context, req RejectAccessRequestReq) error {
	tx, err := uc.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin trasaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := pgstore.New(tx)

	accessRequest, err := qtx.GetAccessRequestById(ctx, req.AccessRequestID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrAccessRequestNotFound
		}

		return err
	}

	if accessRequest.Status != "pending" {
		return ErrRequestNotPending
	}

	role, err := qtx.GetCondominiumMemberRole(ctx, pgstore.GetCondominiumMemberRoleParams{
		CondominiumID: accessRequest.CondominiumID,
		UserID:        req.ReviewerID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoPermission
		}
		return err
	}

	if role != "admin" && role != "syndic" {
		return ErrNoPermission
	}

	params := pgstore.UpdateAccessRequestStatusParams{
		ID:         accessRequest.ID,
		ReviewedBy: utils.ToPtr(req.ReviewerID),
		Status:     "rejected",
	}

	err = qtx.UpdateAccessRequestStatus(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update request status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	go func() {
		bgCtx := context.Background()

		var title, body string
		title = "Acesso Negado"
		body = "A sua solicitação para fazer parte do condomínio foi rejeitada."

		_ = uc.notifier.SendToUser(bgCtx, accessRequest.UserID, title, body)

	}()

	return nil
}
