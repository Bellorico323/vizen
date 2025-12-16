package auth

import (
	"crypto/ecdsa"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenService struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
}

func NewTokenService() (*TokenService, error) {
	privKeyContent := os.Getenv("JWT_PRIVATE_KEY")
	if privKeyContent == "" {
		return nil, fmt.Errorf("JWT_PRIVATE_KEY is missing")
	}

	privKeyBytes, err := base64.StdEncoding.DecodeString(privKeyContent)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key base64: %w", err)
	}

	privateKey, err := jwt.ParseECPrivateKeyFromPEM(privKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	pubKeyContent := os.Getenv("JWT_PUBLIC_KEY")
	if pubKeyContent == "" {
		return nil, fmt.Errorf("JWT_PUBLIC_KEY is missing")
	}

	pubKeyBytes, err := base64.StdEncoding.DecodeString(pubKeyContent)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key base64: %w", err)
	}

	publicKey, err := jwt.ParseECPublicKeyFromPEM(pubKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return &TokenService{
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

func (s *TokenService) GenerateToken(userId uuid.UUID, expTime time.Time) (string, error) {
	claims := jwt.MapClaims{
		"sub": userId.String(),
		"iat": time.Now().Unix(),
		"exp": expTime.Unix(),
		"iss": "vizen-api",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)

	signedToken, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

func (s *TokenService) ValidateToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}

		return s.publicKey, nil
	})

	if err != nil {
		return uuid.UUID{}, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		sub, err := claims.GetSubject()
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("invalid subject")
		}

		uid, err := uuid.Parse(sub)
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("invalid uuid in subject")
		}
		return uid, nil
	}

	return uuid.UUID{}, fmt.Errorf("invalid token")
}
