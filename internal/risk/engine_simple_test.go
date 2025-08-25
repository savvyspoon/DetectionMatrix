package risk

import (
	"testing"
	"time"

	"riskmatrix/pkg/database"
	"riskmatrix/pkg/models"
)

// setupSimpleTestEngine creates a test risk engine with in-memory database
func setupSimpleTestEngine(t *testing.T) *Engine {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	config := DefaultConfig()
	config.RiskThreshold = 100 // Set higher threshold for testing
	config.DecayFactor = 0.1
	config.DecayInterval = time.Hour

	return NewEngine(db, config)
}

// createSimpleTestDetection creates a test detection
func createSimpleTestDetection(t *testing.T, engine *Engine) *models.Detection {
	detection := &models.Detection{
		Name:        "Test Detection",
		Description: "Test description",
		Status:      models.StatusProduction,
		Severity:    models.SeverityHigh,
		RiskPoints:  25,
	}

	query := `INSERT INTO detections (name, description, status, severity, risk_points, event_count_last_30_days, false_positives_last_30_days, created_at, updated_at) 
              VALUES (?, ?, ?, ?, ?, 0, 0, ?, ?)`

	now := time.Now().Format(time.RFC3339)
	result, err := engine.db.Exec(query, detection.Name, detection.Description, detection.Status, detection.Severity, detection.RiskPoints, now, now)
	if err != nil {
		t.Fatalf("Failed to create test detection: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get detection ID: %v", err)
	}
	detection.ID = id

	return detection
}

// createSimpleRiskObject creates a test risk object
func createSimpleRiskObject(t *testing.T, engine *Engine, entityType models.EntityType, entityValue string) *models.RiskObject {
	query := `INSERT INTO risk_objects (entity_type, entity_value, current_score) VALUES (?, ?, ?)`
	result, err := engine.db.Exec(query, entityType, entityValue, 0)
	if err != nil {
		t.Fatalf("Failed to create risk object: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get risk object ID: %v", err)
	}

	return &models.RiskObject{
		ID:           id,
		EntityType:   entityType,
		EntityValue:  entityValue,
		CurrentScore: 0,
	}
}

func TestEngine_BasicProcessEvent(t *testing.T) {
	engine := setupSimpleTestEngine(t)
	detection := createSimpleTestDetection(t, engine)

	// Create event with RiskObject populated
	event := &models.Event{
		DetectionID: detection.ID,
		RiskPoints:  25,
		RawData:     `{"user": "test@example.com", "action": "login"}`,
		Context:     `{"source_ip": "192.168.1.100"}`,
		RiskObject: &models.RiskObject{
			EntityType:  models.EntityTypeUser,
			EntityValue: "test@example.com",
		},
	}

	err := engine.ProcessEvent(event)
	if err != nil {
		t.Fatalf("Failed to process event: %v", err)
	}

	// Verify event was created
	if event.ID == 0 {
		t.Error("Event ID should be set after processing")
	}

	// Verify risk object was created and score was updated
	riskObj, err := engine.repo.GetRiskObjectByEntity(models.EntityTypeUser, "test@example.com")
	if err != nil {
		t.Errorf("Failed to get risk object: %v", err)
	} else if riskObj.CurrentScore != event.RiskPoints {
		t.Errorf("Expected risk score %d, got %d", event.RiskPoints, riskObj.CurrentScore)
	}
}

func TestEngine_ProcessMultipleEvents(t *testing.T) {
	engine := setupSimpleTestEngine(t)
	detection := createSimpleTestDetection(t, engine)

	riskObject := &models.RiskObject{
		EntityType:  models.EntityTypeHost,
		EntityValue: "workstation-001",
	}

	events := []*models.Event{
		{
			DetectionID: detection.ID,
			RiskPoints:  30,
			RawData:     `{"process": "malware.exe"}`,
			RiskObject:  riskObject,
		},
		{
			DetectionID: detection.ID,
			RiskPoints:  40,
			RawData:     `{"process": "suspicious.exe"}`,
			RiskObject:  riskObject,
		},
		{
			DetectionID: detection.ID,
			RiskPoints:  50,
			RawData:     `{"process": "ransomware.exe"}`,
			RiskObject:  riskObject,
		},
	}

	err := engine.ProcessEvents(events)
	if err != nil {
		t.Fatalf("Failed to process events: %v", err)
	}

	// Verify all events were processed
	for _, event := range events {
		if event.ID == 0 {
			t.Error("Event ID should be set after processing")
		}
	}

	// Verify cumulative risk score
	updatedRiskObj, err := engine.repo.GetRiskObjectByEntity(models.EntityTypeHost, "workstation-001")
	if err != nil {
		t.Fatalf("Failed to get updated risk object: %v", err)
	}

	expectedScore := 30 + 40 + 50 // Sum of all risk points
	if updatedRiskObj.CurrentScore != expectedScore {
		t.Errorf("Expected cumulative risk score %d, got %d", expectedScore, updatedRiskObj.CurrentScore)
	}

	// Check if risk alert was generated (score 120 > threshold 100)
	alerts, err := engine.GetRiskAlerts()
	if err != nil {
		t.Fatalf("Failed to get risk alerts: %v", err)
	}

	// Should have at least one alert
	found := false
	for _, alert := range alerts {
		if alert.EntityID == updatedRiskObj.ID {
			found = true
			if alert.TotalScore != expectedScore {
				t.Errorf("Expected alert score %d, got %d", expectedScore, alert.TotalScore)
			}
			break
		}
	}

	if !found {
		t.Error("Expected risk alert to be generated")
	}
}

