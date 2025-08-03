package datasource

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

// setupTestRepo creates a datasource repository with test database
func setupTestRepo(t *testing.T) (*Repository, *database.DB) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	return repo, db
}

// createTestDataSource creates a test data source for use in tests
func createTestDataSource(t *testing.T, repo *Repository) *models.DataSource {
	dataSource := &models.DataSource{
		Name:        "sysmon",
		Description: "System Monitor logs",
		LogFormat:   "XML",
	}

	err := repo.CreateDataSource(dataSource)
	if err != nil {
		t.Fatalf("Failed to create test data source: %v", err)
	}

	return dataSource
}

func TestRepository_CreateDataSource(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	tests := []struct {
		name       string
		dataSource models.DataSource
		wantError  bool
	}{
		{
			name: "Valid data source",
			dataSource: models.DataSource{
				Name:        "cloudtrail",
				Description: "AWS CloudTrail logs",
				LogFormat:   "JSON",
			},
			wantError: false,
		},
		{
			name: "Data source with minimal fields",
			dataSource: models.DataSource{
				Name: "firewall",
			},
			wantError: false,
		},
		{
			name: "Duplicate data source name",
			dataSource: models.DataSource{
				Name:        "cloudtrail", // Same as first test
				Description: "Duplicate data source",
				LogFormat:   "JSON",
			},
			wantError: true,
		},
		{
			name: "Empty name",
			dataSource: models.DataSource{
				Name:        "",
				Description: "Data source with empty name",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.CreateDataSource(&tt.dataSource)

			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.wantError {
				if tt.dataSource.ID == 0 {
					t.Error("Expected non-zero ID after creation")
				}
			}
		})
	}
}

func TestRepository_GetDataSource(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Create test data source
	testDataSource := createTestDataSource(t, repo)

	tests := []struct {
		name      string
		id        int64
		wantError bool
	}{
		{
			name:      "Valid ID",
			id:        testDataSource.ID,
			wantError: false,
		},
		{
			name:      "Non-existent ID",
			id:        9999,
			wantError: true,
		},
		{
			name:      "Zero ID",
			id:        0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataSource, err := repo.GetDataSource(tt.id)

			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.wantError {
				if dataSource == nil {
					t.Error("Expected data source but got nil")
				} else {
					if dataSource.ID != tt.id {
						t.Errorf("Expected ID %d, got %d", tt.id, dataSource.ID)
					}
					if dataSource.Name != testDataSource.Name {
						t.Errorf("Expected name %s, got %s", testDataSource.Name, dataSource.Name)
					}
				}
			}
		})
	}
}

func TestRepository_GetDataSourceByName(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Create test data source
	testDataSource := createTestDataSource(t, repo)

	tests := []struct {
		name      string
		dsName    string
		wantError bool
	}{
		{
			name:      "Valid name",
			dsName:    testDataSource.Name,
			wantError: false,
		},
		{
			name:      "Non-existent name",
			dsName:    "nonexistent",
			wantError: true,
		},
		{
			name:      "Empty name",
			dsName:    "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataSource, err := repo.GetDataSourceByName(tt.dsName)

			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.wantError {
				if dataSource == nil {
					t.Error("Expected data source but got nil")
				} else {
					if dataSource.Name != tt.dsName {
						t.Errorf("Expected name %s, got %s", tt.dsName, dataSource.Name)
					}
					if dataSource.ID != testDataSource.ID {
						t.Errorf("Expected ID %d, got %d", testDataSource.ID, dataSource.ID)
					}
				}
			}
		})
	}
}

func TestRepository_ListDataSources(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Create multiple test data sources
	dataSource1 := createTestDataSource(t, repo)

	dataSource2 := &models.DataSource{
		Name:        "windows_event_logs",
		Description: "Windows Event Logs",
		LogFormat:   "EVTX",
	}
	repo.CreateDataSource(dataSource2)

	dataSources, err := repo.ListDataSources()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(dataSources) != 2 {
		t.Errorf("Expected 2 data sources, got %d", len(dataSources))
	}

	// Verify first data source
	found := false
	for _, ds := range dataSources {
		if ds.ID == dataSource1.ID {
			found = true
			if ds.Name != dataSource1.Name {
				t.Errorf("Expected name %s, got %s", dataSource1.Name, ds.Name)
			}
			break
		}
	}
	if !found {
		t.Error("First data source not found in list")
	}
}

