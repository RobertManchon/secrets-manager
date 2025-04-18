// filepath: d:\go\src\secrets-manager\internal\storage\mysql\users_repository.go

/*************************************************************************/
/*                                                                       */
/*   Ce fichier implémente le repository MySQL pour les utilisateurs     */
/*   Il gère les opérations CRUD pour les utilisateurs dans MySQL        */
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

// ErrUserNotFound indique qu'un utilisateur n'a pas été trouvé
var ErrUserNotFound = errors.New("utilisateur non trouvé")

// ErrEmailAlreadyExists indique qu'un email est déjà utilisé
var ErrEmailAlreadyExists = errors.New("cet email est déjà utilisé")

// UsersRepository gère l'accès aux données utilisateur dans MySQL
type UsersRepository struct {
	db *sql.DB
}

// NewUsersRepository crée un nouveau repository pour les utilisateurs
func NewUsersRepository(db *sql.DB) *UsersRepository {
	return &UsersRepository{
		db: db,
	}
}

// CreateUser crée un nouvel utilisateur dans la base de données
func (r *UsersRepository) CreateUser(ctx context.Context, user *models.User) error {
	// Vérifier si l'email existe déjà
	var exists bool
	err := r.db.QueryRowContext(ctx, 
		"SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", 
		user.Email).Scan(&exists)
	
	if err != nil {
		return err
	}
	
	if exists {
		return ErrEmailAlreadyExists
	}

	// Générer un ID si non fourni
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	// Initialiser les timestamps
	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = now
	}

	// Définir un rôle par défaut si non spécifié
	if user.Role == "" {
		user.Role = "user"
	}

	query := `
		INSERT INTO users (
			id, email, hashed_password, first_name, last_name, 
			role, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.ExecContext(
		ctx,
		query,
		user.ID,
		user.Email,
		user.HashedPassword,
		user.FirstName,
		user.LastName,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	)

	return err
}

// GetUserByID récupère un utilisateur par son ID
func (r *UsersRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, email, hashed_password, first_name, last_name, 
			   role, created_at, updated_at
		FROM users
		WHERE id = ?
	`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.HashedPassword,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// GetUserByEmail récupère un utilisateur par son email
func (r *UsersRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, hashed_password, first_name, last_name, 
			   role, created_at, updated_at
		FROM users
		WHERE email = ?
	`

	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.HashedPassword,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// UpdateUser met à jour les informations d'un utilisateur
func (r *UsersRepository) UpdateUser(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET email = ?, first_name = ?, last_name = ?, role = ?, updated_at = NOW()
		WHERE id = ?
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		user.Email,
		user.FirstName,
		user.LastName,
		user.Role,
		user.ID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdatePassword met à jour le mot de passe d'un utilisateur
func (r *UsersRepository) UpdatePassword(ctx context.Context, userID, hashedPassword string) error {
	query := `
		UPDATE users
		SET hashed_password = ?, updated_at = NOW()
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, hashedPassword, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// DeleteUser supprime un utilisateur
func (r *UsersRepository) DeleteUser(ctx context.Context, id string) error {
	// Vérifier les contraintes de clé étrangère avant la suppression
	// (si l'utilisateur est référencé ailleurs, il faudra gérer ce cas)
	
	// Pour l'instant, on supprime simplement l'utilisateur
	query := "DELETE FROM users WHERE id = ?"
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// ListUsers liste tous les utilisateurs avec pagination
func (r *UsersRepository) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	query := `
		SELECT id, email, hashed_password, first_name, last_name, 
			   role, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.HashedPassword,
			&user.FirstName,
			&user.LastName,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// CountUsers compte le nombre total d'utilisateurs
func (r *UsersRepository) CountUsers(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetUserOrganizations récupère toutes les organisations d'un utilisateur
func (r *UsersRepository) GetUserOrganizations(ctx context.Context, userID string) ([]*models.Organization, error) {
	query := `
		SELECT o.id, o.name, o.description, o.plan_id, o.created_at, o.updated_at, o.owner_id
		FROM organizations o
		JOIN user_organizations uo ON o.id = uo.organization_id
		WHERE uo.user_id = ?
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

// GetUserRole récupère le rôle d'un utilisateur dans une organisation
func (r *UsersRepository) GetUserRole(ctx context.Context, userID, orgID string) (string, error) {
	query := `
		SELECT role
		FROM user_organizations
		WHERE user_id = ? AND organization_id = ?
	`

	var role string
	err := r.db.QueryRowContext(ctx, query, userID, orgID).Scan(&role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", err
	}

	return role, nil
}

// AssignUserToOrganization assigne un utilisateur à une organisation avec un rôle
func (r *UsersRepository) AssignUserToOrganization(ctx context.Context, userID, orgID, role string) error {
	// Vérifier si l'assignation existe déjà
	var exists bool
	err := r.db.QueryRowContext(ctx, 
		"SELECT EXISTS(SELECT 1 FROM user_organizations WHERE user_id = ? AND organization_id = ?)",
		userID, orgID).Scan(&exists)
	
	if err != nil {
		return err
	}
	
	// Si l'assignation existe, mettre à jour le rôle
	if exists {
		query := `
			UPDATE user_organizations
			SET role = ?, updated_at = NOW()
			WHERE user_id = ? AND organization_id = ?
		`
		_, err = r.db.ExecContext(ctx, query, role, userID, orgID)
		return err
	}
	
	// Sinon, créer une nouvelle assignation
	query := `
		INSERT INTO user_organizations (user_id, organization_id, role, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())
	`
	_, err = r.db.ExecContext(ctx, query, userID, orgID, role)
	return err
}

// RemoveUserFromOrganization supprime un utilisateur d'une organisation
func (r *UsersRepository) RemoveUserFromOrganization(ctx context.Context, userID, orgID string) error {
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
		return ErrUserNotFound
	}

	return nil
}
