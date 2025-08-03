package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"riskmatrix/internal/mitre"
	"riskmatrix/pkg/database"
	"riskmatrix/pkg/models"
)

// setupMitreTestHandler creates a MITRE handler with test database
func setupMitreTestHandler(t *testing.T) (*MitreHandler, *database.DB) {
	db := setupTestDB(t)
	repo := mitre.NewRepository(db)
	handler := NewMitreHandler(repo)
	return handler, db
}

// createTestMitreTechnique creates a test MITRE technique in the database
func createTestMitreTechnique(t *testing.T, db *database.DB) *models.MitreTechnique {
	repo := mitre.NewRepository(db)
	technique := &models.MitreTechnique{
		ID:          "T1059",
		Name:        "Command and Scripting Interpreter",
		Description: "Adversaries may abuse command and script interpreters to execute commands, scripts, or binaries.",
		Tactic:      "Execution",
	}

	err := repo.CreateMitreTechnique(technique)
	if err != nil {
		t.Fatalf("Failed to create test MITRE technique: %v", err)
	}

	return technique
}

func TestMitreHandler_GetMitreTechnique(t *testing.T) {
	handler, db := setupMitreTestHandler(t)
	defer db.Close()

	// Create test technique
	testTechnique := createTestMitreTechnique(t, db)

	tests := []struct {
		name           string
		techniqueID    string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Valid technique ID",
			techniqueID:    testTechnique.ID,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Empty technique ID",
			techniqueID:    "",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "Non-existent technique ID",
			techniqueID:    "T9999",
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/mitre/techniques/"+tt.techniqueID, nil)
			req.SetPathValue("id", tt.techniqueID)
			w := httptest.NewRecorder()

			handler.GetMitreTechnique(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.expectError && w.Code == http.StatusOK {
				var technique models.MitreTechnique
				err := json.NewDecoder(w.Body).Decode(&technique)
				if err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if technique.ID != testTechnique.ID {
					t.Errorf("Expected ID %s, got %s", testTechnique.ID, technique.ID)
				}

				if technique.Name != testTechnique.Name {
					t.Errorf("Expected name %s, got %s", testTechnique.Name, technique.Name)
				}
			}
		})
	}
}

func TestMitreHandler_ListMitreTechniques(t *testing.T) {
	handler, db := setupMitreTestHandler(t)
	defer db.Close()

	// Create multiple test techniques
	createTestMitreTechnique(t, db)
	
	// Create another technique with different tactic
	repo := mitre.NewRepository(db)
	technique2 := &models.MitreTechnique{
		ID:          "T1055",
		Name:        "Process Injection",
		Description: "Adversaries may inject code into processes in order to evade process-based defenses.",
		Tactic:      "Defense Evasion",
	}
	repo.CreateMitreTechnique(technique2)

	tests := []struct {
		name           string
		tacticFilter   string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "List all techniques",
			tacticFilter:   "",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "Filter by Execution tactic",
			tacticFilter:   "Execution",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "Filter by Defense Evasion tactic",
			tacticFilter:   "Defense Evasion",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "Filter by non-existent tactic",
			tacticFilter:   "NonExistent",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestURL := "/api/mitre/techniques"
			if tt.tacticFilter != "" {
				requestURL += "?tactic=" + url.QueryEscape(tt.tacticFilter)
			}

			req := httptest.NewRequest("GET", requestURL, nil)
			w := httptest.NewRecorder()

			handler.ListMitreTechniques(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var techniques []*models.MitreTechnique
			err := json.NewDecoder(w.Body).Decode(&techniques)
			if err != nil {
				t.Errorf("Failed to decode response: %v", err)
			}

			if len(techniques) != tt.expectedCount {
				t.Errorf("Expected %d techniques, got %d", tt.expectedCount, len(techniques))
			}
		})
	}
}

