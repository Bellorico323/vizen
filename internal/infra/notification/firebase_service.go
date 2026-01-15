package notification

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"firebase.google.com/go/v4/messaging"
	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/google/uuid"
)

type FirebaseService struct {
	client  *messaging.Client
	querier pgstore.Querier
}

var (
	ErrSendNotification = errors.New("Failed to send notification")
)

func NewFireBaseService(client *messaging.Client, q pgstore.Querier) *FirebaseService {
	return &FirebaseService{client: client, querier: q}
}

func (s *FirebaseService) SendToUser(ctx context.Context, userID uuid.UUID, title, body string) error {
	tokens, err := s.querier.GetUserDeviceTokens(ctx, userID)
	if err != nil {
		return err
	}

	if len(tokens) == 0 {
		return nil
	}

	msg := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: map[string]string{
			"click_action": "ACCESS_REQUEST_UPDATE",
		},
	}

	_, err = s.client.SendEachForMulticast(ctx, msg)
	if err != nil {
		slog.Error("Failed to send FCM message", "error", err)
		return err
	}

	return nil
}

func (s *FirebaseService) SendToCondoAdmins(ctx context.Context, condoID uuid.UUID, title, body string) error {
	tokens, err := s.querier.GetCondoAdminTokens(ctx, condoID)
	if err != nil {
		return err
	}

	if len(tokens) == 0 {
		return nil
	}

	msg := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
	}

	_, err = s.client.SendEachForMulticast(ctx, msg)
	return err
}

func (s *FirebaseService) SendToCondoResidents(ctx context.Context, condoID uuid.UUID, title, body string) error {
	tokens, err := s.querier.GetCondoResidentsTokens(ctx, condoID)
	if err != nil {
		slog.Error("Failed to fetch residents tokens", "condo_id", condoID, "error", err)
		return fmt.Errorf("Error to fetch residents tokens: %w", err)
	}

	if len(tokens) == 0 {
		return nil
	}

	return s.sendChunks(ctx, tokens, title, body, nil)
}

func (s *FirebaseService) SendToApartmentResidents(ctx context.Context, apartmentID, packageID uuid.UUID, title, body string) error {
	tokens, err := s.querier.GetManyTokensByApartmentId(ctx, apartmentID)
	if err != nil {
		slog.Error("Failed to fetch resident tokens for notification", "error", err)
		return nil
	}

	if len(tokens) == 0 {
		return nil
	}

	return s.sendChunks(ctx, tokens, title, body, map[string]string{
		"type":      "PACKAGE_ARRIVED",
		"packageId": packageID.String(),
	})
}

func (s *FirebaseService) sendChunks(ctx context.Context, tokens []string, title, body string, data map[string]string) error {
	if len(tokens) == 0 {
		return nil
	}

	const batchSize = 500

	for i := 0; i < len(tokens); i += batchSize {
		end := min(i+batchSize, len(tokens))

		batchTokens := tokens[i:end]

		msg := &messaging.MulticastMessage{
			Tokens: batchTokens,
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
			Data: data,
		}

		br, err := s.client.SendEachForMulticast(ctx, msg)
		if err != nil {
			slog.Error("failed to send batch FCM message", "error", err)

			return err
		}

		if br.FailureCount > 0 {
			slog.Warn("Some notifications failed", "success", br.SuccessCount, "failure", br.FailureCount)
		}
	}
	return nil
}
