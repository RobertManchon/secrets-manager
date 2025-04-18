// filepath: d:\go\src\secrets-manager\internal\storage\mysql\organizations_repository.go

/*************************************************************************/
/*                                                                       */
/*   Ce fichier implémente le repository MySQL pour les organisations    */
/*   Il gère les opérations CRUD pour les organisations dans MySQL       */
/*                                                                       */
/*************************************************************************/

package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"secrets-manager/internal/models"
)

// ErrOrganizationNotFound indique qu'une organisation n'a pas été trouvée
var ErrOrganizationNotFound = errors.New("organisation non trouvée")

// ErrOrganizationNameExists indique qu'une organisation avec ce nom existe déjà
var ErrOrganizationNameExists = errors.New("une organisation avec ce nom existe déjà")

// OrganizationsRepository gère l'accès aux données d'organisation dans MySQL
type OrganizationsRepository struct {
	db *sql.DB
}

// NewOrganizationsRepository crée un nouveau repository pour les organisations
func NewOrganizationsRepository(db *sql.DB) *OrganizationsRepository {
	return &OrganizationsRepository{
		db: db,
	}
}

// CreateOrganization crée une nouvelle organisation
func (r *OrganizationsRepository) CreateOrganization(ctx context.Context, org *models.Organization) error {
	// Vérifier si le nom existe déjà
	var exists bool
	err := r.db.QueryRowContext(ctx, 
		"SELECT EXISTS(SELECT 1 FROM organizations WHERE name = ?)", 
		org.Name).Scan(&exists)
	
	if err != nil {
		return err
	}
	
	if exists {
		return ErrOrganizationNameExists
	}

	// Générer un ID si non fourni
	if org.ID == "" {
		org.ID = uuid.New().String()
	}

	// Initialiser les timestamps
	now := time.Now()
	if org.CreatedAt.IsZero() {
		org.CreatedAt = now
	}
	if org.UpdatedAt.IsZero() {
		org.UpdatedAt = now
	}

	// Démarrer une transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insérer l'organisation
	query := `
		INSERT INTO organizations (
			id, name, description, plan_id, created_at, updated_at, owner_id
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err = tx.ExecContext(
		ctx,
		query,
		org.ID,
		org.Name,
		org.Description,
		org.PlanID,
		org.CreatedAt,
		org.UpdatedAt,
		org.OwnerID,
	)

	if err != nil {
		return err
	}

	// Ajouter le créateur comme admin de l'organisation
	userOrgQuery := `
		INSERT INTO user_organizations (
			user_id, organization_id, role, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?)
	`

	_, err = tx.ExecContext(
		ctx,
		userOrgQuery,
		org.OwnerID,
		org.ID,
		"admin", // Le créateur est automatiquement admin
		now,
		now,
	)

	if err != nil {
		return err
	}

	// Valider la transaction
	return tx.Commit()
}

// GetOrganizationByID récupère une organisation par son ID
func (r *OrganizationsRepository) GetOrganizationByID(ctx context.Context, id string) (*models.Organization, error) {
	query := `
		SELECT id, name, description, plan_id, created_at, updated_at, owner_id
		FROM organizations
		WHERE id = ?
	`

	org := &models.Organization{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&org.ID,
		&org.Name,
		&org.Description,
		&org.PlanID,
		&org.CreatedAt,
		&org.UpdatedAt,
		&org.OwnerID,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrganizationNotFound
		}
		return nil, err
	}

	return org, nil
}

// ListUserOrganizations liste toutes les organisations d'un utilisateur
func (r *OrganizationsRepository) ListUserOrganizations(ctx context.Context, userID string) ([]*models.Organization, error) {
	query := `
		SELECT o.id, o.name, o.description, o.plan_id, o.created_at, o.updated_at, o.owner_id
		FROM organizations o
		JOIN user_organizations uo ON o.id = uo.organization_id
		WHERE uo.user_id = ?
		ORDER BY o.name
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []*models.Organization
	for rows.Next() {
		org := &models.Organization{}
		err := rows.Scan(
			&org.ID,
			&org.Name,
			&org.Description,
			&org.PlanID,
			&org.CreatedAt,
			&org.UpdatedAt,
			&org.OwnerID,
		)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orgs, nil
}

// UpdateOrganization met à jour une organisation
func (r *OrganizationsRepository) UpdateOrganization(ctx context.Context, org *models.Organization) error {
	// Vérifier si le nom est déjà utilisé par une autre organisation
	var existingID string
	err := r.db.QueryRowContext(ctx, 
		"SELECT id FROM organizations WHERE name = ? AND id != ?", 
		org.Name, org.ID).Scan(&existingID)
	
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	
	if existingID != "" {
		return ErrOrganizationNameExists
	}

	// Mettre à jour l'organisation
	query := `
		UPDATE organizations
		SET name = ?, description = ?, updated_at = NOW()
		WHERE id = ?
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		org.Name,
		org.Description,
		org.ID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrOrganizationNotFound
	}

	return nil
}

// DeleteOrganization supprime une organisation
func (r *OrganizationsRepository) DeleteOrganization(ctx context.Context, id string) error {
	// Vérifier d'abord si l'organisation existe
	_, err := r.GetOrganizationByID(ctx, id)
	if err != nil {
		return err
	}

	// Démarrer une transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Supprimer d'abord les relations user_organizations
	userOrgQuery := "DELETE FROM user_organizations WHERE organization_id = ?"
	_, err = tx.ExecContext(ctx, userOrgQuery, id)
	if err != nil {
		return err
	}

	// Supprimer les projets de l'organisation
	projectsQuery := "DELETE FROM projects WHERE organization_id = ?"
	_, err = tx.ExecContext(ctx, projectsQuery, id)
	if err != nil {
		return err
	}

	// Supprimer les statistiques d'usage
	statsQuery := "DELETE FROM usage_statistics WHERE organization_id = ?"
	_, err = tx.ExecContext(ctx, statsQuery, id)
	if err != nil {
		return err
	}

	// Supprimer les abonnements
	subscriptionsQuery := "DELETE FROM subscriptions WHERE organization_id = ?"
	_, err = tx.ExecContext(ctx, subscriptionsQuery, id)
	if err != nil {
		return err
	}

	// Supprimer les secrets
	secretsQuery := "DELETE FROM secret_metadata WHERE organization_id = ?"
	_, err = tx.ExecContext(ctx, secretsQuery, id)
	if err != nil {
		return err
	}

	// Supprimer l'organisation elle-même
	orgQuery := "DELETE FROM organizations WHERE id = ?"
	_, err = tx.ExecContext(ctx, orgQuery, id)
	if err != nil {
		return err
	}

	// Valider la transaction
	return tx.Commit()
}

// ListOrganizationUsers liste tous les utilisateurs d'une organisation
func (r *OrganizationsRepository) ListOrganizationUsers(ctx context.Context, orgID string) ([]*models.UserOrganization, error) {
	query := `
		SELECT u.id, u.email, u.first_name, u.last_name, u.role, 
			   uo.role, uo.created_at, uo.updated_at
		FROM users u
		JOIN user_organizations uo ON u.id = uo.user_id
		WHERE uo.organization_id = ?
		ORDER BY u.last_name, u.first_name
	`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userOrgs []*models.UserOrganization
	for rows.Next() {
		user := &models.User{}
		userOrg := &models.UserOrganization{
			OrganizationID: orgID,
		}

		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.FirstName,
			&user.LastName,
			&user.Role,
			&userOrg.Role,
			&userOrg.CreatedAt,
			&userOrg.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		userOrg.UserID = user.ID

		userOrgs = append(userOrgs, userOrg)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return userOrgs, nil
}

// AddUserToOrganization ajoute un utilisateur à une organisation
func (r *OrganizationsRepository) AddUserToOrganization(ctx context.Context, userID, orgID, role string) error {
	// Vérifier si l'utilisateur est déjà dans l'organisation
	var exists bool
	err := r.db.QueryRowContext(ctx, 
		"SELECT EXISTS(SELECT 1 FROM user_organizations WHERE user_id = ? AND organization_id = ?)",
		userID, orgID).Scan(&exists)
	
	if err != nil {
		return err
	}
	
	if exists {
		// Mettre à jour le rôle
		query := `
			UPDATE user_organizations
			SET role = ?, updated_at = NOW()
			WHERE user_id = ? AND organization_id = ?
		`
		_, err = r.db.ExecContext(ctx, query, role, userID, orgID)
		return err
	}
	
	// Ajouter l'utilisateur
	now := time.Now()
	query := `
		INSERT INTO user_organizations (
			user_id, organization_id, role, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?)
	`
	_, err = r.db.ExecContext(ctx, query, userID, orgID, role, now, now)
	return err
}

// RemoveUserFromOrganization retire un utilisateur d'une organisation
func (r *OrganizationsRepository) RemoveUserFromOrganization(ctx context.Context, userID, orgID string) error {
	// Vérifier si l'utilisateur est le propriétaire
	var isOwner bool
	err := r.db.QueryRowContext(ctx, 
		"SELECT EXISTS(SELECT 1 FROM organizations WHERE id = ? AND owner_id = ?)",
		orgID, userID).Scan(&isOwner)
	
	if err != nil {
		return err
	}
	
	if isOwner {
		return errors.New("impossible de retirer le propriétaire de l'organisation")
	}
	
	// Supprimer l'utilisateur
	query := "DELETE FROM user_organizations WHERE user_id = ? AND organization_id = ?"
	result, err := r.db.ExecContext(ctx, query, userID, orgID)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return errors.New("l'utilisateur n'appartient pas à cette organisation")
	}
	
	return nil
}

// ChangeOrganizationOwner change le propriétaire d'une organisation
func (r *OrganizationsRepository) ChangeOrganizationOwner(ctx context.Context, orgID, newOwnerID string) error {
	// Vérifier si le nouvel utilisateur appartient à l'organisation
	var isMember bool
	err := r.db.QueryRowContext(ctx, 
		"SELECT EXISTS(SELECT 1 FROM user_organizations WHERE user_id = ? AND organization_id = ?)",
		newOwnerID, orgID).Scan(&isMember)
	
	if err != nil {
		return err
	}
	
	if !isMember {
		return errors.New("le nouvel utilisateur n'appartient pas à cette organisation")
	}
	
	// Démarrer une transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// Mettre à jour le propriétaire
	query := `
		UPDATE organizations
		SET owner_id = ?, updated_at = NOW()
		WHERE id = ?
	`
	_, err = tx.ExecContext(ctx, query, newOwnerID, orgID)
	if err != nil {
		return err
	}
	
	// Assurer que le nouveau propriétaire a les droits d'administrateur
	userOrgQuery := `
		UPDATE user_organizations
		SET role = 'admin', updated_at = NOW()
		WHERE user_id = ? AND organization_id = ?
	`
	_, err = tx.ExecContext(ctx, userOrgQuery, newOwnerID, orgID)
	if err != nil {
		return err
	}
	
	// Valider la transaction
	return tx.Commit()
}

// UpdateOrganizationPlan met à jour le plan d'une organisation
func (r *OrganizationsRepository) UpdateOrganizationPlan(ctx context.Context, orgID, planID string) error {
	query := `
		UPDATE organizations
		SET plan_id = ?, updated_at = NOW()
		WHERE id = ?
	`
	
	result, err := r.db.ExecContext(ctx, query, planID, orgID)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return ErrOrganizationNotFound
	}
	
	return nil
}

// GetOrganizationPlan récupère le plan actuel d'une organisation
func (r *OrganizationsRepository) GetOrganizationPlan(ctx context.Context, orgID string) (string, error) {
	query := "SELECT plan_id FROM organizations WHERE id = ?"
	
	var planID string
	err := r.db.QueryRowContext(ctx, query, orgID).Scan(&planID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrOrganizationNotFound
		}
		return "", err
	}
	
	return planID, nil
}

// CountOrganizationSecrets compte le nombre de secrets d'une organisation
func (r *OrganizationsRepository) CountOrganizationSecrets(ctx context.Context, orgID string) (int, error) {
	query := "SELECT COUNT(*) FROM secret_metadata WHERE organization_id = ?"
	
	var count int
	err := r.db.QueryRowContext(ctx, query, orgID).Scan(&count)
	if err != nil {
		return 0, err
	}
	
	return count, nil
}
