package mitre

import (
	"testing"

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

// setupTestRepo creates a MITRE repository with test database
func setupTestRepo(t *testing.T) (*Repository, *database.DB) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	return repo, db
}

// createTestMitreTechnique creates a test MITRE technique for use in tests
func createTestMitreTechnique(t *testing.T, repo *Repository) *models.MitreTechnique {
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

func TestRepository_CreateMitreTechnique(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	tests := []struct {
		name      string
		technique models.MitreTechnique
		wantError bool
	}{
		{
			name: "Valid technique",
			technique: models.MitreTechnique{
				ID:          "T1001",
				Name:        "Data Obfuscation",
				Description: "Adversaries may obfuscate command and control traffic to make it more difficult to detect.",
				Tactic:      "Command and Control",
			},
			wantError: false,
		},
		{
			name: "Technique with minimal fields",
			technique: models.MitreTechnique{
				ID:     "T1002",
				Name:   "Test Technique",
				Tactic: "Execution",
			},
			wantError: false,
		},
		{
			name: "Duplicate technique ID",
			technique: models.MitreTechnique{
				ID:     "T1001", // Same as first test
				Name:   "Duplicate Technique",
				Tactic: "Execution",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.CreateMitreTechnique(&tt.technique)

			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestRepository_GetMitreTechnique(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Create test technique
	testTechnique := createTestMitreTechnique(t, repo)

	tests := []struct {
		name      string
		id        string
		wantError bool
	}{
		{
			name:      "Valid ID",
			id:        testTechnique.ID,
			wantError: false,
		},
		{
			name:      "Non-existent ID",
			id:        "T9999",
			wantError: true,
		},
		{
			name:      "Empty ID",
			id:        "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			technique, err := repo.GetMitreTechnique(tt.id)

			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.wantError {
				if technique == nil {
					t.Error("Expected technique but got nil")
				} else {
					if technique.ID != tt.id {
						t.Errorf("Expected ID %s, got %s", tt.id, technique.ID)
					}
					if technique.Name != testTechnique.Name {
						t.Errorf("Expected name %s, got %s", testTechnique.Name, technique.Name)
					}
				}
			}
		})
	}
}

func TestRepository_ListMitreTechniques(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Create multiple test techniques
	technique1 := createTestMitreTechnique(t, repo)
	
	technique2 := &models.MitreTechnique{
		ID:          "T1055",
		Name:        "Process Injection",
		Description: "Adversaries may inject code into processes in order to evade process-based defenses.",
		Tactic:      "Defense Evasion",
	}
	repo.CreateMitreTechnique(technique2)

	techniques, err := repo.ListMitreTechniques()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(techniques) != 2 {
		t.Errorf("Expected 2 techniques, got %d", len(techniques))
	}

	// Verify first technique
	found := false
	for _, technique := range techniques {
		if technique.ID == technique1.ID {
			found = true
			if technique.Name != technique1.Name {
				t.Errorf("Expected name %s, got %s", technique1.Name, technique.Name)
			}
			break
		}
	}
	if !found {
		t.Error("First technique not found in list")
	}
}

func TestRepository_ListMitreTechniquesByTactic(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Create techniques with different tactics
	technique1 := &models.MitreTechnique{
		ID:          "T1059",
		Name:        "Command and Scripting Interpreter",
		Description: "Execution technique",
		Tactic:      "Execution",
	}
	repo.CreateMitreTechnique(technique1)

	technique2 := &models.MitreTechnique{
		ID:          "T1055",
		Name:        "Process Injection",
		Description: "Defense Evasion technique",
		Tactic:      "Defense Evasion",
	}
	repo.CreateMitreTechnique(technique2)

	technique3 := &models.MitreTechnique{
		ID:          "T1003",
		Name:        "OS Credential Dumping",
		Description: "Another execution technique",
		Tactic:      "Execution",
	}
	repo.CreateMitreTechnique(technique3)

	// Test filtering by Execution tactic
	executionTechniques, err := repo.ListMitreTechniquesByTactic("Execution")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(executionTechniques) != 2 {
		t.Errorf("Expected 2 execution techniques, got %d", len(executionTechniques))
	}

	for _, technique := range executionTechniques {
		if technique.Tactic != "Execution" {
			t.Errorf("Expected tactic 'Execution', got %s", technique.Tactic)
		}
	}

	// Test filtering by Defense Evasion tactic
	defenseEvasionTechniques, err := repo.ListMitreTechniquesByTactic("Defense Evasion")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(defenseEvasionTechniques) != 1 {
		t.Errorf("Expected 1 defense evasion technique, got %d", len(defenseEvasionTechniques))
	}

	if defenseEvasionTechniques[0].Tactic != "Defense Evasion" {
		t.Errorf("Expected tactic 'Defense Evasion', got %s", defenseEvasionTechniques[0].Tactic)
	}

	// Test filtering by non-existent tactic
	nonExistentTechniques, err := repo.ListMitreTechniquesByTactic("NonExistent")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(nonExistentTechniques) != 0 {
		t.Errorf("Expected 0 techniques for non-existent tactic, got %d", len(nonExistentTechniques))
	}
}

func TestRepository_UpdateMitreTechnique(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Create test technique
	testTechnique := createTestMitreTechnique(t, repo)

	// Update technique
	testTechnique.Name = "Updated Command and Scripting Interpreter"
	testTechnique.Description = "Updated description for command and scripting interpreter technique."
	testTechnique.Tactic = "Execution"

	err := repo.UpdateMitreTechnique(testTechnique)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Retrieve updated technique
	updatedTechnique, err := repo.GetMitreTechnique(testTechnique.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify updates
	if updatedTechnique.Name != "Updated Command and Scripting Interpreter" {
		t.Errorf("Expected name 'Updated Command and Scripting Interpreter', got %s", updatedTechnique.Name)
	}
	if updatedTechnique.Description != "Updated description for command and scripting interpreter technique." {
		t.Errorf("Expected updated description, got %s", updatedTechnique.Description)
	}
	if updatedTechnique.Tactic != "Execution" {
		t.Errorf("Expected tactic 'Execution', got %s", updatedTechnique.Tactic)
	}
}

func TestRepository_DeleteMitreTechnique(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Create test technique
	testTechnique := createTestMitreTechnique(t, repo)

	// Delete technique
	err := repo.DeleteMitreTechnique(testTechnique.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify deletion
	_, err = repo.GetMitreTechnique(testTechnique.ID)
	if err == nil {
		t.Error("Expected error when getting deleted technique")
	}
}

func TestRepository_GetDetectionsByTechnique(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Create test technique
	testTechnique := createTestMitreTechnique(t, repo)

	// Create test detection
	_, err := db.Exec("INSERT INTO detections (name, description, status, severity, risk_points, playbook_link, owner, risk_object, testing_description, event_count_last_30_days, false_positives_last_30_days) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"Test Detection", "Test description", "draft", "medium", 50, nil, nil, nil, nil, 0, 0)
	if err != nil {
		t.Fatalf("Failed to create test detection: %v", err)
	}

	// Create association
	_, err = db.Exec("INSERT INTO detection_mitre_map (detection_id, mitre_id) VALUES (?, ?)",
		1, testTechnique.ID)
	if err != nil {
		t.Fatalf("Failed to create test association: %v", err)
	}

	// Test getting detections for technique
	detections, err := repo.GetDetectionsByTechnique(testTechnique.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(detections) != 1 {
		t.Errorf("Expected 1 detection, got %d", len(detections))
	}

	if detections[0].Name != "Test Detection" {
		t.Errorf("Expected detection name 'Test Detection', got %s", detections[0].Name)
	}

	// Test getting detections for non-existent technique
	nonExistentDetections, err := repo.GetDetectionsByTechnique("T9999")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(nonExistentDetections) != 0 {
		t.Errorf("Expected 0 detections for non-existent technique, got %d", len(nonExistentDetections))
	}
}

func TestRepository_GetCoverageByTactic(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Create techniques with different tactics
	technique1 := &models.MitreTechnique{
		ID:     "T1059",
		Name:   "Command and Scripting Interpreter",
		Tactic: "Execution",
	}
	repo.CreateMitreTechnique(technique1)

	technique2 := &models.MitreTechnique{
		ID:     "T1055",
		Name:   "Process Injection",
		Tactic: "Defense Evasion",
	}
	repo.CreateMitreTechnique(technique2)

	technique3 := &models.MitreTechnique{
		ID:     "T1003",
		Name:   "OS Credential Dumping",
		Tactic: "Execution",
	}
	repo.CreateMitreTechnique(technique3)

	// Create detections
	_, err := db.Exec("INSERT INTO detections (name, description, status, severity, risk_points, playbook_link, owner, risk_object, testing_description, event_count_last_30_days, false_positives_last_30_days) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"Detection 1", "Test description", "draft", "medium", 50, nil, nil, nil, nil, 0, 0)
	if err != nil {
		t.Fatalf("Failed to create test detection: %v", err)
	}

	_, err = db.Exec("INSERT INTO detections (name, description, status, severity, risk_points, playbook_link, owner, risk_object, testing_description, event_count_last_30_days, false_positives_last_30_days) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"Detection 2", "Test description", "draft", "medium", 50, nil, nil, nil, nil, 0, 0)
	if err != nil {
		t.Fatalf("Failed to create test detection: %v", err)
	}

	// Create associations (2 detections for Execution, 0 for Defense Evasion)
	_, err = db.Exec("INSERT INTO detection_mitre_map (detection_id, mitre_id) VALUES (?, ?)", 1, "T1059")
	if err != nil {
		t.Fatalf("Failed to create test association: %v", err)
	}

	_, err = db.Exec("INSERT INTO detection_mitre_map (detection_id, mitre_id) VALUES (?, ?)", 2, "T1003")
	if err != nil {
		t.Fatalf("Failed to create test association: %v", err)
	}

	// Test coverage calculation
	coverage, err := repo.GetCoverageByTactic()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should have coverage data for tactics
	if coverage == nil {
		t.Error("Expected coverage data, got nil")
	}

	// The exact structure depends on implementation, but should contain tactic information
	if len(coverage) == 0 {
		t.Error("Expected non-empty coverage data")
	}
}
