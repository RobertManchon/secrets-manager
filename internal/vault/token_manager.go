// filepath: d:\go\src\secrets-manager\internal\vault\token_manager.go

// Gestionnaire de tokens Vault
package vault

import (
	"context"
	"time"

	vault "github.com/hashicorp/vault/api"
)

type TokenManager struct {
	client *vault.Client
}

// NewTokenManager crée un gestionnaire de tokens
func NewTokenManager(client *vault.Client) *TokenManager {
	return &TokenManager{client: client}
}

// CreateClientToken crée un token client temporaire avec accès limité
func (tm *TokenManager) CreateClientToken(ctx context.Context, policies []string, ttl time.Duration) (string, error) {
	// Créer un token à durée limitée avec des politiques spécifiques
	secret, err := tm.client.Auth().Token().Create(&vault.TokenCreateRequest{
		Policies: policies,
		TTL:      ttl.String(),
	})
	if err != nil {
		return "", err
	}
	return secret.Auth.ClientToken, nil
}
