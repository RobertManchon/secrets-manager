// filepath: d:\go\src\secrets-manager\internal\models\secret.go

package models

import (
	"time"
)

// Secret représente un secret stocké dans le système
type SecretData struct {
	ID             string    `json:"id,omitempty" db:"id"`
	Name           string    `json:"name" db:"name"`
	Value          string    `json:"value,omitempty" db:"-"` // Ne pas stocker dans la BDD
	Description    string    `json:"description" db:"description"`
	OrganizationID string    `json:"organization_id" db:"organization_id"`
	ProjectID      string    `json:"project_id" db:"project_id"`
	Environment    string    `json:"environment" db:"environment"`
	CreatedBy      string    `json:"created_by" db:"created_by"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	Version        int       `json:"version" db:"version"`
}

// SecretMetadata contient les métadonnées d'un secret sans sa valeur
type SecretMetadata struct {
	ID             string    `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	Description    string    `json:"description" db:"description"`
	OrganizationID string    `json:"organization_id" db:"organization_id"`
	ProjectID      string    `json:"project_id" db:"project_id"`
	Environment    string    `json:"environment" db:"environment"`
	CreatedBy      string    `json:"created_by" db:"created_by"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	Version        int       `json:"version" db:"version"`
}

// ToMetadata convertit un Secret en SecretMetadata (sans la valeur)
func (s *SecretData) ToMetadata() *SecretMetadata {
	return &SecretMetadata{
		ID:             s.ID,
		Name:           s.Name,
		Description:    s.Description,
		OrganizationID: s.OrganizationID,
		ProjectID:      s.ProjectID,
		Environment:    s.Environment,
		CreatedBy:      s.CreatedBy,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
		Version:        s.Version,
	}
}
