package usecases

import (
	"context"

	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/google/uuid"
)

type ListAnnouncementsUC interface {
	Exec(ctx context.Context, req ListAnnouncementsReq) ([]pgstore.GetManyAnnouncementsByCondoIdRow, error)
}

type ListAnnouncementsReq struct {
	CondominiumID uuid.UUID
	UserID        uuid.UUID
	Page          int32
	Limit         int32
}

type ListAnnouncementsUseCase struct {
	querier pgstore.Querier
}

func NewListAnnouncementsUseCase(q pgstore.Querier) *ListAnnouncementsUseCase {
	return &ListAnnouncementsUseCase{
		querier: q,
	}
}

func (uc *ListAnnouncementsUseCase) Exec(ctx context.Context, req ListAnnouncementsReq) ([]pgstore.GetManyAnnouncementsByCondoIdRow, error) {
	hasAccess, err := uc.querier.CheckUserAccessToCondo(ctx, pgstore.CheckUserAccessToCondoParams{
		UserID:        req.UserID,
		CondominiumID: req.CondominiumID,
	})
	if err != nil {
		return nil, err
	}

	if !hasAccess {
		return nil, ErrNoPermission
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	offset := (req.Page - 1) * req.Limit

	announcements, err := uc.querier.GetManyAnnouncementsByCondoId(ctx, pgstore.GetManyAnnouncementsByCondoIdParams{
		CondominiumID: req.CondominiumID,
		Limit:         req.Limit,
		Offset:        offset,
	})
	if err != nil {
		return nil, err
	}

	return announcements, nil
}
