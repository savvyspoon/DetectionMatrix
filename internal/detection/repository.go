package detection

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"riskmatrix/pkg/database"
	"riskmatrix/pkg/models"
)

// Repository implements the detection-related data access
type Repository struct {
	db *database.DB
}

// NewRepository creates a new detection repository
func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

// GetDetection retrieves a detection by ID
func (r *Repository) GetDetection(id int64) (*models.Detection, error) {
	query := `SELECT id, name, description, query, status, severity, risk_points, playbook_link, owner, risk_object, testing_description, event_count_last_30_days, false_positives_last_30_days, created_at, updated_at 
              FROM detections WHERE id = ?`
	
	row := r.db.QueryRow(query, id)
	
	var detection models.Detection
	var createdAt, updatedAt string
	
	var playbookLink, owner, riskObject, testingDescription, queryField sql.NullString
	var description sql.NullString
	
	err := row.Scan(
		&detection.ID,
		&detection.Name,
		&description,
		&queryField,
		&detection.Status,
		&detection.Severity,
		&detection.RiskPoints,
		&playbookLink,
		&owner,
		&riskObject,
		&testingDescription,
		&detection.EventCountLast30Days,
		&detection.FalsePositivesLast30Days,
		&createdAt,
		&updatedAt,
	)
	
	// Handle nullable fields
	if description.Valid {
		detection.Description = description.String
	}
	if queryField.Valid {
		detection.Query = queryField.String
	}
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
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("detection not found: %d", id)
		}
		return nil, fmt.Errorf("error scanning detection: %w", err)
	}
	
	// Parse timestamps (SQLite format)
	if parsedTime, err := time.Parse("2006-01-02 15:04:05", createdAt); err == nil {
		detection.CreatedAt = parsedTime
	} else if parsedTime, err := time.Parse(time.RFC3339, createdAt); err == nil {
		detection.CreatedAt = parsedTime
	}
	
	if parsedTime, err := time.Parse("2006-01-02 15:04:05", updatedAt); err == nil {
		detection.UpdatedAt = parsedTime
	} else if parsedTime, err := time.Parse(time.RFC3339, updatedAt); err == nil {
		detection.UpdatedAt = parsedTime
	}
	
	// Load relationships
	if err := r.loadMitreTechniques(&detection); err != nil {
		return nil, err
	}
	
	if err := r.loadDataSources(&detection); err != nil {
		return nil, err
	}
	
	
	return &detection, nil
}

// ListDetections retrieves all detections
func (r *Repository) ListDetections() ([]*models.Detection, error) {
	query := `SELECT id, name, description, query, status, severity, risk_points, playbook_link, owner, risk_object, testing_description, event_count_last_30_days, false_positives_last_30_days, created_at, updated_at 
              FROM detections ORDER BY name`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying detections: %w", err)
	}
	defer rows.Close()
	
	var detections []*models.Detection
	
	for rows.Next() {
		var detection models.Detection
		var createdAt, updatedAt string
		var playbookLink, owner, riskObject, testingDescription, queryField sql.NullString
		var description sql.NullString
		
		err := rows.Scan(
			&detection.ID,
			&detection.Name,
			&description,
			&queryField,
			&detection.Status,
			&detection.Severity,
			&detection.RiskPoints,
			&playbookLink,
			&owner,
			&riskObject,
			&testingDescription,
			&detection.EventCountLast30Days,
			&detection.FalsePositivesLast30Days,
			&createdAt,
			&updatedAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("error scanning detection row: %w", err)
		}
		
		// Handle nullable fields
		if description.Valid {
			detection.Description = description.String
		}
		if queryField.Valid {
			detection.Query = queryField.String
		}
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
		
		// Parse timestamps (SQLite format)
		if parsedTime, err := time.Parse("2006-01-02 15:04:05", createdAt); err == nil {
			detection.CreatedAt = parsedTime
		} else if parsedTime, err := time.Parse(time.RFC3339, createdAt); err == nil {
			detection.CreatedAt = parsedTime
		}
		
		if parsedTime, err := time.Parse("2006-01-02 15:04:05", updatedAt); err == nil {
			detection.UpdatedAt = parsedTime
		} else if parsedTime, err := time.Parse(time.RFC3339, updatedAt); err == nil {
			detection.UpdatedAt = parsedTime
		}
		
		detections = append(detections, &detection)
	}
	
	// Load relationships for each detection
	for _, detection := range detections {
		if err := r.loadMitreTechniques(detection); err != nil {
			return nil, err
		}
		
		if err := r.loadDataSources(detection); err != nil {
			return nil, err
		}
		
	}
	
	return detections, nil
}

