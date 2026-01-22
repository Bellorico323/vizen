package usecases

import (
	"context"
	"fmt"

	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/Bellorico323/vizen/internal/utils"
	"github.com/google/uuid"
)

type ListCommonAreasUC interface {
	Exec(ctx context.Context, req ListCommonAreasReq) ([]pgstore.CommonArea, error)
}

type ListCommonAreasReq struct {
	UserID        uuid.UUID
	CondominiumID uuid.UUID
}

type ListCommonAreasUseCase struct {
	querier pgstore.Querier
}

func NewListCommonAreasUseCase(q pgstore.Querier) *ListCommonAreasUseCase {
	return &ListCommonAreasUseCase{
		querier: q,
	}
}

func (uc *ListCommonAreasUseCase) Exec(ctx context.Context, req ListCommonAreasReq) ([]pgstore.CommonArea, error) {
	_, err := uc.querier.GetCondominiumMemberRole(ctx, pgstore.GetCondominiumMemberRoleParams{
		CondominiumID: req.CondominiumID,
		UserID:        req.UserID,
	})

	isMember := err == nil

	if !isMember {
		userApts, err := uc.querier.GetApartmentsByUserId(ctx, pgstore.GetApartmentsByUserIdParams{
			UserID:        req.UserID,
			CondominiumID: utils.ToPtr(req.CondominiumID),
		})
		if err != nil {
			return nil, err
		}

		if len(userApts) == 0 {
			return nil, ErrNoPermission
		}
	}

	areas, err := uc.querier.ListCommonAreas(ctx, req.CondominiumID)
	if err != nil {
		return nil, fmt.Errorf("failed to list common areas: %w", err)
	}

	return areas, nil
}
