// filepath: d:\go\src\secrets-manager\internal\api\handlers\secrets.go
// filepath: d:\go\src\secrets-manager\internal\api\handlers\secrets.go
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"secrets-manager/internal/models"
	"secrets-manager/internal/vault"
)

// SecretsHandler gère les routes liées aux secrets
type SecretsHandler struct {
	vaultService *vault.Service
}

// NewSecretsHandler crée un nouveau gestionnaire de secrets
func NewSecretsHandler(vaultService *vault.Service) *SecretsHandler {
	return &SecretsHandler{
		vaultService: vaultService,
	}
}

// GetSecret récupère un secret
func (h *SecretsHandler) GetSecret(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgID := vars["orgID"]
	projectID := vars["projectID"]
	env := vars["env"]
	name := vars["name"]

	// Extraire l'ID utilisateur depuis le contexte (mis par middleware auth)
	//userID := r.Context().Value("userID").(string)

	// Vérifier si l'utilisateur a accès à ce secret
	// TODO: implémenter la vérification des permissions

	secret, err := h.vaultService.GetSecret(r.Context(), orgID, projectID, env, name)
	if err != nil {
		http.Error(w, "Secret non trouvé", http.StatusNotFound)
		return
	}

	// Audit de l'accès au secret
	// TODO: journaliser l'accès au secret

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(secret); err != nil {
		http.Error(w, "Erreur lors de l'encodage du secret", http.StatusInternalServerError)
	}
}

// CreateSecret crée un nouveau secret
func (h *SecretsHandler) CreateSecret(w http.ResponseWriter, r *http.Request) {
	var secret models.Secret
	if err := json.NewDecoder(r.Body).Decode(&secret); err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}

	// Extraire l'ID utilisateur depuis le contexte (mis par middleware auth)
	userID := r.Context().Value("userID").(string)
	secret.CreatedBy = userID

	// Vérifier si l'utilisateur a le droit de créer un secret dans ce projet
	// TODO: implémenter la vérification des permissions

	if err := h.vaultService.StoreSecret(r.Context(), &secret); err != nil {
		http.Error(w, "Impossible de créer le secret", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// ListSecrets liste tous les secrets d'un projet
func (h *SecretsHandler) ListSecrets(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgID := vars["orgID"]
	projectID := vars["projectID"]
	env := vars["env"]

	// TODO: vérifier les permissions

	secrets, err := h.vaultService.ListProjectSecrets(r.Context(), orgID, projectID, env)
	if err != nil {
		http.Error(w, "Impossible de lister les secrets", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(secrets); err != nil {
		http.Error(w, "Erreur lors de l'encodage des secrets", http.StatusInternalServerError)
	}
}

// DeleteSecret supprime un secret
func (h *SecretsHandler) DeleteSecret(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgID := vars["orgID"]
	projectID := vars["projectID"]
	env := vars["env"]
	name := vars["name"]

	// TODO: vérifier les permissions

	if err := h.vaultService.DeleteSecret(r.Context(), orgID, projectID, env, name); err != nil {
		http.Error(w, "Impossible de supprimer le secret", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
