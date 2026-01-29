package usecases

import (
	"context"
	"fmt"

	"github.com/Bellorico323/vizen/internal/store/pgstore"
)

type LogoutUC interface {
	Exec(ctx context.Context, req LogoutReq) error
}

type LogoutUseCase struct {
	querier pgstore.Querier
}

func NewLogoutUseCase(querier pgstore.Querier) *LogoutUseCase {
	return &LogoutUseCase{
		querier: querier,
	}
}

type LogoutReq struct {
	RefreshToken string
}

func (uc *LogoutUseCase) Exec(ctx context.Context, req LogoutReq) error {
	err := uc.querier.DeleteSession(ctx, req.RefreshToken)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}
