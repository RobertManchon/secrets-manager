// filepath: internal/api/middleware/middleware.go

package middleware

import (
	"context"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"secrets-manager/internal/auth"
)

// Logger est un middleware pour journaliser les requêtes
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)

		log.Printf("Completed %s %s in %v", r.Method, r.URL.Path, time.Since(start))
	})
}

// Recover est un middleware pour récupérer des paniques
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v\n%s", err, debug.Stack())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// JWTAuth est un middleware pour l'authentification JWT
func JWTAuth(authService *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extraire le token de l'en-tête Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Autorisation requise", http.StatusUnauthorized)
				return
			}

			// Vérifier le format Bearer token
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				http.Error(w, "Format d'autorisation invalide", http.StatusUnauthorized)
				return
			}

			// Vérifier le token
			userID, err := authService.VerifyToken(tokenParts[1])
			if err != nil {
				http.Error(w, "Token invalide", http.StatusUnauthorized)
				return
			}

			// Ajouter l'ID utilisateur au contexte
			ctx := context.WithValue(r.Context(), "userID", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
