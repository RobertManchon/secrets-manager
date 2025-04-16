// filepath: internal/vault/client.go

package vault

import (
	"context"
	"fmt"

	vault "github.com/hashicorp/vault/api"
)

// Client encapsule l'interaction avec Vault
type Client struct {
	client *vault.Client
	config *Config
}

// Config contient la configuration du client Vault
type Config struct {
	Address   string
	Token     string
	Namespace string
	// Autres paramètres de configuration
}

// NewClient crée un nouveau client Vault
func NewClient(config *Config) (*Client, error) {
	cfg := vault.DefaultConfig()
	cfg.Address = config.Address

	client, err := vault.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("impossible de créer le client Vault: %w", err)
	}

	client.SetToken(config.Token)
	if config.Namespace != "" {
		client.SetNamespace(config.Namespace)
	}

	return &Client{
		client: client,
		config: config,
	}, nil
}

// GetSecret récupère un secret de Vault
func (c *Client) GetSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	secret, err := c.client.KVv2("secret").Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("impossible de récupérer le secret: %w", err)
	}

	if secret == nil {
		return nil, fmt.Errorf("secret non trouvé: %s", path)
	}

	return secret.Data, nil
}

// WriteSecret écrit un secret dans Vault
func (c *Client) WriteSecret(ctx context.Context, path string, data map[string]interface{}) error {
	_, err := c.client.KVv2("secret").Put(ctx, path, data)
	if err != nil {
		return fmt.Errorf("impossible d'écrire le secret: %w", err)
	}

	return nil
}

// DeleteSecret supprime un secret de Vault
func (c *Client) DeleteSecret(ctx context.Context, path string) error {
	err := c.client.KVv2("secret").Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("impossible de supprimer le secret: %w", err)
	}

	return nil
}

// ListSecrets liste les secrets d'un chemin
// Note: Cette méthode utilise maintenant la méthode List directement du client Vault
func (c *Client) ListSecrets(ctx context.Context, path string) ([]string, error) {
	// Construire le chemin complet pour le stockage KV v2
	fullPath := fmt.Sprintf("secret/metadata/%s", path)

	// Appeler l'API List directement
	secret, err := c.client.Logical().List(fullPath)
	if err != nil {
		return nil, fmt.Errorf("impossible de lister les secrets: %w", err)
	}

	if secret == nil || secret.Data == nil {
		return []string{}, nil
	}

	// Extraire les clés
	keysInterface, ok := secret.Data["keys"]
	if !ok {
		return []string{}, nil
	}

	keysSlice, ok := keysInterface.([]interface{})
	if !ok {
		return nil, fmt.Errorf("format inattendu pour les clés")
	}

	// Convertir en slice de strings
	result := make([]string, 0, len(keysSlice))
	for _, key := range keysSlice {
		if s, ok := key.(string); ok {
			result = append(result, s)
		}
	}

	return result, nil
}