// ListDetectionsByStatus retrieves detections by status
func (r *Repository) ListDetectionsByStatus(status models.DetectionStatus) ([]*models.Detection, error) {
	query := `SELECT id, name, description, query, status, severity, risk_points, playbook_link, owner, risk_object, testing_description, event_count_last_30_days, false_positives_last_30_days, created_at, updated_at 
              FROM detections WHERE status = ? ORDER BY name`
	
	rows, err := r.db.Query(query, status)
	if err != nil {
		return nil, fmt.Errorf("error querying detections by status: %w", err)
	}
	defer rows.Close()
	
	var detections []*models.Detection
	
	for rows.Next() {
		var detection models.Detection
		var createdAt, updatedAt string
		var playbookLink, owner, riskObject, testingDescription, queryField sql.NullString
		var description sql.NullString
		
		err := rows.Scan(
			&detection.ID,
			&detection.Name,
			&description,
			&queryField,
			&detection.Status,
			&detection.Severity,
			&detection.RiskPoints,
			&playbookLink,
			&owner,
			&riskObject,
			&testingDescription,
			&detection.EventCountLast30Days,
			&detection.FalsePositivesLast30Days,
			&createdAt,
			&updatedAt,
		)
		
		if err != nil {
			return nil, fmt.Errorf("error scanning detection row: %w", err)
		}
		
		// Handle nullable fields
		if description.Valid {
			detection.Description = description.String
		}
		if queryField.Valid {
			detection.Query = queryField.String
		}
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
		
		// Parse timestamps (SQLite format)
		if parsedTime, err := time.Parse("2006-01-02 15:04:05", createdAt); err == nil {
			detection.CreatedAt = parsedTime
		} else if parsedTime, err := time.Parse(time.RFC3339, createdAt); err == nil {
			detection.CreatedAt = parsedTime
		}
		
		if parsedTime, err := time.Parse("2006-01-02 15:04:05", updatedAt); err == nil {
			detection.UpdatedAt = parsedTime
		} else if parsedTime, err := time.Parse(time.RFC3339, updatedAt); err == nil {
			detection.UpdatedAt = parsedTime
		}
		
		detections = append(detections, &detection)
	}
	
	// Load relationships for each detection
	for _, detection := range detections {
		if err := r.loadMitreTechniques(detection); err != nil {
			return nil, err
		}
		
		if err := r.loadDataSources(detection); err != nil {
			return nil, err
		}
		
	}
	
	return detections, nil
}

