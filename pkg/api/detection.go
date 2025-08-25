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
		Error(w, r, http.StatusBadRequest, "Invalid detection ID")
		return
	}

	// Get detection from repository
	detection, err := h.repo.GetDetection(id)
	if err != nil {
		Error(w, r, http.StatusNotFound, "Detection not found")
		return
	}

	// Return detection as JSON
	JSON(w, http.StatusOK, detection)
}

// ListDetections handles GET /api/detections
func (h *DetectionHandler) ListDetections(w http.ResponseWriter, r *http.Request) {
	// Check for filters
	status := r.URL.Query().Get("status")
	classIDStr := r.URL.Query().Get("class_id")

	var detections []*models.Detection
	var err error

	// Apply filters based on query parameters
	if classIDStr != "" {
		// Filter by class ID
		classID, parseErr := strconv.ParseInt(classIDStr, 10, 64)
		if parseErr != nil {
			Error(w, r, http.StatusBadRequest, "Invalid class_id parameter")
			return
		}
		detections, err = h.repo.ListDetectionsByClass(classID)
	} else if status != "" {
		// Filter by status
		detections, err = h.repo.ListDetectionsByStatus(models.DetectionStatus(status))
	} else {
		// List all detections
		detections, err = h.repo.ListDetections()
	}

	if err != nil {
		Error(w, r, http.StatusInternalServerError, "Error retrieving detections")
		return
	}

	// Return detections using the standard list envelope
	List(w, detections, 1, len(detections), len(detections))
}

// CreateDetection handles POST /api/detections
func (h *DetectionHandler) CreateDetection(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var detection models.Detection
	if err := json.NewDecoder(r.Body).Decode(&detection); err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Create detection in repository
	if err := h.repo.CreateDetection(&detection); err != nil {
		Error(w, r, http.StatusInternalServerError, "Error creating detection")
		return
	}

	// Return created detection as JSON
	JSON(w, http.StatusCreated, detection)
}

// UpdateDetection handles PUT /api/detections/{id}
func (h *DetectionHandler) UpdateDetection(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid detection ID")
		return
	}

	// Define a struct to capture the update request including relationship IDs
	var updateRequest struct {
		models.Detection
		DataSourceIDs     []int64 `json:"data_source_ids,omitempty"`
		MitreTechniqueIDs []string `json:"mitre_technique_ids,omitempty"`
	}

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Ensure ID in URL matches
	updateRequest.Detection.ID = id

	// Update detection in repository
	if err := h.repo.UpdateDetection(&updateRequest.Detection); err != nil {
		Error(w, r, http.StatusInternalServerError, "Error updating detection")
		return
	}

	// Update data source relationships if provided
	if updateRequest.DataSourceIDs != nil {
		// Get current data sources
		currentDetection, err := h.repo.GetDetection(id)
		if err != nil {
			Error(w, r, http.StatusInternalServerError, "Error retrieving current detection")
			return
		}

		// Build maps for efficient lookup
		currentDSMap := make(map[int64]bool)
		for _, ds := range currentDetection.DataSources {
			currentDSMap[ds.ID] = true
		}

		newDSMap := make(map[int64]bool)
		for _, dsID := range updateRequest.DataSourceIDs {
			newDSMap[dsID] = true
		}

		// Remove data sources that are no longer in the list
		for _, ds := range currentDetection.DataSources {
			if !newDSMap[ds.ID] {
				if err := h.repo.RemoveDataSource(id, ds.ID); err != nil {
					// Log error but continue
					continue
				}
			}
		}

		// Add new data sources
		for _, dsID := range updateRequest.DataSourceIDs {
			if !currentDSMap[dsID] {
				if err := h.repo.AddDataSource(id, dsID); err != nil {
					// Log error but continue
					continue
				}
			}
		}
	}

	// Update MITRE technique relationships if provided
	if updateRequest.MitreTechniqueIDs != nil {
		// Get current techniques
		currentDetection, err := h.repo.GetDetection(id)
		if err != nil {
			Error(w, r, http.StatusInternalServerError, "Error retrieving current detection")
			return
		}

		// Build maps for efficient lookup
		currentTechMap := make(map[string]bool)
		for _, tech := range currentDetection.MitreTechniques {
			currentTechMap[tech.ID] = true
		}

		newTechMap := make(map[string]bool)
		for _, techID := range updateRequest.MitreTechniqueIDs {
			newTechMap[techID] = true
		}

		// Remove techniques that are no longer in the list
		for _, tech := range currentDetection.MitreTechniques {
			if !newTechMap[tech.ID] {
				if err := h.repo.RemoveMitreTechnique(id, tech.ID); err != nil {
					// Log error but continue
					continue
				}
			}
		}

		// Add new techniques
		for _, techID := range updateRequest.MitreTechniqueIDs {
			if !currentTechMap[techID] {
				if err := h.repo.AddMitreTechnique(id, techID); err != nil {
					// Log error but continue
					continue
				}
			}
		}
	}

	// Get the updated detection with all relationships
	updatedDetection, err := h.repo.GetDetection(id)
	if err != nil {
		Error(w, r, http.StatusInternalServerError, "Error retrieving updated detection")
		return
	}

	// Return updated detection as JSON
	JSON(w, http.StatusOK, updatedDetection)
}

