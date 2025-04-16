// filepath: internal/auth/service.go

package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("identifiants invalides")
	ErrUserExists         = errors.New("l'utilisateur existe déjà")
	ErrInvalidToken       = errors.New("token invalide")
)

// Service fournit des fonctionnalités d'authentification
type Service struct {
	db        *sql.DB
	jwtSecret string
	jwtExpiry time.Duration
}

// NewService crée un nouveau service d'authentification
func NewService(db *sql.DB, jwtSecret string, jwtExpiry time.Duration) *Service {
	return &Service{
		db:        db,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

// Credentials représente les identifiants d'un utilisateur
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TokenResponse représente la réponse avec le token JWT
type TokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Authenticate vérifie les identifiants d'un utilisateur et génère un token JWT
func (s *Service) Authenticate(creds *Credentials) (*TokenResponse, error) {
	// Dans une implémentation réelle, vous vérifieriez les identifiants avec la base de données
	var hashedPassword string
	var userID string

	query := "SELECT id, hashed_password FROM users WHERE email = ?"
	err := s.db.QueryRow(query, creds.Email).Scan(&userID, &hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Vérifier le mot de passe
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(creds.Password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Générer le token JWT
	token, expiresAt, err := s.generateToken(userID)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

// RegisterUser enregistre un nouvel utilisateur
func (s *Service) RegisterUser(creds *Credentials, firstName, lastName string) error {
	// Vérifier si l'utilisateur existe déjà
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", creds.Email).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return ErrUserExists
	}

	// Hasher le mot de passe
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Insérer le nouvel utilisateur
	_, err = s.db.Exec(
		"INSERT INTO users (email, hashed_password, first_name, last_name, role, created_at, updated_at) VALUES (?, ?, ?, ?, ?, NOW(), NOW())",
		creds.Email, hashedPassword, firstName, lastName, "user",
	)
	return err
}

// VerifyToken vérifie la validité d'un token JWT
func (s *Service) VerifyToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Vérifier l'algorithme de signature
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("méthode de signature inattendue: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", ErrInvalidToken
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return "", ErrInvalidToken
	}

	return userID, nil
}

// generateToken génère un nouveau token JWT
func (s *Service) generateToken(userID string) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.jwtExpiry)

	claims := jwt.MapClaims{
		"sub": userID,
		"exp": expiresAt.Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return signedToken, expiresAt, nil
}
