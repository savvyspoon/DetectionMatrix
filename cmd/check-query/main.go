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

	fmt.Println("Checking query field in detections...")

	// Check the last few detections
	rows, err := db.Query("SELECT id, name, query FROM detections WHERE id >= 10 ORDER BY id")
	if err != nil {
		log.Fatal("Error querying detections:", err)
	}
	defer rows.Close()

	fmt.Println("ID | Name | Query")
	fmt.Println("---|------|------")

	for rows.Next() {
		var id int
		var name string
		var query sql.NullString

		err := rows.Scan(&id, &name, &query)
		if err != nil {
			log.Fatal("Error scanning row:", err)
		}

		queryStr := "NULL"
		if query.Valid {
			if len(query.String) > 50 {
				queryStr = query.String[:50] + "..."
			} else {
				queryStr = query.String
			}
		}

		fmt.Printf("%d | %s | %s\n", id, name, queryStr)
	}
}