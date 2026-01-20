package usecases

import (
	"context"
	"errors"

	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/Bellorico323/vizen/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ListInvitesUC interface {
	Exec(ctx context.Context, req ListInvitesReq) ([]pgstore.ListInvitesRow, error)
}

type ListInvitesReq struct {
	CondominiumID uuid.UUID
	UserID        uuid.UUID
	ApartmentID   *uuid.UUID
	OnlyActive    bool
	Limit         int32
	Offset        int32
}

type ListInvitesUseCase struct {
	querier pgstore.Querier
}

func NewListInvitesUseCase(q pgstore.Querier) *ListInvitesUseCase {
	return &ListInvitesUseCase{querier: q}
}

func (uc *ListInvitesUseCase) Exec(ctx context.Context, req ListInvitesReq) ([]pgstore.ListInvitesRow, error) {
	_, err := uc.querier.GetCondominiumMemberRole(ctx, pgstore.GetCondominiumMemberRoleParams{
		CondominiumID: req.CondominiumID,
		UserID:        req.UserID,
	})

	isStaff := false
	if err == nil {
		isStaff = true
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	if !isStaff {
		if req.ApartmentID == nil {
			return nil, ErrNoPermission
		}

		isResident, err := uc.querier.CheckIsResident(ctx, pgstore.CheckIsResidentParams{
			UserID:      req.UserID,
			ApartmentID: *req.ApartmentID,
		})
		if err != nil {
			return nil, err
		}
		if !isResident {
			return nil, ErrNoPermission
		}
	}

	invites, err := uc.querier.ListInvites(ctx, pgstore.ListInvitesParams{
		CondominiumID: req.CondominiumID,
		ApartmentID:   req.ApartmentID,
		OnlyActive:    utils.ToPtr(req.OnlyActive),
		Limit:         req.Limit,
		Offset:        req.Offset,
	})
	if err != nil {
		return nil, err
	}

	return invites, err
}
