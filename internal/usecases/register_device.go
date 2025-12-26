package usecases

import (
	"context"
	"fmt"

	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/Bellorico323/vizen/internal/utils"
	"github.com/google/uuid"
)

type RegisterDeviceUC interface {
	Exec(ctx context.Context, userID uuid.UUID, fcmToken, platform string) error
}

type RegisterDeviceUseCase struct {
	querier pgstore.Querier
}

func NewRegisterDeviceUseCase(q pgstore.Querier) *RegisterDeviceUseCase {
	return &RegisterDeviceUseCase{
		querier: q,
	}
}

func (uc *RegisterDeviceUseCase) Exec(ctx context.Context, userID uuid.UUID, fcmToken, platform string) error {
	if fcmToken == "" {
		return fmt.Errorf("token cannot be empty")
	}

	return uc.querier.SaveUserDevice(ctx, pgstore.SaveUserDeviceParams{
		UserID:   userID,
		FcmToken: fcmToken,
		Platform: utils.ToNullString(platform),
	})
}
