package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/kaihedrick/go-loyalty-benefits/internal/platform/config"
	"github.com/kaihedrick/go-loyalty-benefits/internal/platform/database"
	"github.com/sirupsen/logrus"
)

// Service represents the catalog service
type Service struct {
	config *config.Config
	logger *logrus.Logger
	db     *database.PostgresDB
}

// Benefit represents a loyalty benefit/reward
type Benefit struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Points      int        `json:"points"`
	Partner     string     `json:"partner"`
	Category    string     `json:"category"`
	Active      bool       `json:"active"`
	StartsAt    *time.Time `json:"starts_at,omitempty"`
	EndsAt      *time.Time `json:"ends_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateBenefitRequest represents a request to create a benefit
type CreateBenefitRequest struct {
	Name        string     `json:"name" validate:"required"`
	Description string     `json:"description"`
	Points      int        `json:"points" validate:"required,gt=0"`
	Partner     string     `json:"partner" validate:"required"`
	Category    string     `json:"category"`
	Active      bool       `json:"active"`
	StartsAt    *time.Time `json:"starts_at"`
	EndsAt      *time.Time `json:"ends_at"`
}

// UpdateBenefitRequest represents a request to update a benefit
type UpdateBenefitRequest struct {
	Name        *string     `json:"name"`
	Description *string     `json:"description"`
	Points      *int        `json:"points"`
	Partner     *string     `json:"partner"`
	Category    *string     `json:"category"`
	Active      *bool       `json:"active"`
	StartsAt    *time.Time  `json:"starts_at"`
	EndsAt      *time.Time  `json:"ends_at"`
}

// BenefitListResponse represents a paginated list of benefits
type BenefitListResponse struct {
	Benefits []*Benefit `json:"benefits"`
	Total    int        `json:"total"`
	Page     int        `json:"page"`
	Limit    int        `json:"limit"`
}

// NewService creates a new catalog service
func NewService(cfg *config.Config, logger *logrus.Logger) *Service {
	return &Service{
		config: cfg,
		logger: logger,
	}
}

// SetDatabase sets the database connection
func (s *Service) SetDatabase(db *database.PostgresDB) {
	s.db = db
}

// Routes returns the catalog service routes
func (s *Service) Routes(r chi.Router) {
	r.Route("/v1", func(r chi.Router) {
		r.Route("/benefits", func(r chi.Router) {
			r.Get("/", s.ListBenefits)
			r.Post("/", s.AuthMiddleware(s.CreateBenefit))
			r.Get("/{id}", s.GetBenefit)
			r.Put("/{id}", s.AuthMiddleware(s.UpdateBenefit))
			r.Delete("/{id}", s.AuthMiddleware(s.DeleteBenefit))
		})
		r.Get("/categories", s.GetCategories)
		r.Get("/partners", s.GetPartners)
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

// ListBenefits returns a paginated list of benefits
func (s *Service) ListBenefits(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	status := r.URL.Query().Get("status")
	category := r.URL.Query().Get("category")
	partner := r.URL.Query().Get("partner")
	
	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		pageStr = "1"
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		limitStr = "50"
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 50
	}

	// Get benefits from database
	benefits, total, err := s.getBenefits(status, category, partner, page, limit)
	if err != nil {
		s.logger.Errorf("Failed to get benefits: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Failed to retrieve benefits"})
		return
	}

	response := &BenefitListResponse{
		Benefits: benefits,
		Total:    total,
		Page:     page,
		Limit:    limit,
	}

	render.JSON(w, r, response)
}

// CreateBenefit creates a new benefit
func (s *Service) CreateBenefit(w http.ResponseWriter, r *http.Request) {
	var req CreateBenefitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate request
	if req.Name == "" || req.Points <= 0 || req.Partner == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Name, points, and partner are required"})
		return
	}

	// Create benefit
	benefit := &Benefit{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Points:      req.Points,
		Partner:     req.Partner,
		Category:    req.Category,
		Active:      req.Active,
		StartsAt:    req.StartsAt,
		EndsAt:      req.EndsAt,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save to database
	if err := s.saveBenefit(benefit); err != nil {
		s.logger.Errorf("Failed to save benefit: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Failed to create benefit"})
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, benefit)
}

// GetBenefit returns a specific benefit by ID
func (s *Service) GetBenefit(w http.ResponseWriter, r *http.Request) {
	benefitID := chi.URLParam(r, "id")
	if benefitID == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Benefit ID required"})
		return
	}

	benefit, err := s.getBenefit(benefitID)
	if err != nil {
		s.logger.Errorf("Failed to get benefit %s: %v", benefitID, err)
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": "Benefit not found"})
		return
	}

	render.JSON(w, r, benefit)
}

// UpdateBenefit updates an existing benefit
func (s *Service) UpdateBenefit(w http.ResponseWriter, r *http.Request) {
	benefitID := chi.URLParam(r, "id")
	if benefitID == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Benefit ID required"})
		return
	}

	var req UpdateBenefitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid request body"})
		return
	}

	// Get existing benefit
	existing, err := s.getBenefit(benefitID)
	if err != nil {
		s.logger.Errorf("Failed to get benefit %s: %v", benefitID, err)
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": "Benefit not found"})
		return
	}

	// Update fields if provided
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.Points != nil {
		existing.Points = *req.Points
	}
	if req.Partner != nil {
		existing.Partner = *req.Partner
	}
	if req.Category != nil {
		existing.Category = *req.Category
	}
	if req.Active != nil {
		existing.Active = *req.Active
	}
	if req.StartsAt != nil {
		existing.StartsAt = req.StartsAt
	}
	if req.EndsAt != nil {
		existing.EndsAt = req.EndsAt
	}
	
	existing.UpdatedAt = time.Now()

	// Save to database
	if err := s.updateBenefit(existing); err != nil {
		s.logger.Errorf("Failed to update benefit %s: %v", benefitID, err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Failed to update benefit"})
		return
	}

	render.JSON(w, r, existing)
}

// DeleteBenefit deletes a benefit
func (s *Service) DeleteBenefit(w http.ResponseWriter, r *http.Request) {
	benefitID := chi.URLParam(r, "id")
	if benefitID == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Benefit ID required"})
		return
	}

	// Check if benefit exists
	_, err := s.getBenefit(benefitID)
	if err != nil {
		s.logger.Errorf("Failed to get benefit %s: %v", benefitID, err)
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": "Benefit not found"})
		return
	}

	// Delete from database
	if err := s.deleteBenefit(benefitID); err != nil {
		s.logger.Errorf("Failed to delete benefit %s: %v", benefitID, err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Failed to delete benefit"})
		return
	}

	render.Status(r, http.StatusNoContent)
}

// GetCategories returns all available benefit categories
func (s *Service) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories := []string{
		"Travel",
		"Retail",
		"Dining",
		"Entertainment",
		"Technology",
		"Health & Wellness",
		"Charity",
		"Cash Back",
	}

	render.JSON(w, r, map[string]interface{}{
		"categories": categories,
	})
}

// GetPartners returns all available benefit partners
func (s *Service) GetPartners(w http.ResponseWriter, r *http.Request) {
	partners := []string{
		"GIFTCO",
		"TRAVELCO",
		"RETAILCO",
		"DININGCO",
		"ENTERTAINMENTCO",
	}

	render.JSON(w, r, map[string]interface{}{
		"partners": partners,
	})
}

// Database operations (placeholder implementations)
func (s *Service) getBenefits(status, category, partner string, page, limit int) ([]*Benefit, int, error) {
	if s.db == nil {
		// Return mock data for now
		benefits := []*Benefit{
			{
				ID:          "benefit-1",
				Name:        "$25 Gift Card",
				Description: "Redeemable at major retailers",
				Points:      2000,
				Partner:     "GIFTCO",
				Category:    "Retail",
				Active:      true,
				CreatedAt:   time.Now().Add(-24 * time.Hour),
				UpdatedAt:   time.Now().Add(-24 * time.Hour),
			},
			{
				ID:          "benefit-2",
				Name:        "Free Movie Ticket",
				Description: "Valid at participating theaters",
				Points:      1500,
				Partner:     "ENTERTAINMENTCO",
				Category:    "Entertainment",
				Active:      true,
				CreatedAt:   time.Now().Add(-48 * time.Hour),
				UpdatedAt:   time.Now().Add(-48 * time.Hour),
			},
		}
		return benefits, 2, nil
	}
	
	// TODO: Implement actual database query
	return nil, 0, fmt.Errorf("not implemented")
}

func (s *Service) getBenefit(id string) (*Benefit, error) {
	if s.db == nil {
		// Return mock data for now
		return &Benefit{
			ID:          id,
			Name:        "$25 Gift Card",
			Description: "Redeemable at major retailers",
			Points:      2000,
			Partner:     "GIFTCO",
			Category:    "Retail",
			Active:      true,
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now().Add(-24 * time.Hour),
		}, nil
	}
	
	// TODO: Implement actual database query
	return nil, fmt.Errorf("not implemented")
}

func (s *Service) saveBenefit(benefit *Benefit) error {
	if s.db == nil {
		s.logger.Infof("Would save benefit: %+v", benefit)
		return nil
	}
	
	// TODO: Implement actual database save
	return fmt.Errorf("not implemented")
}

func (s *Service) updateBenefit(benefit *Benefit) error {
	if s.db == nil {
		s.logger.Infof("Would update benefit: %+v", benefit)
		return nil
	}
	
	// TODO: Implement actual database update
	return fmt.Errorf("not implemented")
}

func (s *Service) deleteBenefit(id string) error {
	if s.db == nil {
		s.logger.Infof("Would delete benefit: %s", id)
		return nil
	}
	
	// TODO: Implement actual database delete
	return fmt.Errorf("not implemented")
}
