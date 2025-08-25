package api

import (
	"encoding/json"
	"fmt"
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
	// Parse request body with custom structure to handle entity_type and entity_value
	var requestData struct {
		DetectionID     int64              `json:"detection_id"`
		EntityID        int64              `json:"entity_id,omitempty"`
		EntityType      string             `json:"entity_type,omitempty"`
		EntityValue     string             `json:"entity_value,omitempty"`
		RiskObject      *models.RiskObject `json:"risk_object,omitempty"`
		Timestamp       time.Time          `json:"timestamp,omitempty"`
		RawData         string             `json:"raw_data,omitempty"`
		Context         string             `json:"context,omitempty"`
		RiskPoints      int                `json:"risk_points"`
		IsFalsePositive bool               `json:"is_false_positive"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Build event model
	event := models.Event{
		DetectionID:     requestData.DetectionID,
		EntityID:        requestData.EntityID,
		RawData:         requestData.RawData,
		Context:         requestData.Context,
		RiskPoints:      requestData.RiskPoints,
		IsFalsePositive: requestData.IsFalsePositive,
		Timestamp:       requestData.Timestamp,
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Handle RiskObject from different sources
	if requestData.RiskObject != nil {
		// If RiskObject is directly provided in the request
		event.RiskObject = requestData.RiskObject
	} else if requestData.EntityType != "" && requestData.EntityValue != "" {
		// If entity_type and entity_value are provided, create RiskObject
		event.RiskObject = &models.RiskObject{
			EntityType:  models.EntityType(requestData.EntityType),
			EntityValue: requestData.EntityValue,
		}
	} else if requestData.EntityID > 0 {
		// If entity_id is provided, fetch the risk object
		riskObj, err := h.repo.GetRiskObject(requestData.EntityID)
		if err != nil {
			Error(w, r, http.StatusBadRequest, "Invalid entity_id")
			return
		}
		event.EntityID = requestData.EntityID
		event.RiskObject = riskObj
	} else {
		Error(w, r, http.StatusBadRequest, "Either entity_id, risk_object, or entity_type/entity_value must be provided")
		return
	}

	// Process event
	if err := h.engine.ProcessEvent(&event); err != nil {
		Error(w, r, http.StatusInternalServerError, "Error processing event")
		return
	}

	// Return created event as JSON
	JSON(w, http.StatusCreated, event)
}

// ProcessEvents handles POST /api/events/batch
func (h *RiskHandler) ProcessEvents(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var events []*models.Event
	if err := json.NewDecoder(r.Body).Decode(&events); err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid request body")
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
		Error(w, r, http.StatusInternalServerError, "Error processing events")
		return
	}

	// Return success with count
	JSON(w, http.StatusCreated, map[string]int{"processed": len(events)})
}

// GetRiskObject handles GET /api/risk/objects/{id}
func (h *RiskHandler) GetRiskObject(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid risk object ID")
		return
	}

	// Get risk object from repository
	obj, err := h.repo.GetRiskObject(id)
	if err != nil {
		Error(w, r, http.StatusNotFound, "Risk object not found")
		return
	}

	// Return risk object as JSON
	JSON(w, http.StatusOK, obj)
}

// GetRiskObjectByEntity handles GET /api/risk/objects/entity
func (h *RiskHandler) GetRiskObjectByEntity(w http.ResponseWriter, r *http.Request) {
	// Extract entity type and value from query parameters
	entityType := r.URL.Query().Get("type")
	entityValue := r.URL.Query().Get("value")

	if entityType == "" || entityValue == "" {
		Error(w, r, http.StatusBadRequest, "Missing entity type or value")
		return
	}

	// Get risk object from repository
	obj, err := h.repo.GetRiskObjectByEntity(models.EntityType(entityType), entityValue)
	if err != nil {
		Error(w, r, http.StatusNotFound, "Risk object not found")
		return
	}

	// Return risk object as JSON
	JSON(w, http.StatusOK, obj)
}

// ListRiskObjects handles GET /api/risk/objects
func (h *RiskHandler) ListRiskObjects(w http.ResponseWriter, r *http.Request) {
	// Check for high risk filter
	thresholdStr := r.URL.Query().Get("threshold")
	limitStr := r.URL.Query().Get("limit")

	var objects []*models.RiskObject
	var err error

	if thresholdStr != "" {
		// Parse threshold
		threshold, err := strconv.Atoi(thresholdStr)
		if err != nil {
			Error(w, r, http.StatusBadRequest, "Invalid threshold")
			return
		}

		// List high risk objects
		objects, err = h.repo.ListHighRiskObjects(threshold)
	} else {
		// List all risk objects
		objects, err = h.repo.ListRiskObjects()
	}

	if err != nil {
		Error(w, r, http.StatusInternalServerError, "Error retrieving risk objects")
		return
	}

	// Apply limit if provided
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			Error(w, r, http.StatusBadRequest, "Invalid limit")
			return
		}
		if limit > 0 && len(objects) > limit {
			objects = objects[:limit]
		}
	}

	// Return risk objects using the standard list envelope
	List(w, objects, 1, len(objects), len(objects))
}

// GetEvent handles GET /api/events/{id}
func (h *RiskHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid event ID")
		return
	}

	// Get event from repository
	event, err := h.repo.GetEvent(id)
	if err != nil {
		Error(w, r, http.StatusNotFound, "Event not found")
		return
	}

	// Return event as JSON
	JSON(w, http.StatusOK, event)
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
		Error(w, r, http.StatusInternalServerError, "Error retrieving events")
		return
	}

	// Return events using the standard list envelope
	List(w, events, page, limit, totalCount)
}

// ListEventsByEntity handles GET /api/events/entity/{id}
func (h *RiskHandler) ListEventsByEntity(w http.ResponseWriter, r *http.Request) {
	// Extract entity ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid entity ID")
		return
	}

	// Get events from repository
	events, err := h.repo.ListEventsByEntity(id)
	if err != nil {
		Error(w, r, http.StatusInternalServerError, "Error retrieving events")
		return
	}

	// Return events as JSON
	JSON(w, http.StatusOK, events)
}

// MarkEventAsFalsePositive handles POST /api/events/{id}/false-positive
func (h *RiskHandler) MarkEventAsFalsePositive(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid event ID")
		return
	}

	// Parse request body
	var fpInfo models.FalsePositive
	if err := json.NewDecoder(r.Body).Decode(&fpInfo); err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Set timestamp if not provided
	if fpInfo.Timestamp.IsZero() {
		fpInfo.Timestamp = time.Now()
	}

	// Mark event as false positive
	if err := h.engine.MarkEventAsFalsePositive(id, &fpInfo); err != nil {
		Error(w, r, http.StatusInternalServerError, "Error marking event as false positive")
		return
	}

	// Return success message
	JSON(w, http.StatusOK, map[string]string{"message": "Event marked as false positive successfully"})
}

// UnmarkEventAsFalsePositive handles DELETE /api/events/{id}/false-positive
func (h *RiskHandler) UnmarkEventAsFalsePositive(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid event ID")
		return
	}

	// Unmark event as false positive
	if err := h.engine.UnmarkEventAsFalsePositive(id); err != nil {
		Error(w, r, http.StatusInternalServerError, "Error unmarking event as false positive")
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

	// Always use paginated method for a consistent envelope
	alerts, totalCount, err := h.engine.GetRiskAlertsPaginated(limit, offset, models.AlertStatus(statusFilter))
	if err != nil {
		Error(w, r, http.StatusInternalServerError, "Error retrieving risk alerts")
		return
	}

	// Return alerts using the standard list envelope
	List(w, alerts, page, limit, totalCount)
}

// GetRiskAlert handles GET /api/risk/alerts/{id}
func (h *RiskHandler) GetRiskAlert(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid alert ID")
		return
	}

	// Get risk alert from repository
	alert, err := h.repo.GetRiskAlert(id)
	if err != nil {
		Error(w, r, http.StatusInternalServerError, "Error retrieving risk alert")
		return
	}

	// Return risk alert as JSON
	JSON(w, http.StatusOK, alert)
}

// UpdateRiskAlert handles PUT /api/risk/alerts/{id}
func (h *RiskHandler) UpdateRiskAlert(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid alert ID")
		return
	}

	// Parse request body
	var alert models.RiskAlert
	if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Ensure ID in URL matches ID in body
	alert.ID = id

	// Update risk alert in repository
	if err := h.repo.UpdateRiskAlert(&alert); err != nil {
		Error(w, r, http.StatusInternalServerError, "Error updating risk alert")
		return
	}

	// Return updated alert as JSON
	JSON(w, http.StatusOK, alert)
}

// GetEventsForAlert handles GET /api/risk/alerts/{id}/events
func (h *RiskHandler) GetEventsForAlert(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid alert ID")
		return
	}

	// Get events for alert from repository
	events, err := h.repo.GetEventsForAlert(id)
	if err != nil {
		// If alert not found, return empty list
		if err.Error() == fmt.Sprintf("risk alert not found: %d", id) {
			JSON(w, http.StatusOK, []*models.Event{})
			return
		}
		Error(w, r, http.StatusInternalServerError, "Error retrieving events for alert")
		return
	}

	// Return events as JSON
	JSON(w, http.StatusOK, events)
}

// DecayRiskScores handles POST /api/risk/decay
func (h *RiskHandler) DecayRiskScores(w http.ResponseWriter, r *http.Request) {
	// Decay risk scores
	if err := h.engine.DecayRiskScores(); err != nil {
		Error(w, r, http.StatusInternalServerError, "Error decaying risk scores")
		return
	}

	// Return success message
	JSON(w, http.StatusOK, map[string]string{"message": "Risk scores decayed successfully"})
}

// GetHighRiskEntities handles GET /api/risk/high
func (h *RiskHandler) GetHighRiskEntities(w http.ResponseWriter, r *http.Request) {
	// Get high risk entities from engine
	entities, err := h.engine.GetHighRiskEntities()
	if err != nil {
		Error(w, r, http.StatusInternalServerError, "Error retrieving high risk entities")
		return
	}

	// Return entities as JSON
	JSON(w, http.StatusOK, entities)
}
