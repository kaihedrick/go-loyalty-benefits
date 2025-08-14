package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kaihedrick/go-loyalty-benefits/internal/platform/auth"
	"github.com/kaihedrick/go-loyalty-benefits/internal/platform/config"
	"github.com/kaihedrick/go-loyalty-benefits/internal/platform/database"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// Service represents the authentication service
type Service struct {
	config     *config.Config
	logger     *logrus.Logger
	db         *database.PostgresDB
	jwtManager *auth.JWTManager
}

// User represents a user in the system
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	AccessToken string `json:"access_token"`
	User       *User  `json:"user"`
}

// NewService creates a new authentication service
func NewService(cfg *config.Config, logger *logrus.Logger) *Service {
	// Initialize JWT manager
	jwtConfig := &auth.JWTConfig{
		Secret:     cfg.Security.JWT.Secret,
		Issuer:     cfg.Security.JWT.Issuer,
		Audience:   cfg.Security.JWT.Audience,
		Expiration: cfg.Security.JWT.Expiration,
	}
	jwtManager := auth.NewJWTManager(jwtConfig)

	return &Service{
		config:     cfg,
		logger:     logger,
		jwtManager: jwtManager,
	}
}

// SetDatabase sets the database connection
func (s *Service) SetDatabase(db *database.PostgresDB) {
	s.db = db
}

	// Routes returns the authentication service routes
func (s *Service) Routes(r chi.Router) {
	r.Route("/v1/auth", func(r chi.Router) {
		r.Post("/register", s.Register)
		r.Post("/login", s.Login)
		r.Get("/me", s.AuthMiddleware(s.GetProfile))
	})
}

// Register handles user registration
func (s *Service) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate request
	if req.Email == "" || req.Password == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Email and password are required"})
		return
	}

	// Check if user already exists
	existingUser, err := s.getUserByEmail(r.Context(), req.Email)
	if err != nil && err != sql.ErrNoRows {
		s.logger.Errorf("Failed to check existing user: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Internal server error"})
		return
	}

	if existingUser != nil {
		render.Status(r, http.StatusConflict)
		render.JSON(w, r, map[string]string{"error": "User already exists"})
		return
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Errorf("Failed to hash password: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Internal server error"})
		return
	}

	// Create user
	userID := uuid.New().String()
	now := time.Now()
	user := &User{
		ID:           userID,
		Email:        req.Email,
		PasswordHash: string(passwordHash),
		Role:         "user",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.createUser(r.Context(), user); err != nil {
		s.logger.Errorf("Failed to create user: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Internal server error"})
		return
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		s.logger.Errorf("Failed to generate token: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Internal server error"})
		return
	}

	response := &AuthResponse{
		AccessToken: token,
		User:        user,
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, response)
}

// Login handles user login
func (s *Service) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate request
	if req.Email == "" || req.Password == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Email and password are required"})
		return
	}

	// Get user by email
	user, err := s.getUserByEmail(r.Context(), req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, map[string]string{"error": "Invalid credentials"})
			return
		}
		s.logger.Errorf("Failed to get user: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Internal server error"})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, map[string]string{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		s.logger.Errorf("Failed to generate token: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Internal server error"})
		return
	}

	response := &AuthResponse{
		AccessToken: token,
		User:        user,
	}

	render.JSON(w, r, response)
}

// GetProfile returns the current user's profile
func (s *Service) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	
	user, err := s.getUserByID(r.Context(), userID)
	if err != nil {
		s.logger.Errorf("Failed to get user profile: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Internal server error"})
		return
	}

	render.JSON(w, r, user)
}

// AuthMiddleware validates JWT tokens
func (s *Service) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, map[string]string{"error": "Authorization header required"})
			return
		}

		// Extract token from "Bearer <token>"
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, map[string]string{"error": "Invalid authorization header format"})
			return
		}

		token := authHeader[7:]
		claims, err := s.jwtManager.ValidateToken(token)
		if err != nil {
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, map[string]string{"error": "Invalid token"})
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "user_email", claims.Email)
		ctx = context.WithValue(ctx, "user_role", claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// Database helper methods
func (s *Service) createUser(ctx context.Context, user *User) error {
	query := `
		INSERT INTO users (id, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	
	_, err := s.db.Exec(ctx, query, user.ID, user.Email, user.PasswordHash, user.Role, user.CreatedAt, user.UpdatedAt)
	return err
}

func (s *Service) getUserByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, email, password_hash, role, created_at, updated_at FROM users WHERE email = $1`
	
	var user User
	err := s.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}

func (s *Service) getUserByID(ctx context.Context, userID string) (*User, error) {
	query := `SELECT id, email, password_hash, role, created_at, updated_at FROM users WHERE id = $1`
	
	var user User
	err := s.db.QueryRow(ctx, query, userID).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}