// DeleteDetection handles DELETE /api/detections/{id}
func (h *DetectionHandler) DeleteDetection(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid detection ID")
		return
	}

	// Delete detection from repository
	if err := h.repo.DeleteDetection(id); err != nil {
		Error(w, r, http.StatusInternalServerError, "Error deleting detection")
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
		Error(w, r, http.StatusInternalServerError, "Error retrieving detection count")
		return
	}

	// Return count as JSON
	JSON(w, http.StatusOK, map[string]int{"count": count})
}

// GetDetectionCountByStatus handles GET /api/detections/count/status
func (h *DetectionHandler) GetDetectionCountByStatus(w http.ResponseWriter, r *http.Request) {
	// Get detection count by status from repository
	counts, err := h.repo.GetDetectionCountByStatus()
	if err != nil {
		Error(w, r, http.StatusInternalServerError, "Error retrieving detection counts")
		return
	}

	// Return counts as JSON
	JSON(w, http.StatusOK, counts)
}

// GetFalsePositiveRate handles GET /api/detections/{id}/fp-rate
func (h *DetectionHandler) GetFalsePositiveRate(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid detection ID")
		return
	}

	// Get false positive rate from repository
	rate, err := h.repo.GetFalsePositiveRate(id)
	if err != nil {
		Error(w, r, http.StatusInternalServerError, "Error retrieving false positive rate")
		return
	}

	// Return rate as JSON
	JSON(w, http.StatusOK, map[string]float64{"false_positive_rate": rate})
}

// GetEventCountLast30Days handles GET /api/detections/{id}/events/count/30days
func (h *DetectionHandler) GetEventCountLast30Days(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid detection ID")
		return
	}

	// Get event count from repository
	count, err := h.repo.GetEventCountLast30Days(id)
	if err != nil {
		Error(w, r, http.StatusInternalServerError, "Error retrieving event count")
		return
	}

	// Return count as JSON
	JSON(w, http.StatusOK, map[string]int{"count": count})
}

// GetFalsePositivesLast30Days handles GET /api/detections/{id}/false-positives/count/30days
func (h *DetectionHandler) GetFalsePositivesLast30Days(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid detection ID")
		return
	}

	// Get false positive count from repository
	count, err := h.repo.GetFalsePositivesLast30Days(id)
	if err != nil {
		Error(w, r, http.StatusInternalServerError, "Error retrieving false positive count")
		return
	}

	// Return count as JSON
	JSON(w, http.StatusOK, map[string]int{"count": count})
}

// AddMitreTechnique handles POST /api/detections/{id}/mitre/{technique_id}
func (h *DetectionHandler) AddMitreTechnique(w http.ResponseWriter, r *http.Request) {
	// Extract detection ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid detection ID")
		return
	}

	// Extract technique ID from URL path
	techniqueID := r.PathValue("technique_id")
	if techniqueID == "" {
		Error(w, r, http.StatusBadRequest, "Invalid technique ID")
		return
	}

	// Add technique to detection
	if err := h.repo.AddMitreTechnique(id, techniqueID); err != nil {
		Error(w, r, http.StatusInternalServerError, "Error adding MITRE technique")
		return
	}

	// Return success message
	JSON(w, http.StatusOK, map[string]string{"message": "MITRE technique added successfully"})
}

// RemoveMitreTechnique handles DELETE /api/detections/{id}/mitre/{technique_id}
func (h *DetectionHandler) RemoveMitreTechnique(w http.ResponseWriter, r *http.Request) {
	// Extract detection ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid detection ID")
		return
	}

	// Extract technique ID from URL path
	techniqueID := r.PathValue("technique_id")
	if techniqueID == "" {
		Error(w, r, http.StatusBadRequest, "Invalid technique ID")
		return
	}

	// Remove technique from detection
	if err := h.repo.RemoveMitreTechnique(id, techniqueID); err != nil {
		Error(w, r, http.StatusInternalServerError, "Error removing MITRE technique")
		return
	}

	// Return success message
	JSON(w, http.StatusOK, map[string]string{"message": "MITRE technique removed successfully"})
}

// AddDataSource handles POST /api/detections/{id}/datasource/{datasource_id}
func (h *DetectionHandler) AddDataSource(w http.ResponseWriter, r *http.Request) {
	// Extract detection ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid detection ID")
		return
	}

	// Extract data source ID from URL path
	dataSourceIDStr := r.PathValue("datasource_id")
	dataSourceID, err := strconv.ParseInt(dataSourceIDStr, 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid data source ID")
		return
	}

	// Add data source to detection
	if err := h.repo.AddDataSource(id, dataSourceID); err != nil {
		Error(w, r, http.StatusInternalServerError, "Error adding data source")
		return
	}

	// Return success message
	JSON(w, http.StatusOK, map[string]string{"message": "Data source added successfully"})
}

// RemoveDataSource handles DELETE /api/detections/{id}/datasource/{datasource_id}
func (h *DetectionHandler) RemoveDataSource(w http.ResponseWriter, r *http.Request) {
	// Extract detection ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid detection ID")
		return
	}

	// Extract data source ID from URL path
	dataSourceIDStr := r.PathValue("datasource_id")
	dataSourceID, err := strconv.ParseInt(dataSourceIDStr, 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid data source ID")
		return
	}

	// Remove data source from detection
	if err := h.repo.RemoveDataSource(id, dataSourceID); err != nil {
		Error(w, r, http.StatusInternalServerError, "Error removing data source")
		return
	}

	// Return success message
	JSON(w, http.StatusOK, map[string]string{"message": "Data source removed successfully"})
}
