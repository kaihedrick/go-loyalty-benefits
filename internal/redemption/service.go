package redemption

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

// Service represents the redemption service
type Service struct {
	config *config.Config
	logger *logrus.Logger
	db     *database.PostgresDB
	kafka  *messaging.KafkaProducer
}

// Redemption represents a loyalty redemption
type Redemption struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	BenefitID       string    `json:"benefit_id"`
	Points          int       `json:"points"`
	Status          string    `json:"status"`
	IdempotencyKey  string    `json:"idempotency_key"`
	PartnerRef      string    `json:"partner_ref,omitempty"`
	ErrorMessage    string    `json:"error_message,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

// RedemptionRequest represents a redemption request
type RedemptionRequest struct {
	BenefitID string `json:"benefit_id" validate:"required"`
	Points    int    `json:"points" validate:"required,gt=0"`
}

// RedemptionResponse represents a redemption response
type RedemptionResponse struct {
	RedemptionID string `json:"redemption_id"`
	Status       string `json:"status"`
	Message      string `json:"message"`
}

// RedemptionStatus represents the status of a redemption
type RedemptionStatus struct {
	ID              string     `json:"id"`
	Status          string     `json:"status"`
	Points          int        `json:"points"`
	BenefitName     string     `json:"benefit_name"`
	PartnerRef      string     `json:"partner_ref,omitempty"`
	ErrorMessage    string     `json:"error_message,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

// RedemptionCompletedEvent represents the redemption completed event
type RedemptionCompletedEvent struct {
	EventID     string    `json:"event_id"`
	UserID      string    `json:"user_id"`
	BenefitID   string    `json:"benefit_id"`
	Points      int       `json:"points"`
	PartnerRef  string    `json:"partner_ref"`
	Timestamp   time.Time `json:"ts"`
}

// RedemptionFailedEvent represents the redemption failed event
type RedemptionFailedEvent struct {
	EventID      string    `json:"event_id"`
	UserID       string    `json:"user_id"`
	BenefitID    string    `json:"benefit_id"`
	Points       int       `json:"points"`
	ErrorMessage string    `json:"error_message"`
	Timestamp    time.Time `json:"ts"`
}

// OutboxMessage represents a message in the outbox
type OutboxMessage struct {
	ID        int64           `json:"id"`
	Aggregate string          `json:"aggregate"`
	Payload   json.RawMessage `json:"payload"`
	Topic     string          `json:"topic"`
	CreatedAt time.Time       `json:"created_at"`
}

// NewService creates a new redemption service
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

// Routes returns the redemption service routes
func (s *Service) Routes(r chi.Router) {
	r.Route("/v1", func(r chi.Router) {
		r.Post("/redeem", s.AuthMiddleware(s.CreateRedemption))
		r.Get("/redemptions/{id}", s.AuthMiddleware(s.GetRedemption))
		r.Get("/redemptions", s.AuthMiddleware(s.ListRedemptions))
	})
}

// AuthMiddleware is a placeholder for JWT authentication
func (s *Service) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement JWT validation
		// For now, just check if user ID header is present
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

// CreateRedemption handles creating a new redemption
func (s *Service) CreateRedemption(w http.ResponseWriter, r *http.Request) {
	var req RedemptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate request
	if req.BenefitID == "" || req.Points <= 0 {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Benefit ID and points are required"})
		return
	}

	userID := r.Context().Value("user_id").(string)
	idempotencyKey := r.Header.Get("Idempotency-Key")
	
	if idempotencyKey == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Idempotency-Key header is required"})
		return
	}

	// Check if redemption already exists (idempotency)
	existing, err := s.getRedemptionByKey(idempotencyKey)
	if err == nil && existing != nil {
		// Return existing redemption
		response := &RedemptionResponse{
			RedemptionID: existing.ID,
			Status:       existing.Status,
			Message:      "Redemption already exists",
		}
		render.JSON(w, r, response)
		return
	}

	// Create redemption
	redemption := &Redemption{
		ID:             uuid.New().String(),
		UserID:         userID,
		BenefitID:      req.BenefitID,
		Points:         req.Points,
		Status:         "requested",
		IdempotencyKey: idempotencyKey,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Save redemption to database
	if err := s.saveRedemption(redemption); err != nil {
		s.logger.Errorf("Failed to save redemption: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Failed to create redemption"})
		return
	}

	// Start redemption saga asynchronously
	go s.processRedemptionSaga(redemption)

	// Return immediate response
	response := &RedemptionResponse{
		RedemptionID: redemption.ID,
		Status:       "requested",
		Message:      "Redemption request accepted",
	}

	render.Status(r, http.StatusAccepted)
	render.JSON(w, r, response)
}

