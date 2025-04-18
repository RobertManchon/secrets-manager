// filepath: internal/auth/service.go

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
)

// Erreurs du service d'authentification
var (
	ErrInvalidCredentials = errors.New("identifiants invalides")
	ErrUserExists         = errors.New("l'utilisateur existe déjà")
	ErrInvalidToken       = errors.New("token invalide")
	ErrUserNotFound       = errors.New("utilisateur non trouvé")
	ErrTokenExpired       = errors.New("token expiré")
)

// Service fournit des fonctionnalités d'authentification
type Service struct {
	db          *sql.DB
	jwtSecret   string
	jwtExpiry   time.Duration
	refreshTime time.Duration
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

// NewService crée un nouveau service d'authentification
func NewService(db *sql.DB, jwtSecret string, jwtExpiry, refreshTime time.Duration) *Service {
	return &Service{
		db:          db,
		jwtSecret:   jwtSecret,
		jwtExpiry:   jwtExpiry,
		refreshTime: refreshTime,
	}
}

// Authenticate vérifie les identifiants d'un utilisateur et génère un token JWT
func (s *Service) Authenticate(ctx context.Context, creds *Credentials) (*TokenResponse, *UserDetails, error) {
	var hashedPassword, userID, firstName, lastName, role string

	query := "SELECT id, hashed_password, first_name, last_name, role FROM users WHERE email = ?"
	err := s.db.QueryRowContext(ctx, query, creds.Email).Scan(&userID, &hashedPassword, &firstName, &lastName, &role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	// Vérifier le mot de passe
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(creds.Password))
	if err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	// Générer le token JWT et le token de rafraîchissement
	token, refreshToken, expiresAt, err := s.generateTokenPair(userID)
	if err != nil {
		return nil, nil, err
	}

	return &TokenResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		UserID:       userID,
	}, &UserDetails{
		ID:        userID,
		Email:     creds.Email,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
	}, nil
}

// RegisterUser enregistre un nouvel utilisateur
func (s *Service) RegisterUser(ctx context.Context, creds *Credentials, firstName, lastName string) (*UserDetails, error) {
	// Vérifier si l'utilisateur existe déjà
	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", creds.Email).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}

	// Hasher le mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Insérer le nouvel utilisateur
	userID := uuid.New().String()
	_, err = s.db.ExecContext(ctx,
		"INSERT INTO users (id, email, hashed_password, first_name, last_name, role, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())",
		userID, creds.Email, hashedPassword, firstName, lastName, "user",
	)
	if err != nil {
		return nil, err
	}

	return &UserDetails{
		ID:        userID,
		Email:     creds.Email,
		FirstName: firstName,
		LastName:  lastName,
		Role:      "user",
	}, nil
}

// VerifyToken vérifie la validité d'un token JWT
func (s *Service) VerifyToken(tokenString string) (string, error) {
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

// RefreshToken rafraîchit un token JWT expiré
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	claims, err := s.parseToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Vérifier que c'est un token de rafraîchissement
	if tokenType, ok := claims["type"].(string); !ok || tokenType != "refresh" {
		return nil, ErrInvalidToken
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Générer de nouveaux tokens
	token, newRefreshToken, expiresAt, err := s.generateTokenPair(userID)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		Token:        token,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
		UserID:       userID,
	}, nil
}

// generateToken génère un nouveau token JWT
func (s *Service) generateToken(userID, tokenType string, expiry time.Duration) (string, time.Time, error) {
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
func (s *Service) generateTokenPair(userID string) (string, string, time.Time, error) {
	accessToken, expiresAt, err := s.generateToken(userID, "access", s.jwtExpiry)
	if err != nil {
		return "", "", time.Time{}, err
	}

	refreshToken, _, err := s.generateToken(userID, "refresh", s.refreshTime)
	if err != nil {
		return "", "", time.Time{}, err
	}

	return accessToken, refreshToken, expiresAt, nil
}

// parseToken parse un token JWT et vérifie sa validité
func (s *Service) parseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
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
