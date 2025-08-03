package detection

import (
	"testing"
	"time"

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

// setupTestRepo creates a detection repository with test database
func setupTestRepo(t *testing.T) (*Repository, *database.DB) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	return repo, db
}

// createTestDetection creates a test detection for use in tests
func createTestDetection(t *testing.T, repo *Repository) *models.Detection {
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

	// Reload to get timestamps
	reloaded, err := repo.GetDetection(detection.ID)
	if err != nil {
		t.Fatalf("Failed to reload test detection: %v", err)
	}

	return reloaded
}

func TestRepository_CreateDetection(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	tests := []struct {
		name      string
		detection models.Detection
		wantError bool
	}{
		{
			name: "Valid detection",
			detection: models.Detection{
				Name:        "Test Detection",
				Description: "Test description",
				Status:      models.StatusIdea,
				Severity:    models.SeverityLow,
				RiskPoints:  25,
				RiskObject:  models.RiskObjectHost,
			},
			wantError: false,
		},
		{
			name: "Detection with all fields",
			detection: models.Detection{
				Name:                     "Complete Detection",
				Description:              "Complete description",
				Status:                   models.StatusProduction,
				Severity:                 models.SeverityHigh,
				RiskPoints:               75,
				PlaybookLink:             "https://example.com/playbook",
				Owner:                    "analyst1",
				RiskObject:               models.RiskObjectUser,
				TestingDescription:       "Comprehensive testing",
				EventCountLast30Days:     20,
				FalsePositivesLast30Days: 3,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.CreateDetection(&tt.detection)

			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.wantError {
				if tt.detection.ID == 0 {
					t.Error("Expected non-zero ID after creation")
				}
				if tt.detection.CreatedAt.IsZero() {
					t.Error("Expected CreatedAt to be set")
				}
				if tt.detection.UpdatedAt.IsZero() {
					t.Error("Expected UpdatedAt to be set")
				}
			}
		})
	}
}

func TestRepository_GetDetection(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Create test detection
	testDetection := createTestDetection(t, repo)

	tests := []struct {
		name      string
		id        int64
		wantError bool
	}{
		{
			name:      "Valid ID",
			id:        testDetection.ID,
			wantError: false,
		},
		{
			name:      "Non-existent ID",
			id:        999999,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detection, err := repo.GetDetection(tt.id)

			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.wantError {
				if detection == nil {
					t.Error("Expected detection but got nil")
				} else {
					if detection.ID != tt.id {
						t.Errorf("Expected ID %d, got %d", tt.id, detection.ID)
					}
					if detection.Name != testDetection.Name {
						t.Errorf("Expected name %s, got %s", testDetection.Name, detection.Name)
					}
				}
			}
		})
	}
}

func TestRepository_ListDetections(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Create multiple test detections
	detection1 := createTestDetection(t, repo)
	
	detection2 := &models.Detection{
		Name:        "Second Detection",
		Description: "Second description",
		Status:      models.StatusProduction,
		Severity:    models.SeverityHigh,
		RiskPoints:  75,
	}
	repo.CreateDetection(detection2)

	detections, err := repo.ListDetections()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(detections) != 2 {
		t.Errorf("Expected 2 detections, got %d", len(detections))
	}

	// Check that detections are returned with relationships loaded
	for _, detection := range detections {
		if detection.MitreTechniques == nil {
			t.Error("Expected MitreTechniques to be initialized")
		}
		if detection.DataSources == nil {
			t.Error("Expected DataSources to be initialized")
		}
	}

	// Verify first detection
	found := false
	for _, detection := range detections {
		if detection.ID == detection1.ID {
			found = true
			if detection.Name != detection1.Name {
				t.Errorf("Expected name %s, got %s", detection1.Name, detection.Name)
			}
			break
		}
	}
	if !found {
		t.Error("First detection not found in list")
	}
}

