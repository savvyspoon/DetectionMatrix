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
	query := `SELECT id, name, description, query, status, severity, risk_points, playbook_link, owner, risk_object, testing_description, event_count_last_30_days, false_positives_last_30_days, class_id, created_at, updated_at 
              FROM detections WHERE id = ?`

	row := r.db.QueryRow(query, id)

	var detection models.Detection
	var createdAt, updatedAt string

	var playbookLink, owner, riskObject, testingDescription, queryField sql.NullString
	var description sql.NullString
	var classID sql.NullInt64

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
		&classID,
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
	if classID.Valid {
		detection.ClassID = &classID.Int64
		// Load the class information
		class, err := r.GetDetectionClass(classID.Int64)
		if err == nil {
			detection.Class = class
		}
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
	query := `SELECT id, name, description, query, status, severity, risk_points, playbook_link, owner, risk_object, testing_description, event_count_last_30_days, false_positives_last_30_days, class_id, created_at, updated_at 
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
		var classID sql.NullInt64

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
			&classID,
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
		if classID.Valid {
			detection.ClassID = &classID.Int64
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

		// Load class if ClassID is set
		if detection.ClassID != nil {
			class, err := r.GetDetectionClass(*detection.ClassID)
			if err == nil {
				detection.Class = class
			}
		}
	}

	return detections, nil
}

// ListDetectionsByStatus retrieves detections by status
func (r *Repository) ListDetectionsByStatus(status models.DetectionStatus) ([]*models.Detection, error) {
	query := `SELECT id, name, description, query, status, severity, risk_points, playbook_link, owner, risk_object, testing_description, event_count_last_30_days, false_positives_last_30_days, class_id, created_at, updated_at 
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
		var classID sql.NullInt64

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
			&classID,
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
		if classID.Valid {
			detection.ClassID = &classID.Int64
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

		// Load class if ClassID is set
		if detection.ClassID != nil {
			class, err := r.GetDetectionClass(*detection.ClassID)
			if err == nil {
				detection.Class = class
			}
		}
	}

	return detections, nil
}

// CreateDetection creates a new detection
func (r *Repository) CreateDetection(detection *models.Detection) error {
	query := `INSERT INTO detections (name, description, query, status, severity, risk_points, playbook_link, owner, risk_object, testing_description, event_count_last_30_days, false_positives_last_30_days, class_id, created_at, updated_at) 
              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	detection.CreatedAt = now
	detection.UpdatedAt = now

	var classID sql.NullInt64
	if detection.ClassID != nil {
		classID = sql.NullInt64{Int64: *detection.ClassID, Valid: true}
	}

	// Handle nullable fields
	var riskObject sql.NullString
	if detection.RiskObject != "" {
		riskObject = sql.NullString{String: string(detection.RiskObject), Valid: true}
	}

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
		riskObject,
		detection.TestingDescription,
		detection.EventCountLast30Days,
		detection.FalsePositivesLast30Days,
		classID,
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
              SET name = ?, description = ?, query = ?, status = ?, severity = ?, risk_points = ?, playbook_link = ?, owner = ?, risk_object = ?, testing_description = ?, event_count_last_30_days = ?, false_positives_last_30_days = ?, class_id = ?, updated_at = ? 
              WHERE id = ?`

	detection.UpdatedAt = time.Now()

	var classID sql.NullInt64
	if detection.ClassID != nil {
		classID = sql.NullInt64{Int64: *detection.ClassID, Valid: true}
	}

	// Handle nullable fields
	var riskObject sql.NullString
	if detection.RiskObject != "" {
		riskObject = sql.NullString{String: string(detection.RiskObject), Valid: true}
	}

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
		riskObject,
		detection.TestingDescription,
		detection.EventCountLast30Days,
		detection.FalsePositivesLast30Days,
		classID,
		detection.UpdatedAt.Format(time.RFC3339),
		detection.ID,
	)

	return err
}

// DeleteDetection deletes a detection by ID
func (r *Repository) DeleteDetection(id int64) error {
	query := `DELETE FROM detections WHERE id = ?`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("detection with ID %d not found", id)
	}

	return nil
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

// GetDetectionClass retrieves a detection class by ID
func (r *Repository) GetDetectionClass(id int64) (*models.DetectionClass, error) {
	query := `SELECT id, name, description, color, icon, is_system, display_order, created_at, updated_at 
	          FROM detection_classes WHERE id = ?`

	row := r.db.QueryRow(query, id)

	var class models.DetectionClass
	var createdAt, updatedAt string
	var description, color, icon sql.NullString

	err := row.Scan(
		&class.ID,
		&class.Name,
		&description,
		&color,
		&icon,
		&class.IsSystem,
		&class.DisplayOrder,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("detection class not found")
		}
		return nil, err
	}

	// Handle nullable fields
	if description.Valid {
		class.Description = description.String
	}
	if color.Valid {
		class.Color = color.String
	}
	if icon.Valid {
		class.Icon = icon.String
	}

	// Parse timestamps
	class.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	class.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

	return &class, nil
}

