// filepath: d:\go\src\secrets-manager\internal\models\models.go

package models

import (
	"time"
)

// User représente un utilisateur du système
type User struct {
	ID             string    `json:"id" db:"id"`
	Email          string    `json:"email" db:"email"`
	HashedPassword string    `json:"-" db:"hashed_password"`
	FirstName      string    `json:"first_name" db:"first_name"`
	LastName       string    `json:"last_name" db:"last_name"`
	Role           string    `json:"role" db:"role"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Organization représente une organisation utilisatrice du service
type Organization struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	PlanID      string    `json:"plan_id" db:"plan_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	OwnerID     string    `json:"owner_id" db:"owner_id"`
}

// Project représente un projet contenant des secrets
type Project struct {
	ID             string    `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	Description    string    `json:"description" db:"description"`
	OrganizationID string    `json:"organization_id" db:"organization_id"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy      string    `json:"created_by" db:"created_by"`
}

// Environment représente un environnement (dev, staging, prod, etc.)
type Environment struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"` // dev, staging, prod, etc.
	Description string    `json:"description" db:"description"`
	ProjectID   string    `json:"project_id" db:"project_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Secret représente un secret stocké dans le système
type Secret struct {
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

// Subscription représente un abonnement au service
type Subscription struct {
	ID             string    `json:"id" db:"id"`
	OrganizationID string    `json:"organization_id" db:"organization_id"`
	PlanID         string    `json:"plan_id" db:"plan_id"`
	Status         string    `json:"status" db:"status"` // active, cancelled, trial, etc.
	SecretsLimit   int       `json:"secrets_limit" db:"secrets_limit"`
	StartDate      time.Time `json:"start_date" db:"start_date"`
	EndDate        time.Time `json:"end_date" db:"end_date"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Plan représente un plan d'abonnement
type Plan struct {
	ID           string    `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"` // Micro, Startup, Business, Enterprise
	Description  string    `json:"description" db:"description"`
	Price        float64   `json:"price" db:"price"`
	BillingCycle string    `json:"billing_cycle" db:"billing_cycle"` // monthly, yearly
	SecretsLimit int       `json:"secrets_limit" db:"secrets_limit"`
	Features     []string  `json:"features" db:"features"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// UserOrganization représente la relation entre un utilisateur et une organisation
type UserOrganization struct {
	UserID         string    `json:"user_id" db:"user_id"`
	OrganizationID string    `json:"organization_id" db:"organization_id"`
	Role           string    `json:"role" db:"role"` // admin, member, viewer
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// AuditLog représente une entrée du journal d'audit
type AuditLog struct {
	ID             string    `json:"id" db:"id"`
	UserID         string    `json:"user_id" db:"user_id"`
	OrganizationID string    `json:"organization_id" db:"organization_id"`
	Action         string    `json:"action" db:"action"`               // read, create, update, delete
	ResourceType   string    `json:"resource_type" db:"resource_type"` // secret, project, user, etc.
	ResourceID     string    `json:"resource_id" db:"resource_id"`
	Timestamp      time.Time `json:"timestamp" db:"timestamp"`
	IPAddress      string    `json:"ip_address" db:"ip_address"`
	UserAgent      string    `json:"user_agent" db:"user_agent"`
}
