// filepath: d:\go\src\secrets-manager\cmd\api\main.go

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"secrets-manager/internal/api"
	"secrets-manager/internal/auth"
	"secrets-manager/internal/config"
	mysqldb "secrets-manager/internal/storage/mysql"
	"secrets-manager/internal/vault"
)

func main() {
	// Charger la configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Erreur de chargement de la configuration: %v", err)
	}

	// Initialiser la base de données
	db, err := mysqldb.NewConnection(cfg.Database)
	if err != nil {
		log.Fatalf("Erreur de connexion à la base de données: %v", err)
	}
	defer db.Close()

	// Initialiser le client Vault
	vaultClient, err := vault.NewClient(&vault.Config{
		Address: cfg.Vault.Address,
		Token:   cfg.Vault.Token,
	})
	if err != nil {
		log.Fatalf("Erreur de connexion à Vault: %v", err)
	}

	// Initialiser les services
	vaultService := vault.NewService(vaultClient)
	authService := auth.NewService(db, cfg.JWT.Secret, cfg.JWT.Expiration)

	// Configurer le routeur
	router := mux.NewRouter()
	api.ConfigureRoutes(router, vaultService, authService)

	// Configurer le serveur HTTP
	srv := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Démarrer le serveur dans une goroutine
	go func() {
		log.Printf("Serveur démarré sur %s", cfg.Server.Address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erreur de démarrage du serveur: %v", err)
		}
	}()

	// Attendre le signal d'arrêt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// Arrêt gracieux
	log.Println("Arrêt du serveur...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Erreur lors de l'arrêt du serveur: %v", err)
	}

	log.Println("Serveur arrêté")
}
