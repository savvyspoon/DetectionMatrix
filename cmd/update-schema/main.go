package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Open the database
	db, err := sql.Open("sqlite3", "data/riskmatrix.db")
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	// Check if the database is accessible
	if err := db.Ping(); err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	fmt.Println("Connected to database successfully")

	// Create a backup of the database
	backupPath := "data/riskmatrix_backup.db"
	fmt.Printf("Creating backup at %s\n", backupPath)
	
	// Read the original database
	originalData, err := os.ReadFile("data/riskmatrix.db")
	if err != nil {
		log.Fatalf("Error reading database file: %v", err)
	}
	
	// Write the backup
	err = os.WriteFile(backupPath, originalData, 0644)
	if err != nil {
		log.Fatalf("Error creating backup: %v", err)
	}
	
	fmt.Println("Backup created successfully")

	// Update the schema
	fmt.Println("Updating schema...")

	// Check if the columns already exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('mitre_techniques') WHERE name='tactics'").Scan(&count)
	if err != nil {
		log.Fatalf("Error checking if column exists: %v", err)
	}

	if count > 0 {
		fmt.Println("Schema is already up to date")
		return
	}

	// Add new columns to mitre_techniques table
	_, err = db.Exec(`
		ALTER TABLE mitre_techniques ADD COLUMN tactics TEXT;
		ALTER TABLE mitre_techniques ADD COLUMN domain TEXT;
		ALTER TABLE mitre_techniques ADD COLUMN last_modified TEXT;
		ALTER TABLE mitre_techniques ADD COLUMN detection TEXT;
		ALTER TABLE mitre_techniques ADD COLUMN platforms TEXT;
		ALTER TABLE mitre_techniques ADD COLUMN data_sources TEXT;
		ALTER TABLE mitre_techniques ADD COLUMN is_sub_technique BOOLEAN NOT NULL DEFAULT 0;
		ALTER TABLE mitre_techniques ADD COLUMN sub_technique_of TEXT;
	`)
	if err != nil {
		log.Fatalf("Error updating schema: %v", err)
	}

	fmt.Println("Schema updated successfully")
}