// ListDetectionClasses retrieves all detection classes
func (r *Repository) ListDetectionClasses() ([]*models.DetectionClass, error) {
	query := `SELECT id, name, description, color, icon, is_system, display_order, created_at, updated_at 
	          FROM detection_classes ORDER BY display_order, name`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []*models.DetectionClass

	for rows.Next() {
		var class models.DetectionClass
		var createdAt, updatedAt string
		var description, color, icon sql.NullString

		err := rows.Scan(
			&class.ID,
			&class.Name,
			&description,
			&color,
			&icon,
			&class.IsSystem,
			&class.DisplayOrder,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		if description.Valid {
			class.Description = description.String
		}
		if color.Valid {
			class.Color = color.String
		}
		if icon.Valid {
			class.Icon = icon.String
		}

		// Parse timestamps
		class.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		class.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

		classes = append(classes, &class)
	}

	return classes, nil
}

// CreateDetectionClass creates a new detection class
func (r *Repository) CreateDetectionClass(class *models.DetectionClass) error {
	query := `INSERT INTO detection_classes (name, description, color, icon, is_system, display_order, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	result, err := r.db.Exec(query,
		class.Name,
		sql.NullString{String: class.Description, Valid: class.Description != ""},
		sql.NullString{String: class.Color, Valid: class.Color != ""},
		sql.NullString{String: class.Icon, Valid: class.Icon != ""},
		class.IsSystem,
		class.DisplayOrder,
	)

	if err != nil {
		return fmt.Errorf("error creating detection class: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %w", err)
	}

	class.ID = id
	class.CreatedAt = time.Now()
	class.UpdatedAt = time.Now()

	return nil
}

// UpdateDetectionClass updates an existing detection class
func (r *Repository) UpdateDetectionClass(class *models.DetectionClass) error {
	// Check if class exists and is not a system class
	existing, err := r.GetDetectionClass(class.ID)
	if err != nil {
		return err
	}

	if existing.IsSystem {
		return fmt.Errorf("cannot modify system detection class")
	}

	query := `UPDATE detection_classes 
	          SET name = ?, description = ?, color = ?, icon = ?, display_order = ?, updated_at = CURRENT_TIMESTAMP
	          WHERE id = ? AND is_system = 0`

	_, err = r.db.Exec(query,
		class.Name,
		sql.NullString{String: class.Description, Valid: class.Description != ""},
		sql.NullString{String: class.Color, Valid: class.Color != ""},
		sql.NullString{String: class.Icon, Valid: class.Icon != ""},
		class.DisplayOrder,
		class.ID,
	)

	if err != nil {
		return fmt.Errorf("error updating detection class: %w", err)
	}

	class.UpdatedAt = time.Now()
	return nil
}

// DeleteDetectionClass deletes a detection class
func (r *Repository) DeleteDetectionClass(id int64) error {
	// Check if class exists and is not a system class
	existing, err := r.GetDetectionClass(id)
	if err != nil {
		return err
	}

	if existing.IsSystem {
		return fmt.Errorf("cannot delete system detection class")
	}

	query := `DELETE FROM detection_classes WHERE id = ? AND is_system = 0`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting detection class: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("detection class not found or is a system class")
	}

	return nil
}

// ListDetectionsByClass retrieves all detections for a specific class
func (r *Repository) ListDetectionsByClass(classID int64) ([]*models.Detection, error) {
	query := `SELECT d.id, d.name, d.description, d.query, d.status, d.severity, d.risk_points, 
	                 d.playbook_link, d.owner, d.risk_object, d.testing_description, 
	                 d.event_count_last_30_days, d.false_positives_last_30_days, 
	                 d.class_id, d.created_at, d.updated_at
	          FROM detections d
	          WHERE d.class_id = ?
	          ORDER BY d.name`

	rows, err := r.db.Query(query, classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var detections []*models.Detection

	for rows.Next() {
		var detection models.Detection
		var createdAt, updatedAt string
		var playbookLink, owner, riskObject, testingDescription, queryField sql.NullString
		var description sql.NullString
		var classID sql.NullInt64

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
			&classID,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			return nil, err
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
		if classID.Valid {
			detection.ClassID = &classID.Int64
		}

		// Parse timestamps
		detection.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		detection.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)

		// Load the class information
		if detection.ClassID != nil {
			class, err := r.GetDetectionClass(*detection.ClassID)
			if err == nil {
				detection.Class = class
			}
		}

		detections = append(detections, &detection)
	}

	return detections, nil
}
