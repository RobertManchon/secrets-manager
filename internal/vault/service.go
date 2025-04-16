// filepath: d:\go\src\secrets-manager\internal\vault\service.go

package vault

import (
	"context"
	"fmt"
	"time"

	"secrets-manager/internal/models"
)

// Service fournit une abstraction de haut niveau pour interagir avec Vault
type Service struct {
	client *Client
}

// NewService crée un nouveau service Vault
func NewService(client *Client) *Service {
	return &Service{
		client: client,
	}
}

// StoreSecret stocke un secret dans Vault avec métadonnées
func (s *Service) StoreSecret(ctx context.Context, secret *models.Secret) error {
	// Construire le chemin basé sur org/projet/env
	path := buildSecretPath(secret.OrganizationID, secret.ProjectID, secret.Environment, secret.Name)

	// Préparer les données et métadonnées
	data := map[string]interface{}{
		"value":       secret.Value,
		"created_at":  time.Now().Unix(),
		"created_by":  secret.CreatedBy,
		"description": secret.Description,
	}

	return s.client.WriteSecret(ctx, path, data)
}

// GetSecret récupère un secret et le convertit en modèle Secret
func (s *Service) GetSecret(ctx context.Context, orgID, projectID, env, name string) (*models.Secret, error) {
	path := buildSecretPath(orgID, projectID, env, name)

	data, err := s.client.GetSecret(ctx, path)
	if err != nil {
		return nil, err
	}

	secret := &models.Secret{
		OrganizationID: orgID,
		ProjectID:      projectID,
		Environment:    env,
		Name:           name,
	}

	// Extraction des données
	if value, ok := data["value"].(string); ok {
		secret.Value = value
	}

	if desc, ok := data["description"].(string); ok {
		secret.Description = desc
	}

	if createdBy, ok := data["created_by"].(string); ok {
		secret.CreatedBy = createdBy
	}

	// Autres extractions...

	return secret, nil
}

// ListProjectSecrets liste tous les secrets d'un projet
func (s *Service) ListProjectSecrets(ctx context.Context, orgID, projectID, env string) ([]*models.Secret, error) {
	path := fmt.Sprintf("%s/%s/%s", orgID, projectID, env)

	keys, err := s.client.ListSecrets(ctx, path)
	if err != nil {
		return nil, err
	}

	secrets := make([]*models.Secret, 0, len(keys))
	for _, key := range keys {
		secret, err := s.GetSecret(ctx, orgID, projectID, env, key)
		if err != nil {
			continue // Ignorer les erreurs individuelles
		}
		secrets = append(secrets, secret)
	}

	return secrets, nil
}

// DeleteSecret supprime un secret
func (s *Service) DeleteSecret(ctx context.Context, orgID, projectID, env, name string) error {
	path := buildSecretPath(orgID, projectID, env, name)
	return s.client.DeleteSecret(ctx, path)
}

// Fonction utilitaire pour construire le chemin du secret
func buildSecretPath(orgID, projectID, env, name string) string {
	return fmt.Sprintf("%s/%s/%s/%s", orgID, projectID, env, name)
}
