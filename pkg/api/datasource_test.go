package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"riskmatrix/internal/datasource"
	"riskmatrix/pkg/database"
	"riskmatrix/pkg/models"
)

// setupDataSourceTestHandler creates a datasource handler with test database
func setupDataSourceTestHandler(t *testing.T) (*DataSourceHandler, *database.DB) {
	db := setupTestDB(t)
	repo := datasource.NewRepository(db)
	handler := NewDataSourceHandler(repo)
	return handler, db
}

// createTestDataSource creates a test data source in the database
func createTestDataSource(t *testing.T, db *database.DB) *models.DataSource {
	repo := datasource.NewRepository(db)
	dataSource := &models.DataSource{
		Name:        "Test DataSource",
		Description: "Test data source for testing purposes",
		LogFormat:   "JSON",
	}

	err := repo.CreateDataSource(dataSource)
	if err != nil {
		t.Fatalf("Failed to create test data source: %v", err)
	}

	return dataSource
}

func TestDataSourceHandler_GetDataSource(t *testing.T) {
	handler, db := setupDataSourceTestHandler(t)
	defer db.Close()

	// Create test data source
	testDataSource := createTestDataSource(t, db)

	tests := []struct {
		name           string
		dataSourceID   string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Valid data source ID",
			dataSourceID:   strconv.FormatInt(testDataSource.ID, 10),
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Invalid data source ID",
			dataSourceID:   "invalid",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "Non-existent data source ID",
			dataSourceID:   "999999",
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/datasources/"+tt.dataSourceID, nil)
			req.SetPathValue("id", tt.dataSourceID)
			w := httptest.NewRecorder()

			handler.GetDataSource(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.expectError && w.Code == http.StatusOK {
				var dataSource models.DataSource
				err := json.NewDecoder(w.Body).Decode(&dataSource)
				if err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if dataSource.Name != testDataSource.Name {
					t.Errorf("Expected name %s, got %s", testDataSource.Name, dataSource.Name)
				}
			}
		})
	}
}

func TestDataSourceHandler_GetDataSourceByName(t *testing.T) {
	handler, db := setupDataSourceTestHandler(t)
	defer db.Close()

	// Create test data source
	testDataSource := createTestDataSource(t, db)

	tests := []struct {
		name           string
		dataSourceName string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Valid data source name",
			dataSourceName: testDataSource.Name,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Empty data source name",
			dataSourceName: "",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "Non-existent data source name",
			dataSourceName: "NonExistentDataSource",
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestURL := "/api/datasources/by-name/" + url.PathEscape(tt.dataSourceName)
			if tt.dataSourceName == "" {
				requestURL = "/api/datasources/by-name/"
			}
			req := httptest.NewRequest("GET", requestURL, nil)
			req.SetPathValue("name", tt.dataSourceName)
			w := httptest.NewRecorder()

			handler.GetDataSourceByName(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.expectError && w.Code == http.StatusOK {
				var dataSource models.DataSource
				err := json.NewDecoder(w.Body).Decode(&dataSource)
				if err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if dataSource.Name != testDataSource.Name {
					t.Errorf("Expected name %s, got %s", testDataSource.Name, dataSource.Name)
				}
			}
		})
	}
}

