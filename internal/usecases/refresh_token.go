package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshToken interface {
	Exec(ctx context.Context, refreshToken string) (RefreshTokenResponse, error)
}

type RefreshTokenUseCase struct {
	querier      pgstore.Querier
	pool         *pgxpool.Pool
	tokenService *auth.TokenService
}

func NewRefreshTokenUseCase(pool *pgxpool.Pool, tokenService *auth.TokenService) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		querier:      pgstore.New(pool),
		pool:         pool,
		tokenService: tokenService,
	}
}

type RefreshTokenResponse struct {
	AccessToken  string
	RefreshToken string
}

var ErrInvalidToken = errors.New("invalid or expired refresh token")

func (uc *RefreshTokenUseCase) Exec(ctx context.Context, refreshToken string) (RefreshTokenResponse, error) {
	userIDFromToken, err := uc.tokenService.ValidateToken(refreshToken)
	if err != nil {
		return RefreshTokenResponse{}, ErrInvalidToken
	}

	session, err := uc.querier.GetSessionByToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RefreshTokenResponse{}, ErrInvalidToken
		}
		return RefreshTokenResponse{}, fmt.Errorf("database error: %w", err)
	}

	if session.UserID != userIDFromToken {
		return RefreshTokenResponse{}, ErrInvalidToken
	}

	if session.ExpiresAt.Before(time.Now()) {
		return RefreshTokenResponse{}, ErrInvalidToken
	}

	newAccessToken, err := uc.tokenService.GenerateToken(session.UserID, time.Now().Add(time.Minute*15))
	if err != nil {
		return RefreshTokenResponse{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	newExpiration := time.Now().Add(time.Hour * 24 * 7)
	newRefreshToken, err := uc.tokenService.GenerateToken(session.UserID, newExpiration)
	if err != nil {
		return RefreshTokenResponse{}, err
	}

	err = uc.querier.UpdateRefreshToken(ctx, pgstore.UpdateRefreshTokenParams{
		ID:        session.ID,
		Token:     newRefreshToken,
		ExpiresAt: newExpiration,
	})
	if err != nil {
		return RefreshTokenResponse{}, fmt.Errorf("failed to rotate session: %w", err)
	}

	return RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}
