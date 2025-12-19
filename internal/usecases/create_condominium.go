package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CreateCondominiumUC interface {
	Exec(ctx context.Context, req CreateCondominiumReq) (uuid.UUID, error)
}

type CreateCondominiumReq struct {
	Name     string
	Cnpj     string
	Address  string
	PlanType string
	UserID   uuid.UUID
}

type CreateCondominiumUseCase struct {
	querier pgstore.Querier
}

func NewCreateCondominiumUseCase(q pgstore.Querier) *CreateCondominiumUseCase {
	return &CreateCondominiumUseCase{
		querier: q,
	}
}

var ErrNoPermission = errors.New("user does not have permission")

func (uc *CreateCondominiumUseCase) Exec(ctx context.Context, req CreateCondominiumReq) (uuid.UUID, error) {
	user, err := uc.querier.GetUserByID(ctx, req.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, fmt.Errorf("user not found: %w", err)
		}

		return uuid.Nil, err
	}

	if user.Role != "admin" {
		return uuid.Nil, ErrNoPermission
	}

	if !isValidPlanType(req.PlanType) {
		return uuid.Nil, fmt.Errorf("invalid plan type")
	}

	id, err := uc.querier.CreateCondominium(ctx, pgstore.CreateCondominiumParams{
		Name:     req.Name,
		Cnpj:     req.Cnpj,
		Address:  req.Address,
		PlanType: req.PlanType,
	})
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func isValidPlanType(plan string) bool {
	switch plan {
	case "basic", "pro", "enterprise":
		return true
	default:
		return false
	}
}
