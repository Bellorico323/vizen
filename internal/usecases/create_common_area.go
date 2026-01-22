package usecases

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CreateCommonAreaUC interface {
	Exec(ctx context.Context, req CreateCommonAreaReq) (pgstore.CommonArea, error)
}

type CreateCommonAreaReq struct {
	CondominiumID    uuid.UUID
	UserID           uuid.UUID
	Name             string
	Capacity         *int32
	RequiresApproval bool
}

type CreateCommonAreaUseCase struct {
	querier pgstore.Querier
}

func NewCreateCommonAreaUseCase(q pgstore.Querier) *CreateCommonAreaUseCase {
	return &CreateCommonAreaUseCase{
		querier: q,
	}
}

var ErrInvalidAreaName = errors.New("Invalid common area name")

func (uc *CreateCommonAreaUseCase) Exec(ctx context.Context, req CreateCommonAreaReq) (pgstore.CommonArea, error) {
	userRole, err := uc.querier.GetCondominiumMemberRole(ctx, pgstore.GetCondominiumMemberRoleParams{
		CondominiumID: req.CondominiumID,
		UserID:        req.UserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgstore.CommonArea{}, ErrNoPermission
		}

		return pgstore.CommonArea{}, err
	}

	if userRole != "admin" && userRole != "syndic" {
		return pgstore.CommonArea{}, ErrNoPermission
	}

	if strings.TrimSpace(req.Name) == "" {
		return pgstore.CommonArea{}, ErrInvalidAreaName
	}

	area, err := uc.querier.CreateCommonArea(ctx, pgstore.CreateCommonAreaParams{
		CondominiumID:    req.CondominiumID,
		Name:             req.Name,
		Capacity:         req.Capacity,
		RequiresApproval: req.RequiresApproval,
	})
	if err != nil {
		return pgstore.CommonArea{}, fmt.Errorf("failed to create common area: %w", err)
	}

	return area, nil
}
