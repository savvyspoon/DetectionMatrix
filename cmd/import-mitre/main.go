package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"riskmatrix/internal/mitre"
	"riskmatrix/pkg/database"
	"riskmatrix/pkg/models"
)

func main() {
	// Parse command line flags
	dbPath := flag.String("db", "data/riskmatrix.db", "Path to SQLite database file")
	csvPath := flag.String("csv", "data/mitre.csv", "Path to MITRE CSV file")
	flag.Parse()

	// Ensure data directory exists
	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Open the CSV file
	file, err := os.Open(*csvPath)
	if err != nil {
		log.Fatalf("Error opening CSV file: %v", err)
	}
	defer file.Close()

	// Create a new CSV reader
	reader := csv.NewReader(file)
	reader.Comma = '\t' // Set delimiter to tab
	reader.LazyQuotes = true // Handle quotes in the data
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	// Read the header row
	header, err := reader.Read()
	if err != nil {
		log.Fatalf("Error reading CSV header: %v", err)
	}

	// Create a map of column indices
	columnMap := make(map[string]int)
	for i, column := range header {
		columnMap[column] = i
	}

	// Connect to the database
	db, err := database.New(*dbPath)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Create a new MITRE repository
	repo := mitre.NewRepository(db)

	// Read and process each row
	var count int
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Warning: Error reading row: %v", err)
			continue
		}

		// Create a new MITRE technique
		technique := &models.MitreTechnique{
			ID:          row[columnMap["ID"]],
			Name:        row[columnMap["name"]],
			Description: row[columnMap["description"]],
			Tactic:      extractPrimaryTactic(row[columnMap["tactics"]]),
		}

		// Set additional fields
		if idx, ok := columnMap["last modified"]; ok && idx < len(row) {
			technique.LastModified = row[idx]
		}

		if idx, ok := columnMap["domain"]; ok && idx < len(row) {
			technique.Domain = row[idx]
		}

		if idx, ok := columnMap["tactics"]; ok && idx < len(row) {
			technique.Tactics = splitAndTrim(row[idx], ",")
		}

		if idx, ok := columnMap["detection"]; ok && idx < len(row) {
			technique.Detection = row[idx]
		}

		if idx, ok := columnMap["platforms"]; ok && idx < len(row) {
			technique.Platforms = splitAndTrim(row[idx], ",")
		}

		if idx, ok := columnMap["data sources"]; ok && idx < len(row) {
			technique.DataSources = splitAndTrim(row[idx], ",")
		}

		if idx, ok := columnMap["is sub-technique"]; ok && idx < len(row) {
			isSubTechnique, err := strconv.ParseBool(row[idx])
			if err == nil {
				technique.IsSubTechnique = isSubTechnique
			}
		}

		if idx, ok := columnMap["sub-technique of"]; ok && idx < len(row) {
			technique.SubTechniqueOf = row[idx]
		}

		// Insert the technique into the database
		err = repo.CreateMitreTechnique(technique)
		if err != nil {
			log.Printf("Warning: Error inserting technique %s: %v", technique.ID, err)
			continue
		}

		count++
		if count%100 == 0 {
			fmt.Printf("Imported %d techniques...\n", count)
		}
	}

	fmt.Printf("Successfully imported %d MITRE techniques\n", count)
}

// extractPrimaryTactic extracts the first tactic from a comma-separated list
func extractPrimaryTactic(tactics string) string {
	parts := splitAndTrim(tactics, ",")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// splitAndTrim splits a string by a delimiter and trims whitespace from each part
func splitAndTrim(s string, delimiter string) []string {
	if s == "" {
		return []string{}
	}
	
	parts := strings.Split(s, delimiter)
	result := make([]string, 0, len(parts))
	
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	
	return result
}