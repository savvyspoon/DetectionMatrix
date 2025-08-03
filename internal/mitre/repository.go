package mitre

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"riskmatrix/pkg/database"
	"riskmatrix/pkg/models"
)

// Repository implements the models.MitreRepository interface
type Repository struct {
	db *database.DB
}

// NewRepository creates a new MITRE repository
func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

// GetMitreTechnique retrieves a MITRE technique by ID
func (r *Repository) GetMitreTechnique(id string) (*models.MitreTechnique, error) {
	query := `SELECT id, name, description, tactic, tactics, domain, last_modified, 
	          detection, platforms, data_sources, is_sub_technique, sub_technique_of 
	          FROM mitre_techniques WHERE id = ?`
	
	row := r.db.QueryRow(query, id)
	
	var technique models.MitreTechnique
	var tacticsJSON, platformsJSON, dataSourcesJSON sql.NullString
	
	err := row.Scan(
		&technique.ID,
		&technique.Name,
		&technique.Description,
		&technique.Tactic,
		&tacticsJSON,
		&technique.Domain,
		&technique.LastModified,
		&technique.Detection,
		&platformsJSON,
		&dataSourcesJSON,
		&technique.IsSubTechnique,
		&technique.SubTechniqueOf,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("MITRE technique not found: %s", id)
		}
		return nil, fmt.Errorf("error scanning MITRE technique: %w", err)
	}
	
	// Parse JSON arrays
	if tacticsJSON.Valid && tacticsJSON.String != "" {
		if err := json.Unmarshal([]byte(tacticsJSON.String), &technique.Tactics); err != nil {
			return nil, fmt.Errorf("error parsing tactics JSON: %w", err)
		}
	}
	
	if platformsJSON.Valid && platformsJSON.String != "" {
		if err := json.Unmarshal([]byte(platformsJSON.String), &technique.Platforms); err != nil {
			return nil, fmt.Errorf("error parsing platforms JSON: %w", err)
		}
	}
	
	if dataSourcesJSON.Valid && dataSourcesJSON.String != "" {
		if err := json.Unmarshal([]byte(dataSourcesJSON.String), &technique.DataSources); err != nil {
			return nil, fmt.Errorf("error parsing data sources JSON: %w", err)
		}
	}
	
	return &technique, nil
}

// ListMitreTechniques retrieves all MITRE techniques
func (r *Repository) ListMitreTechniques() ([]*models.MitreTechnique, error) {
	query := `SELECT id, name, description, tactic, tactics, domain, last_modified, 
	          detection, platforms, data_sources, is_sub_technique, sub_technique_of 
	          FROM mitre_techniques ORDER BY tactic, id`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying MITRE techniques: %w", err)
	}
	defer rows.Close()
	
	var techniques []*models.MitreTechnique
	
	for rows.Next() {
		var technique models.MitreTechnique
		var tacticsJSON, platformsJSON, dataSourcesJSON sql.NullString
		
		err := rows.Scan(
			&technique.ID,
			&technique.Name,
			&technique.Description,
			&technique.Tactic,
			&tacticsJSON,
			&technique.Domain,
			&technique.LastModified,
			&technique.Detection,
			&platformsJSON,
			&dataSourcesJSON,
			&technique.IsSubTechnique,
			&technique.SubTechniqueOf,
		)
		
		if err != nil {
			return nil, fmt.Errorf("error scanning MITRE technique row: %w", err)
		}
		
		// Parse JSON arrays
		if tacticsJSON.Valid && tacticsJSON.String != "" {
			if err := json.Unmarshal([]byte(tacticsJSON.String), &technique.Tactics); err != nil {
				return nil, fmt.Errorf("error parsing tactics JSON: %w", err)
			}
		}
		
		if platformsJSON.Valid && platformsJSON.String != "" {
			if err := json.Unmarshal([]byte(platformsJSON.String), &technique.Platforms); err != nil {
				return nil, fmt.Errorf("error parsing platforms JSON: %w", err)
			}
		}
		
		if dataSourcesJSON.Valid && dataSourcesJSON.String != "" {
			if err := json.Unmarshal([]byte(dataSourcesJSON.String), &technique.DataSources); err != nil {
				return nil, fmt.Errorf("error parsing data sources JSON: %w", err)
			}
		}
		
		techniques = append(techniques, &technique)
	}
	
	return techniques, nil
}