func TestRepository_UpdateDataSource(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Create test data source
	testDataSource := createTestDataSource(t, repo)

	// Update data source
	testDataSource.Name = "updated_sysmon"
	testDataSource.Description = "Updated System Monitor logs"
	testDataSource.LogFormat = "JSON"

	err := repo.UpdateDataSource(testDataSource)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Retrieve updated data source
	updatedDataSource, err := repo.GetDataSource(testDataSource.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify updates
	if updatedDataSource.Name != "updated_sysmon" {
		t.Errorf("Expected name 'updated_sysmon', got %s", updatedDataSource.Name)
	}
	if updatedDataSource.Description != "Updated System Monitor logs" {
		t.Errorf("Expected updated description, got %s", updatedDataSource.Description)
	}
	if updatedDataSource.LogFormat != "JSON" {
		t.Errorf("Expected log format 'JSON', got %s", updatedDataSource.LogFormat)
	}
}

func TestRepository_DeleteDataSource(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Create test data source
	testDataSource := createTestDataSource(t, repo)

	// Delete data source
	err := repo.DeleteDataSource(testDataSource.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify deletion
	_, err = repo.GetDataSource(testDataSource.ID)
	if err == nil {
		t.Error("Expected error when getting deleted data source")
	}

	// Test deleting non-existent data source
	err = repo.DeleteDataSource(9999)
	if err == nil {
		t.Error("Expected error when deleting non-existent data source")
	}
}

func TestRepository_GetDetectionsByDataSource(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Create test data source
	testDataSource := createTestDataSource(t, repo)

	// Create test detection
	_, err := db.Exec("INSERT INTO detections (name, description, status, severity, risk_points, playbook_link, owner, risk_object, testing_description, event_count_last_30_days, false_positives_last_30_days) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"Test Detection", "Test description", "draft", "medium", 50, nil, nil, nil, nil, 0, 0)
	if err != nil {
		t.Fatalf("Failed to create test detection: %v", err)
	}

	// Create association
	_, err = db.Exec("INSERT INTO detection_datasource (detection_id, datasource_id) VALUES (?, ?)",
		1, testDataSource.ID)
	if err != nil {
		t.Fatalf("Failed to create test association: %v", err)
	}

	// Test getting detections for data source
	detections, err := repo.GetDetectionsByDataSource(testDataSource.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(detections) != 1 {
		t.Errorf("Expected 1 detection, got %d", len(detections))
	}

	if detections[0].Name != "Test Detection" {
		t.Errorf("Expected detection name 'Test Detection', got %s", detections[0].Name)
	}

	// Test getting detections for non-existent data source
	nonExistentDetections, err := repo.GetDetectionsByDataSource(9999)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(nonExistentDetections) != 0 {
		t.Errorf("Expected 0 detections for non-existent data source, got %d", len(nonExistentDetections))
	}
}

func TestRepository_GetMitreTechniquesByDataSource(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Create test data source
	testDataSource := createTestDataSource(t, repo)

	// Create test MITRE technique
	_, err := db.Exec("INSERT INTO mitre_techniques (id, name, description, tactic) VALUES (?, ?, ?, ?)",
		"T1059", "Command and Scripting Interpreter", "Test description", "Execution")
	if err != nil {
		t.Fatalf("Failed to create test MITRE technique: %v", err)
	}

	// Create test detection
	_, err = db.Exec("INSERT INTO detections (name, description, status, severity, risk_points, playbook_link, owner, risk_object, testing_description, event_count_last_30_days, false_positives_last_30_days) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"Test Detection", "Test description", "draft", "medium", 50, nil, nil, nil, nil, 0, 0)
	if err != nil {
		t.Fatalf("Failed to create test detection: %v", err)
	}

	// Create associations
	_, err = db.Exec("INSERT INTO detection_mitre_map (detection_id, mitre_id) VALUES (?, ?)",
		1, "T1059")
	if err != nil {
		t.Fatalf("Failed to create detection-mitre association: %v", err)
	}

	_, err = db.Exec("INSERT INTO detection_datasource (detection_id, datasource_id) VALUES (?, ?)",
		1, testDataSource.ID)
	if err != nil {
		t.Fatalf("Failed to create detection-datasource association: %v", err)
	}

	// Test getting MITRE techniques for data source
	techniques, err := repo.GetMitreTechniquesByDataSource(testDataSource.ID)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(techniques) != 1 {
		t.Errorf("Expected 1 technique, got %d", len(techniques))
	}

	if techniques[0].ID != "T1059" {
		t.Errorf("Expected technique ID 'T1059', got %s", techniques[0].ID)
	}

	if techniques[0].Name != "Command and Scripting Interpreter" {
		t.Errorf("Expected technique name 'Command and Scripting Interpreter', got %s", techniques[0].Name)
	}

	// Test getting techniques for non-existent data source
	nonExistentTechniques, err := repo.GetMitreTechniquesByDataSource(9999)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(nonExistentTechniques) != 0 {
		t.Errorf("Expected 0 techniques for non-existent data source, got %d", len(nonExistentTechniques))
	}
}

func TestRepository_GetDataSourceUtilization(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Create test data sources
	dataSource1 := createTestDataSource(t, repo)

	dataSource2 := &models.DataSource{
		Name:        "windows_event_logs",
		Description: "Windows Event Logs",
		LogFormat:   "EVTX",
	}
	repo.CreateDataSource(dataSource2)

	dataSource3 := &models.DataSource{
		Name:        "unused_source",
		Description: "Unused data source",
		LogFormat:   "JSON",
	}
	repo.CreateDataSource(dataSource3)

	// Create test detections
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

	_, err = db.Exec("INSERT INTO detections (name, description, status, severity, risk_points, playbook_link, owner, risk_object, testing_description, event_count_last_30_days, false_positives_last_30_days) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		"Detection 3", "Test description", "draft", "medium", 50, nil, nil, nil, nil, 0, 0)
	if err != nil {
		t.Fatalf("Failed to create test detection: %v", err)
	}

	// Create associations (2 detections for dataSource1, 1 for dataSource2, 0 for dataSource3)
	_, err = db.Exec("INSERT INTO detection_datasource (detection_id, datasource_id) VALUES (?, ?)", 1, dataSource1.ID)
	if err != nil {
		t.Fatalf("Failed to create test association: %v", err)
	}

	_, err = db.Exec("INSERT INTO detection_datasource (detection_id, datasource_id) VALUES (?, ?)", 2, dataSource1.ID)
	if err != nil {
		t.Fatalf("Failed to create test association: %v", err)
	}

	_, err = db.Exec("INSERT INTO detection_datasource (detection_id, datasource_id) VALUES (?, ?)", 3, dataSource2.ID)
	if err != nil {
		t.Fatalf("Failed to create test association: %v", err)
	}

	// Test utilization calculation
	utilization, err := repo.GetDataSourceUtilization()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if utilization == nil {
		t.Error("Expected utilization data, got nil")
	}

	if utilization[dataSource1.Name] != 2 {
		t.Errorf("Expected 2 detections for %s, got %d", dataSource1.Name, utilization[dataSource1.Name])
	}

	if utilization[dataSource2.Name] != 1 {
		t.Errorf("Expected 1 detection for %s, got %d", dataSource2.Name, utilization[dataSource2.Name])
	}

	if utilization[dataSource3.Name] != 0 {
		t.Errorf("Expected 0 detections for %s, got %d", dataSource3.Name, utilization[dataSource3.Name])
	}
}

func TestRepository_ErrorHandling(t *testing.T) {
	repo, db := setupTestRepo(t)
	defer db.Close()

	// Test creating data source with invalid data
	invalidDataSource := &models.DataSource{
		Name: "", // Empty name should cause error
	}

	err := repo.CreateDataSource(invalidDataSource)
	if err == nil {
		t.Error("Expected error when creating data source with empty name")
	}

	// Test updating non-existent data source
	nonExistentDataSource := &models.DataSource{
		ID:   9999,
		Name: "nonexistent",
	}

	err = repo.UpdateDataSource(nonExistentDataSource)
	if err == nil {
		t.Error("Expected error when updating non-existent data source")
	}
}