// GetRedemption returns a specific redemption by ID
func (s *Service) GetRedemption(w http.ResponseWriter, r *http.Request) {
	redemptionID := chi.URLParam(r, "id")
	if redemptionID == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Redemption ID required"})
		return
	}

	redemption, err := s.getRedemption(redemptionID)
	if err != nil {
		s.logger.Errorf("Failed to get redemption %s: %v", redemptionID, err)
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": "Redemption not found"})
		return
	}

	// Convert to status response
	status := &RedemptionStatus{
		ID:           redemption.ID,
		Status:       redemption.Status,
		Points:       redemption.Points,
		BenefitName:  "Unknown Benefit", // TODO: Get from catalog service
		PartnerRef:   redemption.PartnerRef,
		ErrorMessage: redemption.ErrorMessage,
		CreatedAt:    redemption.CreatedAt,
		CompletedAt:  redemption.CompletedAt,
	}

	render.JSON(w, r, status)
}

// ListRedemptions returns the user's redemption history
func (s *Service) ListRedemptions(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	
	redemptions, err := s.getRedemptionsByUser(userID)
	if err != nil {
		s.logger.Errorf("Failed to get redemptions: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Failed to retrieve redemptions"})
		return
	}

	render.JSON(w, r, redemptions)
}

// processRedemptionSaga processes the redemption saga
func (s *Service) processRedemptionSaga(redemption *Redemption) {
	// Step 1: Validate benefit and check availability
	if err := s.validateBenefit(redemption.BenefitID); err != nil {
		s.failRedemption(redemption, err.Error())
		return
	}

	// Step 2: Check user has enough points
	if err := s.checkUserPoints(redemption.UserID, redemption.Points); err != nil {
		s.failRedemption(redemption, err.Error())
		return
	}

	// Step 3: Deduct points from user balance
	if err := s.deductPoints(redemption.UserID, redemption.Points); err != nil {
		s.failRedemption(redemption, err.Error())
		return
	}

	// Step 4: Call partner gateway to fulfill benefit
	partnerRef, err := s.callPartnerGateway(redemption)
	if err != nil {
		// Try to reverse points deduction
		s.reversePointsDeduction(redemption.UserID, redemption.Points)
		s.failRedemption(redemption, err.Error())
		return
	}

	// Step 5: Mark redemption as completed
	redemption.Status = "completed"
	redemption.PartnerRef = partnerRef
	redemption.CompletedAt = &time.Time{}
	*redemption.CompletedAt = time.Now()
	redemption.UpdatedAt = time.Now()

	if err := s.updateRedemption(redemption); err != nil {
		s.logger.Errorf("Failed to update redemption status: %v", err)
		// Don't fail the saga at this point
	}

	// Step 6: Emit completion event
	event := &RedemptionCompletedEvent{
		EventID:    uuid.New().String(),
		UserID:     redemption.UserID,
		BenefitID:  redemption.BenefitID,
		Points:     redemption.Points,
		PartnerRef: partnerRef,
		Timestamp:  time.Now(),
	}

	if err := s.emitRedemptionCompletedEvent(event); err != nil {
		s.logger.Errorf("Failed to emit redemption completed event: %v", err)
		// Don't fail the saga for event emission failure
	}

	s.logger.Infof("Redemption %s completed successfully", redemption.ID)
}

// failRedemption marks a redemption as failed
func (s *Service) failRedemption(redemption *Redemption, errorMessage string) {
	redemption.Status = "failed"
	redemption.ErrorMessage = errorMessage
	redemption.UpdatedAt = time.Now()

	if err := s.updateRedemption(redemption); err != nil {
		s.logger.Errorf("Failed to update redemption status: %v", err)
	}

	// Emit failure event
	event := &RedemptionFailedEvent{
		EventID:      uuid.New().String(),
		UserID:       redemption.UserID,
		BenefitID:    redemption.BenefitID,
		Points:       redemption.Points,
		ErrorMessage: errorMessage,
		Timestamp:    time.Now(),
	}

	if err := s.emitRedemptionFailedEvent(event); err != nil {
		s.logger.Errorf("Failed to emit redemption failed event: %v", err)
	}

	s.logger.Errorf("Redemption %s failed: %s", redemption.ID, errorMessage)
}

