// filepath: internal/config/config.go

package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config contient toutes les configurations de l'application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Vault    VaultConfig
	JWT      JWTConfig
}

// ServerConfig contient la configuration du serveur HTTP
type ServerConfig struct {
	Address string
	Port    int
}

// DatabaseConfig contient la configuration de la base de données
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// VaultConfig contient la configuration de Vault
type VaultConfig struct {
	Address string
	Token   string
}

// JWTConfig contient la configuration JWT
type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

// Load charge la configuration depuis les variables d'environnement
func Load() (*Config, error) {
	// Charger le fichier .env s'il existe
	_ = godotenv.Load()

	config := &Config{}

	// Configuration du serveur
	config.Server.Address = getEnv("SERVER_ADDRESS", "0.0.0.0")
	port, err := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("SERVER_PORT invalide: %w", err)
	}
	config.Server.Port = port

	// Configuration de la base de données
	config.Database.Host = getEnv("DB_HOST", "localhost")
	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "3306"))
	if err != nil {
		return nil, fmt.Errorf("DB_PORT invalide: %w", err)
	}
	config.Database.Port = dbPort
	config.Database.User = getEnv("DB_USER", "root")
	config.Database.Password = getEnv("DB_PASSWORD", "")
	config.Database.DBName = getEnv("DB_NAME", "secrets_manager")

	// Configuration de Vault
	config.Vault.Address = getEnv("VAULT_ADDR", "http://localhost:8200")
	config.Vault.Token = getEnv("VAULT_TOKEN", "")

	// Configuration JWT
	config.JWT.Secret = getEnv("JWT_SECRET", "votre_secret_jwt_très_sécurisé")
	jwtExp, err := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "24"))
	if err != nil {
		return nil, fmt.Errorf("JWT_EXPIRATION_HOURS invalide: %w", err)
	}
	config.JWT.Expiration = time.Duration(jwtExp) * time.Hour

	return config, nil
}

// getEnv récupère une variable d'environnement ou renvoie une valeur par défaut
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
