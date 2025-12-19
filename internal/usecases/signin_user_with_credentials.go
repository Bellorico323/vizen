package usecases

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/Bellorico323/vizen/internal/auth"
	"github.com/Bellorico323/vizen/internal/store/pgstore"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type SigninWithCredentials interface {
	Exec(ctx context.Context, req SigninUserWithCredentialsReq) (string, error)
}

type SigninUserWithCredentials struct {
	querier      pgstore.Querier
	tokenService *auth.TokenService
}

type SigninUserWithCredentialsReq struct {
	Email     string
	Password  string
	IpAddress string
	UserAgent string
}

type SigninUserWithCredentialsRes struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func NewSigninUserWithCredentials(querier pgstore.Querier, tokenService *auth.TokenService) *SigninUserWithCredentials {
	return &SigninUserWithCredentials{
		querier:      querier,
		tokenService: tokenService,
	}
}

var (
	ErrInvalidCredentials = errors.New("Invalid credentials")
)

func (si *SigninUserWithCredentials) Exec(ctx context.Context, req SigninUserWithCredentialsReq) (SigninUserWithCredentialsRes, error) {
	user, err := si.querier.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SigninUserWithCredentialsRes{}, ErrInvalidCredentials
		}

		return SigninUserWithCredentialsRes{}, fmt.Errorf("error searching for user: %w", err)
	}

	userAccount, err := si.querier.GetAccountByUserId(ctx, user.ID)
	if err != nil {
		return SigninUserWithCredentialsRes{}, err
	}

	err = bcrypt.CompareHashAndPassword(userAccount.PasswordHash, []byte(req.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return SigninUserWithCredentialsRes{}, ErrInvalidCredentials
		}
		return SigninUserWithCredentialsRes{}, err
	}

	refreshTokenExpiration := time.Now().Add(time.Hour * 24 * 7)

	refreshToken, err := si.tokenService.GenerateToken(user.ID, refreshTokenExpiration)
	if err != nil {
		return SigninUserWithCredentialsRes{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	accessToken, err := si.tokenService.GenerateToken(user.ID, time.Now().Add(time.Minute*15))
	if err != nil {
		return SigninUserWithCredentialsRes{}, fmt.Errorf("failed to generate access token: %w", err)
	}

	var ipAddress *netip.Addr
	if parsedIp, err := netip.ParseAddr(req.IpAddress); err == nil {
		ipAddress = &parsedIp
	}

	var userAgent *string
	if req.UserAgent != "" {
		userAgent = &req.UserAgent
	}

	args := pgstore.CreateSessionParams{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: refreshTokenExpiration,
		IpAddress: ipAddress,
		UserAgent: userAgent,
	}

	if err := si.querier.CreateSession(ctx, args); err != nil {
		return SigninUserWithCredentialsRes{}, fmt.Errorf("failed to create session: %w", err)
	}

	return SigninUserWithCredentialsRes{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
