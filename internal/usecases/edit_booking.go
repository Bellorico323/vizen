package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bellorico323/vizen/internal/services"
	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type EditBookingtUC interface {
	Exec(ctx context.Context, req EditBookingReq) error
}

type EditBookingReq struct {
	BookingID uuid.UUID
	UserID    uuid.UUID
	Status    string
}

type EditBookingUseCase struct {
	querier  pgstore.Querier
	notifier services.NotificationService
}

func NewEditBookingUseCase(q pgstore.Querier, n services.NotificationService) *EditBookingUseCase {
	return &EditBookingUseCase{
		querier:  q,
		notifier: n,
	}
}

var (
	ErrBookingNotFound      = errors.New("Booking not found")
	ErrBookingNotPending    = errors.New("Booking is not pending")
	ErrInvalidBookingStatus = errors.New("Invalid booking status")
)

func (uc *EditBookingUseCase) Exec(ctx context.Context, req EditBookingReq) error {
	booking, err := uc.querier.GetBookingById(ctx, req.BookingID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrBookingNotFound
		}

		return err
	}

	validStatus := checkBookingStatusExists(req.Status)
	if !validStatus {
		return ErrInvalidBookingStatus
	}

	if booking.Status != "pending" {
		return ErrBookingNotPending
	}

	role, err := uc.querier.GetCondominiumMemberRole(ctx, pgstore.GetCondominiumMemberRoleParams{
		CondominiumID: booking.CondominiumID,
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

	params := pgstore.UpdateBookingStatusParams{
		Status: req.Status,
		ID:     booking.ID,
	}

	_, err = uc.querier.UpdateBookingStatus(ctx, params)
	if err != nil {
		return err
	}

	go func() {
		bgCtx := context.Background()

		var title, body string

		switch req.Status {
		case "confirmed":
			title = "✅ Agendamento Aprovado"
			body = fmt.Sprintf("Seu agendamento para '%s' foi confirmado!", booking.CommonAreaName)
		case "denied":
			title = "❌ Agendamento Recusado"
			body = fmt.Sprintf("Seu agendamento para '%s' foi recusado.", booking.CommonAreaName)

		default:
			title = "Status de agendamento alterado"
			body = fmt.Sprintf("Seu agendamento para '%s' teve o status alterado.", booking.CommonAreaName)
		}

		_ = uc.notifier.SendToUser(bgCtx, booking.UserID, title, body)
	}()

	return nil
}

func checkBookingStatusExists(status string) bool {
	switch status {
	case "pending", "confirmed", "cancelled", "denied":
		return true
	default:
		return false
	}
}
