package services

import (
	"context"

	"github.com/google/uuid"
)

type NotificationService interface {
	SendToUser(ctx context.Context, userID uuid.UUID, title, body string) error
	SendToCondoAdmins(ctx context.Context, condoID uuid.UUID, title, body string) error
	SendToCondoResidents(ctx context.Context, condoID uuid.UUID, title, body string) error
}
