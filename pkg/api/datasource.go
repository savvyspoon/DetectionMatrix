package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"riskmatrix/internal/datasource"
	"riskmatrix/pkg/models"
)

// DataSourceHandler handles HTTP requests for data source endpoints
type DataSourceHandler struct {
	repo *datasource.Repository
}

// NewDataSourceHandler creates a new data source handler
func NewDataSourceHandler(repo *datasource.Repository) *DataSourceHandler {
	return &DataSourceHandler{repo: repo}
}

// GetDataSource handles GET /api/datasources/{id}
func (h *DataSourceHandler) GetDataSource(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid data source ID", http.StatusBadRequest)
		return
	}

	// Get data source from repository
	dataSource, err := h.repo.GetDataSource(id)
	if err != nil {
		http.Error(w, "Data source not found", http.StatusNotFound)
		return
	}

	// Return data source as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dataSource)
}

// GetDataSourceByName handles GET /api/datasources/by-name/{name}
func (h *DataSourceHandler) GetDataSourceByName(w http.ResponseWriter, r *http.Request) {
	// Extract name from URL path
	name := r.PathValue("name")
	if name == "" {
		http.Error(w, "Invalid data source name", http.StatusBadRequest)
		return
	}

	// Get data source from repository
	dataSource, err := h.repo.GetDataSourceByName(name)
	if err != nil {
		http.Error(w, "Data source not found", http.StatusNotFound)
		return
	}

	// Return data source as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dataSource)
}

// ListDataSources handles GET /api/datasources
func (h *DataSourceHandler) ListDataSources(w http.ResponseWriter, r *http.Request) {
	// Get data sources from repository
	dataSources, err := h.repo.ListDataSources()
	if err != nil {
		http.Error(w, "Error retrieving data sources", http.StatusInternalServerError)
		return
	}

	// Return data sources as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dataSources)
}

// CreateDataSource handles POST /api/datasources
func (h *DataSourceHandler) CreateDataSource(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var dataSource models.DataSource
	if err := json.NewDecoder(r.Body).Decode(&dataSource); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if dataSource.Name == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Create data source in repository
	if err := h.repo.CreateDataSource(&dataSource); err != nil {
		http.Error(w, "Error creating data source", http.StatusInternalServerError)
		return
	}

	// Return created data source as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dataSource)
}

// UpdateDataSource handles PUT /api/datasources/{id}
func (h *DataSourceHandler) UpdateDataSource(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid data source ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var dataSource models.DataSource
	if err := json.NewDecoder(r.Body).Decode(&dataSource); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Ensure ID in URL matches ID in body
	dataSource.ID = id

	// Update data source in repository
	if err := h.repo.UpdateDataSource(&dataSource); err != nil {
		http.Error(w, "Error updating data source", http.StatusInternalServerError)
		return
	}

	// Return updated data source as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dataSource)
}

// DeleteDataSource handles DELETE /api/datasources/{id}
func (h *DataSourceHandler) DeleteDataSource(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid data source ID", http.StatusBadRequest)
		return
	}

	// Delete data source from repository
	if err := h.repo.DeleteDataSource(id); err != nil {
		http.Error(w, "Error deleting data source", http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// GetDetectionsByDataSource handles GET /api/datasources/id/{id}/detections
func (h *DataSourceHandler) GetDetectionsByDataSource(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid data source ID", http.StatusBadRequest)
		return
	}

	// Get detections from repository
	detections, err := h.repo.GetDetectionsByDataSource(id)
	if err != nil {
		http.Error(w, "Error retrieving detections", http.StatusInternalServerError)
		return
	}

	// Return detections as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(detections)
}

// GetDataSourceUtilization handles GET /api/datasources/utilization
func (h *DataSourceHandler) GetDataSourceUtilization(w http.ResponseWriter, r *http.Request) {
	// Get data source utilization from repository
	utilization, err := h.repo.GetDataSourceUtilization()
	if err != nil {
		http.Error(w, "Error retrieving data source utilization", http.StatusInternalServerError)
		return
	}

	// Return utilization as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(utilization)
}

// GetMitreTechniquesByDataSource handles GET /api/datasources/id/{id}/techniques
func (h *DataSourceHandler) GetMitreTechniquesByDataSource(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid data source ID", http.StatusBadRequest)
		return
	}

	// Get MITRE techniques from repository
	techniques, err := h.repo.GetMitreTechniquesByDataSource(id)
	if err != nil {
		http.Error(w, "Error retrieving MITRE techniques", http.StatusInternalServerError)
		return
	}

	// Return techniques as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(techniques)
}