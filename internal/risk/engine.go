package risk

import (
	"fmt"
	"log"
	"time"

	"riskmatrix/pkg/database"
	"riskmatrix/pkg/models"
)

// Config holds configuration for the risk engine
type Config struct {
	// Threshold at which to generate risk alerts
	RiskThreshold int
	
	// Decay factor for risk scores (0-1, where 0 means no decay)
	DecayFactor float64
	
	// How often to run the decay process
	DecayInterval time.Duration
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		RiskThreshold: 50,
		DecayFactor:   0.1,
		DecayInterval: 24 * time.Hour,
	}
}

// Engine is responsible for processing events and managing risk scores
type Engine struct {
	db     *database.DB
	repo   *Repository
	config Config
}

// NewEngine creates a new risk engine
func NewEngine(db *database.DB, config Config) *Engine {
	return &Engine{
		db:     db,
		repo:   NewRepository(db),
		config: config,
	}
}

// ProcessEvent processes a security event and updates risk scores
func (e *Engine) ProcessEvent(event *models.Event) error {
	// Begin transaction
	tx, err := e.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get or create risk object
	riskObject, err := e.repo.GetRiskObjectByEntityTx(tx, event.RiskObject.EntityType, event.RiskObject.EntityValue)
	if err != nil {
		// Create new risk object if not found
		riskObject = &models.RiskObject{
			EntityType:   event.RiskObject.EntityType,
			EntityValue:  event.RiskObject.EntityValue,
			CurrentScore: 0,
			LastSeen:     time.Now(),
		}
		
		if err := e.repo.CreateRiskObjectTx(tx, riskObject); err != nil {
			return fmt.Errorf("failed to create risk object: %w", err)
		}
	}
	
	// Set entity ID in event
	event.EntityID = riskObject.ID
	
	// Save event
	if err := e.repo.CreateEventTx(tx, event); err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}
	
	// Update risk score
	oldScore := riskObject.CurrentScore
	riskObject.CurrentScore += event.RiskPoints
	riskObject.LastSeen = time.Now()
	
	if err := e.repo.UpdateRiskObjectTx(tx, riskObject); err != nil {
		return fmt.Errorf("failed to update risk object: %w", err)
	}
	
	// Check if threshold crossed
	if oldScore < e.config.RiskThreshold && riskObject.CurrentScore >= e.config.RiskThreshold {
		// Create risk alert
		alert := &models.RiskAlert{
			EntityID:    riskObject.ID,
			TriggeredAt: time.Now(),
			TotalScore:  riskObject.CurrentScore,
			Status:      models.AlertStatusNew,
			Notes:       "",
			Owner:       "",
		}
		
		if err := e.repo.CreateRiskAlertTx(tx, alert); err != nil {
			return fmt.Errorf("failed to create risk alert: %w", err)
		}
		
		log.Printf("Risk alert generated for %s '%s' with score %d", 
			riskObject.EntityType, riskObject.EntityValue, riskObject.CurrentScore)
	}
	
	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// ProcessEvents processes multiple events in batch
func (e *Engine) ProcessEvents(events []*models.Event) error {
	for _, event := range events {
		if err := e.ProcessEvent(event); err != nil {
			return fmt.Errorf("failed to process event: %w", err)
		}
	}
	return nil
}

// DecayRiskScores reduces all risk scores by the decay factor
func (e *Engine) DecayRiskScores() error {
	if e.config.DecayFactor <= 0 {
		return nil // No decay needed
	}
	
	return e.repo.DecayRiskScores(e.config.DecayFactor)
}

// StartDecayProcess starts a background process to decay risk scores periodically
func (e *Engine) StartDecayProcess(stop <-chan struct{}) {
	ticker := time.NewTicker(e.config.DecayInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if err := e.DecayRiskScores(); err != nil {
				log.Printf("Error decaying risk scores: %v", err)
			} else {
				log.Printf("Risk scores decayed by factor %.2f", e.config.DecayFactor)
			}
		case <-stop:
			return
		}
	}
}