// ListMitreTechniquesByTactic retrieves MITRE techniques by tactic
func (r *Repository) ListMitreTechniquesByTactic(tactic string) ([]*models.MitreTechnique, error) {
	query := `SELECT id, name, description, tactic, tactics, domain, last_modified, 
	          detection, platforms, data_sources, is_sub_technique, sub_technique_of 
	          FROM mitre_techniques WHERE tactic = ? ORDER BY id`
	
	rows, err := r.db.Query(query, tactic)
	if err != nil {
		return nil, fmt.Errorf("error querying MITRE techniques by tactic: %w", err)
	}
	defer rows.Close()
	
	var techniques []*models.MitreTechnique
	
	for rows.Next() {
		var technique models.MitreTechnique
		var tacticsJSON, platformsJSON, dataSourcesJSON sql.NullString
		
		err := rows.Scan(
			&technique.ID,
			&technique.Name,
			&technique.Description,
			&technique.Tactic,
			&tacticsJSON,
			&technique.Domain,
			&technique.LastModified,
			&technique.Detection,
			&platformsJSON,
			&dataSourcesJSON,
			&technique.IsSubTechnique,
			&technique.SubTechniqueOf,
		)
		
		if err != nil {
			return nil, fmt.Errorf("error scanning MITRE technique row: %w", err)
		}
		
		// Parse JSON arrays
		if tacticsJSON.Valid && tacticsJSON.String != "" {
			if err := json.Unmarshal([]byte(tacticsJSON.String), &technique.Tactics); err != nil {
				return nil, fmt.Errorf("error parsing tactics JSON: %w", err)
			}
		}
		
		if platformsJSON.Valid && platformsJSON.String != "" {
			if err := json.Unmarshal([]byte(platformsJSON.String), &technique.Platforms); err != nil {
				return nil, fmt.Errorf("error parsing platforms JSON: %w", err)
			}
		}
		
		if dataSourcesJSON.Valid && dataSourcesJSON.String != "" {
			if err := json.Unmarshal([]byte(dataSourcesJSON.String), &technique.DataSources); err != nil {
				return nil, fmt.Errorf("error parsing data sources JSON: %w", err)
			}
		}
		
		techniques = append(techniques, &technique)
	}
	
	return techniques, nil
}

