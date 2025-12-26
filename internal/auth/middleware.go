package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/Bellorico323/vizen/internal/jsonutils"
	"github.com/google/uuid"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func Auth(tokenService *TokenService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				jsonutils.EncodeJson(w, r, http.StatusUnauthorized, map[string]any{
					"message": "Missing authorization header",
				})
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			tokenString = strings.TrimSpace(tokenString)

			if tokenString == "" {
				jsonutils.EncodeJson(w, r, http.StatusUnauthorized, map[string]any{
					"message": "Malformed authorization header",
				})
				return
			}

			userID, err := tokenService.ValidateToken(tokenString)
			if err != nil {
				jsonutils.EncodeJson(w, r, http.StatusUnauthorized, map[string]any{
					"message": "Invalid or expired token",
				})
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}
