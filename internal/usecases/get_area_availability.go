package usecases

import (
	"context"
	"time"

	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/google/uuid"
)

type GetAreaAvailabilityUC interface {
	Exec(ctx context.Context, req GetAreaAvailabilityReq) ([]pgstore.GetAreaAvailabilityRow, error)
}

type GetAreaAvailabilityReq struct {
	CommonAreaID uuid.UUID
	FromDate     time.Time
	ToDate       time.Time
}

type GetAreaAvailabilityUseCase struct {
	querier pgstore.Querier
}

func NewGetAreaAvailabilityUseCase(q pgstore.Querier) *GetAreaAvailabilityUseCase {
	return &GetAreaAvailabilityUseCase{querier: q}
}

func (uc *GetAreaAvailabilityUseCase) Exec(ctx context.Context, req GetAreaAvailabilityReq) ([]pgstore.GetAreaAvailabilityRow, error) {
	slots, err := uc.querier.GetAreaAvailability(ctx, pgstore.GetAreaAvailabilityParams{
		CommonAreaID: req.CommonAreaID,
		StartsAt:     req.FromDate,
		EndsAt:       req.ToDate,
	})
	if err != nil {
		return nil, err
	}
	return slots, nil
}