// CreateDetection creates a new detection
func (r *Repository) CreateDetection(detection *models.Detection) error {
	query := `INSERT INTO detections (name, description, query, status, severity, risk_points, playbook_link, owner, risk_object, testing_description, event_count_last_30_days, false_positives_last_30_days, created_at, updated_at) 
              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	now := time.Now()
	detection.CreatedAt = now
	detection.UpdatedAt = now
	
	result, err := r.db.Exec(
		query,
		detection.Name,
		detection.Description,
		detection.Query,
		detection.Status,
		detection.Severity,
		detection.RiskPoints,
		detection.PlaybookLink,
		detection.Owner,
		detection.RiskObject,
		detection.TestingDescription,
		detection.EventCountLast30Days,
		detection.FalsePositivesLast30Days,
		detection.CreatedAt.Format(time.RFC3339),
		detection.UpdatedAt.Format(time.RFC3339),
	)
	
	if err != nil {
		return fmt.Errorf("error creating detection: %w", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %w", err)
	}
	
	detection.ID = id
	return nil
}

// UpdateDetection updates an existing detection
func (r *Repository) UpdateDetection(detection *models.Detection) error {
	query := `UPDATE detections 
              SET name = ?, description = ?, query = ?, status = ?, severity = ?, risk_points = ?, playbook_link = ?, owner = ?, risk_object = ?, testing_description = ?, event_count_last_30_days = ?, false_positives_last_30_days = ?, updated_at = ? 
              WHERE id = ?`
	
	detection.UpdatedAt = time.Now()
	
	_, err := r.db.Exec(
		query,
		detection.Name,
		detection.Description,
		detection.Query,
		detection.Status,
		detection.Severity,
		detection.RiskPoints,
		detection.PlaybookLink,
		detection.Owner,
		detection.RiskObject,
		detection.TestingDescription,
		detection.EventCountLast30Days,
		detection.FalsePositivesLast30Days,
		detection.UpdatedAt.Format(time.RFC3339),
		detection.ID,
	)
	
	return err
}

// DeleteDetection deletes a detection by ID
func (r *Repository) DeleteDetection(id int64) error {
	query := `DELETE FROM detections WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

// AddMitreTechnique adds a MITRE technique to a detection
func (r *Repository) AddMitreTechnique(detectionID int64, mitreID string) error {
	query := `INSERT OR IGNORE INTO detection_mitre_map (detection_id, mitre_id) VALUES (?, ?)`
	_, err := r.db.Exec(query, detectionID, mitreID)
	return err
}

// RemoveMitreTechnique removes a MITRE technique from a detection
func (r *Repository) RemoveMitreTechnique(detectionID int64, mitreID string) error {
	query := `DELETE FROM detection_mitre_map WHERE detection_id = ? AND mitre_id = ?`
	_, err := r.db.Exec(query, detectionID, mitreID)
	return err
}

// AddDataSource adds a data source to a detection
func (r *Repository) AddDataSource(detectionID int64, dataSourceID int64) error {
	query := `INSERT OR IGNORE INTO detection_datasource (detection_id, datasource_id) VALUES (?, ?)`
	_, err := r.db.Exec(query, detectionID, dataSourceID)
	return err
}

// RemoveDataSource removes a data source from a detection
func (r *Repository) RemoveDataSource(detectionID int64, dataSourceID int64) error {
	query := `DELETE FROM detection_datasource WHERE detection_id = ? AND datasource_id = ?`
	_, err := r.db.Exec(query, detectionID, dataSourceID)
	return err
}

// GetDetectionCount returns the total number of detections
func (r *Repository) GetDetectionCount() (int, error) {
	query := `SELECT COUNT(*) FROM detections`
	var count int
	err := r.db.QueryRow(query).Scan(&count)
	return count, err
}

// GetDetectionCountByStatus returns the count of detections by status
func (r *Repository) GetDetectionCountByStatus() (map[models.DetectionStatus]int, error) {
	query := `SELECT status, COUNT(*) FROM detections GROUP BY status`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying detection counts by status: %w", err)
	}
	defer rows.Close()
	
	counts := make(map[models.DetectionStatus]int)
	
	for rows.Next() {
		var status models.DetectionStatus
		var count int
		
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("error scanning detection count row: %w", err)
		}
		
		counts[status] = count
	}
	
	return counts, nil
}

// GetFalsePositiveRate calculates the false positive rate for a detection
func (r *Repository) GetFalsePositiveRate(detectionID int64) (float64, error) {
	query := `SELECT 
                COUNT(*) as total_events,
                COUNT(CASE WHEN is_false_positive = 1 THEN 1 END) as false_positives
              FROM events 
              WHERE detection_id = ?`
	
	var totalEvents, falsePositives int
	err := r.db.QueryRow(query, detectionID).Scan(&totalEvents, &falsePositives)
	if err != nil {
		return 0, fmt.Errorf("error calculating false positive rate: %w", err)
	}
	
	if totalEvents == 0 {
		return 0, nil
	}
	
	return float64(falsePositives) / float64(totalEvents), nil
}

// GetEventCountLast30Days returns the count of events for a detection in the last 30 days
func (r *Repository) GetEventCountLast30Days(detectionID int64) (int, error) {
	query := `SELECT COUNT(*) FROM events 
              WHERE detection_id = ? 
              AND timestamp >= datetime('now', '-30 days')`
	
	var count int
	err := r.db.QueryRow(query, detectionID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error counting events for last 30 days: %w", err)
	}
	
	return count, nil
}

