package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type SignupWithCredentials interface {
	Exec(ctx context.Context, req SignupUserwithCredentialsReq) (uuid.UUID, error)
}

type SignupUserWithCredentials struct {
	querier pgstore.Querier
	pool    *pgxpool.Pool
}

type SignupUserwithCredentialsReq struct {
	Name       string
	Role       string
	Avatar_url *string
	Email      string
	Password   string
}

func NewSignupWithCredentialsUseCase(pool *pgxpool.Pool) *SignupUserWithCredentials {
	return &SignupUserWithCredentials{
		pool:    pool,
		querier: pgstore.New(pool),
	}
}

var (
	ErrDuplicatedEmail = errors.New("email already exists")
)

func (su *SignupUserWithCredentials) Exec(ctx context.Context, payload SignupUserwithCredentialsReq) (uuid.UUID, error) {
	user, err := su.querier.GetUserByEmail(ctx, payload.Email)
	if err == nil {
		return user.ID, ErrDuplicatedEmail
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.UUID{}, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), 12)
	if err != nil {
		return uuid.UUID{}, err
	}

	tx, err := su.pool.Begin(ctx)
	if err != nil {
		return uuid.UUID{}, err
	}

	defer tx.Rollback(ctx)

	qtx := pgstore.New(tx)

	userId, err := qtx.CreateUser(ctx, pgstore.CreateUserParams{
		Name:      payload.Name,
		Email:     payload.Email,
		AvatarUrl: payload.Avatar_url,
		Role:      payload.Role,
	})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("Failed to create user: %w", err)
	}

	err = qtx.CreateAccountWithCredentials(ctx, pgstore.CreateAccountWithCredentialsParams{
		UserID:            userId,
		ProviderAccountID: userId.String(),
		ProviderID:        "credentials",
		PasswordHash:      hashedPassword,
	})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("Failed to create account: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.UUID{}, fmt.Errorf("Failed to commit transaction: %w", err)
	}

	return userId, nil
}
