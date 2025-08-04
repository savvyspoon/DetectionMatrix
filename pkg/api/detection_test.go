package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"riskmatrix/internal/detection"
	"riskmatrix/pkg/database"
	"riskmatrix/pkg/models"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *database.DB {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	return db
}

// setupTestHandler creates a detection handler with test database
func setupTestHandler(t *testing.T) (*DetectionHandler, *database.DB) {
	db := setupTestDB(t)
	repo := detection.NewRepository(db)
	handler := NewDetectionHandler(repo)
	return handler, db
}

// createTestDetection creates a test detection in the database
func createTestDetection(t *testing.T, db *database.DB) *models.Detection {
	repo := detection.NewRepository(db)
	detection := &models.Detection{
		Name:                     "Test Detection",
		Description:              "Test description",
		Status:                   models.StatusDraft,
		Severity:                 models.SeverityMedium,
		RiskPoints:               50,
		PlaybookLink:             "https://example.com/playbook",
		Owner:                    "test-owner",
		RiskObject:               models.RiskObjectHost,
		TestingDescription:       "Test testing description",
		EventCountLast30Days:     10,
		FalsePositivesLast30Days: 2,
	}

	err := repo.CreateDetection(detection)
	if err != nil {
		t.Fatalf("Failed to create test detection: %v", err)
	}

	return detection
}

func TestDetectionHandler_GetDetection(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer db.Close()

	// Create test detection
	testDetection := createTestDetection(t, db)

	tests := []struct {
		name           string
		detectionID    string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Valid detection ID",
			detectionID:    strconv.FormatInt(testDetection.ID, 10),
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Invalid detection ID",
			detectionID:    "invalid",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "Non-existent detection ID",
			detectionID:    "999999",
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/detections/"+tt.detectionID, nil)
			req.SetPathValue("id", tt.detectionID)
			w := httptest.NewRecorder()

			handler.GetDetection(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.expectError && w.Code == http.StatusOK {
				var detection models.Detection
				err := json.NewDecoder(w.Body).Decode(&detection)
				if err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if detection.Name != testDetection.Name {
					t.Errorf("Expected name %s, got %s", testDetection.Name, detection.Name)
				}
			}
		})
	}
}

func TestDetectionHandler_ListDetections(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer db.Close()

	// Create multiple test detections
	createTestDetection(t, db)
	
	// Create another detection with different status
	repo := detection.NewRepository(db)
	detection2 := &models.Detection{
		Name:        "Test Detection 2",
		Description: "Test description 2",
		Status:      models.StatusProduction,
		Severity:    models.SeverityHigh,
		RiskPoints:  75,
		RiskObject:  models.RiskObjectUser,
	}
	err := repo.CreateDetection(detection2)
	if err != nil {
		t.Fatalf("Failed to create second detection: %v", err)
	}

	tests := []struct {
		name           string
		statusFilter   string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "List all detections",
			statusFilter:   "",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "Filter by draft status",
			statusFilter:   "draft",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "Filter by production status",
			statusFilter:   "production",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/detections"
			if tt.statusFilter != "" {
				url += "?status=" + tt.statusFilter
			}

			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			handler.ListDetections(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var detections []*models.Detection
			err := json.NewDecoder(w.Body).Decode(&detections)
			if err != nil {
				t.Errorf("Failed to decode response: %v", err)
			}

			if len(detections) != tt.expectedCount {
				t.Errorf("Expected %d detections, got %d", tt.expectedCount, len(detections))
			}
		})
	}
}

