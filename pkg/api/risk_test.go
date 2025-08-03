package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"riskmatrix/internal/risk"
	"riskmatrix/pkg/database"
	"riskmatrix/pkg/models"
)

// setupRiskTestHandler creates a risk handler with test database
func setupRiskTestHandler(t *testing.T) (*RiskHandler, *database.DB) {
	db := setupTestDB(t)
	repo := risk.NewRepository(db)
	engine := risk.NewEngine(db, risk.DefaultConfig())
	handler := NewRiskHandler(engine, repo)
	return handler, db
}

// createTestRiskObject creates a test risk object in the database using direct SQL
func createTestRiskObject(t *testing.T, db *database.DB) *models.RiskObject {
	query := `INSERT INTO risk_objects (entity_type, entity_value, current_score, last_seen) VALUES (?, ?, ?, ?)`
	
	now := time.Now()
	result, err := db.Exec(query, models.EntityTypeUser, "test-user", 25, now.Format(time.RFC3339))
	if err != nil {
		t.Fatalf("Failed to create test risk object: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get last insert ID: %v", err)
	}

	return &models.RiskObject{
		ID:           id,
		EntityType:   models.EntityTypeUser,
		EntityValue:  "test-user",
		CurrentScore: 25,
		LastSeen:     now,
	}
}

// createTestEvent creates a test event in the database using direct SQL
func createTestEvent(t *testing.T, db *database.DB, detectionID int64, riskObjectID int64) *models.Event {
	query := `INSERT INTO events (detection_id, entity_id, timestamp, raw_data, risk_points, is_false_positive) VALUES (?, ?, ?, ?, ?, ?)`
	
	now := time.Now()
	result, err := db.Exec(query, detectionID, riskObjectID, now.Format(time.RFC3339), "test event data", 10, false)
	if err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get last insert ID: %v", err)
	}

	return &models.Event{
		ID:          id,
		DetectionID: detectionID,
		EntityID:    riskObjectID,
		Timestamp:   now,
		RawData:     "test event data",
		RiskPoints:  10,
	}
}