// GetFalsePositivesLast30Days returns the count of false positive events for a detection in the last 30 days
func (r *Repository) GetFalsePositivesLast30Days(detectionID int64) (int, error) {
	query := `SELECT COUNT(*) FROM events 
              WHERE detection_id = ? 
              AND is_false_positive = 1 
              AND timestamp >= datetime('now', '-30 days')`
	
	var count int
	err := r.db.QueryRow(query, detectionID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error counting false positives for last 30 days: %w", err)
	}
	
	return count, nil
}


// loadMitreTechniques loads MITRE techniques for a detection
func (r *Repository) loadMitreTechniques(detection *models.Detection) error {
	// Initialize empty slice to avoid nil
	detection.MitreTechniques = []models.MitreTechnique{}
	
	query := `SELECT mt.id, mt.name, mt.description, mt.tactic, mt.tactics, mt.domain, mt.last_modified, mt.detection, mt.platforms, mt.data_sources, mt.is_sub_technique, mt.sub_technique_of
              FROM mitre_techniques mt
              JOIN detection_mitre_map dmm ON mt.id = dmm.mitre_id
              WHERE dmm.detection_id = ?`
	
	rows, err := r.db.Query(query, detection.ID)
	if err != nil {
		return fmt.Errorf("error loading MITRE techniques: %w", err)
	}
	defer rows.Close()
	
	var techniques []models.MitreTechnique
	
	for rows.Next() {
		var technique models.MitreTechnique
		var tactics, platforms, dataSources, detectionField, lastModified, subTechniqueOf, domain sql.NullString
		
		err := rows.Scan(
			&technique.ID,
			&technique.Name,
			&technique.Description,
			&technique.Tactic,
			&tactics,
			&domain,
			&lastModified,
			&detectionField,
			&platforms,
			&dataSources,
			&technique.IsSubTechnique,
			&subTechniqueOf,
		)
		
		if err != nil {
			return fmt.Errorf("error scanning MITRE technique: %w", err)
		}
		
		// Handle nullable fields
		if domain.Valid {
			technique.Domain = domain.String
		}
		if tactics.Valid {
			json.Unmarshal([]byte(tactics.String), &technique.Tactics)
		}
		if platforms.Valid {
			json.Unmarshal([]byte(platforms.String), &technique.Platforms)
		}
		if dataSources.Valid {
			json.Unmarshal([]byte(dataSources.String), &technique.DataSources)
		}
		if detectionField.Valid {
			technique.Detection = detectionField.String
		}
		if lastModified.Valid {
			technique.LastModified = lastModified.String
		}
		if subTechniqueOf.Valid {
			technique.SubTechniqueOf = subTechniqueOf.String
		}
		
		techniques = append(techniques, technique)
	}
	
	if len(techniques) > 0 {
		detection.MitreTechniques = techniques
	}
	return nil
}

// loadDataSources loads data sources for a detection
func (r *Repository) loadDataSources(detection *models.Detection) error {
	// Initialize empty slice to avoid nil
	detection.DataSources = []models.DataSource{}
	
	query := `SELECT ds.id, ds.name, ds.description, ds.log_format
              FROM data_sources ds
              JOIN detection_datasource dd ON ds.id = dd.datasource_id
              WHERE dd.detection_id = ?`
	
	rows, err := r.db.Query(query, detection.ID)
	if err != nil {
		return fmt.Errorf("error loading data sources: %w", err)
	}
	defer rows.Close()
	
	var dataSources []models.DataSource
	
	for rows.Next() {
		var dataSource models.DataSource
		var description, logFormat sql.NullString
		
		err := rows.Scan(
			&dataSource.ID,
			&dataSource.Name,
			&description,
			&logFormat,
		)
		
		if err != nil {
			return fmt.Errorf("error scanning data source: %w", err)
		}
		
		// Handle nullable fields
		if description.Valid {
			dataSource.Description = description.String
		}
		if logFormat.Valid {
			dataSource.LogFormat = logFormat.String
		}
		
		dataSources = append(dataSources, dataSource)
	}
	
	if len(dataSources) > 0 {
		detection.DataSources = dataSources
	}
	return nil
}