func TestDataSourceHandler_ListDataSources(t *testing.T) {
	handler, db := setupDataSourceTestHandler(t)
	defer db.Close()

	// Create multiple test data sources
	createTestDataSource(t, db)

	// Create another data source
	repo := datasource.NewRepository(db)
	dataSource2 := &models.DataSource{
		Name:        "Test DataSource 2",
		Description: "Second test data source",
		LogFormat:   "XML",
	}
	repo.CreateDataSource(dataSource2)

	req := httptest.NewRequest("GET", "/api/datasources", nil)
	w := httptest.NewRecorder()

	handler.ListDataSources(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp struct {
		Items    []*models.DataSource `json:"items"`
		Page     int                  `json:"page"`
		PageSize int                  `json:"page_size"`
		Total    int                  `json:"total"`
	}
	err := json.NewDecoder(w.Body).Decode(&resp)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if len(resp.Items) != 2 {
		t.Errorf("Expected 2 data sources, got %d", len(resp.Items))
	}
}

func TestDataSourceHandler_CreateDataSource(t *testing.T) {
	handler, db := setupDataSourceTestHandler(t)
	defer db.Close()

	tests := []struct {
		name           string
		dataSource     models.DataSource
		expectedStatus int
		expectError    bool
	}{
		{
			name: "Valid data source",
			dataSource: models.DataSource{
				Name:        "New DataSource",
				Description: "New data source for testing",
				LogFormat:   "CEF",
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "Data source with minimal fields",
			dataSource: models.DataSource{
				Name: "Minimal DataSource",
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "Data source without name",
			dataSource: models.DataSource{
				Description: "Data source without name",
				LogFormat:   "JSON",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.dataSource)
			req := httptest.NewRequest("POST", "/api/datasources", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateDataSource(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.expectError && w.Code == http.StatusCreated {
				var createdDataSource models.DataSource
				err := json.NewDecoder(w.Body).Decode(&createdDataSource)
				if err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if createdDataSource.Name != tt.dataSource.Name {
					t.Errorf("Expected name %s, got %s", tt.dataSource.Name, createdDataSource.Name)
				}

				if createdDataSource.ID == 0 {
					t.Error("Expected non-zero ID for created data source")
				}
			}
		})
	}
}

func TestDataSourceHandler_CreateDataSource_InvalidJSON(t *testing.T) {
	handler, db := setupDataSourceTestHandler(t)
	defer db.Close()

	req := httptest.NewRequest("POST", "/api/datasources", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateDataSource(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDataSourceHandler_UpdateDataSource(t *testing.T) {
	handler, db := setupDataSourceTestHandler(t)
	defer db.Close()

	// Create test data source
	testDataSource := createTestDataSource(t, db)

	tests := []struct {
		name           string
		dataSourceID   string
		updatedData    models.DataSource
		expectedStatus int
		expectError    bool
	}{
		{
			name:         "Valid update",
			dataSourceID: strconv.FormatInt(testDataSource.ID, 10),
			updatedData: models.DataSource{
				Name:        "Updated DataSource",
				Description: "Updated description",
				LogFormat:   "LEEF",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:         "Invalid data source ID",
			dataSourceID: "invalid",
			updatedData: models.DataSource{
				Name: "Updated DataSource",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.updatedData)
			req := httptest.NewRequest("PUT", "/api/datasources/"+tt.dataSourceID, bytes.NewBuffer(body))
			req.SetPathValue("id", tt.dataSourceID)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.UpdateDataSource(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.expectError && w.Code == http.StatusOK {
				var updatedDataSource models.DataSource
				err := json.NewDecoder(w.Body).Decode(&updatedDataSource)
				if err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if updatedDataSource.Name != tt.updatedData.Name {
					t.Errorf("Expected name %s, got %s", tt.updatedData.Name, updatedDataSource.Name)
				}
			}
		})
	}
}

func TestDataSourceHandler_DeleteDataSource(t *testing.T) {
	handler, db := setupDataSourceTestHandler(t)
	defer db.Close()

	// Create test data source
	testDataSource := createTestDataSource(t, db)

	tests := []struct {
		name           string
		dataSourceID   string
		expectedStatus int
	}{
		{
			name:           "Valid deletion",
			dataSourceID:   strconv.FormatInt(testDataSource.ID, 10),
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Invalid data source ID",
			dataSourceID:   "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Non-existent data source ID",
			dataSourceID:   "999999",
			expectedStatus: http.StatusInternalServerError, // Repository will return error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/api/datasources/"+tt.dataSourceID, nil)
			req.SetPathValue("id", tt.dataSourceID)
			w := httptest.NewRecorder()

			handler.DeleteDataSource(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestDataSourceHandler_GetDetectionsByDataSource(t *testing.T) {
	handler, db := setupDataSourceTestHandler(t)
	defer db.Close()

	// Create test data source
	testDataSource := createTestDataSource(t, db)

	tests := []struct {
		name           string
		dataSourceID   string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Valid data source ID",
			dataSourceID:   strconv.FormatInt(testDataSource.ID, 10),
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Invalid data source ID",
			dataSourceID:   "invalid",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "Non-existent data source ID",
			dataSourceID:   "999999",
			expectedStatus: http.StatusOK, // Should return empty array, not error
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/datasources/"+tt.dataSourceID+"/detections", nil)
			req.SetPathValue("id", tt.dataSourceID)
			w := httptest.NewRecorder()

			handler.GetDetectionsByDataSource(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.expectError && w.Code == http.StatusOK {
				var detections []*models.Detection
				err := json.NewDecoder(w.Body).Decode(&detections)
				if err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				// Should return empty array for data source with no detections
				if detections == nil {
					t.Error("Expected detections array, got nil")
				}
			}
		})
	}
}

func TestDataSourceHandler_GetDataSourceUtilization(t *testing.T) {
	handler, db := setupDataSourceTestHandler(t)
	defer db.Close()

	// Create test data sources
	createTestDataSource(t, db)

	req := httptest.NewRequest("GET", "/api/datasources/utilization", nil)
	w := httptest.NewRecorder()

	handler.GetDataSourceUtilization(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var utilization map[string]int
	err := json.NewDecoder(w.Body).Decode(&utilization)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	// Utilization should be a map (exact structure depends on repository implementation)
	if utilization == nil {
		t.Error("Expected utilization data, got nil")
	}
}

func TestDataSourceHandler_GetMitreTechniquesByDataSource(t *testing.T) {
	handler, db := setupDataSourceTestHandler(t)
	defer db.Close()

	// Create test data source
	testDataSource := createTestDataSource(t, db)

	tests := []struct {
		name           string
		dataSourceID   string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Valid data source ID",
			dataSourceID:   strconv.FormatInt(testDataSource.ID, 10),
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Invalid data source ID",
			dataSourceID:   "invalid",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "Non-existent data source ID",
			dataSourceID:   "999999",
			expectedStatus: http.StatusOK, // Should return empty array, not error
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/datasources/"+tt.dataSourceID+"/techniques", nil)
			req.SetPathValue("id", tt.dataSourceID)
			w := httptest.NewRecorder()

			handler.GetMitreTechniquesByDataSource(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.expectError && w.Code == http.StatusOK {
				var techniques []*models.MitreTechnique
				err := json.NewDecoder(w.Body).Decode(&techniques)
				if err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				// Should return empty array for data source with no techniques
				if techniques == nil {
					t.Error("Expected techniques array, got nil")
				}
			}
		})
	}
}

func TestDataSourceHandler_UpdateDataSource_InvalidJSON(t *testing.T) {
	handler, db := setupDataSourceTestHandler(t)
	defer db.Close()

	req := httptest.NewRequest("PUT", "/api/datasources/1", bytes.NewBufferString("invalid json"))
	req.SetPathValue("id", "1")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateDataSource(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// Test edge cases and error conditions
func TestDataSourceHandler_CreateDataSource_DuplicateName(t *testing.T) {
	handler, db := setupDataSourceTestHandler(t)
	defer db.Close()

	// Create first data source
	dataSource1 := models.DataSource{
		Name:        "Duplicate Name Test",
		Description: "First data source",
		LogFormat:   "JSON",
	}

	body1, _ := json.Marshal(dataSource1)
	req1 := httptest.NewRequest("POST", "/api/datasources", bytes.NewBuffer(body1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()

	handler.CreateDataSource(w1, req1)

	if w1.Code != http.StatusCreated {
		t.Errorf("Expected status %d for first creation, got %d", http.StatusCreated, w1.Code)
	}

	// Try to create second data source with same name
	dataSource2 := models.DataSource{
		Name:        "Duplicate Name Test",
		Description: "Second data source with same name",
		LogFormat:   "XML",
	}

	body2, _ := json.Marshal(dataSource2)
	req2 := httptest.NewRequest("POST", "/api/datasources", bytes.NewBuffer(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()

	handler.CreateDataSource(w2, req2)

	// Should fail due to unique constraint on name
	if w2.Code == http.StatusCreated {
		t.Error("Expected creation to fail due to duplicate name, but it succeeded")
	}
}
