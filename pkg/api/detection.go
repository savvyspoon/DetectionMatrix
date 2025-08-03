package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"riskmatrix/internal/detection"
	"riskmatrix/pkg/models"
)

// DetectionHandler handles HTTP requests for detection endpoints
type DetectionHandler struct {
	repo *detection.Repository
}

// NewDetectionHandler creates a new detection handler
func NewDetectionHandler(repo *detection.Repository) *DetectionHandler {
	return &DetectionHandler{repo: repo}
}

// GetDetection handles GET /api/detections/{id}
func (h *DetectionHandler) GetDetection(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid detection ID", http.StatusBadRequest)
		return
	}

	// Get detection from repository
	detection, err := h.repo.GetDetection(id)
	if err != nil {
		http.Error(w, "Detection not found", http.StatusNotFound)
		return
	}

	// Return detection as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detection)
}

// ListDetections handles GET /api/detections
func (h *DetectionHandler) ListDetections(w http.ResponseWriter, r *http.Request) {
	// Check for status filter
	status := r.URL.Query().Get("status")

	var detections []*models.Detection
	var err error

	if status != "" {
		// List detections by status
		detections, err = h.repo.ListDetectionsByStatus(models.DetectionStatus(status))
	} else {
		// List all detections
		detections, err = h.repo.ListDetections()
	}

	if err != nil {
		http.Error(w, "Error retrieving detections", http.StatusInternalServerError)
		return
	}

	// Return detections as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detections)
}

// CreateDetection handles POST /api/detections
func (h *DetectionHandler) CreateDetection(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var detection models.Detection
	if err := json.NewDecoder(r.Body).Decode(&detection); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create detection in repository
	if err := h.repo.CreateDetection(&detection); err != nil {
		http.Error(w, "Error creating detection", http.StatusInternalServerError)
		return
	}

	// Return created detection as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(detection)
}

// UpdateDetection handles PUT /api/detections/{id}
func (h *DetectionHandler) UpdateDetection(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid detection ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var detection models.Detection
	if err := json.NewDecoder(r.Body).Decode(&detection); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure ID in URL matches ID in body
	detection.ID = id

	// Update detection in repository
	if err := h.repo.UpdateDetection(&detection); err != nil {
		http.Error(w, "Error updating detection", http.StatusInternalServerError)
		return
	}

	// Return updated detection as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detection)
}

// DeleteDetection handles DELETE /api/detections/{id}
func (h *DetectionHandler) DeleteDetection(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid detection ID", http.StatusBadRequest)
		return
	}

	// Delete detection from repository
	if err := h.repo.DeleteDetection(id); err != nil {
		http.Error(w, "Error deleting detection", http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// GetDetectionCount handles GET /api/detections/count
func (h *DetectionHandler) GetDetectionCount(w http.ResponseWriter, r *http.Request) {
	// Get detection count from repository
	count, err := h.repo.GetDetectionCount()
	if err != nil {
		http.Error(w, "Error retrieving detection count", http.StatusInternalServerError)
		return
	}

	// Return count as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"count": count})
}

// GetDetectionCountByStatus handles GET /api/detections/count/status
func (h *DetectionHandler) GetDetectionCountByStatus(w http.ResponseWriter, r *http.Request) {
	// Get detection count by status from repository
	counts, err := h.repo.GetDetectionCountByStatus()
	if err != nil {
		http.Error(w, "Error retrieving detection counts", http.StatusInternalServerError)
		return
	}

	// Return counts as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(counts)
}

// GetFalsePositiveRate handles GET /api/detections/{id}/fp-rate
func (h *DetectionHandler) GetFalsePositiveRate(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid detection ID", http.StatusBadRequest)
		return
	}

	// Get false positive rate from repository
	rate, err := h.repo.GetFalsePositiveRate(id)
	if err != nil {
		http.Error(w, "Error retrieving false positive rate", http.StatusInternalServerError)
		return
	}

	// Return rate as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{"rate": rate})
}

// GetEventCountLast30Days handles GET /api/detections/{id}/events/count/30days
func (h *DetectionHandler) GetEventCountLast30Days(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid detection ID", http.StatusBadRequest)
		return
	}

	// Get event count from repository
	count, err := h.repo.GetEventCountLast30Days(id)
	if err != nil {
		http.Error(w, "Error retrieving event count", http.StatusInternalServerError)
		return
	}

	// Return count as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"count": count})
}

// GetFalsePositivesLast30Days handles GET /api/detections/{id}/false-positives/count/30days
func (h *DetectionHandler) GetFalsePositivesLast30Days(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid detection ID", http.StatusBadRequest)
		return
	}

	// Get false positive count from repository
	count, err := h.repo.GetFalsePositivesLast30Days(id)
	if err != nil {
		http.Error(w, "Error retrieving false positive count", http.StatusInternalServerError)
		return
	}

	// Return count as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"count": count})
}

// AddMitreTechnique handles POST /api/detections/{id}/mitre/{technique_id}
func (h *DetectionHandler) AddMitreTechnique(w http.ResponseWriter, r *http.Request) {
	// Extract detection ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid detection ID", http.StatusBadRequest)
		return
	}

	// Extract technique ID from URL path
	techniqueID := r.PathValue("technique_id")
	if techniqueID == "" {
		http.Error(w, "Invalid technique ID", http.StatusBadRequest)
		return
	}

	// Add technique to detection
	if err := h.repo.AddMitreTechnique(id, techniqueID); err != nil {
		http.Error(w, "Error adding MITRE technique", http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// RemoveMitreTechnique handles DELETE /api/detections/{id}/mitre/{technique_id}
func (h *DetectionHandler) RemoveMitreTechnique(w http.ResponseWriter, r *http.Request) {
	// Extract detection ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid detection ID", http.StatusBadRequest)
		return
	}

	// Extract technique ID from URL path
	techniqueID := r.PathValue("technique_id")
	if techniqueID == "" {
		http.Error(w, "Invalid technique ID", http.StatusBadRequest)
		return
	}

	// Remove technique from detection
	if err := h.repo.RemoveMitreTechnique(id, techniqueID); err != nil {
		http.Error(w, "Error removing MITRE technique", http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// AddDataSource handles POST /api/detections/{id}/datasource/{datasource_id}
func (h *DetectionHandler) AddDataSource(w http.ResponseWriter, r *http.Request) {
	// Extract detection ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid detection ID", http.StatusBadRequest)
		return
	}

	// Extract data source ID from URL path
	dataSourceIDStr := r.PathValue("datasource_id")
	dataSourceID, err := strconv.ParseInt(dataSourceIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid data source ID", http.StatusBadRequest)
		return
	}

	// Add data source to detection
	if err := h.repo.AddDataSource(id, dataSourceID); err != nil {
		http.Error(w, "Error adding data source", http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// RemoveDataSource handles DELETE /api/detections/{id}/datasource/{datasource_id}
func (h *DetectionHandler) RemoveDataSource(w http.ResponseWriter, r *http.Request) {
	// Extract detection ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid detection ID", http.StatusBadRequest)
		return
	}

	// Extract data source ID from URL path
	dataSourceIDStr := r.PathValue("datasource_id")
	dataSourceID, err := strconv.ParseInt(dataSourceIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid data source ID", http.StatusBadRequest)
		return
	}

	// Remove data source from detection
	if err := h.repo.RemoveDataSource(id, dataSourceID); err != nil {
		http.Error(w, "Error removing data source", http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}