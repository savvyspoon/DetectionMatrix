package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"riskmatrix/internal/datasource"
	validation "riskmatrix/pkg"
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
		Error(w, r, http.StatusBadRequest, "Invalid data source ID")
		return
	}

	// Get data source from repository
	dataSource, err := h.repo.GetDataSource(id)
	if err != nil {
		Error(w, r, http.StatusNotFound, "Data source not found")
		return
	}

	// Return data source as JSON
	JSON(w, http.StatusOK, dataSource)
}

// GetDataSourceByName handles GET /api/datasources/lookup?name=<name>
func (h *DataSourceHandler) GetDataSourceByName(w http.ResponseWriter, r *http.Request) {
	// Extract name from query parameters
	name := r.URL.Query().Get("name")
	if name == "" {
		Error(w, r, http.StatusBadRequest, "Missing 'name' query parameter")
		return
	}

	// Get data source from repository
	dataSource, err := h.repo.GetDataSourceByName(name)
	if err != nil {
		Error(w, r, http.StatusNotFound, "Data source not found")
		return
	}

	// Return data source as JSON
	JSON(w, http.StatusOK, dataSource)
}

// ListDataSources handles GET /api/datasources
func (h *DataSourceHandler) ListDataSources(w http.ResponseWriter, r *http.Request) {
	// Get data sources from repository
	dataSources, err := h.repo.ListDataSources()
	if err != nil {
		Error(w, r, http.StatusInternalServerError, "Error retrieving data sources")
		return
	}

	// Return data sources using the standard list envelope
	List(w, dataSources, 1, len(dataSources), len(dataSources))
}

// CreateDataSource handles POST /api/datasources
func (h *DataSourceHandler) CreateDataSource(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var dataSource models.DataSource
	if err := json.NewDecoder(r.Body).Decode(&dataSource); err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate payload
	if err := validation.ValidateDataSource(&dataSource); err != nil {
		Error(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// Create data source in repository
	if err := h.repo.CreateDataSource(&dataSource); err != nil {
		Error(w, r, http.StatusInternalServerError, "Error creating data source")
		return
	}

	// Return created data source as JSON
	JSON(w, http.StatusCreated, dataSource)
}

// UpdateDataSource handles PUT /api/datasources/{id}
func (h *DataSourceHandler) UpdateDataSource(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid data source ID")
		return
	}

	// Parse request body
	var dataSource models.DataSource
	if err := json.NewDecoder(r.Body).Decode(&dataSource); err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Ensure ID in URL matches ID in body
	dataSource.ID = id

	// Validate payload
	if err := validation.ValidateDataSource(&dataSource); err != nil {
		Error(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// Update data source in repository
	if err := h.repo.UpdateDataSource(&dataSource); err != nil {
		Error(w, r, http.StatusInternalServerError, "Error updating data source")
		return
	}

	// Return updated data source as JSON
	JSON(w, http.StatusOK, dataSource)
}

// DeleteDataSource handles DELETE /api/datasources/{id}
func (h *DataSourceHandler) DeleteDataSource(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid data source ID")
		return
	}

	// Delete data source from repository
	if err := h.repo.DeleteDataSource(id); err != nil {
		Error(w, r, http.StatusInternalServerError, "Error deleting data source")
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// GetDetectionsByDataSource handles GET /api/datasources/{id}/detections
func (h *DataSourceHandler) GetDetectionsByDataSource(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid data source ID")
		return
	}

	// Get detections from repository
	detections, err := h.repo.GetDetectionsByDataSource(id)
	if err != nil {
		Error(w, r, http.StatusInternalServerError, "Error retrieving detections")
		return
	}

	// Return detections as JSON
	JSON(w, http.StatusOK, detections)
}

// GetDataSourceUtilization handles GET /api/datasources/utilization
func (h *DataSourceHandler) GetDataSourceUtilization(w http.ResponseWriter, r *http.Request) {
	// Get data source utilization from repository
	utilization, err := h.repo.GetDataSourceUtilization()
	if err != nil {
		Error(w, r, http.StatusInternalServerError, "Error retrieving data source utilization")
		return
	}

	// Return utilization as JSON
	JSON(w, http.StatusOK, utilization)
}

// GetMitreTechniquesByDataSource handles GET /api/datasources/{id}/techniques
func (h *DataSourceHandler) GetMitreTechniquesByDataSource(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "Invalid data source ID")
		return
	}

	// Get MITRE techniques from repository
	techniques, err := h.repo.GetMitreTechniquesByDataSource(id)
	if err != nil {
		Error(w, r, http.StatusInternalServerError, "Error retrieving MITRE techniques")
		return
	}

	// Return techniques as JSON
	JSON(w, http.StatusOK, techniques)
}
