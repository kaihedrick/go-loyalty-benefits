package loyalty

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/kaihedrick/go-loyalty-benefits/internal/platform/auth"
	"github.com/kaihedrick/go-loyalty-benefits/internal/platform/config"
	"github.com/kaihedrick/go-loyalty-benefits/internal/platform/database"
	"github.com/sirupsen/logrus"
)

// Service represents the loyalty service
type Service struct {
	config     *config.Config
	logger     *logrus.Logger
	db         *database.PostgresDB
	jwtManager *auth.JWTManager
}

// User represents a user's loyalty profile
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Points    int       `json:"points"`
	Tier      string    `json:"tier"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Transaction represents a loyalty transaction
type Transaction struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Type        string    `json:"type"` // "earn" or "spend"
	Amount      int       `json:"amount"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// Reward represents an available reward
type Reward struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PointsCost  int    `json:"points_cost"`
	Category    string `json:"category"`
	IsActive    bool   `json:"is_active"`
}

// EarnRequest represents a points earning request
type EarnRequest struct {
	UserID      string `json:"user_id" validate:"required"`
	Amount      int    `json:"amount" validate:"required,min=1"`
	Description string `json:"description" validate:"required"`
}

// SpendRequest represents a points spending request
type SpendRequest struct {
	UserID      string `json:"user_id" validate:"required"`
	Amount      int    `json:"amount" validate:"required,min=1"`
	Description string `json:"description" validate:"required"`
}

// LoyaltyResponse represents a loyalty service response
type LoyaltyResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// NewService creates a new loyalty service
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

// Routes returns the loyalty service routes
func (s *Service) Routes(r *chi.Mux) {
	r.Route("/v1/loyalty", func(r chi.Router) {
		r.Post("/earn", s.AuthMiddleware(s.EarnPoints))
		r.Post("/spend", s.AuthMiddleware(s.SpendPoints))
		r.Get("/balance", s.AuthMiddleware(s.GetBalance))
		r.Get("/history", s.AuthMiddleware(s.GetHistory))
		r.Get("/rewards", s.GetRewards)
	})
}

