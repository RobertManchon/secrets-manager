// filepath: d:\go\src\secrets-manager\internal\storage\mysql\subscription_service.go

/*************************************************************************/
/*                                                                       */
/*   Ce fichier implémente le service de gestion des abonnements         */
/*   Il vérifie les limites des plans et gère les quotas d'utilisation   */
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

// ErrSubscriptionLimitReached indique que la limite d'un abonnement a été atteinte
var ErrSubscriptionLimitReached = errors.New("limite de secrets atteinte pour cet abonnement")

// SubscriptionService gère les abonnements et leurs limites
type SubscriptionService struct {
	db            *sql.DB
	secretsRepo   *SecretCountRepository
}

// NewSubscriptionService crée un nouveau service d'abonnement
func NewSubscriptionService(db *sql.DB) *SubscriptionService {
	return &SubscriptionService{
		db:          db,
		secretsRepo: NewSecretCountRepository(db),
	}
}

// GetActiveSubscription récupère l'abonnement actif pour une organisation
func (s *SubscriptionService) GetActiveSubscription(ctx context.Context, orgID string) (*models.Subscription, error) {
	query := `
		SELECT id, organization_id, plan_id, status, secrets_limit, 
			   start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE organization_id = ? 
		  AND status = 'active'
		  AND end_date > NOW()
		ORDER BY end_date DESC
		LIMIT 1
	`

	subscription := &models.Subscription{}
	err := s.db.QueryRowContext(ctx, query, orgID).Scan(
		&subscription.ID,
		&subscription.OrganizationID,
		&subscription.PlanID,
		&subscription.Status,
		&subscription.SecretsLimit,
		&subscription.StartDate,
		&subscription.EndDate,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Pas d'abonnement actif
		}
		return nil, err
	}

	return subscription, nil
}

// CreateSubscription crée un nouvel abonnement
func (s *SubscriptionService) CreateSubscription(ctx context.Context, subscription *models.Subscription) error {
	// Générer un ID si non fourni
	if subscription.ID == "" {
		subscription.ID = uuid.New().String()
	}
	
	// Vérifier si un abonnement actif existe déjà
	existingSub, err := s.GetActiveSubscription(ctx, subscription.OrganizationID)
	if err != nil {
		return err
	}
	
	// Si un abonnement actif existe, le désactiver
	if existingSub != nil {
		err = s.cancelSubscription(ctx, existingSub.ID)
		if err != nil {
			return err
		}
	}

	// Insérer le nouvel abonnement
	query := `
		INSERT INTO subscriptions (
			id, organization_id, plan_id, status, secrets_limit,
			start_date, end_date, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`

	_, err = s.db.ExecContext(
		ctx,
		query,
		subscription.ID,
		subscription.OrganizationID,
		subscription.PlanID,
		subscription.Status,
		subscription.SecretsLimit,
		subscription.StartDate,
		subscription.EndDate,
	)

	return err
}

// CancelSubscription annule un abonnement
func (s *SubscriptionService) cancelSubscription(ctx context.Context, subscriptionID string) error {
	query := `
		UPDATE subscriptions
		SET status = 'cancelled', updated_at = NOW()
		WHERE id = ?
	`

	_, err := s.db.ExecContext(ctx, query, subscriptionID)
	return err
}

// CanCreateSecret vérifie si l'organisation peut créer un nouveau secret
func (s *SubscriptionService) CanCreateSecret(ctx context.Context, orgID string) (bool, error) {
	// Obtenir le nombre actuel de secrets
	count, err := s.secretsRepo.GetSecretsCount(ctx, orgID)
	if err != nil {
		return false, err
	}

	// Obtenir la limite de secrets
	limit, err := s.secretsRepo.GetSecretsLimit(ctx, orgID)
	if err != nil {
		return false, err
	}

	// Vérifier si on peut créer un nouveau secret
	return count < limit, nil
}

// GetPlan récupère les détails d'un plan d'abonnement
func (s *SubscriptionService) GetPlan(ctx context.Context, planID string) (*models.Plan, error) {
	query := `
		SELECT id, name, description, price, billing_cycle, secrets_limit, 
		       created_at, updated_at
		FROM plans
		WHERE id = ?
	`

	plan := &models.Plan{}
	// features JSON parsing logic can be added here if needed in the future

	err := s.db.QueryRowContext(ctx, query, planID).Scan(
		&plan.ID,
		&plan.Name,
		&plan.Description,
		&plan.Price,
		&plan.BillingCycle,
		&plan.SecretsLimit,
		&plan.CreatedAt,
		&plan.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Ici, tu pourrais ajouter une logique pour parser les features JSON en []string
	// Pour l'instant, on laisse features vide

	return plan, nil
}

// ListAvailablePlans liste tous les plans disponibles
func (s *SubscriptionService) ListAvailablePlans(ctx context.Context) ([]*models.Plan, error) {
	query := `
		SELECT id, name, description, price, billing_cycle, secrets_limit, 
		       created_at, updated_at
		FROM plans
		ORDER BY price ASC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*models.Plan
	for rows.Next() {
		plan := &models.Plan{}
		// features JSON parsing logic can be added here if needed in the future

		err := rows.Scan(
			&plan.ID,
			&plan.Name,
			&plan.Description,
			&plan.Price,
			&plan.BillingCycle,
			&plan.SecretsLimit,
			&plan.CreatedAt,
			&plan.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Ici aussi, on pourrait parser les features JSON

		plans = append(plans, plan)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return plans, nil
}

// UpdateSubscriptionLimit met à jour la limite de secrets d'un abonnement
func (s *SubscriptionService) UpdateSubscriptionLimit(ctx context.Context, subscriptionID string, newLimit int) error {
	query := `
		UPDATE subscriptions
		SET secrets_limit = ?, updated_at = NOW()
		WHERE id = ?
	`

	_, err := s.db.ExecContext(ctx, query, newLimit, subscriptionID)
	return err
}

// UpgradeToPlan met à niveau l'abonnement d'une organisation vers un nouveau plan
func (s *SubscriptionService) UpgradeToPlan(ctx context.Context, orgID string, planID string, durationMonths int) error {
	// Récupérer les détails du nouveau plan
	newPlan, err := s.GetPlan(ctx, planID)
	if err != nil {
		return err
	}

	// Créer un nouvel abonnement
	startDate := time.Now()
	endDate := startDate.AddDate(0, durationMonths, 0)

	subscription := &models.Subscription{
		ID:             uuid.New().String(),
		OrganizationID: orgID,
		PlanID:         planID,
		Status:         "active",
		SecretsLimit:   newPlan.SecretsLimit,
		StartDate:      startDate,
		EndDate:        endDate,
	}

	return s.CreateSubscription(ctx, subscription)
}

// CheckSubscriptionStatus vérifie si une organisation a un abonnement actif
func (s *SubscriptionService) CheckSubscriptionStatus(ctx context.Context, orgID string) (bool, error) {
	subscription, err := s.GetActiveSubscription(ctx, orgID)
	if err != nil {
		return false, err
	}

	return subscription != nil, nil
}

// GetUsagePercentage calcule le pourcentage d'utilisation des secrets
func (s *SubscriptionService) GetUsagePercentage(ctx context.Context, orgID string) (float64, error) {
	count, err := s.secretsRepo.GetSecretsCount(ctx, orgID)
	if err != nil {
		return 0, err
	}

	limit, err := s.secretsRepo.GetSecretsLimit(ctx, orgID)
	if err != nil {
		return 0, err
	}

	if limit == 0 {
		return 100, nil // Éviter la division par zéro
	}

	return float64(count) * 100 / float64(limit), nil
}
