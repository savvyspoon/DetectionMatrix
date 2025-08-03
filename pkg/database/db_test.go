package database

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNew_InMemory(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory database: %v", err)
	}
	defer db.Close()

	// Verify database is functional
	var result int
	err = db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		t.Errorf("Failed to query database: %v", err)
	}
	if result != 1 {
		t.Errorf("Expected result 1, got %d", result)
	}
}

func TestNew_FileDatabase(t *testing.T) {
	// Create temporary database file
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create file database: %v", err)
	}
	defer db.Close()

	// Verify file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}

	// Verify database is functional
	var result int
	err = db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		t.Errorf("Failed to query database: %v", err)
	}
}

func TestNew_InvalidPath(t *testing.T) {
	// Try to create database in non-existent directory without creating it
	invalidPath := "/non/existent/path/test.db"
	
	_, err := New(invalidPath)
	if err == nil {
		t.Error("Expected error for invalid database path")
	}
}

func TestDatabase_SchemaInitialization(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Test that all expected tables exist
	expectedTables := []string{
		"detections",
		"mitre_techniques", 
		"detection_mitre_map",
		"data_sources",
		"detection_datasource",
		"risk_objects",
		"events",
		"risk_alerts",
		"false_positives",
	}

	for _, table := range expectedTables {
		t.Run("Table_"+table, func(t *testing.T) {
			query := "SELECT name FROM sqlite_master WHERE type='table' AND name=?"
			var tableName string
			err := db.QueryRow(query, table).Scan(&tableName)
			if err != nil {
				t.Errorf("Table %s does not exist: %v", table, err)
			}
			if tableName != table {
				t.Errorf("Expected table %s, got %s", table, tableName)
			}
		})
	}
}

func TestDatabase_ForeignKeyConstraints(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Verify foreign keys are enabled
	var fkEnabled int
	err = db.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled)
	if err != nil {
		t.Fatalf("Failed to check foreign key setting: %v", err)
	}
	if fkEnabled != 1 {
		t.Error("Foreign key constraints should be enabled")
	}

	// Test foreign key constraint enforcement
	t.Run("Detection_to_MITRE_FK", func(t *testing.T) {
		// Try to insert detection_mitre_map with non-existent detection
		_, err := db.Exec("INSERT INTO detection_mitre_map (detection_id, mitre_id) VALUES (?, ?)", 
			99999, "T1059")
		if err == nil {
			t.Error("Expected foreign key constraint violation")
		}
		if !strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			t.Errorf("Expected foreign key error, got: %v", err)
		}
	})

	t.Run("Event_to_Detection_FK", func(t *testing.T) {
		// Insert a valid risk object first
		_, err := db.Exec("INSERT INTO risk_objects (entity_type, entity_value, current_score) VALUES (?, ?, ?)",
			"user", "test@example.com", 0)
		if err != nil {
			t.Fatalf("Failed to insert risk object: %v", err)
		}

		// Try to insert event with non-existent detection
		_, err = db.Exec("INSERT INTO events (detection_id, entity_id, risk_points) VALUES (?, ?, ?)",
			99999, 1, 10)
		if err == nil {
			t.Error("Expected foreign key constraint violation")
		}
	})
}

func TestDatabase_Indexes(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	expectedIndexes := []string{
		"idx_detections_status",
		"idx_events_detection_id", 
		"idx_events_entity_id",
		"idx_events_timestamp",
		"idx_risk_objects_entity",
		"idx_risk_alerts_entity_id",
		"idx_false_positives_event_id",
	}

	for _, index := range expectedIndexes {
		t.Run("Index_"+index, func(t *testing.T) {
			query := "SELECT name FROM sqlite_master WHERE type='index' AND name=?"
			var indexName string
			err := db.QueryRow(query, index).Scan(&indexName)
			if err != nil {
				t.Errorf("Index %s does not exist: %v", index, err)
			}
		})
	}
}