func TestRepository_ListDetectionsByStatus(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Create detections with different statuses
	detection1 := &models.Detection{
		Name:        "Draft Detection",
		Description: "Draft",
		Status:      models.StatusDraft,
		Severity:    models.SeverityMedium,
		RiskPoints:  50,
	}
	repo.CreateDetection(detection1)

	detection2 := &models.Detection{
		Name:        "Production Detection",
		Description: "Production",
		Status:      models.StatusProduction,
		Severity:    models.SeverityHigh,
		RiskPoints:  75,
	}
	repo.CreateDetection(detection2)

	// Test filtering by draft status
	draftDetections, err := repo.ListDetectionsByStatus(models.StatusDraft)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(draftDetections) != 1 {
		t.Errorf("Expected 1 draft detection, got %d", len(draftDetections))
	} else {
		if draftDetections[0].Status != models.StatusDraft {
			t.Errorf("Expected status %s, got %s", models.StatusDraft, draftDetections[0].Status)
		}
	}

	// Test filtering by production status
	prodDetections, err := repo.ListDetectionsByStatus(models.StatusProduction)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(prodDetections) != 1 {
		t.Errorf("Expected 1 production detection, got %d", len(prodDetections))
	} else {
		if prodDetections[0].Status != models.StatusProduction {
			t.Errorf("Expected status %s, got %s", models.StatusProduction, prodDetections[0].Status)
		}
	}
}

