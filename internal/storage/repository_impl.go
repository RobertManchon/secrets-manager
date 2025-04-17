// filepath: d:\go\src\secrets-manager\internal\storage\mysql\repository_impl.go

/*************************************************************************/
/*                                                                       */
/*   Ce fichier implémente l'interface SecretsCountRepository            */
/*   Il gère les compteurs et limites de secrets pour les organisations  */
/*                                                                       */
/*************************************************************************/

package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

// SecretCountRepository implémente l'interface storage.SecretsCountRepository
type SecretCountRepository struct {
	db *sql.DB
}

// NewSecretCountRepository crée un nouveau repository pour la gestion des compteurs
func NewSecretCountRepository(db *sql.DB) *SecretCountRepository {
	return &SecretCountRepository{
		db: db,
	}
}

// GetSecretsCount récupère le nombre de secrets pour une organisation
func (r *SecretCountRepository) GetSecretsCount(ctx context.Context, orgID string) (int, error) {
	query := "SELECT secret_count FROM usage_statistics WHERE organization_id = ?"

	var count int
	err := r.db.QueryRowContext(ctx, query, orgID).Scan(&count)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil // Pas d'erreur, juste pas d'enregistrement trouvé
		}
		return 0, err
	}

	return count, nil
}

// IncrementSecretsCount incrémente le compteur de secrets pour une organisation
func (r *SecretCountRepository) IncrementSecretsCount(ctx context.Context, orgID string) error {
	// Tentative de mise à jour
	query := `
		UPDATE usage_statistics 
		SET secret_count = secret_count + 1, last_updated = NOW() 
		WHERE organization_id = ?
	`

	result, err := r.db.ExecContext(ctx, query, orgID)
	if err != nil {
		return err
	}

	// Si aucune ligne n'a été mise à jour, insérer un nouveau record
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		insertQuery := `
			INSERT INTO usage_statistics (id, organization_id, secret_count, api_calls, last_updated)
			VALUES (?, ?, 1, 0, NOW())
		`
		_, err = r.db.ExecContext(ctx, insertQuery, uuid.New().String(), orgID)
		return err
	}

	return nil
}

// DecrementSecretsCount décrémente le compteur de secrets pour une organisation
func (r *SecretCountRepository) DecrementSecretsCount(ctx context.Context, orgID string) error {
	query := `
		UPDATE usage_statistics 
		SET secret_count = GREATEST(0, secret_count - 1), last_updated = NOW() 
		WHERE organization_id = ?
	`

	_, err := r.db.ExecContext(ctx, query, orgID)
	return err
}

// GetSecretsLimit récupère la limite de secrets pour une organisation
func (r *SecretCountRepository) GetSecretsLimit(ctx context.Context, orgID string) (int, error) {
	query := `
		SELECT s.secrets_limit 
		FROM subscriptions s
		WHERE s.organization_id = ? AND s.status = 'active'
		ORDER BY s.end_date DESC
		LIMIT 1
	`

	var limit int
	err := r.db.QueryRowContext(ctx, query, orgID).Scan(&limit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Pas d'abonnement actif trouvé, utiliser une limite par défaut
			return 5, nil // Limite gratuite par défaut (5 secrets)
		}
		return 0, err
	}

	return limit, nil
}

// IncrementAPICallCount incrémente le compteur d'appels API pour une organisation
func (r *SecretCountRepository) IncrementAPICallCount(ctx context.Context, orgID string) error {
	// Tentative de mise à jour
	query := `
		UPDATE usage_statistics 
		SET api_calls = api_calls + 1, last_updated = NOW() 
		WHERE organization_id = ?
	`

	result, err := r.db.ExecContext(ctx, query, orgID)
	if err != nil {
		return err
	}

	// Si aucune ligne n'a été mise à jour, insérer un nouveau record
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		insertQuery := `
			INSERT INTO usage_statistics (id, organization_id, secret_count, api_calls, last_updated)
			VALUES (?, ?, 0, 1, NOW())
		`
		_, err = r.db.ExecContext(ctx, insertQuery, uuid.New().String(), orgID)
		return err
	}

	return nil
}

// GetUsageStatistics récupère les statistiques d'usage pour une organisation
func (r *SecretCountRepository) GetUsageStatistics(ctx context.Context, orgID string) (int, int, error) {
	query := `
		SELECT secret_count, api_calls
		FROM usage_statistics 
		WHERE organization_id = ?
	`

	var secretCount, apiCalls int
	err := r.db.QueryRowContext(ctx, query, orgID).Scan(&secretCount, &apiCalls)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, nil // Pas d'erreur, juste pas d'enregistrement trouvé
		}
		return 0, 0, err
	}

	return secretCount, apiCalls, nil
}
