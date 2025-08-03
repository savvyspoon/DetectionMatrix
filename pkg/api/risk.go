package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"riskmatrix/internal/risk"
	"riskmatrix/pkg/models"
)

// RiskHandler handles HTTP requests for risk-related endpoints
type RiskHandler struct {
	engine *risk.Engine
	repo   *risk.Repository
}

// NewRiskHandler creates a new risk handler
func NewRiskHandler(engine *risk.Engine, repo *risk.Repository) *RiskHandler {
	return &RiskHandler{
		engine: engine,
		repo:   repo,
	}
}

// ProcessEvent handles POST /api/events
func (h *RiskHandler) ProcessEvent(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var event models.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Process event
	if err := h.engine.ProcessEvent(&event); err != nil {
		http.Error(w, "Error processing event", http.StatusInternalServerError)
		return
	}

	// Return created event as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(event)
}

// ProcessEvents handles POST /api/events/batch
func (h *RiskHandler) ProcessEvents(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var events []*models.Event
	if err := json.NewDecoder(r.Body).Decode(&events); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set timestamp for events if not provided
	now := time.Now()
	for _, event := range events {
		if event.Timestamp.IsZero() {
			event.Timestamp = now
		}
	}

	// Process events
	if err := h.engine.ProcessEvents(events); err != nil {
		http.Error(w, "Error processing events", http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// GetRiskObject handles GET /api/risk/objects/{id}
func (h *RiskHandler) GetRiskObject(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid risk object ID", http.StatusBadRequest)
		return
	}

	// Get risk object from repository
	obj, err := h.repo.GetRiskObject(id)
	if err != nil {
		http.Error(w, "Risk object not found", http.StatusNotFound)
		return
	}

	// Return risk object as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(obj)
}

// GetRiskObjectByEntity handles GET /api/risk/objects/entity
func (h *RiskHandler) GetRiskObjectByEntity(w http.ResponseWriter, r *http.Request) {
	// Extract entity type and value from query parameters
	entityType := r.URL.Query().Get("type")
	entityValue := r.URL.Query().Get("value")

	if entityType == "" || entityValue == "" {
		http.Error(w, "Missing entity type or value", http.StatusBadRequest)
		return
	}

	// Get risk object from repository
	obj, err := h.repo.GetRiskObjectByEntity(models.EntityType(entityType), entityValue)
	if err != nil {
		http.Error(w, "Risk object not found", http.StatusNotFound)
		return
	}

	// Return risk object as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(obj)
}

// ListRiskObjects handles GET /api/risk/objects
func (h *RiskHandler) ListRiskObjects(w http.ResponseWriter, r *http.Request) {
	// Check for high risk filter
	thresholdStr := r.URL.Query().Get("threshold")

	var objects []*models.RiskObject
	var err error

	if thresholdStr != "" {
		// Parse threshold
		threshold, err := strconv.Atoi(thresholdStr)
		if err != nil {
			http.Error(w, "Invalid threshold", http.StatusBadRequest)
			return
		}

		// List high risk objects
		objects, err = h.repo.ListHighRiskObjects(threshold)
	} else {
		// List all risk objects
		objects, err = h.repo.ListRiskObjects()
	}

	if err != nil {
		http.Error(w, "Error retrieving risk objects", http.StatusInternalServerError)
		return
	}

	// Return risk objects as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(objects)
}

// GetEvent handles GET /api/events/{id}
func (h *RiskHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	// Get event from repository
	event, err := h.repo.GetEvent(id)
	if err != nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	// Return event as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}

// ListEvents handles GET /api/events
func (h *RiskHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	
	// Default values
	page := 1
	limit := 20
	
	// Parse page parameter
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	// Parse limit parameter
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	// Calculate offset
	offset := (page - 1) * limit
	
	// Get paginated events from repository
	events, totalCount, err := h.repo.ListEventsPaginated(limit, offset)
	if err != nil {
		http.Error(w, "Error retrieving events", http.StatusInternalServerError)
		return
	}
	
	// Calculate pagination metadata
	totalPages := (totalCount + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1
	
	// Create response with pagination metadata
	response := map[string]interface{}{
		"events": events,
		"pagination": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total_count": totalCount,
			"total_pages": totalPages,
			"has_next":    hasNext,
			"has_prev":    hasPrev,
		},
	}

	// Return paginated events as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListEventsByEntity handles GET /api/events/entity/{id}
func (h *RiskHandler) ListEventsByEntity(w http.ResponseWriter, r *http.Request) {
	// Extract entity ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid entity ID", http.StatusBadRequest)
		return
	}

	// Get events from repository
	events, err := h.repo.ListEventsByEntity(id)
	if err != nil {
		http.Error(w, "Error retrieving events", http.StatusInternalServerError)
		return
	}

	// Return events as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// MarkEventAsFalsePositive handles POST /api/events/{id}/false-positive
func (h *RiskHandler) MarkEventAsFalsePositive(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var fpInfo models.FalsePositive
	if err := json.NewDecoder(r.Body).Decode(&fpInfo); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set timestamp if not provided
	if fpInfo.Timestamp.IsZero() {
		fpInfo.Timestamp = time.Now()
	}

	// Mark event as false positive
	if err := h.engine.MarkEventAsFalsePositive(id, &fpInfo); err != nil {
		http.Error(w, "Error marking event as false positive", http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// RiskAlertsResponse represents the paginated response for risk alerts
type RiskAlertsResponse struct {
	Alerts     []*models.RiskAlert `json:"alerts"`
	TotalCount int                 `json:"total_count"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
}

// ListRiskAlerts handles GET /api/risk/alerts
func (h *RiskHandler) ListRiskAlerts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	statusFilter := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	pageStr := r.URL.Query().Get("page")

	// Default pagination values
	limit := 50
	page := 1

	// Parse limit parameter
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		} else if parsedLimit > 100 {
			limit = 100 // Cap at 100
		}
	}

	// Parse page parameter
	if pageStr != "" {
		if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Check if pagination is requested
	if limitStr != "" || pageStr != "" {
		// Use paginated method
		alerts, totalCount, err := h.engine.GetRiskAlertsPaginated(limit, offset, models.AlertStatus(statusFilter))
		if err != nil {
			http.Error(w, "Error retrieving risk alerts", http.StatusInternalServerError)
			return
		}

		// Calculate total pages
		totalPages := (totalCount + limit - 1) / limit

		response := RiskAlertsResponse{
			Alerts:     alerts,
			TotalCount: totalCount,
			Page:       page,
			PageSize:   limit,
			TotalPages: totalPages,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		// Legacy non-paginated response for backward compatibility
		var alerts []*models.RiskAlert
		var err error

		if statusFilter != "" {
			// List alerts filtered by status
			alerts, err = h.engine.GetRiskAlertsByStatus(models.AlertStatus(statusFilter))
		} else {
			// List all risk alerts
			alerts, err = h.engine.GetRiskAlerts()
		}

		if err != nil {
			http.Error(w, "Error retrieving risk alerts", http.StatusInternalServerError)
			return
		}

		// Return risk alerts as JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(alerts)
	}
}

// GetRiskAlert handles GET /api/risk/alerts/{id}
func (h *RiskHandler) GetRiskAlert(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid alert ID", http.StatusBadRequest)
		return
	}

	// Get risk alert from repository
	alert, err := h.repo.GetRiskAlert(id)
	if err != nil {
		http.Error(w, "Error retrieving risk alert", http.StatusInternalServerError)
		return
	}

	// Return risk alert as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alert)
}

// UpdateRiskAlert handles PUT /api/risk/alerts/{id}
func (h *RiskHandler) UpdateRiskAlert(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid alert ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var alert models.RiskAlert
	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure ID in URL matches ID in body
	alert.ID = id

	// Update risk alert in repository
	if err := h.repo.UpdateRiskAlert(&alert); err != nil {
		http.Error(w, "Error updating risk alert", http.StatusInternalServerError)
		return
	}

	// Return updated alert as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alert)
}

// GetEventsForAlert handles GET /api/risk/alerts/{id}/events
func (h *RiskHandler) GetEventsForAlert(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid alert ID", http.StatusBadRequest)
		return
	}

	// Get events for alert from repository
	events, err := h.repo.GetEventsForAlert(id)
	if err != nil {
		http.Error(w, "Error retrieving events for alert", http.StatusInternalServerError)
		return
	}

	// Return events as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// DecayRiskScores handles POST /api/risk/decay
func (h *RiskHandler) DecayRiskScores(w http.ResponseWriter, r *http.Request) {
	// Decay risk scores
	if err := h.engine.DecayRiskScores(); err != nil {
		http.Error(w, "Error decaying risk scores", http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// GetHighRiskEntities handles GET /api/risk/high
func (h *RiskHandler) GetHighRiskEntities(w http.ResponseWriter, r *http.Request) {
	// Get high risk entities from engine
	entities, err := h.engine.GetHighRiskEntities()
	if err != nil {
		http.Error(w, "Error retrieving high risk entities", http.StatusInternalServerError)
		return
	}

	// Return entities as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entities)
}