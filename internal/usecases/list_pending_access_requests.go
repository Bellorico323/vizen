package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ListPendingAccessRequestUC interface {
	Exec(ctx context.Context, req ListPendingAccessRequestReq) ([]pgstore.ListPendingRequestsByCondoRow, error)
}

type ListPendingAccessRequestReq struct {
	UserID        uuid.UUID
	CondominiumID uuid.UUID
}

type ListPendingAccessRequestsUseCase struct {
	querier pgstore.Querier
}

func NewListPendingAccessRequestsUseCase(q pgstore.Querier) *ListPendingAccessRequestsUseCase {
	return &ListPendingAccessRequestsUseCase{
		querier: q,
	}
}
func (uc *ListPendingAccessRequestsUseCase) Exec(ctx context.Context, req ListPendingAccessRequestReq) ([]pgstore.ListPendingRequestsByCondoRow, error) {
	role, err := uc.querier.GetCondominiumMemberRole(ctx, pgstore.GetCondominiumMemberRoleParams{
		CondominiumID: req.CondominiumID,
		UserID:        req.UserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []pgstore.ListPendingRequestsByCondoRow{}, ErrNoPermission
		}
		return []pgstore.ListPendingRequestsByCondoRow{}, err
	}

	if role != "admin" && role != "syndic" {
		return []pgstore.ListPendingRequestsByCondoRow{}, ErrNoPermission
	}

	pendingRequests, err := uc.querier.ListPendingRequestsByCondo(ctx, req.CondominiumID)
	if err != nil {
		return []pgstore.ListPendingRequestsByCondoRow{}, fmt.Errorf("failed to fetch pending access requests: %w", err)
	}

	return pendingRequests, nil
}
