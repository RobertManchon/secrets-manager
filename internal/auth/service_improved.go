// filepath: d:\go\src\secrets-manager\internal\auth\service_improved.go

/*************************************************************************/
/*                                                                       */
/*   Ce fichier implémente le service d'authentification amélioré        */
/*   Il utilise le repository des utilisateurs pour la gestion des       */
/*   identités et fournit des fonctionnalités JWT complètes              */
/*                                                                       */
/*************************************************************************/

package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"secrets-manager/internal/models"
	mysqlstorage "secrets-manager/internal/storage/mysql"
)

// Erreurs du service d'authentification
var (
	ErrInvalidCredentials = errors.New("identifiants invalides")
	ErrUserExists         = errors.New("l'utilisateur existe déjà")
	ErrInvalidToken       = errors.New("token invalide")
	ErrUserNotFound       = errors.New("utilisateur non trouvé")
	ErrTokenExpired       = errors.New("token expiré")
)

// ServiceImproved fournit des fonctionnalités d'authentification avancées
type ServiceImproved struct {
	usersRepo   *mysqlstorage.UsersRepository
	jwtSecret   string
	jwtExpiry   time.Duration
	refreshTime time.Duration
}

// NewServiceImproved crée un nouveau service d'authentification amélioré
func NewServiceImproved(
	db *sql.DB,
	jwtSecret string,
	jwtExpiry time.Duration,
	refreshTime time.Duration,
) *ServiceImproved {
	return &ServiceImproved{
		usersRepo:   mysqlstorage.NewUsersRepository(db),
		jwtSecret:   jwtSecret,
		jwtExpiry:   jwtExpiry,
		refreshTime: refreshTime,
	}
}

// Credentials représente les identifiants d'un utilisateur
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TokenResponse représente la réponse avec le token JWT
type TokenResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	UserID       string    `json:"user_id"`
}

// UserDetails représente les informations renvoyées lors de l'authentification
type UserDetails struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
}

// RegisterUser enregistre un nouvel utilisateur
func (s *ServiceImproved) RegisterUser(ctx context.Context, creds *Credentials, firstName, lastName string) (*UserDetails, error) {
	// Vérifier si l'utilisateur existe déjà
	existingUser, err := s.usersRepo.GetUserByEmail(ctx, creds.Email)
	if err != nil && !errors.Is(err, mysqlstorage.ErrUserNotFound) {
		return nil, err
	}

	if existingUser != nil {
		return nil, ErrUserExists
	}

	// Hasher le mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Créer le nouvel utilisateur
	newUser := &models.User{
		ID:             uuid.New().String(),
		Email:          creds.Email,
		HashedPassword: string(hashedPassword),
		FirstName:      firstName,
		LastName:       lastName,
		Role:           "user",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Enregistrer l'utilisateur
	err = s.usersRepo.CreateUser(ctx, newUser)
	if err != nil {
		return nil, err
	}

	// Renvoyer les détails de l'utilisateur créé
	return &UserDetails{
		ID:        newUser.ID,
		Email:     newUser.Email,
		FirstName: newUser.FirstName,
		LastName:  newUser.LastName,
		Role:      newUser.Role,
	}, nil
}

// Authenticate vérifie les identifiants d'un utilisateur et génère un token JWT
func (s *ServiceImproved) Authenticate(ctx context.Context, creds *Credentials) (*TokenResponse, *UserDetails, error) {
	// Récupérer l'utilisateur par email
	user, err := s.usersRepo.GetUserByEmail(ctx, creds.Email)
	if err != nil {
		if errors.Is(err, mysqlstorage.ErrUserNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	// Vérifier le mot de passe
	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(creds.Password))
	if err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	// Générer le token JWT et le token de rafraîchissement
	token, refreshToken, expiresAt, err := s.generateTokenPair(user.ID)
	if err != nil {
		return nil, nil, err
	}

	// Renvoyer le token et les détails utilisateur
	return &TokenResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		UserID:       user.ID,
	}, &UserDetails{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
	}, nil
}

// RefreshToken rafraîchit un token JWT expiré
func (s *ServiceImproved) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	// Vérifier le token de rafraîchissement
	claims, err := s.parseToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Vérifier que c'est bien un token de rafraîchissement
	if tokenType, ok := claims["type"].(string); !ok || tokenType != "refresh" {
		return nil, ErrInvalidToken
	}

	// Extraire l'ID utilisateur
	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Vérifier que l'utilisateur existe toujours
	user, err := s.usersRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, mysqlstorage.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Générer de nouveaux tokens
	token, newRefreshToken, expiresAt, err := s.generateTokenPair(user.ID)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		Token:        token,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
		UserID:       user.ID,
	}, nil
}

// VerifyToken vérifie la validité d'un token JWT
func (s *ServiceImproved) VerifyToken(tokenString string) (string, error) {
	claims, err := s.parseToken(tokenString)
	if err != nil {
		return "", err
	}

	// Vérifier que c'est un token d'accès
	if tokenType, ok := claims["type"].(string); !ok || tokenType != "access" {
		return "", ErrInvalidToken
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return "", ErrInvalidToken
	}

	return userID, nil
}

// GetUserByID récupère un utilisateur par son ID
func (s *ServiceImproved) GetUserByID(ctx context.Context, userID string) (*UserDetails, error) {
	user, err := s.usersRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, mysqlstorage.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &UserDetails{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
	}, nil
}

// UpdateUserPassword met à jour le mot de passe d'un utilisateur
func (s *ServiceImproved) UpdateUserPassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	// Récupérer l'utilisateur
	user, err := s.usersRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, mysqlstorage.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	// Vérifier le mot de passe actuel
	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(currentPassword))
	if err != nil {
		return ErrInvalidCredentials
	}

	// Hasher le nouveau mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Mettre à jour le mot de passe
	return s.usersRepo.UpdatePassword(ctx, userID, string(hashedPassword))
}

// UpdateUserProfile met à jour le profil d'un utilisateur
func (s *ServiceImproved) UpdateUserProfile(ctx context.Context, userID, firstName, lastName string) (*UserDetails, error) {
	// Récupérer l'utilisateur
	user, err := s.usersRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, mysqlstorage.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Mettre à jour les informations
	user.FirstName = firstName
	user.LastName = lastName
	user.UpdatedAt = time.Now()

	// Sauvegarder les modifications
	err = s.usersRepo.UpdateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	// Renvoyer les détails mis à jour
	return &UserDetails{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
	}, nil
}

// parseToken parse un token JWT et vérifie sa validité
func (s *ServiceImproved) parseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Vérifier l'algorithme de signature
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("méthode de signature inattendue: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// generateToken génère un nouveau token JWT
func (s *ServiceImproved) generateToken(userID, tokenType string, expiry time.Duration) (string, time.Time, error) {
	expiresAt := time.Now().Add(expiry)

	claims := jwt.MapClaims{
		"sub":  userID,
		"type": tokenType,
		"exp":  expiresAt.Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return signedToken, expiresAt, nil
}

// generateTokenPair génère un token d'accès et un token de rafraîchissement
func (s *ServiceImproved) generateTokenPair(userID string) (string, string, time.Time, error) {
	// Générer le token d'accès
	accessToken, expiresAt, err := s.generateToken(userID, "access", s.jwtExpiry)
	if err != nil {
		return "", "", time.Time{}, err
	}

	// Générer le token de rafraîchissement (validité plus longue)
	refreshToken, _, err := s.generateToken(userID, "refresh", s.refreshTime)
	if err != nil {
		return "", "", time.Time{}, err
	}

	return accessToken, refreshToken, expiresAt, nil
}
