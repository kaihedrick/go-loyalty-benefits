package loyalty

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/kaihedrick/go-loyalty-benefits/internal/platform/config"
	"github.com/kaihedrick/go-loyalty-benefits/internal/platform/database"
	"github.com/kaihedrick/go-loyalty-benefits/internal/platform/messaging"
	"github.com/sirupsen/logrus"
)

// Service represents the loyalty service
type Service struct {
	config     *config.Config
	logger     *logrus.Logger
	db         *database.PostgresDB
	kafka      *messaging.KafkaProducer
}

// Transaction represents a loyalty transaction
type Transaction struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Amount      float64   `json:"amount"`
	MCC         string    `json:"mcc"`           // Merchant Category Code
	MerchantID  string    `json:"merchant_id"`
	Points      int       `json:"points"`
	Multiplier  float64   `json:"multiplier"`
	CreatedAt   time.Time `json:"created_at"`
}

// Balance represents a user's loyalty balance
type Balance struct {
	UserID         string `json:"user_id"`
	AvailablePoints int64  `json:"available_points"`
	LifetimePoints  int64  `json:"lifetime_points"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TransactionRequest represents a transaction request
type TransactionRequest struct {
	Amount     float64 `json:"amount" validate:"required,gt=0"`
	MCC        string  `json:"mcc" validate:"required"`
	MerchantID string  `json:"merchant_id" validate:"required"`
}

// TransactionResponse represents a transaction response
type TransactionResponse struct {
	TransactionID string  `json:"txn_id"`
	Points        int     `json:"points"`
	Multiplier    float64 `json:"multiplier"`
	Message       string  `json:"message"`
}

// PointsEarnedEvent represents the points earned event
type PointsEarnedEvent struct {
	EventID    string    `json:"event_id"`
	UserID     string    `json:"user_id"`
	TxnID      string    `json:"txn_id"`
	Points     int       `json:"points"`
	Multiplier float64   `json:"multiplier"`
	MCC        string    `json:"mcc"`
	Timestamp  time.Time `json:"ts"`
}

// NewService creates a new loyalty service
func NewService(cfg *config.Config, logger *logrus.Logger) *Service {
	// Initialize Kafka producer
	kafkaConfig := &messaging.KafkaConfig{
		Brokers:  cfg.Kafka.Brokers,
		ClientID: cfg.Kafka.ClientID,
	}
	kafkaProducer := messaging.NewKafkaProducer(kafkaConfig, logger)

	return &Service{
		config: cfg,
		logger: logger,
		kafka:  kafkaProducer,
	}
}

// SetDatabase sets the database connection
func (s *Service) SetDatabase(db *database.PostgresDB) {
	s.db = db
}

// Routes returns the loyalty service routes
func (s *Service) Routes(r chi.Router) {
	r.Route("/v1", func(r chi.Router) {
		r.Post("/transactions", s.AuthMiddleware(s.CreateTransaction))
		r.Get("/balance", s.AuthMiddleware(s.GetBalance))
		r.Get("/transactions", s.AuthMiddleware(s.GetTransactions))
	})
}

// AuthMiddleware is a placeholder for JWT authentication
func (s *Service) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement JWT validation
		// For now, just extract user ID from header
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			render.Status(r, http.StatusUnauthorized)
			render.JSON(w, r, map[string]string{"error": "User ID required"})
			return
		}
		// Add user ID to context
		ctx := context.WithValue(r.Context(), "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// CreateTransaction handles creating a new loyalty transaction
func (s *Service) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate request
	if req.Amount <= 0 || req.MCC == "" || req.MerchantID == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Amount, MCC, and Merchant ID are required"})
		return
	}

	userID := r.Context().Value("user_id").(string)
	
	// Calculate points based on amount and MCC
	points, multiplier := s.calculatePoints(req.Amount, req.MCC)
	
	// Create transaction
	txn := &Transaction{
		ID:         uuid.New().String(),
		UserID:     userID,
		Amount:     req.Amount,
		MCC:        req.MCC,
		MerchantID: req.MerchantID,
		Points:     points,
		Multiplier: multiplier,
		CreatedAt:  time.Now(),
	}

	// Save transaction to database
	if err := s.saveTransaction(txn); err != nil {
		s.logger.Errorf("Failed to save transaction: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Failed to process transaction"})
		return
	}

	// Update user balance
	if err := s.updateBalance(userID, points); err != nil {
		s.logger.Errorf("Failed to update balance: %v", err)
		// Transaction was saved, so we should still return success
	}

	// Emit points earned event
	event := &PointsEarnedEvent{
		EventID:    uuid.New().String(),
		UserID:     userID,
		TxnID:      txn.ID,
		Points:     points,
		Multiplier: multiplier,
		MCC:        req.MCC,
		Timestamp:  time.Now(),
	}

	if err := s.emitPointsEarnedEvent(event); err != nil {
		s.logger.Errorf("Failed to emit points earned event: %v", err)
		// Don't fail the request for event emission failure
	}

	// Return response
	response := &TransactionResponse{
		TransactionID: txn.ID,
		Points:        points,
		Multiplier:    multiplier,
		Message:       fmt.Sprintf("Earned %d points with %.1fx multiplier", points, multiplier),
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, response)
}

// GetBalance returns the user's current loyalty balance
func (s *Service) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	
	balance, err := s.getBalance(userID)
	if err != nil {
		s.logger.Errorf("Failed to get balance: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Failed to retrieve balance"})
		return
	}

	render.JSON(w, r, balance)
}

// GetTransactions returns the user's transaction history
func (s *Service) GetTransactions(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	
	transactions, err := s.getTransactions(userID)
	if err != nil {
		s.logger.Errorf("Failed to get transactions: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Failed to retrieve transactions"})
		return
	}

	render.JSON(w, r, transactions)
}

// calculatePoints calculates points based on amount and MCC
func (s *Service) calculatePoints(amount float64, mcc string) (int, float64) {
	// Base rate: 1 point per dollar
	basePoints := int(amount)
	
	// MCC-based multipliers
	multipliers := map[string]float64{
		"5812": 3.0,  // Eating places and restaurants
		"5411": 2.5,  // Grocery stores, supermarkets
		"5541": 2.0,  // Service stations
		"5311": 1.5,  // Department stores
		"5999": 1.0,  // Miscellaneous retail
	}
	
	multiplier, exists := multipliers[mcc]
	if !exists {
		multiplier = 1.0
	}
	
	points := int(float64(basePoints) * multiplier)
	return points, multiplier
}

// saveTransaction saves a transaction to the database
func (s *Service) saveTransaction(txn *Transaction) error {
	if s.db == nil {
		// For now, just log - in production this would fail
		s.logger.Warn("Database not initialized, skipping transaction save")
		return nil
	}
	
	// TODO: Implement actual database save
	s.logger.Infof("Would save transaction: %+v", txn)
	return nil
}

// updateBalance updates the user's balance
func (s *Service) updateBalance(userID string, points int) error {
	if s.db == nil {
		s.logger.Warn("Database not initialized, skipping balance update")
		return nil
	}
	
	// TODO: Implement actual balance update
	s.logger.Infof("Would update balance for user %s: +%d points", userID, points)
	return nil
}

// getBalance retrieves the user's balance
func (s *Service) getBalance(userID string) (*Balance, error) {
	if s.db == nil {
		// Return mock data for now
		return &Balance{
			UserID:         userID,
			AvailablePoints: 1500,
			LifetimePoints:  5000,
			UpdatedAt:      time.Now(),
		}, nil
	}
	
	// TODO: Implement actual database query
	return nil, fmt.Errorf("not implemented")
}

// getTransactions retrieves the user's transaction history
func (s *Service) getTransactions(userID string) ([]*Transaction, error) {
	if s.db == nil {
		// Return mock data for now
		return []*Transaction{
			{
				ID:         "mock-1",
				UserID:     userID,
				Amount:     100.00,
				MCC:        "5812",
				MerchantID: "REST-001",
				Points:     300,
				Multiplier: 3.0,
				CreatedAt:  time.Now().Add(-24 * time.Hour),
			},
		}, nil
	}
	
	// TODO: Implement actual database query
	return nil, fmt.Errorf("not implemented")
}

// emitPointsEarnedEvent emits a points earned event to Kafka
func (s *Service) emitPointsEarnedEvent(event *PointsEarnedEvent) error {
	if s.kafka == nil {
		s.logger.Warn("Kafka not initialized, skipping event emission")
		return nil
	}
	
	// TODO: Implement actual Kafka event emission
	s.logger.Infof("Would emit event: %+v", event)
	return nil
}