func TestRiskHandler_ProcessEvent(t *testing.T) {
	handler, db := setupRiskTestHandler(t)
	defer db.Close()

	// Create test detection first
	testDetection := createTestDetection(t, db)

	tests := []struct {
		name           string
		event          models.Event
		expectedStatus int
		expectError    bool
	}{
		{
			name: "Valid event",
			event: models.Event{
				DetectionID: testDetection.ID,
				RiskObject: &models.RiskObject{
					EntityType:  models.EntityTypeUser,
					EntityValue: "test-user",
				},
				Timestamp:  time.Now(),
				RawData:    "test event data",
				RiskPoints: 15,
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "Event with IP risk object",
			event: models.Event{
				DetectionID: testDetection.ID,
				RiskObject: &models.RiskObject{
					EntityType:  models.EntityTypeIP,
					EntityValue: "192.168.1.100",
				},
				Timestamp:  time.Now(),
				RawData:    "IP-based event",
				RiskPoints: 20,
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.event)
			req := httptest.NewRequest("POST", "/api/events", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ProcessEvent(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.expectError && w.Code == http.StatusCreated {
				var processedEvent models.Event
				err := json.NewDecoder(w.Body).Decode(&processedEvent)
				if err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if processedEvent.ID == 0 {
					t.Error("Expected non-zero ID for processed event")
				}

				if processedEvent.DetectionID != tt.event.DetectionID {
					t.Errorf("Expected detection ID %d, got %d", tt.event.DetectionID, processedEvent.DetectionID)
				}
			}
		})
	}
}

func TestRiskHandler_ProcessEvent_InvalidJSON(t *testing.T) {
	handler, db := setupRiskTestHandler(t)
	defer db.Close()

	req := httptest.NewRequest("POST", "/api/events", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ProcessEvent(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestRiskHandler_ProcessEvents(t *testing.T) {
	handler, db := setupRiskTestHandler(t)
	defer db.Close()

	// Create test detection
	testDetection := createTestDetection(t, db)

	events := []models.Event{
		{
			DetectionID: testDetection.ID,
			RiskObject: &models.RiskObject{
				EntityType:  models.EntityTypeUser,
				EntityValue: "batch-user-1",
			},
			Timestamp:  time.Now(),
			RawData:    "batch event 1",
			RiskPoints: 10,
		},
		{
			DetectionID: testDetection.ID,
			RiskObject: &models.RiskObject{
				EntityType:  models.EntityTypeUser,
				EntityValue: "batch-user-2",
			},
			Timestamp:  time.Now(),
			RawData:    "batch event 2",
			RiskPoints: 15,
		},
	}

	body, _ := json.Marshal(events)
	req := httptest.NewRequest("POST", "/api/events/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ProcessEvents(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if response["processed"] != float64(2) {
		t.Errorf("Expected 2 processed events, got %v", response["processed"])
	}
}

func TestRiskHandler_GetRiskObject(t *testing.T) {
	handler, db := setupRiskTestHandler(t)
	defer db.Close()

	// Create test risk object
	testRiskObject := createTestRiskObject(t, db)

	tests := []struct {
		name           string
		riskObjectID   string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Valid risk object ID",
			riskObjectID:   strconv.FormatInt(testRiskObject.ID, 10),
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Invalid risk object ID",
			riskObjectID:   "invalid",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "Non-existent risk object ID",
			riskObjectID:   "999999",
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/risk/objects/"+tt.riskObjectID, nil)
			req.SetPathValue("id", tt.riskObjectID)
			w := httptest.NewRecorder()

			handler.GetRiskObject(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.expectError && w.Code == http.StatusOK {
				var riskObject models.RiskObject
				err := json.NewDecoder(w.Body).Decode(&riskObject)
				if err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if riskObject.EntityValue != testRiskObject.EntityValue {
					t.Errorf("Expected entity value %s, got %s", testRiskObject.EntityValue, riskObject.EntityValue)
				}
			}
		})
	}
}

func TestRiskHandler_GetRiskObjectByEntity(t *testing.T) {
	handler, db := setupRiskTestHandler(t)
	defer db.Close()

	// Create test risk object
	testRiskObject := createTestRiskObject(t, db)

	tests := []struct {
		name           string
		entityType     string
		entityValue    string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Valid entity",
			entityType:     string(testRiskObject.EntityType),
			entityValue:    testRiskObject.EntityValue,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Missing entity type",
			entityType:     "",
			entityValue:    testRiskObject.EntityValue,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "Missing entity value",
			entityType:     string(testRiskObject.EntityType),
			entityValue:    "",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "Non-existent entity",
			entityType:     "user",
			entityValue:    "non-existent-user",
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/risk/objects/entity?type=" + tt.entityType + "&value=" + tt.entityValue
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			handler.GetRiskObjectByEntity(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.expectError && w.Code == http.StatusOK {
				var riskObject models.RiskObject
				err := json.NewDecoder(w.Body).Decode(&riskObject)
				if err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if riskObject.EntityValue != tt.entityValue {
					t.Errorf("Expected entity value %s, got %s", tt.entityValue, riskObject.EntityValue)
				}
			}
		})
	}
}

func TestRiskHandler_ListRiskObjects(t *testing.T) {
	handler, db := setupRiskTestHandler(t)
	defer db.Close()

	// Create multiple test risk objects
	createTestRiskObject(t, db)
	
	// Create another risk object using direct SQL
	query2 := `INSERT INTO risk_objects (entity_type, entity_value, current_score, last_seen) VALUES (?, ?, ?, ?)`
	now2 := time.Now()
	_, err := db.Exec(query2, models.EntityTypeHost, "test-host", 35, now2.Format(time.RFC3339))
	if err != nil {
		t.Fatalf("Failed to create second test risk object: %v", err)
	}

	tests := []struct {
		name           string
		limit          string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "List all risk objects",
			limit:          "",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "List with limit",
			limit:          "1",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "List with invalid limit",
			limit:          "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/risk/objects"
			if tt.limit != "" {
				url += "?limit=" + tt.limit
			}

			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			handler.ListRiskObjects(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var riskObjects []*models.RiskObject
				err := json.NewDecoder(w.Body).Decode(&riskObjects)
				if err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if len(riskObjects) != tt.expectedCount {
					t.Errorf("Expected %d risk objects, got %d", tt.expectedCount, len(riskObjects))
				}
			}
		})
	}
}

func TestRiskHandler_GetEvent(t *testing.T) {
	handler, db := setupRiskTestHandler(t)
	defer db.Close()

	// Create test data
	testDetection := createTestDetection(t, db)
	testRiskObject := createTestRiskObject(t, db)
	testEvent := createTestEvent(t, db, testDetection.ID, testRiskObject.ID)

	tests := []struct {
		name           string
		eventID        string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Valid event ID",
			eventID:        strconv.FormatInt(testEvent.ID, 10),
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Invalid event ID",
			eventID:        "invalid",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "Non-existent event ID",
			eventID:        "999999",
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/events/"+tt.eventID, nil)
			req.SetPathValue("id", tt.eventID)
			w := httptest.NewRecorder()

			handler.GetEvent(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.expectError && w.Code == http.StatusOK {
				var event models.Event
				err := json.NewDecoder(w.Body).Decode(&event)
				if err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if event.DetectionID != testEvent.DetectionID {
					t.Errorf("Expected detection ID %d, got %d", testEvent.DetectionID, event.DetectionID)
				}
			}
		})
	}
}