func TestDetectionHandler_CreateDetection(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer db.Close()

	tests := []struct {
		name           string
		detection      models.Detection
		expectedStatus int
		expectError    bool
	}{
		{
			name: "Valid detection",
			detection: models.Detection{
				Name:        "New Detection",
				Description: "New description",
				Status:      models.StatusIdea,
				Severity:    models.SeverityLow,
				RiskPoints:  25,
				RiskObject:  models.RiskObjectHost,
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "Detection with all fields",
			detection: models.Detection{
				Name:                     "Complete Detection",
				Description:              "Complete description",
				Status:                   models.StatusTest,
				Severity:                 models.SeverityHigh,
				RiskPoints:               80,
				PlaybookLink:             "https://example.com/playbook",
				Owner:                    "analyst1",
				RiskObject:               models.RiskObjectUser,
				TestingDescription:       "Comprehensive testing",
				EventCountLast30Days:     15,
				FalsePositivesLast30Days: 1,
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.detection)
			req := httptest.NewRequest("POST", "/api/detections", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateDetection(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.expectError && w.Code == http.StatusCreated {
				var createdDetection models.Detection
				err := json.NewDecoder(w.Body).Decode(&createdDetection)
				if err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if createdDetection.Name != tt.detection.Name {
					t.Errorf("Expected name %s, got %s", tt.detection.Name, createdDetection.Name)
				}

				if createdDetection.ID == 0 {
					t.Error("Expected non-zero ID for created detection")
				}
			}
		})
	}
}

func TestDetectionHandler_CreateDetection_InvalidJSON(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer db.Close()

	req := httptest.NewRequest("POST", "/api/detections", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateDetection(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDetectionHandler_UpdateDetection(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer db.Close()

	// Create test detection
	testDetection := createTestDetection(t, db)

	tests := []struct {
		name           string
		detectionID    string
		updatedData    models.Detection
		expectedStatus int
		expectError    bool
	}{
		{
			name:        "Valid update",
			detectionID: strconv.FormatInt(testDetection.ID, 10),
			updatedData: models.Detection{
				Name:        "Updated Detection",
				Description: "Updated description",
				Status:      models.StatusProduction,
				Severity:    models.SeverityHigh,
				RiskPoints:  75,
				Owner:       "updated-owner",
				RiskObject:  models.RiskObjectUser,
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:        "Invalid detection ID",
			detectionID: "invalid",
			updatedData: models.Detection{
				Name: "Updated Detection",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.updatedData)
			req := httptest.NewRequest("PUT", "/api/detections/"+tt.detectionID, bytes.NewBuffer(body))
			req.SetPathValue("id", tt.detectionID)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.UpdateDetection(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.expectError && w.Code == http.StatusOK {
				var updatedDetection models.Detection
				err := json.NewDecoder(w.Body).Decode(&updatedDetection)
				if err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if updatedDetection.Name != tt.updatedData.Name {
					t.Errorf("Expected name %s, got %s", tt.updatedData.Name, updatedDetection.Name)
				}
			}
		})
	}
}

func TestDetectionHandler_DeleteDetection(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer db.Close()

	// Create test detection
	testDetection := createTestDetection(t, db)

	tests := []struct {
		name           string
		detectionID    string
		expectedStatus int
	}{
		{
			name:           "Valid deletion",
			detectionID:    strconv.FormatInt(testDetection.ID, 10),
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Invalid detection ID",
			detectionID:    "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Non-existent detection ID",
			detectionID:    "999999",
			expectedStatus: http.StatusInternalServerError, // Repository will return error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/api/detections/"+tt.detectionID, nil)
			req.SetPathValue("id", tt.detectionID)
			w := httptest.NewRecorder()

			handler.DeleteDetection(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestDetectionHandler_GetDetectionCount(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer db.Close()

	// Create test detections
	createTestDetection(t, db)
	createTestDetection(t, db)

	req := httptest.NewRequest("GET", "/api/detections/count", nil)
	w := httptest.NewRecorder()

	handler.GetDetectionCount(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]int
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if response["count"] != 2 {
		t.Errorf("Expected count 2, got %d", response["count"])
	}
}

func TestDetectionHandler_GetDetectionCountByStatus(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer db.Close()

	// Create test detections with different statuses
	repo := detection.NewRepository(db)
	
	detection1 := &models.Detection{
		Name:        "Draft Detection",
		Description: "Draft",
		Status:      models.StatusDraft,
		Severity:    models.SeverityMedium,
		RiskPoints:  50,
		RiskObject:  models.RiskObjectHost,
	}
	var err error
	err = repo.CreateDetection(detection1)
	if err != nil {
		t.Fatalf("Failed to create draft detection: %v", err)
	}

	detection2 := &models.Detection{
		Name:        "Production Detection",
		Description: "Production",
		Status:      models.StatusProduction,
		Severity:    models.SeverityHigh,
		RiskPoints:  75,
		RiskObject:  models.RiskObjectUser,
	}
	err = repo.CreateDetection(detection2)
	if err != nil {
		t.Fatalf("Failed to create production detection: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/detections/count/status", nil)
	w := httptest.NewRecorder()

	handler.GetDetectionCountByStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[models.DetectionStatus]int
	err = json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if response[models.StatusDraft] != 1 {
		t.Errorf("Expected 1 draft detection, got %d", response[models.StatusDraft])
	}

	if response[models.StatusProduction] != 1 {
		t.Errorf("Expected 1 production detection, got %d", response[models.StatusProduction])
	}
}

func TestDetectionHandler_GetFalsePositiveRate(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer db.Close()

	// Create test detection
	testDetection := createTestDetection(t, db)

	tests := []struct {
		name           string
		detectionID    string
		expectedStatus int
	}{
		{
			name:           "Valid detection ID",
			detectionID:    strconv.FormatInt(testDetection.ID, 10),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid detection ID",
			detectionID:    "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/detections/"+tt.detectionID+"/fp-rate", nil)
			req.SetPathValue("id", tt.detectionID)
			w := httptest.NewRecorder()

			handler.GetFalsePositiveRate(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]float64
				err := json.NewDecoder(w.Body).Decode(&response)
				if err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if _, exists := response["false_positive_rate"]; !exists {
					t.Error("Expected false_positive_rate in response")
				}
			}
		})
	}
}

func TestDetectionHandler_AddMitreTechnique(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer db.Close()

	// Create test detection
	testDetection := createTestDetection(t, db)

	// Create test MITRE technique
	_, err := db.Exec("INSERT INTO mitre_techniques (id, name, description, tactic) VALUES (?, ?, ?, ?)",
		"T1059", "Command and Scripting Interpreter", "Test description", "Execution")
	if err != nil {
		t.Fatalf("Failed to create test MITRE technique: %v", err)
	}

	tests := []struct {
		name           string
		detectionID    string
		techniqueID    string
		expectedStatus int
	}{
		{
			name:           "Valid addition",
			detectionID:    strconv.FormatInt(testDetection.ID, 10),
			techniqueID:    "T1059",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid detection ID",
			detectionID:    "invalid",
			techniqueID:    "T1059",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/detections/"+tt.detectionID+"/mitre/"+tt.techniqueID, nil)
			req.SetPathValue("id", tt.detectionID)
			req.SetPathValue("technique_id", tt.techniqueID)
			w := httptest.NewRecorder()

			handler.AddMitreTechnique(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestDetectionHandler_RemoveMitreTechnique(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer db.Close()

	// Create test detection
	testDetection := createTestDetection(t, db)

	// Create test MITRE technique and association
	_, err := db.Exec("INSERT INTO mitre_techniques (id, name, description, tactic) VALUES (?, ?, ?, ?)",
		"T1059", "Command and Scripting Interpreter", "Test description", "Execution")
	if err != nil {
		t.Fatalf("Failed to create test MITRE technique: %v", err)
	}

	_, err = db.Exec("INSERT INTO detection_mitre_map (detection_id, mitre_id) VALUES (?, ?)",
		testDetection.ID, "T1059")
	if err != nil {
		t.Fatalf("Failed to create test association: %v", err)
	}

	tests := []struct {
		name           string
		detectionID    string
		techniqueID    string
		expectedStatus int
	}{
		{
			name:           "Valid removal",
			detectionID:    strconv.FormatInt(testDetection.ID, 10),
			techniqueID:    "T1059",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid detection ID",
			detectionID:    "invalid",
			techniqueID:    "T1059",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/api/detections/"+tt.detectionID+"/mitre/"+tt.techniqueID, nil)
			req.SetPathValue("id", tt.detectionID)
			req.SetPathValue("technique_id", tt.techniqueID)
			w := httptest.NewRecorder()

			handler.RemoveMitreTechnique(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestDetectionHandler_AddDataSource(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer db.Close()

	// Create test detection
	testDetection := createTestDetection(t, db)

	// Create test data source
	_, err := db.Exec("INSERT INTO data_sources (id, name, description) VALUES (?, ?, ?)",
		1, "Test DataSource", "Test description")
	if err != nil {
		t.Fatalf("Failed to create test data source: %v", err)
	}

	tests := []struct {
		name           string
		detectionID    string
		dataSourceID   string
		expectedStatus int
	}{
		{
			name:           "Valid addition",
			detectionID:    strconv.FormatInt(testDetection.ID, 10),
			dataSourceID:   "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid detection ID",
			detectionID:    "invalid",
			dataSourceID:   "1",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid data source ID",
			detectionID:    strconv.FormatInt(testDetection.ID, 10),
			dataSourceID:   "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/detections/"+tt.detectionID+"/datasource/"+tt.dataSourceID, nil)
			req.SetPathValue("id", tt.detectionID)
			req.SetPathValue("datasource_id", tt.dataSourceID)
			w := httptest.NewRecorder()

			handler.AddDataSource(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestDetectionHandler_RemoveDataSource(t *testing.T) {
	handler, db := setupTestHandler(t)
	defer db.Close()

	// Create test detection
	testDetection := createTestDetection(t, db)

	// Create test data source and association
	_, err := db.Exec("INSERT INTO data_sources (id, name, description) VALUES (?, ?, ?)",
		1, "Test DataSource", "Test description")
	if err != nil {
		t.Fatalf("Failed to create test data source: %v", err)
	}

	_, err = db.Exec("INSERT INTO detection_datasource (detection_id, datasource_id) VALUES (?, ?)",
		testDetection.ID, 1)
	if err != nil {
		t.Fatalf("Failed to create test association: %v", err)
	}

	tests := []struct {
		name           string
		detectionID    string
		dataSourceID   string
		expectedStatus int
	}{
		{
			name:           "Valid removal",
			detectionID:    strconv.FormatInt(testDetection.ID, 10),
			dataSourceID:   "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid detection ID",
			detectionID:    "invalid",
			dataSourceID:   "1",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid data source ID",
			detectionID:    strconv.FormatInt(testDetection.ID, 10),
			dataSourceID:   "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/api/detections/"+tt.detectionID+"/datasource/"+tt.dataSourceID, nil)
			req.SetPathValue("id", tt.detectionID)
			req.SetPathValue("datasource_id", tt.dataSourceID)
			w := httptest.NewRecorder()

			handler.RemoveDataSource(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}