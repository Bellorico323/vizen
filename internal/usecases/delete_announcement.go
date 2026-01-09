package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type DeleteAnnouncementUC interface {
	Exec(ctx context.Context, req DeleteAnnouncementReq) error
}

type DeleteAnnouncementReq struct {
	AnnouncementID uuid.UUID
	UserID         uuid.UUID
}

type DeleteAnnouncementUseCase struct {
	querier pgstore.Querier
}

func NewDeleteAnnouncementUseCase(q pgstore.Querier) *DeleteAnnouncementUseCase {
	return &DeleteAnnouncementUseCase{
		querier: q,
	}
}

var ErrAnnouncementNotFound = errors.New("announcement not found")

func (uc *DeleteAnnouncementUseCase) Exec(ctx context.Context, req DeleteAnnouncementReq) error {
	announcement, err := uc.querier.GetAnnouncementById(ctx, req.AnnouncementID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrAccessRequestNotFound
		}
		return fmt.Errorf("failed to fetch announcement: %w", err)
	}

	role, err := uc.querier.GetCondominiumMemberRole(ctx, pgstore.GetCondominiumMemberRoleParams{
		CondominiumID: announcement.CondominiumID,
		UserID:        req.UserID,
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

	err = uc.querier.DeleteAnnouncement(ctx, pgstore.DeleteAnnouncementParams{
		CondominiumID: announcement.CondominiumID,
		ID:            announcement.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete announcement: %w", err)
	}

	return nil
}
