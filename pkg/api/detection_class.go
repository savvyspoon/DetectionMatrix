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
		Error(w, r, http.StatusInternalServerError, "Error fetching detection classes")
		return
	}

	JSON(w, http.StatusOK, classes)
}

// GetDetectionClass handles GET /api/detection-classes/{id}
func (h *DetectionClassHandler) GetDetectionClass(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid class ID")
		return
	}

	class, err := h.repo.GetDetectionClass(id)
	if err != nil {
		Error(w, r, http.StatusNotFound, "Detection class not found")
		return
	}

	JSON(w, http.StatusOK, class)
}

// CreateDetectionClass handles POST /api/detection-classes
func (h *DetectionClassHandler) CreateDetectionClass(w http.ResponseWriter, r *http.Request) {
	var class models.DetectionClass
	if err := json.NewDecoder(r.Body).Decode(&class); err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if class.Name == "" {
		Error(w, r, http.StatusBadRequest, "Class name is required")
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
		Error(w, r, http.StatusInternalServerError, "Error creating detection class")
		return
	}

	JSON(w, http.StatusCreated, class)
}

// UpdateDetectionClass handles PUT /api/detection-classes/{id}
func (h *DetectionClassHandler) UpdateDetectionClass(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid class ID")
		return
	}

	// Get existing class to check if it's a system class
	existing, err := h.repo.GetDetectionClass(id)
	if err != nil {
		Error(w, r, http.StatusNotFound, "Detection class not found")
		return
	}

	// Prevent editing system classes
	if existing.IsSystem {
		Error(w, r, http.StatusForbidden, "Cannot modify system detection classes")
		return
	}

	var class models.DetectionClass
	if err := json.NewDecoder(r.Body).Decode(&class); err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if class.Name == "" {
		Error(w, r, http.StatusBadRequest, "Class name is required")
		return
	}

	// Ensure we're updating the correct class
	class.ID = id
	class.IsSystem = false // Ensure it remains a non-system class
	class.UpdatedAt = time.Now()

	if err := h.repo.UpdateDetectionClass(&class); err != nil {
		Error(w, r, http.StatusInternalServerError, "Error updating detection class")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteDetectionClass handles DELETE /api/detection-classes/{id}
func (h *DetectionClassHandler) DeleteDetectionClass(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid class ID")
		return
	}

	// The repository method already checks if it's a system class
	if err := h.repo.DeleteDetectionClass(id); err != nil {
		if err.Error() == "cannot delete system detection class" {
			Error(w, r, http.StatusForbidden, err.Error())
		} else if err.Error() == "detection class not found" {
			Error(w, r, http.StatusNotFound, err.Error())
		} else {
			Error(w, r, http.StatusInternalServerError, "Error deleting detection class")
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
		Error(w, r, http.StatusBadRequest, "Invalid class ID")
		return
	}

	// Verify the class exists
	_, err = h.repo.GetDetectionClass(id)
	if err != nil {
		Error(w, r, http.StatusNotFound, "Detection class not found")
		return
	}

	detections, err := h.repo.ListDetectionsByClass(id)
	if err != nil {
		Error(w, r, http.StatusInternalServerError, "Error fetching detections by class")
		return
	}

	JSON(w, http.StatusOK, detections)
}
