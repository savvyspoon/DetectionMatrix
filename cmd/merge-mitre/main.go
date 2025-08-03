package main

import (
	"fmt"
	"log"

	"riskmatrix/internal/mitre"
	"riskmatrix/pkg/database"
)

func main() {
	// Connect to the source database (mitre_import.db)
	sourceDB, err := database.New("data/mitre_import.db")
	if err != nil {
		log.Fatalf("Error connecting to source database: %v", err)
	}
	defer sourceDB.Close()

	// Connect to the target database (riskmatrix.db)
	targetDB, err := database.New("data/riskmatrix.db")
	if err != nil {
		log.Fatalf("Error connecting to target database: %v", err)
	}
	defer targetDB.Close()

	// Create repositories
	sourceRepo := mitre.NewRepository(sourceDB)
	targetRepo := mitre.NewRepository(targetDB)

	// Get all techniques from source database
	techniques, err := sourceRepo.ListMitreTechniques()
	if err != nil {
		log.Fatalf("Error retrieving techniques from source database: %v", err)
	}

	fmt.Printf("Found %d techniques in source database\n", len(techniques))

	// Insert techniques into target database
	var count int
	for _, technique := range techniques {
		// Check if technique already exists in target database
		existingTechnique, err := targetRepo.GetMitreTechnique(technique.ID)
		if err == nil && existingTechnique != nil {
			// Technique exists, update it
			err = targetRepo.UpdateMitreTechnique(technique)
			if err != nil {
				log.Printf("Warning: Error updating technique %s: %v", technique.ID, err)
				continue
			}
			fmt.Printf("Updated technique %s\n", technique.ID)
		} else {
			// Technique doesn't exist, create it
			err = targetRepo.CreateMitreTechnique(technique)
			if err != nil {
				log.Printf("Warning: Error inserting technique %s: %v", technique.ID, err)
				continue
			}
			fmt.Printf("Inserted technique %s\n", technique.ID)
		}

		count++
		if count%100 == 0 {
			fmt.Printf("Processed %d techniques...\n", count)
		}
	}

	fmt.Printf("Successfully processed %d MITRE techniques\n", count)
}