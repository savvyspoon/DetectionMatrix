package main

import (
	"fmt"
	"log"

	"riskmatrix/internal/mitre"
	"riskmatrix/pkg/database"
)

func main() {
	// Connect to the database
	db, err := database.New("data/riskmatrix.db")
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Create a new MITRE repository
	repo := mitre.NewRepository(db)

	// Get all techniques
	techniques, err := repo.ListMitreTechniques()
	if err != nil {
		log.Fatalf("Error retrieving techniques: %v", err)
	}

	fmt.Printf("Found %d MITRE techniques in the database\n", len(techniques))

	// Display a few sample techniques
	if len(techniques) > 0 {
		fmt.Println("\nSample techniques:")
		for i, technique := range techniques {
			if i >= 5 {
				break
			}
			fmt.Printf("\nID: %s\n", technique.ID)
			fmt.Printf("Name: %s\n", technique.Name)
			fmt.Printf("Tactic: %s\n", technique.Tactic)
			fmt.Printf("Domain: %s\n", technique.Domain)
			fmt.Printf("Last Modified: %s\n", technique.LastModified)
			fmt.Printf("Is Sub-Technique: %v\n", technique.IsSubTechnique)
			if technique.IsSubTechnique {
				fmt.Printf("Sub-Technique Of: %s\n", technique.SubTechniqueOf)
			}
			fmt.Printf("Tactics: %v\n", technique.Tactics)
			fmt.Printf("Platforms: %v\n", technique.Platforms)
			fmt.Printf("Data Sources: %v\n", technique.DataSources)
		}
	}

	// Count sub-techniques
	var subTechniqueCount int
	for _, technique := range techniques {
		if technique.IsSubTechnique {
			subTechniqueCount++
		}
	}
	fmt.Printf("\nFound %d sub-techniques\n", subTechniqueCount)

	// Count techniques by domain
	domainCounts := make(map[string]int)
	for _, technique := range techniques {
		domainCounts[technique.Domain]++
	}
	fmt.Println("\nTechniques by domain:")
	for domain, count := range domainCounts {
		fmt.Printf("%s: %d\n", domain, count)
	}

	fmt.Println("\nVerification complete!")
}