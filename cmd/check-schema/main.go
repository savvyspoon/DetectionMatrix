package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Open database
	db, err := sql.Open("sqlite3", "data/riskmatrix.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	fmt.Println("Checking detections table schema...")

	// Check table schema
	rows, err := db.Query("PRAGMA table_info(detections)")
	if err != nil {
		log.Fatal("Error getting table info:", err)
	}
	defer rows.Close()

	fmt.Println("Column | Type | NotNull | Default | PrimaryKey")
	fmt.Println("-------|------|---------|---------|----------")

	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var defaultValue sql.NullString

		err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			log.Fatal("Error scanning row:", err)
		}

		defaultStr := "NULL"
		if defaultValue.Valid {
			defaultStr = defaultValue.String
		}

		fmt.Printf("%s | %s | %d | %s | %d\n", name, dataType, notNull, defaultStr, pk)
	}

	// Check if query column exists by trying to select it
	fmt.Println("\nTesting query column...")
	_, err = db.Query("SELECT query FROM detections LIMIT 1")
	if err != nil {
		fmt.Printf("Query column test failed: %v\n", err)
	} else {
		fmt.Println("Query column exists and is accessible")
	}
}