package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
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
	pool *pgxpool.Pool
}

func NewCreateCondominiumUseCase(p *pgxpool.Pool) *CreateCondominiumUseCase {
	return &CreateCondominiumUseCase{
		pool: p,
	}
}

var (
	ErrUserNotFound = errors.New("user not found")
	ErrNoPermission = errors.New("user does not have permission")
)

func (uc *CreateCondominiumUseCase) Exec(ctx context.Context, req CreateCondominiumReq) (uuid.UUID, error) {
	if !isValidPlanType(req.PlanType) {
		return uuid.Nil, fmt.Errorf("invalid plan type")
	}

	tx, err := uc.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer tx.Rollback(ctx)

	qtx := pgstore.New(tx)

	_, err = qtx.GetUserByID(ctx, req.UserID)
	if err != nil {
		return uuid.Nil, ErrUserNotFound
	}

	condoID, err := qtx.CreateCondominium(ctx, pgstore.CreateCondominiumParams{
		Name:     req.Name,
		Cnpj:     req.Cnpj,
		Address:  req.Address,
		PlanType: req.PlanType,
	})
	if err != nil {
		return uuid.Nil, err
	}

	err = qtx.CreateCondominiumMember(ctx, pgstore.CreateCondominiumMemberParams{
		CondominiumID: condoID,
		UserID:        req.UserID,
		Role:          "admin",
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to add admin member: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return condoID, nil
}

func isValidPlanType(plan string) bool {
	switch plan {
	case "basic", "pro", "enterprise":
		return true
	default:
		return false
	}
}