func TestDatabase_DataTypes(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Test detection insertion with all field types
	now := time.Now().Format(time.RFC3339)
	_, err = db.Exec(`INSERT INTO detections 
		(name, description, query, status, severity, risk_points, playbook_link, owner, risk_object, 
		 testing_description, event_count_last_30_days, false_positives_last_30_days, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"Test Detection", "Description", "SELECT * FROM logs", "production", "high", 50,
		"https://playbook.example.com", "owner@example.com", "Host", "Testing notes", 10, 2, now, now)
	
	if err != nil {
		t.Errorf("Failed to insert detection with all fields: %v", err)
	}

	// Test risk object with different entity types
	entityTypes := []string{"user", "host", "IP"}
	for _, entityType := range entityTypes {
		_, err = db.Exec("INSERT INTO risk_objects (entity_type, entity_value, current_score) VALUES (?, ?, ?)",
			entityType, "test-"+entityType, 25)
		if err != nil {
			t.Errorf("Failed to insert risk object with entity type %s: %v", entityType, err)
		}
	}
}

func TestDatabase_Constraints(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	t.Run("RiskObject_CHECK_Constraint", func(t *testing.T) {
		// Test valid risk object types
		validTypes := []string{"IP", "Host", "User"}
		for _, riskType := range validTypes {
			now := time.Now().Format(time.RFC3339)
			_, err := db.Exec(`INSERT INTO detections 
				(name, status, severity, risk_points, risk_object, event_count_last_30_days, false_positives_last_30_days, created_at, updated_at) 
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				"Test "+riskType, "draft", "medium", 10, riskType, 0, 0, now, now)
			if err != nil {
				t.Errorf("Failed to insert detection with valid risk_object %s: %v", riskType, err)
			}
		}

		// Test invalid risk object type
		now := time.Now().Format(time.RFC3339)
		_, err = db.Exec(`INSERT INTO detections 
			(name, status, severity, risk_points, risk_object, event_count_last_30_days, false_positives_last_30_days, created_at, updated_at) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			"Test Invalid", "draft", "medium", 10, "InvalidType", 0, 0, now, now)
		if err == nil {
			t.Error("Expected CHECK constraint violation for invalid risk_object type")
		}
	})

	t.Run("AlertStatus_CHECK_Constraint", func(t *testing.T) {
		// Insert risk object first
		_, err := db.Exec("INSERT INTO risk_objects (entity_type, entity_value, current_score) VALUES (?, ?, ?)",
			"user", "test@example.com", 100)
		if err != nil {
			t.Fatalf("Failed to insert risk object: %v", err)
		}

		// Test valid alert statuses
		validStatuses := []string{"New", "Triage", "Investigation", "On Hold", "Incident", "Closed"}
		for _, status := range validStatuses {
			_, err := db.Exec("INSERT INTO risk_alerts (entity_id, total_score, status) VALUES (?, ?, ?)",
				1, 100, status)
			if err != nil {
				t.Errorf("Failed to insert alert with valid status %s: %v", status, err)
			}
		}

		// Test invalid alert status
		_, err = db.Exec("INSERT INTO risk_alerts (entity_id, total_score, status) VALUES (?, ?, ?)",
			1, 100, "InvalidStatus")
		if err == nil {
			t.Error("Expected CHECK constraint violation for invalid alert status")
		}
	})
}

func TestDatabase_UniqueConstraints(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	t.Run("DataSource_Name_Unique", func(t *testing.T) {
		// Insert first data source
		_, err := db.Exec("INSERT INTO data_sources (name, description) VALUES (?, ?)",
			"Windows Event Logs", "Windows security events")
		if err != nil {
			t.Fatalf("Failed to insert first data source: %v", err)
		}

		// Try to insert duplicate name
		_, err = db.Exec("INSERT INTO data_sources (name, description) VALUES (?, ?)",
			"Windows Event Logs", "Duplicate description")
		if err == nil {
			t.Error("Expected unique constraint violation for duplicate data source name")
		}
	})

	t.Run("RiskObject_Entity_Unique", func(t *testing.T) {
		// Insert first risk object
		_, err := db.Exec("INSERT INTO risk_objects (entity_type, entity_value, current_score) VALUES (?, ?, ?)",
			"user", "john@example.com", 50)
		if err != nil {
			t.Fatalf("Failed to insert first risk object: %v", err)
		}

		// Try to insert duplicate entity
		_, err = db.Exec("INSERT INTO risk_objects (entity_type, entity_value, current_score) VALUES (?, ?, ?)",
			"user", "john@example.com", 75)
		if err == nil {
			t.Error("Expected unique constraint violation for duplicate risk object entity")
		}
	})

	t.Run("MITRE_ID_Primary_Key", func(t *testing.T) {
		// Insert first MITRE technique
		_, err := db.Exec("INSERT INTO mitre_techniques (id, name, tactic) VALUES (?, ?, ?)",
			"T1059", "Command and Scripting Interpreter", "Execution")
		if err != nil {
			t.Fatalf("Failed to insert first MITRE technique: %v", err)
		}

		// Try to insert duplicate ID
		_, err = db.Exec("INSERT INTO mitre_techniques (id, name, tactic) VALUES (?, ?, ?)",
			"T1059", "Duplicate Technique", "Defense Evasion")
		if err == nil {
			t.Error("Expected primary key constraint violation for duplicate MITRE technique ID")
		}
	})
}

func TestDatabase_Transactions(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	t.Run("Successful_Transaction", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// Insert data in transaction
		_, err = tx.Exec("INSERT INTO data_sources (name, description) VALUES (?, ?)", 
			"Test Source 1", "Description 1")
		if err != nil {
			tx.Rollback()
			t.Fatalf("Failed to insert in transaction: %v", err)
		}

		_, err = tx.Exec("INSERT INTO data_sources (name, description) VALUES (?, ?)", 
			"Test Source 2", "Description 2")
		if err != nil {
			tx.Rollback()
			t.Fatalf("Failed to insert second record in transaction: %v", err)
		}

		// Commit transaction
		err = tx.Commit()
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// Verify data was committed
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM data_sources WHERE name LIKE 'Test Source%'").Scan(&count)
		if err != nil {
			t.Fatalf("Failed to count records: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected 2 records after commit, got %d", count)
		}
	})

	t.Run("Rolled_Back_Transaction", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		// Insert data in transaction
		_, err = tx.Exec("INSERT INTO data_sources (name, description) VALUES (?, ?)", 
			"Rollback Test 1", "Description 1")
		if err != nil {
			tx.Rollback()
			t.Fatalf("Failed to insert in transaction: %v", err)
		}

		// Rollback transaction
		err = tx.Rollback()
		if err != nil {
			t.Fatalf("Failed to rollback transaction: %v", err)
		}

		// Verify data was not committed
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM data_sources WHERE name = 'Rollback Test 1'").Scan(&count)
		if err != nil {
			t.Fatalf("Failed to count records: %v", err)
		}
		if count != 0 {
			t.Errorf("Expected 0 records after rollback, got %d", count)
		}
	})
}

func TestDatabase_Close(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	// Verify database is functional before closing
	var result int
	err = db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		t.Errorf("Failed to query database before close: %v", err)
	}

	// Close database
	err = db.Close()
	if err != nil {
		t.Errorf("Failed to close database: %v", err)
	}

	// Verify database is no longer functional after closing
	err = db.QueryRow("SELECT 1").Scan(&result)
	if err == nil {
		t.Error("Expected error querying closed database")
	}
}