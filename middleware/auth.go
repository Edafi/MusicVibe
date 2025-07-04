package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	ContextUserIDKey contextKey = "userID"
	JWT_SECRET                  = "1111"
)

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		authHeader := request.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			log.Println("Middleware: Missing or invalid Authorization header")
			http.Error(response, "Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte(JWT_SECRET), nil
		})
		if err != nil || !token.Valid {
			log.Println("Middleware: Missing or invalid Authorization header")
			http.Error(response, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Println("Middleware: Invalid token claims")
			http.Error(response, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			log.Println("Middleware: Invalid user ID")
			http.Error(response, "Invalid user ID", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(request.Context(), ContextUserIDKey, userID)
		next.ServeHTTP(response, request.WithContext(ctx))
	})
}