func TestRepository_UpdateDetection(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Create test detection
	testDetection := createTestDetection(t, repo)
	originalUpdatedAt := testDetection.UpdatedAt

	// Wait a moment to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Update detection
	testDetection.Name = "Updated Detection"
	testDetection.Description = "Updated description"
	testDetection.Status = models.StatusProduction
	testDetection.Owner = "updated-owner"

	err := repo.UpdateDetection(testDetection)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Retrieve updated detection
	updatedDetection, err := repo.GetDetection(testDetection.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify updates
	if updatedDetection.Name != "Updated Detection" {
		t.Errorf("Expected name 'Updated Detection', got %s", updatedDetection.Name)
	}
	if updatedDetection.Description != "Updated description" {
		t.Errorf("Expected description 'Updated description', got %s", updatedDetection.Description)
	}
	if updatedDetection.Status != models.StatusProduction {
		t.Errorf("Expected status %s, got %s", models.StatusProduction, updatedDetection.Status)
	}
	if updatedDetection.Owner != "updated-owner" {
		t.Errorf("Expected owner 'updated-owner', got %s", updatedDetection.Owner)
	}

	// Verify UpdatedAt was changed
	if !updatedDetection.UpdatedAt.After(originalUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated")
	}
}

func TestRepository_DeleteDetection(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Create test detection
	testDetection := createTestDetection(t, repo)

	// Delete detection
	err := repo.DeleteDetection(testDetection.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify deletion
	_, err = repo.GetDetection(testDetection.ID)
	if err == nil {
		t.Error("Expected error when getting deleted detection")
	}
}

func TestRepository_GetDetectionCount(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Initially should be 0
	count, err := repo.GetDetectionCount()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}

	// Create test detections
	createTestDetection(t, repo)
	createTestDetection(t, repo)

	// Should now be 2
	count, err = repo.GetDetectionCount()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

func TestRepository_GetDetectionCountByStatus(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Create detections with different statuses
	detection1 := &models.Detection{
		Name:        "Draft Detection 1",
		Description: "Draft 1",
		Status:      models.StatusDraft,
		Severity:    models.SeverityMedium,
		RiskPoints:  50,
	}
	repo.CreateDetection(detection1)

	detection2 := &models.Detection{
		Name:        "Draft Detection 2",
		Description: "Draft 2",
		Status:      models.StatusDraft,
		Severity:    models.SeverityMedium,
		RiskPoints:  50,
	}
	repo.CreateDetection(detection2)

	detection3 := &models.Detection{
		Name:        "Production Detection",
		Description: "Production",
		Status:      models.StatusProduction,
		Severity:    models.SeverityHigh,
		RiskPoints:  75,
	}
	repo.CreateDetection(detection3)

	counts, err := repo.GetDetectionCountByStatus()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if counts[models.StatusDraft] != 2 {
		t.Errorf("Expected 2 draft detections, got %d", counts[models.StatusDraft])
	}

	if counts[models.StatusProduction] != 1 {
		t.Errorf("Expected 1 production detection, got %d", counts[models.StatusProduction])
	}

	if counts[models.StatusIdea] != 0 {
		t.Errorf("Expected 0 idea detections, got %d", counts[models.StatusIdea])
	}
}

func TestRepository_GetFalsePositiveRate(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Create test detection
	testDetection := createTestDetection(t, repo)

	// Initially should be 0 (no events)
	rate, err := repo.GetFalsePositiveRate(testDetection.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if rate != 0.0 {
		t.Errorf("Expected rate 0.0, got %f", rate)
	}

	// Create test risk object for events
	_, err = db.Exec("INSERT INTO risk_objects (entity_type, entity_value, current_score, last_seen) VALUES (?, ?, ?, ?)",
		"user", "test-user", 25, time.Now().Format(time.RFC3339))
	if err != nil {
		t.Fatalf("Failed to create test risk object: %v", err)
	}

	// Create some events (2 total, 1 false positive)
	now := time.Now().Format(time.RFC3339)
	_, err = db.Exec("INSERT INTO events (detection_id, entity_id, timestamp, risk_points, is_false_positive) VALUES (?, ?, ?, ?, ?)",
		testDetection.ID, 1, now, 10, false)
	if err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	_, err = db.Exec("INSERT INTO events (detection_id, entity_id, timestamp, risk_points, is_false_positive) VALUES (?, ?, ?, ?, ?)",
		testDetection.ID, 1, now, 10, true)
	if err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}

	// Should now be 0.5 (1 FP out of 2 total)
	rate, err = repo.GetFalsePositiveRate(testDetection.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if rate != 0.5 {
		t.Errorf("Expected rate 0.5, got %f", rate)
	}
}

func TestRepository_AddRemoveMitreTechnique(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Create test detection
	testDetection := createTestDetection(t, repo)

	// Create test MITRE technique
	_, err := db.Exec("INSERT INTO mitre_techniques (id, name, description, tactic) VALUES (?, ?, ?, ?)",
		"T1059", "Command and Scripting Interpreter", "Test description", "Execution")
	if err != nil {
		t.Fatalf("Failed to create test MITRE technique: %v", err)
	}

	// Add MITRE technique
	err = repo.AddMitreTechnique(testDetection.ID, "T1059")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify association
	detection, err := repo.GetDetection(testDetection.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(detection.MitreTechniques) != 1 {
		t.Errorf("Expected 1 MITRE technique, got %d", len(detection.MitreTechniques))
	}

	if detection.MitreTechniques[0].ID != "T1059" {
		t.Errorf("Expected technique ID T1059, got %s", detection.MitreTechniques[0].ID)
	}

	// Remove MITRE technique
	err = repo.RemoveMitreTechnique(testDetection.ID, "T1059")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify removal
	detection, err = repo.GetDetection(testDetection.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(detection.MitreTechniques) != 0 {
		t.Errorf("Expected 0 MITRE techniques, got %d", len(detection.MitreTechniques))
	}
}

func TestRepository_AddRemoveDataSource(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Create test detection
	testDetection := createTestDetection(t, repo)

	// Create test data source
	_, err := db.Exec("INSERT INTO data_sources (id, name, description) VALUES (?, ?, ?)",
		1, "Test DataSource", "Test description")
	if err != nil {
		t.Fatalf("Failed to create test data source: %v", err)
	}

	// Add data source
	err = repo.AddDataSource(testDetection.ID, 1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify association
	detection, err := repo.GetDetection(testDetection.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(detection.DataSources) != 1 {
		t.Errorf("Expected 1 data source, got %d", len(detection.DataSources))
	}

	if detection.DataSources[0].ID != 1 {
		t.Errorf("Expected data source ID 1, got %d", detection.DataSources[0].ID)
	}

	// Remove data source
	err = repo.RemoveDataSource(testDetection.ID, 1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify removal
	detection, err = repo.GetDetection(testDetection.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(detection.DataSources) != 0 {
		t.Errorf("Expected 0 data sources, got %d", len(detection.DataSources))
	}
}