// Database operations (placeholder implementations)
func (s *Service) getRedemptionByKey(idempotencyKey string) (*Redemption, error) {
	if s.db == nil {
		// For now, return nil (no existing redemption)
		return nil, fmt.Errorf("not implemented")
	}
	
	// TODO: Implement actual database query
	return nil, fmt.Errorf("not implemented")
}

func (s *Service) saveRedemption(redemption *Redemption) error {
	if s.db == nil {
		s.logger.Infof("Would save redemption: %+v", redemption)
		return nil
	}
	
	// TODO: Implement actual database save
	return fmt.Errorf("not implemented")
}

func (s *Service) getRedemption(id string) (*Redemption, error) {
	if s.db == nil {
		// Return mock data for now
		return &Redemption{
			ID:         id,
			UserID:     "user-123",
			BenefitID:  "benefit-1",
			Points:     2000,
			Status:     "completed",
			PartnerRef: "VENDOR-12345",
			CreatedAt:  time.Now().Add(-1 * time.Hour),
			UpdatedAt:  time.Now().Add(-30 * time.Minute),
		}, nil
	}
	
	// TODO: Implement actual database query
	return nil, fmt.Errorf("not implemented")
}

func (s *Service) getRedemptionsByUser(userID string) ([]*Redemption, error) {
	if s.db == nil {
		// Return mock data for now
		return []*Redemption{
			{
				ID:         "redemption-1",
				UserID:     userID,
				BenefitID:  "benefit-1",
				Points:     2000,
				Status:     "completed",
				PartnerRef: "VENDOR-12345",
				CreatedAt:  time.Now().Add(-24 * time.Hour),
				UpdatedAt:  time.Now().Add(-24 * time.Hour),
			},
		}, nil
	}
	
	// TODO: Implement actual database query
	return nil, fmt.Errorf("not implemented")
}

func (s *Service) updateRedemption(redemption *Redemption) error {
	if s.db == nil {
		s.logger.Infof("Would update redemption: %+v", redemption)
		return nil
	}
	
	// TODO: Implement actual database update
	return fmt.Errorf("not implemented")
}

// Saga step implementations (placeholder)
func (s *Service) validateBenefit(benefitID string) error {
	// TODO: Call catalog service to validate benefit
	s.logger.Infof("Would validate benefit: %s", benefitID)
	return nil
}

func (s *Service) checkUserPoints(userID string, points int) error {
	// TODO: Call loyalty service to check user points
	s.logger.Infof("Would check user %s has %d points", userID, points)
	return nil
}

func (s *Service) deductPoints(userID string, points int) error {
	// TODO: Call loyalty service to deduct points
	s.logger.Infof("Would deduct %d points from user %s", points, userID)
	return nil
}

func (s *Service) callPartnerGateway(redemption *Redemption) (string, error) {
	// TODO: Call partner gateway service
	s.logger.Infof("Would call partner gateway for redemption: %s", redemption.ID)
	return "VENDOR-" + uuid.New().String()[:8], nil
}

func (s *Service) reversePointsDeduction(userID string, points int) error {
	// TODO: Call loyalty service to reverse points deduction
	s.logger.Infof("Would reverse %d points deduction for user %s", points, userID)
	return nil
}

// Event emission (placeholder implementations)
func (s *Service) emitRedemptionCompletedEvent(event *RedemptionCompletedEvent) error {
	if s.kafka == nil {
		s.logger.Warn("Kafka not initialized, skipping event emission")
		return nil
	}
	
	// TODO: Implement actual Kafka event emission
	s.logger.Infof("Would emit redemption completed event: %+v", event)
	return nil
}

func (s *Service) emitRedemptionFailedEvent(event *RedemptionFailedEvent) error {
	if s.kafka == nil {
		s.logger.Warn("Kafka not initialized, skipping event emission")
		return nil
	}
	
	// TODO: Implement actual Kafka event emission
	s.logger.Infof("Would emit redemption failed event: %+v", event)
	return nil
}
