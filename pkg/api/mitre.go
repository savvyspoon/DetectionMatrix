package api

import (
	"encoding/json"
	"net/http"

	"riskmatrix/internal/mitre"
	validation "riskmatrix/pkg"
	"riskmatrix/pkg/models"
)

// MitreHandler handles HTTP requests for MITRE ATT&CK endpoints
type MitreHandler struct {
	repo *mitre.Repository
}

// NewMitreHandler creates a new MITRE handler
func NewMitreHandler(repo *mitre.Repository) *MitreHandler {
	return &MitreHandler{repo: repo}
}

// GetMitreTechnique handles GET /api/mitre/techniques/{id}
func (h *MitreHandler) GetMitreTechnique(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	id := r.PathValue("id")
	if id == "" {
		Error(w, r, http.StatusBadRequest, "Invalid technique ID")
		return
	}

	// Get technique from repository
	technique, err := h.repo.GetMitreTechnique(id)
	if err != nil {
		Error(w, r, http.StatusNotFound, "Technique not found")
		return
	}

	// Return technique as JSON
	JSON(w, http.StatusOK, technique)
}

// ListMitreTechniques handles GET /api/mitre/techniques
func (h *MitreHandler) ListMitreTechniques(w http.ResponseWriter, r *http.Request) {
	// Check for tactic filter
	tactic := r.URL.Query().Get("tactic")

	var techniques []*models.MitreTechnique
	var err error

	if tactic != "" {
		// List techniques by tactic
		techniques, err = h.repo.ListMitreTechniquesByTactic(tactic)
	} else {
		// List all techniques
		techniques, err = h.repo.ListMitreTechniques()
	}

	if err != nil {
		Error(w, r, http.StatusInternalServerError, "Error retrieving techniques")
		return
	}

	// Return techniques using the standard list envelope
	List(w, techniques, 1, len(techniques), len(techniques))
}

// CreateMitreTechnique handles POST /api/mitre/techniques
func (h *MitreHandler) CreateMitreTechnique(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var technique models.MitreTechnique
	if err := json.NewDecoder(r.Body).Decode(&technique); err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate payload
	if err := validation.ValidateMitreTechnique(&technique); err != nil {
		Error(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// Create technique in repository
	if err := h.repo.CreateMitreTechnique(&technique); err != nil {
		Error(w, r, http.StatusInternalServerError, "Error creating technique")
		return
	}

	// Return created technique as JSON
	JSON(w, http.StatusCreated, technique)
}

// UpdateMitreTechnique handles PUT /api/mitre/techniques/{id}
func (h *MitreHandler) UpdateMitreTechnique(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	id := r.PathValue("id")
	if id == "" {
		Error(w, r, http.StatusBadRequest, "Invalid technique ID")
		return
	}

	// Parse request body
	var technique models.MitreTechnique
	if err := json.NewDecoder(r.Body).Decode(&technique); err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Ensure ID in URL matches ID in body
	technique.ID = id

	// Validate payload
	if err := validation.ValidateMitreTechnique(&technique); err != nil {
		Error(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// Update technique in repository
	if err := h.repo.UpdateMitreTechnique(&technique); err != nil {
		Error(w, r, http.StatusInternalServerError, "Error updating technique")
		return
	}

	// Return updated technique as JSON
	JSON(w, http.StatusOK, technique)
}

// DeleteMitreTechnique handles DELETE /api/mitre/techniques/{id}
func (h *MitreHandler) DeleteMitreTechnique(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	id := r.PathValue("id")
	if id == "" {
		Error(w, r, http.StatusBadRequest, "Invalid technique ID")
		return
	}

	// Delete technique from repository
	if err := h.repo.DeleteMitreTechnique(id); err != nil {
		Error(w, r, http.StatusInternalServerError, "Error deleting technique")
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// GetCoverageByTactic handles GET /api/mitre/coverage
func (h *MitreHandler) GetCoverageByTactic(w http.ResponseWriter, r *http.Request) {
	// Get coverage by tactic from repository
	coverage, err := h.repo.GetCoverageByTactic()
	if err != nil {
		Error(w, r, http.StatusInternalServerError, "Error retrieving coverage")
		return
	}

	// Return coverage as JSON
	JSON(w, http.StatusOK, coverage)
}

// GetDetectionsByTechnique handles GET /api/mitre/techniques/{id}/detections
func (h *MitreHandler) GetDetectionsByTechnique(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	id := r.PathValue("id")
	if id == "" {
		Error(w, r, http.StatusBadRequest, "Invalid technique ID")
		return
	}

	// Get detections from repository
	detections, err := h.repo.GetDetectionsByTechnique(id)
	if err != nil {
		Error(w, r, http.StatusInternalServerError, "Error retrieving detections")
		return
	}

	// Return detections as JSON
	JSON(w, http.StatusOK, detections)
}

// GetCoverageSummary handles GET /api/mitre/coverage/summary
func (h *MitreHandler) GetCoverageSummary(w http.ResponseWriter, r *http.Request) {
	// Get coverage summary from repository
	summary, err := h.repo.GetCoverageSummary()
	if err != nil {
		Error(w, r, http.StatusInternalServerError, "Error retrieving coverage summary")
		return
	}

	// Return summary as JSON
	JSON(w, http.StatusOK, summary)
}
