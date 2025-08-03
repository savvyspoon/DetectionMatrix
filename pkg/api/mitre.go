package api

import (
	"encoding/json"
	"net/http"

	"riskmatrix/internal/mitre"
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
		http.Error(w, "Invalid technique ID", http.StatusBadRequest)
		return
	}

	// Get technique from repository
	technique, err := h.repo.GetMitreTechnique(id)
	if err != nil {
		http.Error(w, "Technique not found", http.StatusNotFound)
		return
	}

	// Return technique as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(technique)
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
		http.Error(w, "Error retrieving techniques", http.StatusInternalServerError)
		return
	}

	// Return techniques as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(techniques)
}

// CreateMitreTechnique handles POST /api/mitre/techniques
func (h *MitreHandler) CreateMitreTechnique(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var technique models.MitreTechnique
	if err := json.NewDecoder(r.Body).Decode(&technique); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if technique.ID == "" || technique.Tactic == "" || technique.Name == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Create technique in repository
	if err := h.repo.CreateMitreTechnique(&technique); err != nil {
		http.Error(w, "Error creating technique", http.StatusInternalServerError)
		return
	}

	// Return created technique as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(technique)
}

// UpdateMitreTechnique handles PUT /api/mitre/techniques/{id}
func (h *MitreHandler) UpdateMitreTechnique(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Invalid technique ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var technique models.MitreTechnique
	if err := json.NewDecoder(r.Body).Decode(&technique); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure ID in URL matches ID in body
	technique.ID = id

	// Update technique in repository
	if err := h.repo.UpdateMitreTechnique(&technique); err != nil {
		http.Error(w, "Error updating technique", http.StatusInternalServerError)
		return
	}

	// Return updated technique as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(technique)
}

// DeleteMitreTechnique handles DELETE /api/mitre/techniques/{id}
func (h *MitreHandler) DeleteMitreTechnique(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Invalid technique ID", http.StatusBadRequest)
		return
	}

	// Delete technique from repository
	if err := h.repo.DeleteMitreTechnique(id); err != nil {
		http.Error(w, "Error deleting technique", http.StatusInternalServerError)
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
		http.Error(w, "Error retrieving coverage", http.StatusInternalServerError)
		return
	}

	// Return coverage as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(coverage)
}

// GetDetectionsByTechnique handles GET /api/mitre/techniques/{id}/detections
func (h *MitreHandler) GetDetectionsByTechnique(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Invalid technique ID", http.StatusBadRequest)
		return
	}

	// Get detections from repository
	detections, err := h.repo.GetDetectionsByTechnique(id)
	if err != nil {
		http.Error(w, "Error retrieving detections", http.StatusInternalServerError)
		return
	}

	// Return detections as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detections)
}

// GetCoverageSummary handles GET /api/mitre/coverage/summary
func (h *MitreHandler) GetCoverageSummary(w http.ResponseWriter, r *http.Request) {
	// Get coverage summary from repository
	summary, err := h.repo.GetCoverageSummary()
	if err != nil {
		http.Error(w, "Error retrieving coverage summary", http.StatusInternalServerError)
		return
	}

	// Return summary as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}
