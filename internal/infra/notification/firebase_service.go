package notification

import (
	"context"
	"log/slog"

	"firebase.google.com/go/v4/messaging"
	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/google/uuid"
)

type FirebaseService struct {
	client  *messaging.Client
	querier pgstore.Querier
}

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
