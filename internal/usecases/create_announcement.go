package usecases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Bellorico323/vizen/internal/services"
	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/Bellorico323/vizen/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CreateAnnouncementUC interface {
	Exec(ctx context.Context, req *CreateAnnouncementReq) error
}

type CreateAnnouncementReq struct {
	AuthorID      uuid.UUID
	CondominiumID uuid.UUID
	Title         string
	Content       string
}

type CreateAnnouncementUseCase struct {
	querier  pgstore.Querier
	notifier services.NotificationService
}

func NewCreateAnnouncementUseCase(q pgstore.Querier, n services.NotificationService) *CreateAnnouncementUseCase {
	return &CreateAnnouncementUseCase{
		querier:  q,
		notifier: n,
	}
}

var (
	ErrEmptyTitle   = errors.New("title cannot be empty")
	ErrEmptyContent = errors.New("content cannot be empty")
)

func (uc *CreateAnnouncementUseCase) Exec(ctx context.Context, req *CreateAnnouncementReq) error {
	if req.Title == "" {
		return ErrEmptyTitle
	}
	if req.Content == "" {
		return ErrEmptyContent
	}

	role, err := uc.querier.GetCondominiumMemberRole(ctx, pgstore.GetCondominiumMemberRoleParams{
		CondominiumID: req.CondominiumID,
		UserID:        req.AuthorID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoPermission
		}
		return fmt.Errorf("failed to check permission: %w", err)
	}

	if role != "admin" && role != "syndic" {
		return ErrNoPermission
	}

	_, err = uc.querier.CreateAnnouncement(ctx, pgstore.CreateAnnouncementParams{
		CondominiumID: req.CondominiumID,
		AuthorID:      utils.ToPtr(req.AuthorID),
		Title:         req.Title,
		Content:       req.Content,
	})
	if err != nil {
		return fmt.Errorf("failed to create announcement: %w", err)
	}

	go func() {
		bgCtx := context.Background()

		err := uc.notifier.SendToCondoResidents(
			bgCtx,
			req.CondominiumID,
			fmt.Sprintf("Novo aviso: %s", req.Title),
			req.Content,
		)
		if err != nil {
			slog.Error("Failed to notify residents about new announcement", "error", err)
		}
	}()

	return nil
}