// MarkEventAsFalsePositive marks an event as a false positive and adjusts risk scores
func (e *Engine) MarkEventAsFalsePositive(eventID int64, fpInfo *models.FalsePositive) error {
	// Begin transaction
	tx, err := e.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Get event
	event, err := e.repo.GetEventTx(tx, eventID)
	if err != nil {
		return fmt.Errorf("failed to get event: %w", err)
	}
	
	// Check if already marked as false positive
	if event.IsFalsePositive {
		return fmt.Errorf("event already marked as false positive")
	}
	
	// Mark event as false positive
	event.IsFalsePositive = true
	if err := e.repo.UpdateEventTx(tx, event); err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}
	
	// Create false positive record
	fpInfo.EventID = eventID
	if err := e.repo.CreateFalsePositiveTx(tx, fpInfo); err != nil {
		return fmt.Errorf("failed to create false positive record: %w", err)
	}
	
	// Adjust risk score for entity
	riskObject, err := e.repo.GetRiskObjectTx(tx, event.EntityID)
	if err != nil {
		return fmt.Errorf("failed to get risk object: %w", err)
	}
	
	// Subtract risk points
	riskObject.CurrentScore -= event.RiskPoints
	if riskObject.CurrentScore < 0 {
		riskObject.CurrentScore = 0
	}
	
	if err := e.repo.UpdateRiskObjectTx(tx, riskObject); err != nil {
		return fmt.Errorf("failed to update risk object: %w", err)
	}
	
	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// UnmarkEventAsFalsePositive unmarks an event as a false positive and re-adds risk points
func (e *Engine) UnmarkEventAsFalsePositive(eventID int64) error {
	// Begin transaction
	tx, err := e.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Get event
	event, err := e.repo.GetEventTx(tx, eventID)
	if err != nil {
		return fmt.Errorf("failed to get event: %w", err)
	}
	
	// Check if event is marked as false positive
	if !event.IsFalsePositive {
		return fmt.Errorf("event is not marked as false positive")
	}
	
	// Unmark event as false positive
	event.IsFalsePositive = false
	if err := e.repo.UpdateEventTx(tx, event); err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}
	
	// Delete false positive record
	if err := e.repo.DeleteFalsePositiveByEventTx(tx, eventID); err != nil {
		return fmt.Errorf("failed to delete false positive record: %w", err)
	}
	
	// Re-add risk points to entity
	riskObject, err := e.repo.GetRiskObjectTx(tx, event.EntityID)
	if err != nil {
		return fmt.Errorf("failed to get risk object: %w", err)
	}
	
	// Add back risk points
	riskObject.CurrentScore += event.RiskPoints
	
	if err := e.repo.UpdateRiskObjectTx(tx, riskObject); err != nil {
		return fmt.Errorf("failed to update risk object: %w", err)
	}
	
	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// GetHighRiskEntities returns entities with risk scores above the threshold
func (e *Engine) GetHighRiskEntities() ([]*models.RiskObject, error) {
	return e.repo.ListHighRiskObjects(e.config.RiskThreshold)
}

// GetRiskAlerts returns all risk alerts
func (e *Engine) GetRiskAlerts() ([]*models.RiskAlert, error) {
	return e.repo.ListRiskAlerts()
}

// GetRiskAlertsByStatus returns risk alerts filtered by status
func (e *Engine) GetRiskAlertsByStatus(status models.AlertStatus) ([]*models.RiskAlert, error) {
	return e.repo.ListRiskAlertsByStatus(status)
}

// GetRiskAlertsPaginated returns paginated risk alerts with total count
func (e *Engine) GetRiskAlertsPaginated(limit, offset int, status models.AlertStatus) ([]*models.RiskAlert, int, error) {
	return e.repo.ListRiskAlertsPaginated(limit, offset, status)
}

// GetEventsForAlert returns events that contributed to a risk alert
func (e *Engine) GetEventsForAlert(alertID int64) ([]*models.Event, error) {
	return e.repo.GetEventsForAlert(alertID)
}