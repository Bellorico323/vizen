package usecases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Bellorico323/vizen/internal/services"
	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CreatePackageUC interface {
	Exec(ctx context.Context, req CreatePackageReq) (uuid.UUID, error)
}

type CreatePackageReq struct {
	CondominiumID uuid.UUID
	ApartmentID   uuid.UUID
	ReceivedBy    uuid.UUID
	RecipientName *string
	PhotoUrl      *string
}

type CreatePackageUseCase struct {
	querier  pgstore.Querier
	notifier services.NotificationService
}

func NewCreatePackageUseCase(q pgstore.Querier, n services.NotificationService) *CreatePackageUseCase {
	return &CreatePackageUseCase{
		querier:  q,
		notifier: n,
	}
}

var ErrApartmentIsNotFromCondominium = errors.New("Apartment is not from condominium")

func (uc *CreatePackageUseCase) Exec(ctx context.Context, req CreatePackageReq) (uuid.UUID, error) {
	_, err := uc.querier.GetCondominiumMemberRole(ctx, pgstore.GetCondominiumMemberRoleParams{
		CondominiumID: req.CondominiumID,
		UserID:        req.ReceivedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrNoPermission
		}
		return uuid.Nil, err
	}

	apartment, err := uc.querier.GetApartmentById(ctx, req.ApartmentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrApartmentNotFound
		}
		return uuid.Nil, fmt.Errorf("failed to find apartment: %w", err)
	}

	if apartment.CondominiumID != req.CondominiumID {
		return uuid.Nil, ErrApartmentIsNotFromCondominium
	}

	packg, err := uc.querier.CreatePackage(ctx, pgstore.CreatePackageParams{
		CondominiumID: req.CondominiumID,
		ApartmentID:   apartment.ID,
		ReceivedBy:    req.ReceivedBy,
		RecipientName: req.RecipientName,
		PhotoUrl:      req.PhotoUrl,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create package: %w", err)
	}

	title := "📦 Chegou Encomenda!"
	body := fmt.Sprintf("Uma nova encomenda para %s chegou na portaria.", apartment.Number)
	if req.RecipientName != nil {
		body = fmt.Sprintf("Encomenda para %s chegou na portaria.", *req.RecipientName)
	}

	go func() {
		bgCtx := context.Background()

		err = uc.notifier.SendToApartmentResidents(bgCtx, apartment.ID, packg.ID, title, body)
		if err != nil {
			slog.Error("Failed to send async notification", "package_id", packg.ID, "error", err)
		}
	}()

	return packg.ID, nil
}
