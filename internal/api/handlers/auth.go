// filepath: internal/api/handlers/auth.go

package handlers

import (
	"encoding/json"
	"net/http"

	"secrets-manager/internal/auth"
)

// AuthHandler gère les routes liées à l'authentification
type AuthHandler struct {
	authService *auth.Service
}

// NewAuthHandler crée un nouveau gestionnaire d'authentification
func NewAuthHandler(authService *auth.Service) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// UserRegistration représente les données pour l'inscription
type UserRegistration struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// Login gère la connexion d'un utilisateur
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds auth.Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	// Authentifier l'utilisateur
	token, err := h.authService.Authenticate(&creds)
	if err != nil {
		if err == auth.ErrInvalidCredentials {
			http.Error(w, "Identifiants invalides", http.StatusUnauthorized)
		} else {
			http.Error(w, "Erreur d'authentification", http.StatusInternalServerError)
		}
		return
	}

	// Répondre avec le token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)
}

// Register gère l'inscription d'un utilisateur
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var reg UserRegistration
	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	// Valider les données
	if reg.Email == "" || reg.Password == "" {
		http.Error(w, "Email et mot de passe requis", http.StatusBadRequest)
		return
	}

	// Créer l'utilisateur
	creds := auth.Credentials{
		Email:    reg.Email,
		Password: reg.Password,
	}
	err := h.authService.RegisterUser(&creds, reg.FirstName, reg.LastName)
	if err != nil {
		if err == auth.ErrUserExists {
			http.Error(w, "L'utilisateur existe déjà", http.StatusConflict)
		} else {
			http.Error(w, "Erreur d'inscription", http.StatusInternalServerError)
		}
		return
	}

	// Répondre avec succès
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Utilisateur créé avec succès",
	})
}