// CreateMitreTechnique creates a new MITRE technique
func (r *Repository) CreateMitreTechnique(technique *models.MitreTechnique) error {
	query := `INSERT INTO mitre_techniques (
		id, name, description, tactic, tactics, domain, last_modified, 
		detection, platforms, data_sources, is_sub_technique, sub_technique_of
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	// Convert slice fields to JSON
	var tacticsJSON, platformsJSON, dataSourcesJSON []byte
	var err error
	
	if len(technique.Tactics) > 0 {
		tacticsJSON, err = json.Marshal(technique.Tactics)
		if err != nil {
			return fmt.Errorf("error marshaling tactics to JSON: %w", err)
		}
	}
	
	if len(technique.Platforms) > 0 {
		platformsJSON, err = json.Marshal(technique.Platforms)
		if err != nil {
			return fmt.Errorf("error marshaling platforms to JSON: %w", err)
		}
	}
	
	if len(technique.DataSources) > 0 {
		dataSourcesJSON, err = json.Marshal(technique.DataSources)
		if err != nil {
			return fmt.Errorf("error marshaling data sources to JSON: %w", err)
		}
	}
	
	_, err = r.db.Exec(
		query,
		technique.ID,
		technique.Name,
		technique.Description,
		technique.Tactic,
		string(tacticsJSON),
		technique.Domain,
		technique.LastModified,
		technique.Detection,
		string(platformsJSON),
		string(dataSourcesJSON),
		technique.IsSubTechnique,
		technique.SubTechniqueOf,
	)
	
	if err != nil {
		return fmt.Errorf("error creating MITRE technique: %w", err)
	}
	
	return nil
}

// UpdateMitreTechnique updates an existing MITRE technique
func (r *Repository) UpdateMitreTechnique(technique *models.MitreTechnique) error {
	query := `UPDATE mitre_techniques SET 
		name = ?, description = ?, tactic = ?, tactics = ?, domain = ?, 
		last_modified = ?, detection = ?, platforms = ?, data_sources = ?, 
		is_sub_technique = ?, sub_technique_of = ? 
		WHERE id = ?`
	
	// Convert slice fields to JSON
	var tacticsJSON, platformsJSON, dataSourcesJSON []byte
	var err error
	
	if len(technique.Tactics) > 0 {
		tacticsJSON, err = json.Marshal(technique.Tactics)
		if err != nil {
			return fmt.Errorf("error marshaling tactics to JSON: %w", err)
		}
	}
	
	if len(technique.Platforms) > 0 {
		platformsJSON, err = json.Marshal(technique.Platforms)
		if err != nil {
			return fmt.Errorf("error marshaling platforms to JSON: %w", err)
		}
	}
	
	if len(technique.DataSources) > 0 {
		dataSourcesJSON, err = json.Marshal(technique.DataSources)
		if err != nil {
			return fmt.Errorf("error marshaling data sources to JSON: %w", err)
		}
	}
	
	result, err := r.db.Exec(
		query,
		technique.Name,
		technique.Description,
		technique.Tactic,
		string(tacticsJSON),
		technique.Domain,
		technique.LastModified,
		technique.Detection,
		string(platformsJSON),
		string(dataSourcesJSON),
		technique.IsSubTechnique,
		technique.SubTechniqueOf,
		technique.ID,
	)
	
	if err != nil {
		return fmt.Errorf("error updating MITRE technique: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("MITRE technique not found: %s", technique.ID)
	}
	
	return nil
}

// DeleteMitreTechnique deletes a MITRE technique
func (r *Repository) DeleteMitreTechnique(id string) error {
	query := `DELETE FROM mitre_techniques WHERE id = ?`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting MITRE technique: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("MITRE technique not found: %s", id)
	}
	
	return nil
}

// GetCoverageByTactic returns the coverage percentage by tactic
func (r *Repository) GetCoverageByTactic() (map[string]float64, error) {
	query := `
		SELECT 
			mt.tactic,
			COUNT(DISTINCT mt.id) as total,
			COUNT(DISTINCT dmm.mitre_id) as covered
		FROM 
			mitre_techniques mt
		LEFT JOIN 
			detection_mitre_map dmm ON mt.id = dmm.mitre_id
		LEFT JOIN
			detections d ON dmm.detection_id = d.id AND d.status = 'production'
		GROUP BY 
			mt.tactic
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying coverage by tactic: %w", err)
	}
	defer rows.Close()
	
	coverage := make(map[string]float64)
	
	for rows.Next() {
		var tactic string
		var total, covered int
		
		if err := rows.Scan(&tactic, &total, &covered); err != nil {
			return nil, fmt.Errorf("error scanning coverage row: %w", err)
		}
		
		if total > 0 {
			coverage[tactic] = float64(covered) / float64(total) * 100
		} else {
			coverage[tactic] = 0
		}
	}
	
	return coverage, nil
}

