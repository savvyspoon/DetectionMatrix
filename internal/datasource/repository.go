package datasource

import (
	"database/sql"
	"fmt"

	"riskmatrix/pkg/database"
	"riskmatrix/pkg/models"
)

// Repository implements the models.DataSourceRepository interface
type Repository struct {
	db *database.DB
}

// NewRepository creates a new data source repository
func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

// GetDataSource retrieves a data source by ID
func (r *Repository) GetDataSource(id int64) (*models.DataSource, error) {
	query := `SELECT id, name, description, log_format FROM data_sources WHERE id = ?`
	
	row := r.db.QueryRow(query, id)
	
	var dataSource models.DataSource
	
	err := row.Scan(
		&dataSource.ID,
		&dataSource.Name,
		&dataSource.Description,
		&dataSource.LogFormat,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("data source not found: %d", id)
		}
		return nil, fmt.Errorf("error scanning data source: %w", err)
	}
	
	return &dataSource, nil
}

// GetDataSourceByName retrieves a data source by name
func (r *Repository) GetDataSourceByName(name string) (*models.DataSource, error) {
	query := `SELECT id, name, description, log_format FROM data_sources WHERE name = ?`
	
	row := r.db.QueryRow(query, name)
	
	var dataSource models.DataSource
	
	err := row.Scan(
		&dataSource.ID,
		&dataSource.Name,
		&dataSource.Description,
		&dataSource.LogFormat,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("data source not found: %s", name)
		}
		return nil, fmt.Errorf("error scanning data source: %w", err)
	}
	
	return &dataSource, nil
}

// ListDataSources retrieves all data sources
func (r *Repository) ListDataSources() ([]*models.DataSource, error) {
	query := `SELECT id, name, description, log_format FROM data_sources ORDER BY name`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying data sources: %w", err)
	}
	defer rows.Close()
	
	var dataSources []*models.DataSource
	
	for rows.Next() {
		var dataSource models.DataSource
		
		err := rows.Scan(
			&dataSource.ID,
			&dataSource.Name,
			&dataSource.Description,
			&dataSource.LogFormat,
		)
		
		if err != nil {
			return nil, fmt.Errorf("error scanning data source row: %w", err)
		}
		
		dataSources = append(dataSources, &dataSource)
	}
	
	return dataSources, nil
}

// CreateDataSource creates a new data source
func (r *Repository) CreateDataSource(dataSource *models.DataSource) error {
	// Validate required fields
	if dataSource.Name == "" {
		return fmt.Errorf("data source name cannot be empty")
	}
	
	query := `INSERT INTO data_sources (name, description, log_format) VALUES (?, ?, ?)`
	
	result, err := r.db.Exec(
		query,
		dataSource.Name,
		dataSource.Description,
		dataSource.LogFormat,
	)
	
	if err != nil {
		return fmt.Errorf("error creating data source: %w", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %w", err)
	}
	
	dataSource.ID = id
	return nil
}

// UpdateDataSource updates an existing data source
func (r *Repository) UpdateDataSource(dataSource *models.DataSource) error {
	query := `UPDATE data_sources SET name = ?, description = ?, log_format = ? WHERE id = ?`
	
	result, err := r.db.Exec(
		query,
		dataSource.Name,
		dataSource.Description,
		dataSource.LogFormat,
		dataSource.ID,
	)
	
	if err != nil {
		return fmt.Errorf("error updating data source: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("data source not found: %d", dataSource.ID)
	}
	
	return nil
}

// DeleteDataSource deletes a data source
func (r *Repository) DeleteDataSource(id int64) error {
	query := `DELETE FROM data_sources WHERE id = ?`
	
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting data source: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("data source not found: %d", id)
	}
	
	return nil
}

// GetDetectionsByDataSource returns detections that use a specific data source
func (r *Repository) GetDetectionsByDataSource(dataSourceID int64) ([]*models.Detection, error) {
	query := `
		SELECT 
			d.id, d.name, d.description, d.status, d.severity, d.risk_points, d.playbook_link, d.owner, d.risk_object, d.testing_description, d.created_at, d.updated_at
		FROM 
			detections d
		JOIN 
			detection_datasource dd ON d.id = dd.detection_id
		WHERE 
			dd.datasource_id = ?
		ORDER BY 
			d.name
	`
	
	rows, err := r.db.Query(query, dataSourceID)
	if err != nil {
		return nil, fmt.Errorf("error querying detections by data source: %w", err)
	}
	defer rows.Close()
	
	// Initialize empty slice to avoid nil
	detections := make([]*models.Detection, 0)
	
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

// GetMitreTechniquesByDataSource returns MITRE techniques associated with a specific data source
func (r *Repository) GetMitreTechniquesByDataSource(dataSourceID int64) ([]*models.MitreTechnique, error) {
	query := `
		SELECT DISTINCT
			mt.id, mt.tactic, mt.name, mt.description
		FROM 
			mitre_techniques mt
		JOIN 
			detection_mitre_map dmm ON mt.id = dmm.mitre_id
		JOIN 
			detections d ON dmm.detection_id = d.id
		JOIN 
			detection_datasource dd ON d.id = dd.detection_id
		WHERE 
			dd.datasource_id = ?
		ORDER BY 
			mt.tactic, mt.name
	`
	
	rows, err := r.db.Query(query, dataSourceID)
	if err != nil {
		return nil, fmt.Errorf("error querying MITRE techniques by data source: %w", err)
	}
	defer rows.Close()
	
	// Initialize empty slice to avoid nil
	techniques := make([]*models.MitreTechnique, 0)
	
	for rows.Next() {
		var technique models.MitreTechnique
		
		err := rows.Scan(
			&technique.ID,
			&technique.Tactic,
			&technique.Name,
			&technique.Description,
		)
		
		if err != nil {
			return nil, fmt.Errorf("error scanning MITRE technique row: %w", err)
		}
		
		techniques = append(techniques, &technique)
	}
	
	return techniques, nil
}

// GetDataSourceUtilization returns the count of detections per data source
func (r *Repository) GetDataSourceUtilization() (map[string]int, error) {
	query := `
		SELECT 
			ds.name,
			COUNT(DISTINCT dd.detection_id) as detection_count
		FROM 
			data_sources ds
		LEFT JOIN 
			detection_datasource dd ON ds.id = dd.datasource_id
		GROUP BY 
			ds.id
		ORDER BY 
			detection_count DESC
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying data source utilization: %w", err)
	}
	defer rows.Close()
	
	utilization := make(map[string]int)
	
	for rows.Next() {
		var name string
		var count int
		
		if err := rows.Scan(&name, &count); err != nil {
			return nil, fmt.Errorf("error scanning utilization row: %w", err)
		}
		
		utilization[name] = count
	}
	
	return utilization, nil
}