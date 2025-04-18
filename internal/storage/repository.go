// filepath: d:\go\src\secrets-manager\internal\storage\repository.go

package storage

import (
	"context"
)

// SecretsCountRepository gère le comptage et la limitation des secrets
type SecretsCountRepository interface {
	// GetSecretsCount récupère le nombre de secrets pour une organisation
	GetSecretsCount(ctx context.Context, orgID string) (int, error)

	// IncrementSecretsCount incrémente le compteur de secrets
	IncrementSecretsCount(ctx context.Context, orgID string) error

	// DecrementSecretsCount décrémente le compteur de secrets
	DecrementSecretsCount(ctx context.Context, orgID string) error

	// GetSecretsLimit récupère la limite de secrets pour une organisation
	GetSecretsLimit(ctx context.Context, orgID string) (int, error)
}

// SubscriptionService gère les abonnements et limites
type SecretsSubscriptionService struct {
	repo SecretsCountRepository
}

// CanCreateSecret vérifie si l'organisation peut créer un nouveau secret
func (s *SecretsSubscriptionService) CanCreateSecret(ctx context.Context, orgID string) (bool, error) {
	count, err := s.repo.GetSecretsCount(ctx, orgID)
	if err != nil {
		return false, err
	}

	limit, err := s.repo.GetSecretsLimit(ctx, orgID)
	if err != nil {
		return false, err
	}

	return count < limit, nil
}