// GetDetectionsByTechnique returns detections that cover a specific technique
func (r *Repository) GetDetectionsByTechnique(techniqueID string) ([]*models.Detection, error) {
	query := `
		SELECT 
			d.id, d.name, d.description, d.status, d.severity, d.risk_points, d.playbook_link, d.owner, d.risk_object, d.testing_description, d.created_at, d.updated_at
		FROM 
			detections d
		JOIN 
			detection_mitre_map dmm ON d.id = dmm.detection_id
		WHERE 
			dmm.mitre_id = ?
		ORDER BY 
			d.name
	`
	
	rows, err := r.db.Query(query, techniqueID)
	if err != nil {
		return nil, fmt.Errorf("error querying detections by technique: %w", err)
	}
	defer rows.Close()
	
	var detections []*models.Detection
	
	for rows.Next() {
		var detection models.Detection
		var createdAt, updatedAt string
		var playbookLink, owner, riskObject, testingDescription sql.NullString
		
		err := rows.Scan(
			&detection.ID,
			&detection.Name,
			&detection.Description,
			&detection.Status,
			&detection.Severity,
			&detection.RiskPoints,
			&playbookLink,
			&owner,
			&riskObject,
			&testingDescription,
			&createdAt,
			&updatedAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("error scanning detection row: %w", err)
		}
		
		// Handle nullable fields
		if playbookLink.Valid {
			detection.PlaybookLink = playbookLink.String
		}
		if owner.Valid {
			detection.Owner = owner.String
		}
		if riskObject.Valid {
			detection.RiskObject = models.RiskObjectType(riskObject.String)
		}
		if testingDescription.Valid {
			detection.TestingDescription = testingDescription.String
		}
		
		detections = append(detections, &detection)
	}
	
	return detections, nil
}

// GetCoverageSummary returns overall coverage statistics
func (r *Repository) GetCoverageSummary() (map[string]int, error) {
	// Get total techniques count
	totalTechniquesQuery := `SELECT COUNT(*) FROM mitre_techniques`
	var totalTechniques int
	err := r.db.QueryRow(totalTechniquesQuery).Scan(&totalTechniques)
	if err != nil {
		return nil, fmt.Errorf("error querying total techniques: %w", err)
	}

	// Get covered techniques count (techniques with production detections)
	coveredTechniquesQuery := `
		SELECT COUNT(DISTINCT mt.id) 
		FROM mitre_techniques mt
		JOIN detection_mitre_map dmm ON mt.id = dmm.mitre_id
		JOIN detections d ON dmm.detection_id = d.id AND d.status = 'production'
	`
	var coveredTechniques int
	err = r.db.QueryRow(coveredTechniquesQuery).Scan(&coveredTechniques)
	if err != nil {
		return nil, fmt.Errorf("error querying covered techniques: %w", err)
	}

	// Get total tactics count
	totalTacticsQuery := `SELECT COUNT(DISTINCT tactic) FROM mitre_techniques`
	var totalTactics int
	err = r.db.QueryRow(totalTacticsQuery).Scan(&totalTactics)
	if err != nil {
		return nil, fmt.Errorf("error querying total tactics: %w", err)
	}

	// Get covered tactics count (tactics with at least one covered technique)
	coveredTacticsQuery := `
		SELECT COUNT(DISTINCT mt.tactic) 
		FROM mitre_techniques mt
		JOIN detection_mitre_map dmm ON mt.id = dmm.mitre_id
		JOIN detections d ON dmm.detection_id = d.id AND d.status = 'production'
	`
	var coveredTactics int
	err = r.db.QueryRow(coveredTacticsQuery).Scan(&coveredTactics)
	if err != nil {
		return nil, fmt.Errorf("error querying covered tactics: %w", err)
	}

	summary := map[string]int{
		"totalTechniques":   totalTechniques,
		"coveredTechniques": coveredTechniques,
		"totalTactics":      totalTactics,
		"coveredTactics":    coveredTactics,
	}

	return summary, nil
}