func TestRiskHandler_ListEventsByEntity(t *testing.T) {
	handler, db := setupRiskTestHandler(t)
	defer db.Close()

	// Create test data
	testDetection := createTestDetection(t, db)
	testRiskObject := createTestRiskObject(t, db)
	createTestEvent(t, db, testDetection.ID, testRiskObject.ID)

	tests := []struct {
		name           string
		entityID       string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Valid entity ID",
			entityID:       strconv.FormatInt(testRiskObject.ID, 10),
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Invalid entity ID",
			entityID:       "invalid",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "Non-existent entity ID",
			entityID:       "999999",
			expectedStatus: http.StatusOK, // Should return empty array
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/events/entity/"+tt.entityID, nil)
			req.SetPathValue("id", tt.entityID)
			w := httptest.NewRecorder()

			handler.ListEventsByEntity(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.expectError && w.Code == http.StatusOK {
				var events []*models.Event
				err := json.NewDecoder(w.Body).Decode(&events)
				if err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if events == nil {
					t.Error("Expected events array, got nil")
				}
			}
		})
	}
}

func TestRiskHandler_MarkEventAsFalsePositive(t *testing.T) {
	handler, db := setupRiskTestHandler(t)
	defer db.Close()

	// Create test data
	testDetection := createTestDetection(t, db)
	testRiskObject := createTestRiskObject(t, db)
	testEvent := createTestEvent(t, db, testDetection.ID, testRiskObject.ID)

	falsePositive := models.FalsePositive{
		Reason:       "Test false positive",
		AnalystName:  "test-analyst",
	}

	tests := []struct {
		name           string
		eventID        string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Valid false positive marking",
			eventID:        strconv.FormatInt(testEvent.ID, 10),
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Invalid event ID",
			eventID:        "invalid",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(falsePositive)
			req := httptest.NewRequest("POST", "/api/events/"+tt.eventID+"/false-positive", bytes.NewBuffer(body))
			req.SetPathValue("id", tt.eventID)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.MarkEventAsFalsePositive(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestRiskHandler_ListRiskAlerts(t *testing.T) {
	handler, db := setupRiskTestHandler(t)
	defer db.Close()

	req := httptest.NewRequest("GET", "/api/risk/alerts", nil)
	w := httptest.NewRecorder()

	handler.ListRiskAlerts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var alerts []*models.RiskAlert
	err := json.NewDecoder(w.Body).Decode(&alerts)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	// Should return empty array initially
	if alerts == nil {
		t.Error("Expected alerts array, got nil")
	}
}

func TestRiskHandler_GetEventsForAlert(t *testing.T) {
	handler, db := setupRiskTestHandler(t)
	defer db.Close()

	tests := []struct {
		name           string
		alertID        string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Valid alert ID",
			alertID:        "1",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Invalid alert ID",
			alertID:        "invalid",
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/risk/alerts/"+tt.alertID+"/events", nil)
			req.SetPathValue("id", tt.alertID)
			w := httptest.NewRecorder()

			handler.GetEventsForAlert(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if !tt.expectError && w.Code == http.StatusOK {
				var events []*models.Event
				err := json.NewDecoder(w.Body).Decode(&events)
				if err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				if events == nil {
					t.Error("Expected events array, got nil")
				}
			}
		})
	}
}

func TestRiskHandler_DecayRiskScores(t *testing.T) {
	handler, db := setupRiskTestHandler(t)
	defer db.Close()

	req := httptest.NewRequest("POST", "/api/risk/decay", nil)
	w := httptest.NewRecorder()

	handler.DecayRiskScores(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if response["message"] == nil {
		t.Error("Expected message in response")
	}
}

func TestRiskHandler_GetHighRiskEntities(t *testing.T) {
	handler, db := setupRiskTestHandler(t)
	defer db.Close()

	req := httptest.NewRequest("GET", "/api/risk/high", nil)
	w := httptest.NewRecorder()

	handler.GetHighRiskEntities(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var entities []*models.RiskObject
	err := json.NewDecoder(w.Body).Decode(&entities)
	if err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	// Should return empty array initially
	if entities == nil {
		t.Error("Expected entities array, got nil")
	}
}

func TestRiskHandler_ProcessEvents_InvalidJSON(t *testing.T) {
	handler, db := setupRiskTestHandler(t)
	defer db.Close()

	req := httptest.NewRequest("POST", "/api/events/batch", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ProcessEvents(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestRiskHandler_MarkEventAsFalsePositive_InvalidJSON(t *testing.T) {
	handler, db := setupRiskTestHandler(t)
	defer db.Close()

	req := httptest.NewRequest("POST", "/api/events/1/false-positive", bytes.NewBufferString("invalid json"))
	req.SetPathValue("id", "1")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.MarkEventAsFalsePositive(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}