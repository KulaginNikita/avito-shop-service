package handler

import (
	"context"
	"net/http"
	"strings"
	"github.com/golang-jwt/jwt/v4"
)

func JwtMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, http.StatusUnauthorized, "Authorization header missing")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				writeError(w, http.StatusUnauthorized, "Invalid Authorization header format")
				return
			}

			tokenString := parts[1]
			claims := jwt.MapClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				writeError(w, http.StatusUnauthorized, "Invalid token")
				return
			}

			// Извлекаем user_id
			userID, ok := claims["user_id"].(float64)
			if !ok {
				writeError(w, http.StatusUnauthorized, "Invalid user_id in token")
				return
			}
			ctx := context.WithValue(r.Context(), "user_id", int(userID))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
