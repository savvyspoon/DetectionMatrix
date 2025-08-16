package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"riskmatrix/pkg/models"
)

// DetectionClassHandler handles detection class management endpoints
type DetectionClassHandler struct {
	repo models.DetectionRepository
}

// NewDetectionClassHandler creates a new detection class handler
func NewDetectionClassHandler(repo models.DetectionRepository) *DetectionClassHandler {
	return &DetectionClassHandler{
		repo: repo,
	}
}

// ListDetectionClasses handles GET /api/detection-classes
func (h *DetectionClassHandler) ListDetectionClasses(w http.ResponseWriter, r *http.Request) {
	classes, err := h.repo.ListDetectionClasses()
	if err != nil {
		http.Error(w, "Error fetching detection classes", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(classes)
}

// GetDetectionClass handles GET /api/detection-classes/{id}
func (h *DetectionClassHandler) GetDetectionClass(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid class ID", http.StatusBadRequest)
		return
	}
	
	class, err := h.repo.GetDetectionClass(id)
	if err != nil {
		http.Error(w, "Detection class not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(class)
}

// CreateDetectionClass handles POST /api/detection-classes
func (h *DetectionClassHandler) CreateDetectionClass(w http.ResponseWriter, r *http.Request) {
	var class models.DetectionClass
	if err := json.NewDecoder(r.Body).Decode(&class); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate required fields
	if class.Name == "" {
		http.Error(w, "Class name is required", http.StatusBadRequest)
		return
	}
	
	// Set defaults
	class.IsSystem = false
	class.CreatedAt = time.Now()
	class.UpdatedAt = time.Now()
	
	// Set default display order if not provided
	if class.DisplayOrder == 0 {
		class.DisplayOrder = 999
	}
	
	// Set default color if not provided
	if class.Color == "" {
		class.Color = "#6B7280" // Default gray color
	}
	
	if err := h.repo.CreateDetectionClass(&class); err != nil {
		http.Error(w, "Error creating detection class: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(class)
}

// UpdateDetectionClass handles PUT /api/detection-classes/{id}
func (h *DetectionClassHandler) UpdateDetectionClass(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid class ID", http.StatusBadRequest)
		return
	}
	
	// Get existing class to check if it's a system class
	existing, err := h.repo.GetDetectionClass(id)
	if err != nil {
		http.Error(w, "Detection class not found", http.StatusNotFound)
		return
	}
	
	// Prevent editing system classes
	if existing.IsSystem {
		http.Error(w, "Cannot modify system detection classes", http.StatusForbidden)
		return
	}
	
	var class models.DetectionClass
	if err := json.NewDecoder(r.Body).Decode(&class); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate required fields
	if class.Name == "" {
		http.Error(w, "Class name is required", http.StatusBadRequest)
		return
	}
	
	// Ensure we're updating the correct class
	class.ID = id
	class.IsSystem = false // Ensure it remains a non-system class
	class.UpdatedAt = time.Now()
	
	if err := h.repo.UpdateDetectionClass(&class); err != nil {
		http.Error(w, "Error updating detection class: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// DeleteDetectionClass handles DELETE /api/detection-classes/{id}
func (h *DetectionClassHandler) DeleteDetectionClass(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid class ID", http.StatusBadRequest)
		return
	}
	
	// The repository method already checks if it's a system class
	if err := h.repo.DeleteDetectionClass(id); err != nil {
		if err.Error() == "cannot delete system detection class" {
			http.Error(w, err.Error(), http.StatusForbidden)
		} else if err.Error() == "detection class not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Error deleting detection class: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	w.WriteHeader(http.StatusNoContent)
}

// ListDetectionsByClass handles GET /api/detection-classes/{id}/detections
func (h *DetectionClassHandler) ListDetectionsByClass(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid class ID", http.StatusBadRequest)
		return
	}
	
	// Verify the class exists
	_, err = h.repo.GetDetectionClass(id)
	if err != nil {
		http.Error(w, "Detection class not found", http.StatusNotFound)
		return
	}
	
	detections, err := h.repo.ListDetectionsByClass(id)
	if err != nil {
		http.Error(w, "Error fetching detections by class", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detections)
}