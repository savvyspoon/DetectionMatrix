package risk

import (
	"fmt"
	"testing"
	"time"

	"riskmatrix/pkg/database"
	"riskmatrix/pkg/models"
)

func setupTestDB(t *testing.T) *database.DB {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	return db
}

func setupTestRepo(t *testing.T) (*Repository, *database.DB) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	return repo, db
}

func createTestRiskObject(t *testing.T, db *database.DB) *models.RiskObject {
	// Create risk object directly in database for testing with unique value
	query := `INSERT INTO risk_objects (entity_type, entity_value, current_score, last_seen) VALUES (?, ?, ?, ?)`
	now := time.Now()
	uniqueValue := fmt.Sprintf("test-user-%d", now.UnixNano())
	result, err := db.Exec(query, models.EntityTypeUser, uniqueValue, 25, now.Format(time.RFC3339))
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
		EntityValue:  uniqueValue,
		CurrentScore: 25,
		LastSeen:     now,
	}
}

func TestRepository_DecayRiskScores(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Create test risk object
	testObj := createTestRiskObject(t, db)

	// Test decay
	err := repo.DecayRiskScores(0.1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify score was decayed
	updated, err := repo.GetRiskObject(testObj.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedScore := int(float64(testObj.CurrentScore) * (1 - 0.1))
	if updated.CurrentScore != expectedScore {
		t.Errorf("Expected score %d, got %d", expectedScore, updated.CurrentScore)
	}
}

func TestRepository_GetRiskObject(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	testObj := createTestRiskObject(t, db)

	obj, err := repo.GetRiskObject(testObj.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if obj.EntityValue != testObj.EntityValue {
		t.Errorf("Expected entity value %s, got %s", testObj.EntityValue, obj.EntityValue)
	}
}

func TestRepository_GetRiskObjectByEntity(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	testObj := createTestRiskObject(t, db)

	obj, err := repo.GetRiskObjectByEntity(testObj.EntityType, testObj.EntityValue)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if obj.ID != testObj.ID {
		t.Errorf("Expected ID %d, got %d", testObj.ID, obj.ID)
	}
}

func TestRepository_ListRiskObjects(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Create first risk object
	createTestRiskObject(t, db)
	
	// Create second risk object with different entity type
	query := `INSERT INTO risk_objects (entity_type, entity_value, current_score, last_seen) VALUES (?, ?, ?, ?)`
	now := time.Now()
	_, err := db.Exec(query, models.EntityTypeHost, "test-host", 30, now.Format(time.RFC3339))
	if err != nil {
		t.Fatalf("Failed to create second test risk object: %v", err)
	}

	objects, err := repo.ListRiskObjects()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(objects) != 2 {
		t.Errorf("Expected 2 objects, got %d", len(objects))
	}
}

func TestRepository_ListHighRiskObjects(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Create high risk object
	query := `INSERT INTO risk_objects (entity_type, entity_value, current_score, last_seen) VALUES (?, ?, ?, ?)`
	now := time.Now()
	_, err := db.Exec(query, models.EntityTypeUser, "high-risk-user", 75, now.Format(time.RFC3339))
	if err != nil {
		t.Fatalf("Failed to create high risk object: %v", err)
	}

	// Create low risk object
	_, err = db.Exec(query, models.EntityTypeUser, "low-risk-user", 25, now.Format(time.RFC3339))
	if err != nil {
		t.Fatalf("Failed to create low risk object: %v", err)
	}

	objects, err := repo.ListHighRiskObjects(50)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(objects) != 1 {
		t.Errorf("Expected 1 high risk object, got %d", len(objects))
	}

	if objects[0].CurrentScore < 50 {
		t.Errorf("Expected high risk object with score >= 50, got %d", objects[0].CurrentScore)
	}
}

func TestRepository_ListEventsByEntity(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Create detection
	_, err := db.Exec("INSERT INTO detections (name, description, status, severity, risk_points) VALUES (?, ?, ?, ?, ?)",
		"Test Detection", "Test", "draft", "medium", 50)
	if err != nil {
		t.Fatalf("Failed to create test detection: %v", err)
	}

	testObj := createTestRiskObject(t, db)

	// Create event directly in database
	now := time.Now()
	_, err = db.Exec("INSERT INTO events (detection_id, entity_id, timestamp, raw_data, risk_points) VALUES (?, ?, ?, ?, ?)",
		1, testObj.ID, now.Format(time.RFC3339), "test event data", 10)
	if err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	events, err := repo.ListEventsByEntity(testObj.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
}

func TestRepository_ListRiskAlerts(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	testObj := createTestRiskObject(t, db)

	// Create risk alert directly in database
	now := time.Now()
	_, err := db.Exec("INSERT INTO risk_alerts (entity_id, triggered_at, total_score) VALUES (?, ?, ?)",
		testObj.ID, now.Format(time.RFC3339), 75)
	if err != nil {
		t.Fatalf("Failed to create test risk alert: %v", err)
	}

	alerts, err := repo.ListRiskAlerts()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(alerts) != 1 {
		t.Errorf("Expected 1 alert, got %d", len(alerts))
	}
}