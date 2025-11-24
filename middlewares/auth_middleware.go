package middlewares

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const uidKey contextKey = "uid"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization header", http.StatusUnauthorized)
			return
		}

		// idToken := parts[1]
		

		// token, err := firebase.Auth.VerifyIDToken(context.Background(), idToken)
		// if err != nil {
		// 	http.Error(w, "Invalid ID token", http.StatusUnauthorized)
		// 	return
		// }
		
		ctx := context.WithValue(r.Context(), uidKey, "daw")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
