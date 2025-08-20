package notify

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/kaihedrick/go-loyalty-benefits/internal/platform/config"
	"github.com/kaihedrick/go-loyalty-benefits/internal/platform/messaging"
	"github.com/sirupsen/logrus"
)

// Service represents the notification service
type Service struct {
	config *config.Config
	logger *logrus.Logger
	kafka  *messaging.KafkaConsumer
}

// Notification represents a notification
type Notification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Type      string    `json:"type"`      // email, sms, push
	Subject   string    `json:"subject"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`    // pending, sent, failed
	Channel   string    `json:"channel"`   // email, sms, push
	CreatedAt time.Time `json:"created_at"`
	SentAt    *time.Time `json:"sent_at,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// NotificationRequest represents a request to send a notification
type NotificationRequest struct {
	UserID  string            `json:"user_id" validate:"required"`
	Type    string            `json:"type" validate:"required,oneof=email sms push"`
	Subject string            `json:"subject"`
	Message string            `json:"message" validate:"required"`
	Channel string            `json:"channel" validate:"required,oneof=email sms push"`
	Data    map[string]string `json:"data,omitempty"`
}

// NotificationResponse represents a notification response
type NotificationResponse struct {
	NotificationID string `json:"notification_id"`
	Status         string `json:"status"`
	Message        string `json:"message"`
}

// EmailTemplate represents an email template
type EmailTemplate struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Subject string            `json:"subject"`
	Body    string            `json:"body"`
	Variables []string        `json:"variables"`
}

// SMSTemplate represents an SMS template
type SMSTemplate struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Message string   `json:"message"`
	Variables []string `json:"variables"`
}

// NewService creates a new notification service
func NewService(cfg *config.Config, logger *logrus.Logger) *Service {
	// Initialize Kafka consumer for redemption events
	kafkaConfig := &messaging.KafkaConfig{
		Brokers:  cfg.Kafka.Brokers,
		ClientID: cfg.Kafka.ClientID,
		GroupID:  cfg.Kafka.GroupID,
	}
	kafkaConsumer := messaging.NewKafkaConsumer(kafkaConfig, "redemption.completed.v1", logger)

	service := &Service{
		config: cfg,
		logger: logger,
		kafka:  kafkaConsumer,
	}

	// Start consuming Kafka events
	go service.consumeRedemptionEvents()

	return service
}

// Routes returns the notification service routes
func (s *Service) Routes(r chi.Router) {
	r.Route("/v1", func(r chi.Router) {
		r.Route("/notifications", func(r chi.Router) {
			r.Post("/", s.AuthMiddleware(s.SendNotification))
			r.Get("/{id}", s.AuthMiddleware(s.GetNotification))
			r.Get("/", s.AuthMiddleware(s.ListNotifications))
		})
		r.Route("/templates", func(r chi.Router) {
			r.Get("/email", s.GetEmailTemplates)
			r.Get("/sms", s.GetSMSTemplates)
		})
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

// SendNotification handles sending a notification
func (s *Service) SendNotification(w http.ResponseWriter, r *http.Request) {
	var req NotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate request
	if req.UserID == "" || req.Type == "" || req.Message == "" || req.Channel == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "User ID, type, message, and channel are required"})
		return
	}

	// Create notification
	notification := &Notification{
		ID:        uuid.New().String(),
		UserID:    req.UserID,
		Type:      req.Type,
		Subject:   req.Subject,
		Message:   req.Message,
		Status:    "pending",
		Channel:   req.Channel,
		CreatedAt: time.Now(),
	}

	// Send notification asynchronously
	go s.sendNotification(notification)

	// Return immediate response
	response := &NotificationResponse{
		NotificationID: notification.ID,
		Status:         "pending",
		Message:        "Notification queued for delivery",
	}

	render.Status(r, http.StatusAccepted)
	render.JSON(w, r, response)
}

// GetNotification returns a specific notification by ID
func (s *Service) GetNotification(w http.ResponseWriter, r *http.Request) {
	notificationID := chi.URLParam(r, "id")
	if notificationID == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Notification ID required"})
		return
	}

	notification, err := s.getNotification(notificationID)
	if err != nil {
		s.logger.Errorf("Failed to get notification %s: %v", notificationID, err)
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": "Notification not found"})
		return
	}

	render.JSON(w, r, notification)
}

// ListNotifications returns the user's notification history
func (s *Service) ListNotifications(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)
	
	notifications, err := s.getNotificationsByUser(userID)
	if err != nil {
		s.logger.Errorf("Failed to get notifications: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Failed to retrieve notifications"})
		return
	}

	render.JSON(w, r, notifications)
}

// GetEmailTemplates returns available email templates
func (s *Service) GetEmailTemplates(w http.ResponseWriter, r *http.Request) {
	templates := []*EmailTemplate{
		{
			ID:        "redemption-completed",
			Name:      "Redemption Completed",
			Subject:   "Your reward has been fulfilled!",
			Body:      "Dear {{user_name}}, your {{benefit_name}} has been successfully fulfilled. Reference: {{partner_ref}}",
			Variables: []string{"user_name", "benefit_name", "partner_ref"},
		},
		{
			ID:        "points-earned",
			Name:      "Points Earned",
			Subject:   "You've earned {{points}} points!",
			Body:      "Congratulations! You've earned {{points}} points from your recent transaction at {{merchant}}.",
			Variables: []string{"points", "merchant"},
		},
		{
			ID:        "welcome",
			Name:      "Welcome",
			Subject:   "Welcome to our loyalty program!",
			Body:      "Welcome {{user_name}}! Start earning points with every purchase.",
			Variables: []string{"user_name"},
		},
	}

	render.JSON(w, r, map[string]interface{}{
		"templates": templates,
		"total":     len(templates),
	})
}

// GetSMSTemplates returns available SMS templates
func (s *Service) GetSMSTemplates(w http.ResponseWriter, r *http.Request) {
	templates := []*SMSTemplate{
		{
			ID:        "redemption-completed-sms",
			Name:      "Redemption Completed SMS",
			Message:   "Your {{benefit_name}} has been fulfilled! Ref: {{partner_ref}}",
			Variables: []string{"benefit_name", "partner_ref"},
		},
		{
			ID:        "points-earned-sms",
			Name:      "Points Earned SMS",
			Message:   "You earned {{points}} points! Keep shopping to earn more.",
			Variables: []string{"points"},
		},
	}

	render.JSON(w, r, map[string]interface{}{
		"templates": templates,
		"total":     len(templates),
	})
}

// consumeRedemptionEvents consumes redemption events from Kafka
func (s *Service) consumeRedemptionEvents() {
	if s.kafka == nil {
		s.logger.Warn("Kafka consumer not initialized, skipping event consumption")
		return
	}

	s.logger.Info("Starting to consume redemption events...")
	
	// TODO: Implement actual Kafka event consumption
	// For now, just log that we would consume events
	s.logger.Info("Would consume redemption.completed.v1 events from Kafka")
}

// sendNotification sends a notification through the appropriate channel
func (s *Service) sendNotification(notification *Notification) {
	s.logger.Infof("Sending notification %s to user %s via %s", notification.ID, notification.UserID, notification.Channel)

	// Simulate sending delay
	time.Sleep(100 * time.Millisecond)

	// Simulate success (in real implementation, this would call actual email/SMS services)
	notification.Status = "sent"
	sentAt := time.Now()
	notification.SentAt = &sentAt

	s.logger.Infof("Notification %s sent successfully", notification.ID)
	
	// TODO: Save notification status to database
	// TODO: Emit notification sent event
}

// Database operations (placeholder implementations)
func (s *Service) getNotification(id string) (*Notification, error) {
	// Return mock data for now
	return &Notification{
		ID:        id,
		UserID:    "user-123",
		Type:      "email",
		Subject:   "Your reward has been fulfilled!",
		Message:   "Dear User, your $25 Gift Card has been successfully fulfilled. Reference: VENDOR-12345",
		Status:    "sent",
		Channel:   "email",
		CreatedAt: time.Now().Add(-1 * time.Hour),
		SentAt:    &time.Time{},
	}, nil
}

func (s *Service) getNotificationsByUser(userID string) ([]*Notification, error) {
	// Return mock data for now
	return []*Notification{
		{
			ID:        "notif-1",
			UserID:    userID,
			Type:      "email",
			Subject:   "Your reward has been fulfilled!",
			Message:   "Dear User, your $25 Gift Card has been successfully fulfilled. Reference: VENDOR-12345",
			Status:    "sent",
			Channel:   "email",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			SentAt:    &time.Time{},
		},
		{
			ID:        "notif-2",
			UserID:    userID,
			Type:      "sms",
			Subject:   "",
			Message:   "You earned 300 points! Keep shopping to earn more.",
			Status:    "sent",
			Channel:   "sms",
			CreatedAt: time.Now().Add(-48 * time.Hour),
			SentAt:    &time.Time{},
		},
	}, nil
}