// EarnPoints handles points earning
func (s *Service) EarnPoints(w http.ResponseWriter, r *http.Request) {
	var req EarnRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, LoyaltyResponse{Success: false, Message: "Invalid request body"})
		return
	}

	// Validate request
	if req.UserID == "" || req.Amount <= 0 || req.Description == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, LoyaltyResponse{Success: false, Message: "User ID, amount, and description are required"})
		return
	}

	// Get user from context (set by auth middleware)
	userID := r.Context().Value("user_id").(string)
	if userID != req.UserID {
		render.Status(r, http.StatusForbidden)
		render.JSON(w, r, LoyaltyResponse{Success: false, Message: "Can only earn points for your own account"})
		return
	}

	// Ensure user exists in loyalty_users (auto-create if needed)
	_, err := s.getUserByID(r.Context(), userID)
	if err != nil {
		s.logger.Errorf("Failed to get/create user: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, LoyaltyResponse{Success: false, Message: "Failed to get user info"})
		return
	}

	// Create transaction
	txID := uuid.New().String()
	now := time.Now()
	transaction := &Transaction{
		ID:          txID,
		UserID:      userID,
		Type:        "earn",
		Amount:      req.Amount,
		Description: req.Description,
		CreatedAt:   now,
	}

	if err := s.createTransaction(r.Context(), transaction); err != nil {
		s.logger.Errorf("Failed to create transaction: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, LoyaltyResponse{Success: false, Message: "Failed to process points earning"})
		return
	}

	// Update user points
	if err := s.updateUserPoints(r.Context(), userID, req.Amount); err != nil {
		s.logger.Errorf("Failed to update user points: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, LoyaltyResponse{Success: false, Message: "Failed to update user points"})
		return
	}

	// Get updated user info
	updatedUser, err := s.getUserByID(r.Context(), userID)
	if err != nil {
		s.logger.Errorf("Failed to get updated user: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, LoyaltyResponse{Success: false, Message: "Failed to get updated user info"})
		return
	}

	response := LoyaltyResponse{
		Success: true,
		Message: "Points earned successfully",
		Data: map[string]interface{}{
			"transaction": transaction,
			"user":        updatedUser,
		},
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, response)
}

// SpendPoints handles points spending
func (s *Service) SpendPoints(w http.ResponseWriter, r *http.Request) {
	var req SpendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, LoyaltyResponse{Success: false, Message: "Invalid request body"})
		return
	}

	// Validate request
	if req.UserID == "" || req.Amount <= 0 || req.Description == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, LoyaltyResponse{Success: false, Message: "User ID, amount, and description are required"})
		return
	}

	// Get user from context (set by auth middleware)
	userID := r.Context().Value("user_id").(string)
	if userID != req.UserID {
		render.Status(r, http.StatusForbidden)
		render.JSON(w, r, LoyaltyResponse{Success: false, Message: "Can only spend points from your own account"})
		return
	}

	// Check if user has enough points
	user, err := s.getUserByID(r.Context(), userID)
	if err != nil {
		s.logger.Errorf("Failed to get user: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, LoyaltyResponse{Success: false, Message: "Failed to get user info"})
		return
	}

	if user.Points < req.Amount {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, LoyaltyResponse{Success: false, Message: "Insufficient points"})
		return
	}

	// Create transaction
	txID := uuid.New().String()
	now := time.Now()
	transaction := &Transaction{
		ID:          txID,
		UserID:      userID,
		Type:        "spend",
		Amount:      req.Amount,
		Description: req.Description,
		CreatedAt:   now,
	}

	if err := s.createTransaction(r.Context(), transaction); err != nil {
		s.logger.Errorf("Failed to create transaction: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, LoyaltyResponse{Success: false, Message: "Failed to process points spending"})
		return
	}

	// Update user points (subtract)
	if err := s.updateUserPoints(r.Context(), userID, -req.Amount); err != nil {
		s.logger.Errorf("Failed to update user points: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, LoyaltyResponse{Success: false, Message: "Failed to update user points"})
		return
	}

	// Get updated user info
	updatedUser, err := s.getUserByID(r.Context(), userID)
	if err != nil {
		s.logger.Errorf("Failed to get updated user: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, LoyaltyResponse{Success: false, Message: "Failed to get updated user info"})
		return
	}

	response := LoyaltyResponse{
		Success: true,
		Message: "Points spent successfully",
		Data: map[string]interface{}{
			"transaction": transaction,
			"user":        updatedUser,
		},
	}

	render.JSON(w, r, response)
}

// GetBalance returns the current user's loyalty balance
func (s *Service) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	user, err := s.getUserByID(r.Context(), userID)
	if err != nil {
		s.logger.Errorf("Failed to get user balance: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, LoyaltyResponse{Success: false, Message: "Failed to get user balance"})
		return
	}

	response := LoyaltyResponse{
		Success: true,
		Message: "Balance retrieved successfully",
		Data:    user,
	}

	render.JSON(w, r, response)
}

// GetHistory returns the user's transaction history
func (s *Service) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	transactions, err := s.getUserTransactions(r.Context(), userID)
	if err != nil {
		s.logger.Errorf("Failed to get user history: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, LoyaltyResponse{Success: false, Message: "Failed to get transaction history"})
		return
	}

	response := LoyaltyResponse{
		Success: true,
		Message: "History retrieved successfully",
		Data:    transactions,
	}

	render.JSON(w, r, response)
}

// GetRewards returns available rewards
func (s *Service) GetRewards(w http.ResponseWriter, r *http.Request) {
	rewards, err := s.getActiveRewards(r.Context())
	if err != nil {
		s.logger.Errorf("Failed to get rewards: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, LoyaltyResponse{Success: false, Message: "Failed to get rewards"})
		return
	}

	response := LoyaltyResponse{
		Success: true,
		Message: "Rewards retrieved successfully",
		Data:    rewards,
	}

	render.JSON(w, r, response)
}

// AuthMiddleware validates JWT tokens
func (s *Service) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, LoyaltyResponse{Success: false, Message: "Authorization header required"})
			return
		}

		// Extract token from "Bearer <token>"
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, LoyaltyResponse{Success: false, Message: "Invalid authorization header format"})
			return
		}

		token := authHeader[7:]
		claims, err := s.jwtManager.ValidateToken(token)
		if err != nil {
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, LoyaltyResponse{Success: false, Message: "Invalid token"})
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
func (s *Service) createTransaction(ctx context.Context, tx *Transaction) error {
	query := `
		INSERT INTO loyalty_transactions (id, user_id, type, amount, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	err := s.db.Exec(ctx, query, tx.ID, tx.UserID, tx.Type, tx.Amount, tx.Description, tx.CreatedAt)
	return err
}

func (s *Service) updateUserPoints(ctx context.Context, userID string, pointsChange int) error {
	query := `
		UPDATE loyalty_users 
		SET points = points + $1, updated_at = $2
		WHERE id = $3
	`

	err := s.db.Exec(ctx, query, pointsChange, time.Now(), userID)
	return err
}

// createLoyaltyUser creates a new loyalty user record
func (s *Service) createLoyaltyUser(ctx context.Context, userID string, email string) error {
	query := `
		INSERT INTO loyalty_users (id, email, points, tier, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	now := time.Now()
	err := s.db.Exec(ctx, query, userID, email, 0, "Bronze", now, now)
	return err
}

// getUserByID gets a user from loyalty_users, auto-creating if they don't exist
func (s *Service) getUserByID(ctx context.Context, userID string) (*User, error) {
	query := `SELECT id, email, points, tier, created_at, updated_at FROM loyalty_users WHERE id = $1`

	var user User
	err := s.db.QueryRow(ctx, query, userID).Scan(
		&user.ID, &user.Email, &user.Points, &user.Tier, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		// User doesn't exist in loyalty_users, try to get their email from auth context
		userEmail, ok := ctx.Value("user_email").(string)
		if !ok {
			return nil, err
		}

		// Auto-create the loyalty user
		if err := s.createLoyaltyUser(ctx, userID, userEmail); err != nil {
			s.logger.Errorf("Failed to auto-create loyalty user: %v", err)
			return nil, err
		}

		// Now get the newly created user
		err = s.db.QueryRow(ctx, query, userID).Scan(
			&user.ID, &user.Email, &user.Points, &user.Tier, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		s.logger.Infof("Auto-created loyalty user: %s (%s)", userID, userEmail)
	}

	return &user, nil
}

func (s *Service) getUserTransactions(ctx context.Context, userID string) ([]*Transaction, error) {
	query := `SELECT id, user_id, type, amount, description, created_at FROM loyalty_transactions WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		var tx Transaction
		err := rows.Scan(&tx.ID, &tx.UserID, &tx.Type, &tx.Amount, &tx.Description, &tx.CreatedAt)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, &tx)
	}

	return transactions, nil
}

func (s *Service) getActiveRewards(ctx context.Context) ([]*Reward, error) {
	query := `SELECT id, name, description, points_cost, category, is_active FROM loyalty_rewards WHERE is_active = true ORDER BY points_cost ASC`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rewards []*Reward
	for rows.Next() {
		var reward Reward
		err := rows.Scan(&reward.ID, &reward.Name, &reward.Description, &reward.PointsCost, &reward.Category, &reward.IsActive)
		if err != nil {
			return nil, err
		}
		rewards = append(rewards, &reward)
	}

	return rewards, nil
}