func TestMitreHandler_CreateMitreTechnique(t *testing.T) {
	handler, db := setupMitreTestHandler(t)
	defer db.Close()

	tests := []struct {
		name           string
		technique      models.MitreTechnique
		expectedStatus int
		expectError    bool
	}{
		{
			name: "Valid technique",
			technique: models.MitreTechnique{
				ID:          "T1001",
				Name:        "Data Obfuscation",
				Description: "Adversaries may obfuscate command and control traffic to make it more difficult to detect.",
				Tactic:      "Command and Control",
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "Missing required ID",
			technique: models.MitreTechnique{
				Name:        "Test Technique",
				Description: "Test description",
				Tactic:      "Execution",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "Missing required Name",
			technique: models.MitreTechnique{
				ID:          "T1002",
				Description: "Test description",
				Tactic:      "Execution",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "Missing required Tactic",
			technique: models.MitreTechnique{
				ID:          "T1003",
				Name:        "Test Technique",
				Description: "Test description",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.technique)
			req := httptest.NewRequest("POST", "/api/mitre/techniques", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateMitreTechnique(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.expectError && w.Code == http.StatusCreated {
				var createdTechnique models.MitreTechnique
				err := json.NewDecoder(w.Body).Decode(&createdTechnique)
				if err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if createdTechnique.ID != tt.technique.ID {
					t.Errorf("Expected ID %s, got %s", tt.technique.ID, createdTechnique.ID)
				}

				if createdTechnique.Name != tt.technique.Name {
					t.Errorf("Expected name %s, got %s", tt.technique.Name, createdTechnique.Name)
				}
			}
		})
	}
}

func TestMitreHandler_CreateMitreTechnique_InvalidJSON(t *testing.T) {
	handler, db := setupMitreTestHandler(t)
	defer db.Close()

	req := httptest.NewRequest("POST", "/api/mitre/techniques", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.CreateMitreTechnique(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestMitreHandler_UpdateMitreTechnique(t *testing.T) {
	handler, db := setupMitreTestHandler(t)
	defer db.Close()

	// Create test technique
	testTechnique := createTestMitreTechnique(t, db)

	tests := []struct {
		name           string
		techniqueID    string
		updatedData    models.MitreTechnique
		expectedStatus int
		expectError    bool
	}{
		{
			name:        "Valid update",
			techniqueID: testTechnique.ID,
			updatedData: models.MitreTechnique{
				Name:        "Updated Command and Scripting Interpreter",
				Description: "Updated description for command and scripting interpreter technique.",
				Tactic:      "Execution",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:        "Empty technique ID",
			techniqueID: "",
			updatedData: models.MitreTechnique{
				Name: "Updated Technique",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.updatedData)
			req := httptest.NewRequest("PUT", "/api/mitre/techniques/"+tt.techniqueID, bytes.NewBuffer(body))
			req.SetPathValue("id", tt.techniqueID)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.UpdateMitreTechnique(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.expectError && w.Code == http.StatusOK {
				var updatedTechnique models.MitreTechnique
				err := json.NewDecoder(w.Body).Decode(&updatedTechnique)
				if err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if updatedTechnique.Name != tt.updatedData.Name {
					t.Errorf("Expected name %s, got %s", tt.updatedData.Name, updatedTechnique.Name)
				}

				// ID should be set from URL path
				if updatedTechnique.ID != tt.techniqueID {
					t.Errorf("Expected ID %s, got %s", tt.techniqueID, updatedTechnique.ID)
				}
			}
		})
	}
}

func TestMitreHandler_DeleteMitreTechnique(t *testing.T) {
	handler, db := setupMitreTestHandler(t)
	defer db.Close()

	// Create test technique
	testTechnique := createTestMitreTechnique(t, db)

	tests := []struct {
		name           string
		techniqueID    string
		expectedStatus int
	}{
		{
			name:           "Valid deletion",
			techniqueID:    testTechnique.ID,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Empty technique ID",
			techniqueID:    "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Non-existent technique ID",
			techniqueID:    "T9999",
			expectedStatus: http.StatusInternalServerError, // Repository will return error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/api/mitre/techniques/"+tt.techniqueID, nil)
			req.SetPathValue("id", tt.techniqueID)
			w := httptest.NewRecorder()

			handler.DeleteMitreTechnique(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestMitreHandler_GetCoverageByTactic(t *testing.T) {
	handler, db := setupMitreTestHandler(t)
	defer db.Close()

	// Create test techniques
	createTestMitreTechnique(t, db)

	req := httptest.NewRequest("GET", "/api/mitre/coverage", nil)
	w := httptest.NewRecorder()

	handler.GetCoverageByTactic(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var coverage map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&coverage)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	// Coverage should be a map (exact structure depends on repository implementation)
	if coverage == nil {
		t.Error("Expected coverage data, got nil")
	}
}

func TestMitreHandler_GetDetectionsByTechnique(t *testing.T) {
	handler, db := setupMitreTestHandler(t)
	defer db.Close()

	// Create test technique
	testTechnique := createTestMitreTechnique(t, db)

	tests := []struct {
		name           string
		techniqueID    string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Valid technique ID",
			techniqueID:    testTechnique.ID,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Empty technique ID",
			techniqueID:    "",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "Non-existent technique ID",
			techniqueID:    "T9999",
			expectedStatus: http.StatusOK, // Should return empty array, not error
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/mitre/techniques/"+tt.techniqueID+"/detections", nil)
			req.SetPathValue("id", tt.techniqueID)
			w := httptest.NewRecorder()

			handler.GetDetectionsByTechnique(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.expectError && w.Code == http.StatusOK {
				var detections []*models.Detection
				err := json.NewDecoder(w.Body).Decode(&detections)
				if err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				// Should return empty array for technique with no detections
				if detections == nil {
					t.Error("Expected detections array, got nil")
				}
			}
		})
	}
}

func TestMitreHandler_UpdateMitreTechnique_InvalidJSON(t *testing.T) {
	handler, db := setupMitreTestHandler(t)
	defer db.Close()

	req := httptest.NewRequest("PUT", "/api/mitre/techniques/T1059", bytes.NewBufferString("invalid json"))
	req.SetPathValue("id", "T1059")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateMitreTechnique(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}