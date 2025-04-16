// filepath: internal/storage/mysql/connection.go
package storage

import (
	"database/sql"
	"fmt"
	"time"

	"secrets-manager/internal/config"

	_ "github.com/go-sql-driver/mysql"
)

// NewConnection établit une nouvelle connexion à la base de données MySQL
func NewConnection(cfg config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("erreur d'ouverture de la connexion: %w", err)
	}

	// Configurer le pool de connexions
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Vérifier la connexion
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("erreur de ping à la base de données: %w", err)
	}

	return db, nil
}