func TestEngine_DecayRiskScores(t *testing.T) {
	engine := setupSimpleTestEngine(t)

	// Create risk objects with different scores
	riskObjects := []*models.RiskObject{
		createSimpleRiskObject(t, engine, models.EntityTypeUser, "user1@example.com"),
		createSimpleRiskObject(t, engine, models.EntityTypeHost, "host1"),
		createSimpleRiskObject(t, engine, models.EntityTypeIP, "192.168.1.10"),
	}

	// Set initial scores
	initialScores := []int{100, 50, 200}
	for i, obj := range riskObjects {
		query := `UPDATE risk_objects SET current_score = ? WHERE id = ?`
		_, err := engine.db.Exec(query, initialScores[i], obj.ID)
		if err != nil {
			t.Fatalf("Failed to set initial score: %v", err)
		}
	}

	// Apply decay
	err := engine.DecayRiskScores()
	if err != nil {
		t.Fatalf("Failed to decay risk scores: %v", err)
	}

	// Verify scores were decayed by the configured factor (0.1)
	expectedScores := []int{90, 45, 180} // 100*0.9, 50*0.9, 200*0.9

	for i, obj := range riskObjects {
		updatedObj, err := engine.repo.GetRiskObject(obj.ID)
		if err != nil {
			t.Fatalf("Failed to get updated risk object: %v", err)
		}

		// Allow for small rounding differences
		if abs(updatedObj.CurrentScore-expectedScores[i]) > 1 {
			t.Errorf("Expected score around %d for object %d, got %d",
				expectedScores[i], obj.ID, updatedObj.CurrentScore)
		}
	}
}

func TestEngine_AlertGeneration(t *testing.T) {
	engine := setupSimpleTestEngine(t)
	detection := createSimpleTestDetection(t, engine)

	// Create a risk object with initial score just below threshold
	riskObj := createSimpleRiskObject(t, engine, models.EntityTypeUser, "highrisk@example.com")
	query := `UPDATE risk_objects SET current_score = ? WHERE id = ?`
	_, err := engine.db.Exec(query, 95, riskObj.ID)
	if err != nil {
		t.Fatalf("Failed to set initial score: %v", err)
	}

	// Process event that should trigger alert (95 + 10 = 105 > threshold 100)
	event := &models.Event{
		DetectionID: detection.ID,
		RiskPoints:  10,
		RawData:     `{"action": "privilege_escalation"}`,
		RiskObject: &models.RiskObject{
			EntityType:  models.EntityTypeUser,
			EntityValue: "highrisk@example.com",
		},
	}

	err = engine.ProcessEvent(event)
	if err != nil {
		t.Fatalf("Failed to process event: %v", err)
	}

	// Verify alert was generated
	alerts, err := engine.GetRiskAlerts()
	if err != nil {
		t.Fatalf("Failed to get risk alerts: %v", err)
	}

	found := false
	for _, alert := range alerts {
		if alert.EntityID == riskObj.ID {
			found = true
			if alert.TotalScore != 105 {
				t.Errorf("Expected alert score 105, got %d", alert.TotalScore)
			}
			break
		}
	}

	if !found {
		t.Error("Expected risk alert to be generated")
	}
}

func TestEngine_Configuration(t *testing.T) {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Test with custom configuration
	customConfig := Config{
		RiskThreshold: 200,
		DecayFactor:   0.2,
		DecayInterval: 2 * time.Hour,
	}

	engine := NewEngine(db, customConfig)

	if engine.config.RiskThreshold != 200 {
		t.Errorf("Expected threshold 200, got %d", engine.config.RiskThreshold)
	}
	if engine.config.DecayFactor != 0.2 {
		t.Errorf("Expected decay factor 0.2, got %f", engine.config.DecayFactor)
	}
	if engine.config.DecayInterval != 2*time.Hour {
		t.Errorf("Expected decay interval 2h, got %v", engine.config.DecayInterval)
	}
}

func TestEngine_GetHighRiskEntities(t *testing.T) {
	engine := setupSimpleTestEngine(t)

	// Create risk objects with different scores
	riskObjects := []*models.RiskObject{
		createSimpleRiskObject(t, engine, models.EntityTypeUser, "lowrisk@example.com"),
		createSimpleRiskObject(t, engine, models.EntityTypeHost, "highrisk-host"),
		createSimpleRiskObject(t, engine, models.EntityTypeIP, "192.168.1.100"),
	}

	// Set scores - some above threshold (100), some below
	scores := []int{50, 150, 120}
	for i, obj := range riskObjects {
		query := `UPDATE risk_objects SET current_score = ? WHERE id = ?`
		_, err := engine.db.Exec(query, scores[i], obj.ID)
		if err != nil {
			t.Fatalf("Failed to set score: %v", err)
		}
	}

	// Get high-risk entities
	highRisk, err := engine.GetHighRiskEntities()
	if err != nil {
		t.Fatalf("Failed to get high-risk entities: %v", err)
	}

	// Should have 2 high-risk entities (scores 150 and 120 > threshold 100)
	if len(highRisk) != 2 {
		t.Errorf("Expected 2 high-risk entities, got %d", len(highRisk))
	}

	// Verify the high-risk entities have the expected scores
	foundScores := make(map[int]bool)
	for _, entity := range highRisk {
		foundScores[entity.CurrentScore] = true
	}

	if !foundScores[150] || !foundScores[120] {
		t.Error("Expected to find entities with scores 150 and 120 in high-risk list")
	}
}

// Helper function for absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
