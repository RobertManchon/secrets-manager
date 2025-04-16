/* filepath: internal/storage/mysql/secrets_repository.go */

/*************************************************************************/
/*                                                                       */
/*   Ce fichier implémente le repository MySQL pour les secrets          */
/*   Il gère la persistance des métadonnées de secrets dans MySQL        */
/*                                                                       */
/*************************************************************************/

package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"secrets-manager/internal/models"
)

// SecretsRepository gère l'accès aux métadonnées des secrets dans MySQL
type SecretsRepository struct {
	db *sql.DB
}

// NewSecretsRepository crée un nouveau repository pour les secrets
func NewSecretsRepository(db *sql.DB) *SecretsRepository {
	return &SecretsRepository{
		db: db,
	}
}

// CreateSecretMetadata crée les métadonnées d'un secret
func (r *SecretsRepository) CreateSecretMetadata(ctx context.Context, metadata *models.SecretMetadata) error {
	// Générer un UUID si non fourni
	if metadata.ID == "" {
		metadata.ID = uuid.New().String()
	}

	query := `
		INSERT INTO secret_metadata (
			id, name, description, organization_id, project_id, 
			environment, created_by, created_at, updated_at, version
		) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW(), ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		metadata.ID,
		metadata.Name,
		metadata.Description,
		metadata.OrganizationID,
		metadata.ProjectID,
		metadata.Environment,
		metadata.CreatedBy,
		metadata.Version,
	)

	if err != nil {
		return err
	}

	// Mettre à jour les statistiques d'usage
	return r.incrementSecretsCount(ctx, metadata.OrganizationID)
}

// GetSecretMetadata récupère les métadonnées d'un secret par son ID
func (r *SecretsRepository) GetSecretMetadata(ctx context.Context, id string) (*models.SecretMetadata, error) {
	query := `
		SELECT id, name, description, organization_id, project_id, 
			   environment, created_by, created_at, updated_at, version
		FROM secret_metadata
		WHERE id = ?
	`

	metadata := &models.SecretMetadata{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&metadata.ID,
		&metadata.Name,
		&metadata.Description,
		&metadata.OrganizationID,
		&metadata.ProjectID,
		&metadata.Environment,
		&metadata.CreatedBy,
		&metadata.CreatedAt,
		&metadata.UpdatedAt,
		&metadata.Version,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Pas d'erreur, juste pas de résultat
		}
		return nil, err
	}

	return metadata, nil
}

// GetSecretMetadataByPath récupère les métadonnées d'un secret par son chemin complet
func (r *SecretsRepository) GetSecretMetadataByPath(
	ctx context.Context,
	orgID, projectID, env, name string,
) (*models.SecretMetadata, error) {
	query := `
		SELECT id, name, description, organization_id, project_id, 
			   environment, created_by, created_at, updated_at, version
		FROM secret_metadata
		WHERE organization_id = ? AND project_id = ? AND environment = ? AND name = ?
	`

	metadata := &models.SecretMetadata{}
	err := r.db.QueryRowContext(ctx, query, orgID, projectID, env, name).Scan(
		&metadata.ID,
		&metadata.Name,
		&metadata.Description,
		&metadata.OrganizationID,
		&metadata.ProjectID,
		&metadata.Environment,
		&metadata.CreatedBy,
		&metadata.CreatedAt,
		&metadata.UpdatedAt,
		&metadata.Version,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Pas d'erreur, juste pas de résultat
		}
		return nil, err
	}

	return metadata, nil
}

// ListProjectSecrets liste tous les secrets d'un projet et environnement
func (r *SecretsRepository) ListProjectSecrets(
	ctx context.Context,
	orgID, projectID, env string,
) ([]*models.SecretMetadata, error) {
	query := `
		SELECT id, name, description, organization_id, project_id, 
			   environment, created_by, created_at, updated_at, version
		FROM secret_metadata
		WHERE organization_id = ? AND project_id = ? AND environment = ?
	`

	rows, err := r.db.QueryContext(ctx, query, orgID, projectID, env)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var secrets []*models.SecretMetadata
	for rows.Next() {
		metadata := &models.SecretMetadata{}
		err := rows.Scan(
			&metadata.ID,
			&metadata.Name,
			&metadata.Description,
			&metadata.OrganizationID,
			&metadata.ProjectID,
			&metadata.Environment,
			&metadata.CreatedBy,
			&metadata.CreatedAt,
			&metadata.UpdatedAt,
			&metadata.Version,
		)
		if err != nil {
			return nil, err
		}
		secrets = append(secrets, metadata)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return secrets, nil
}

// UpdateSecretMetadata met à jour les métadonnées d'un secret
func (r *SecretsRepository) UpdateSecretMetadata(ctx context.Context, metadata *models.SecretMetadata) error {
	query := `
		UPDATE secret_metadata
		SET name = ?, description = ?, updated_at = NOW(), version = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		metadata.Name,
		metadata.Description,
		metadata.Version,
		metadata.ID,
	)

	return err
}

// DeleteSecretMetadata supprime les métadonnées d'un secret
func (r *SecretsRepository) DeleteSecretMetadata(ctx context.Context, id string, orgID string) error {
	query := "DELETE FROM secret_metadata WHERE id = ?"

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// Mettre à jour les statistiques d'usage
	return r.decrementSecretsCount(ctx, orgID)
}

// DeleteSecretMetadataByPath supprime les métadonnées d'un secret par son chemin
func (r *SecretsRepository) DeleteSecretMetadataByPath(
	ctx context.Context,
	orgID, projectID, env, name string,
) error {
	// D'abord récupérer les métadonnées pour avoir l'ID
	metadata, err := r.GetSecretMetadataByPath(ctx, orgID, projectID, env, name)
	if err != nil {
		return err
	}

	if metadata == nil {
		return nil // Rien à supprimer
	}

	return r.DeleteSecretMetadata(ctx, metadata.ID, orgID)
}

// Méthodes pour la gestion des statistiques

func (r *SecretsRepository) incrementSecretsCount(ctx context.Context, orgID string) error {
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

func (r *SecretsRepository) decrementSecretsCount(ctx context.Context, orgID string) error {
	query := `
		UPDATE usage_statistics 
		SET secret_count = GREATEST(0, secret_count - 1), last_updated = NOW() 
		WHERE organization_id = ?
	`

	_, err := r.db.ExecContext(ctx, query, orgID)
	return err
}

// GetSecretsCount obtient le nombre de secrets pour une organisation
func (r *SecretsRepository) GetSecretsCount(ctx context.Context, orgID string) (int, error) {
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

// GetSecretsLimit obtient la limite de secrets pour une organisation
func (r *SecretsRepository) GetSecretsLimit(ctx context.Context, orgID string) (int, error) {
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
			return 0, nil // Pas d'erreur, juste pas d'abonnement actif
		}
		return 0, err
	}

	return limit, nil
}
