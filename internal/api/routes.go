// filepath: d:\go\src\secrets-manager\internal\api\routes.go
// filepath: d:\go\src\secrets-manager\internal\api\routes.go
package api

import (
	"github.com/gorilla/mux"

	"secrets-manager/internal/api/handlers"
	"secrets-manager/internal/api/middleware"
	"secrets-manager/internal/auth"
	"secrets-manager/internal/vault"
)

// ConfigureRoutes configure les routes de l'API
func ConfigureRoutes(
	router *mux.Router,
	vaultService *vault.Service,
	authService *auth.Service,
) {
	// Middleware pour toutes les routes
	router.Use(middleware.Logger)
	router.Use(middleware.Recover)

	// Gestionnaires
	secretsHandler := handlers.NewSecretsHandler(vaultService)
	authHandler := handlers.NewAuthHandler(authService)

	// Routes d'authentification (non protégées)
	router.HandleFunc("/api/v1/auth/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/api/v1/auth/register", authHandler.Register).Methods("POST")

	// Routes API protégées
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	apiRouter.Use(middleware.JWTAuth(authService))

	// Routes pour les secrets
	apiRouter.HandleFunc("/organizations/{orgID}/projects/{projectID}/environments/{env}/secrets",
		secretsHandler.ListSecrets).Methods("GET")
	apiRouter.HandleFunc("/organizations/{orgID}/projects/{projectID}/environments/{env}/secrets",
		secretsHandler.CreateSecret).Methods("POST")
	apiRouter.HandleFunc("/organizations/{orgID}/projects/{projectID}/environments/{env}/secrets/{name}",
		secretsHandler.GetSecret).Methods("GET")
	apiRouter.HandleFunc("/organizations/{orgID}/projects/{projectID}/environments/{env}/secrets/{name}",
		secretsHandler.DeleteSecret).Methods("DELETE")

	// Routes pour projets, organisations, etc.
	// ...
}
