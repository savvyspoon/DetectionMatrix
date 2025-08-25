package risk

import (
	"database/sql"
	"fmt"
	"time"

	"riskmatrix/pkg/database"
	"riskmatrix/pkg/models"
)

// Repository implements the risk-related data access
type Repository struct {
	db *database.DB
}

// NewRepository creates a new risk repository
func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

// Transaction methods

// GetRiskObjectByEntityTx gets a risk object by entity type and value within a transaction
func (r *Repository) GetRiskObjectByEntityTx(tx *sql.Tx, entityType models.EntityType, entityValue string) (*models.RiskObject, error) {
	query := `SELECT id, entity_type, entity_value, current_score, last_seen 
              FROM risk_objects 
              WHERE entity_type = ? AND entity_value = ?`

	row := tx.QueryRow(query, entityType, entityValue)

	var obj models.RiskObject
	var lastSeen string

	err := row.Scan(
		&obj.ID,
		&obj.EntityType,
		&obj.EntityValue,
		&obj.CurrentScore,
		&lastSeen,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("risk object not found for %s '%s'", entityType, entityValue)
		}
		return nil, fmt.Errorf("error scanning risk object: %w", err)
	}

	// Parse timestamp
	obj.LastSeen, _ = time.Parse(time.RFC3339, lastSeen)

	return &obj, nil
}

// CreateRiskObjectTx creates a risk object within a transaction
func (r *Repository) CreateRiskObjectTx(tx *sql.Tx, obj *models.RiskObject) error {
	query := `INSERT INTO risk_objects (entity_type, entity_value, current_score, last_seen) 
              VALUES (?, ?, ?, ?)`

	result, err := tx.Exec(
		query,
		obj.EntityType,
		obj.EntityValue,
		obj.CurrentScore,
		obj.LastSeen.Format(time.RFC3339),
	)

	if err != nil {
		return fmt.Errorf("error creating risk object: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %w", err)
	}

	obj.ID = id
	return nil
}

// UpdateRiskObjectTx updates a risk object within a transaction
func (r *Repository) UpdateRiskObjectTx(tx *sql.Tx, obj *models.RiskObject) error {
	query := `UPDATE risk_objects 
              SET current_score = ?, last_seen = ? 
              WHERE id = ?`

	_, err := tx.Exec(
		query,
		obj.CurrentScore,
		obj.LastSeen.Format(time.RFC3339),
		obj.ID,
	)

	if err != nil {
		return fmt.Errorf("error updating risk object: %w", err)
	}

	return nil
}

// GetRiskObjectTx gets a risk object by ID within a transaction
func (r *Repository) GetRiskObjectTx(tx *sql.Tx, id int64) (*models.RiskObject, error) {
	query := `SELECT id, entity_type, entity_value, current_score, last_seen 
              FROM risk_objects 
              WHERE id = ?`

	row := tx.QueryRow(query, id)

	var obj models.RiskObject
	var lastSeen string

	err := row.Scan(
		&obj.ID,
		&obj.EntityType,
		&obj.EntityValue,
		&obj.CurrentScore,
		&lastSeen,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("risk object not found: %d", id)
		}
		return nil, fmt.Errorf("error scanning risk object: %w", err)
	}

	// Parse timestamp
	obj.LastSeen, _ = time.Parse(time.RFC3339, lastSeen)

	return &obj, nil
}

// CreateEventTx creates an event within a transaction
func (r *Repository) CreateEventTx(tx *sql.Tx, event *models.Event) error {
	query := `INSERT INTO events (detection_id, entity_id, timestamp, raw_data, context, risk_points, is_false_positive) 
              VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := tx.Exec(
		query,
		event.DetectionID,
		event.EntityID,
		event.Timestamp.Format(time.RFC3339),
		event.RawData,
		event.Context,
		event.RiskPoints,
		event.IsFalsePositive,
	)

	if err != nil {
		return fmt.Errorf("error creating event: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %w", err)
	}

	event.ID = id
	return nil
}

// GetEventTx gets an event by ID within a transaction
func (r *Repository) GetEventTx(tx *sql.Tx, id int64) (*models.Event, error) {
	query := `SELECT id, detection_id, entity_id, timestamp, raw_data, context, risk_points, is_false_positive 
              FROM events 
              WHERE id = ?`

	row := tx.QueryRow(query, id)

	var event models.Event
	var timestamp string
	var context sql.NullString

	err := row.Scan(
		&event.ID,
		&event.DetectionID,
		&event.EntityID,
		&timestamp,
		&event.RawData,
		&context,
		&event.RiskPoints,
		&event.IsFalsePositive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("event not found: %d", id)
		}
		return nil, fmt.Errorf("error scanning event: %w", err)
	}

	// Parse timestamp
	event.Timestamp, _ = time.Parse(time.RFC3339, timestamp)

	// Handle nullable context field
	if context.Valid {
		event.Context = context.String
	}

	return &event, nil
}

// UpdateEventTx updates an event within a transaction
func (r *Repository) UpdateEventTx(tx *sql.Tx, event *models.Event) error {
	query := `UPDATE events 
              SET is_false_positive = ? 
              WHERE id = ?`

	_, err := tx.Exec(
		query,
		event.IsFalsePositive,
		event.ID,
	)

	if err != nil {
		return fmt.Errorf("error updating event: %w", err)
	}

	return nil
}

// CreateRiskAlertTx creates a risk alert within a transaction
func (r *Repository) CreateRiskAlertTx(tx *sql.Tx, alert *models.RiskAlert) error {
	query := `INSERT INTO risk_alerts (entity_id, triggered_at, total_score, status, notes, owner) 
              VALUES (?, ?, ?, ?, ?, ?)`

	result, err := tx.Exec(
		query,
		alert.EntityID,
		alert.TriggeredAt.Format(time.RFC3339),
		alert.TotalScore,
		alert.Status,
		alert.Notes,
		alert.Owner,
	)

	if err != nil {
		return fmt.Errorf("error creating risk alert: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %w", err)
	}

	alert.ID = id
	return nil
}

// CreateFalsePositiveTx creates a false positive record within a transaction
func (r *Repository) CreateFalsePositiveTx(tx *sql.Tx, fp *models.FalsePositive) error {
	query := `INSERT INTO false_positives (event_id, reason, analyst_name, timestamp) 
              VALUES (?, ?, ?, ?)`

	result, err := tx.Exec(
		query,
		fp.EventID,
		fp.Reason,
		fp.AnalystName,
		fp.Timestamp.Format(time.RFC3339),
	)

	if err != nil {
		return fmt.Errorf("error creating false positive record: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %w", err)
	}

	fp.ID = id
	return nil
}

// DeleteFalsePositiveByEventTx deletes a false positive record by event ID within a transaction
func (r *Repository) DeleteFalsePositiveByEventTx(tx *sql.Tx, eventID int64) error {
	query := `DELETE FROM false_positives WHERE event_id = ?`

	_, err := tx.Exec(query, eventID)
	if err != nil {
		return fmt.Errorf("error deleting false positive record: %w", err)
	}

	return nil
}

// Non-transaction methods

// GetRiskObject gets a risk object by ID
func (r *Repository) GetRiskObject(id int64) (*models.RiskObject, error) {
	query := `SELECT id, entity_type, entity_value, current_score, last_seen 
              FROM risk_objects 
              WHERE id = ?`

	row := r.db.QueryRow(query, id)

	var obj models.RiskObject
	var lastSeen string

	err := row.Scan(
		&obj.ID,
		&obj.EntityType,
		&obj.EntityValue,
		&obj.CurrentScore,
		&lastSeen,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("risk object not found: %d", id)
		}
		return nil, fmt.Errorf("error scanning risk object: %w", err)
	}

	// Parse timestamp
	obj.LastSeen, _ = time.Parse(time.RFC3339, lastSeen)

	return &obj, nil
}

// GetRiskObjectByEntity gets a risk object by entity type and value
func (r *Repository) GetRiskObjectByEntity(entityType models.EntityType, entityValue string) (*models.RiskObject, error) {
	query := `SELECT id, entity_type, entity_value, current_score, last_seen 
              FROM risk_objects 
              WHERE entity_type = ? AND entity_value = ?`

	row := r.db.QueryRow(query, entityType, entityValue)

	var obj models.RiskObject
	var lastSeen string

	err := row.Scan(
		&obj.ID,
		&obj.EntityType,
		&obj.EntityValue,
		&obj.CurrentScore,
		&lastSeen,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("risk object not found for %s '%s'", entityType, entityValue)
		}
		return nil, fmt.Errorf("error scanning risk object: %w", err)
	}

	// Parse timestamp
	obj.LastSeen, _ = time.Parse(time.RFC3339, lastSeen)

	return &obj, nil
}

// ListRiskObjects lists all risk objects
func (r *Repository) ListRiskObjects() ([]*models.RiskObject, error) {
	query := `SELECT id, entity_type, entity_value, current_score, last_seen 
              FROM risk_objects 
              ORDER BY current_score DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying risk objects: %w", err)
	}
	defer rows.Close()

	// Initialize empty slice to avoid nil
	objects := make([]*models.RiskObject, 0)

	for rows.Next() {
		var obj models.RiskObject
		var lastSeen string

		err := rows.Scan(
			&obj.ID,
			&obj.EntityType,
			&obj.EntityValue,
			&obj.CurrentScore,
			&lastSeen,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning risk object row: %w", err)
		}

		// Parse timestamp
		obj.LastSeen, _ = time.Parse(time.RFC3339, lastSeen)

		objects = append(objects, &obj)
	}

	return objects, nil
}

// ListHighRiskObjects lists risk objects with scores above the threshold
func (r *Repository) ListHighRiskObjects(threshold int) ([]*models.RiskObject, error) {
	query := `SELECT id, entity_type, entity_value, current_score, last_seen 
              FROM risk_objects 
              WHERE current_score >= ? 
              ORDER BY current_score DESC`

	rows, err := r.db.Query(query, threshold)
	if err != nil {
		return nil, fmt.Errorf("error querying high risk objects: %w", err)
	}
	defer rows.Close()

	// Initialize empty slice to avoid nil
	objects := make([]*models.RiskObject, 0)

	for rows.Next() {
		var obj models.RiskObject
		var lastSeen string

		err := rows.Scan(
			&obj.ID,
			&obj.EntityType,
			&obj.EntityValue,
			&obj.CurrentScore,
			&lastSeen,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning risk object row: %w", err)
		}

		// Parse timestamp
		obj.LastSeen, _ = time.Parse(time.RFC3339, lastSeen)

		objects = append(objects, &obj)
	}

	return objects, nil
}

// GetEvent gets an event by ID
func (r *Repository) GetEvent(id int64) (*models.Event, error) {
	query := `SELECT id, detection_id, entity_id, timestamp, raw_data, context, risk_points, is_false_positive 
              FROM events 
              WHERE id = ?`

	row := r.db.QueryRow(query, id)

	var event models.Event
	var timestamp string
	var context sql.NullString

	err := row.Scan(
		&event.ID,
		&event.DetectionID,
		&event.EntityID,
		&timestamp,
		&event.RawData,
		&context,
		&event.RiskPoints,
		&event.IsFalsePositive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("event not found: %d", id)
		}
		return nil, fmt.Errorf("error scanning event: %w", err)
	}

	// Parse timestamp
	event.Timestamp, _ = time.Parse(time.RFC3339, timestamp)

	// Handle nullable context field
	if context.Valid {
		event.Context = context.String
	}

	return &event, nil
}

// ListEvents lists all events
func (r *Repository) ListEvents() ([]*models.Event, error) {
	query := `SELECT id, detection_id, entity_id, timestamp, raw_data, context, risk_points, is_false_positive 
              FROM events 
              ORDER BY timestamp DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying events: %w", err)
	}
	defer rows.Close()

	events := make([]*models.Event, 0)

	for rows.Next() {
		var event models.Event
		var timestamp string
		var context sql.NullString

		err := rows.Scan(
			&event.ID,
			&event.DetectionID,
			&event.EntityID,
			&timestamp,
			&event.RawData,
			&context,
			&event.RiskPoints,
			&event.IsFalsePositive,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning event row: %w", err)
		}

		// Handle nullable context field
		if context.Valid {
			event.Context = context.String
		}

		// Parse timestamp
		event.Timestamp, _ = time.Parse(time.RFC3339, timestamp)

		events = append(events, &event)
	}

	return events, nil
}

// ListEventsPaginated lists events with pagination support
func (r *Repository) ListEventsPaginated(limit, offset int) ([]*models.Event, int, error) {
	// Get total count first
	countQuery := `SELECT COUNT(*) FROM events`
	var totalCount int
	err := r.db.QueryRow(countQuery).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting events: %w", err)
	}

	// Get paginated events
	query := `SELECT id, detection_id, entity_id, timestamp, raw_data, context, risk_points, is_false_positive 
              FROM events 
              ORDER BY timestamp DESC
              LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying events: %w", err)
	}
	defer rows.Close()

	events := make([]*models.Event, 0)

	for rows.Next() {
		var event models.Event
		var timestamp string
		var context sql.NullString

		err := rows.Scan(
			&event.ID,
			&event.DetectionID,
			&event.EntityID,
			&timestamp,
			&event.RawData,
			&context,
			&event.RiskPoints,
			&event.IsFalsePositive,
		)

		if err != nil {
			return nil, 0, fmt.Errorf("error scanning event row: %w", err)
		}

		// Handle nullable context field
		if context.Valid {
			event.Context = context.String
		}

		// Parse timestamp
		event.Timestamp, _ = time.Parse(time.RFC3339, timestamp)

		events = append(events, &event)
	}

	return events, totalCount, nil
}

// ListEventsByEntity lists events for an entity
func (r *Repository) ListEventsByEntity(entityID int64) ([]*models.Event, error) {
	query := `SELECT id, detection_id, entity_id, timestamp, raw_data, context, risk_points, is_false_positive 
              FROM events 
              WHERE entity_id = ? 
              ORDER BY timestamp DESC`

	rows, err := r.db.Query(query, entityID)
	if err != nil {
		return nil, fmt.Errorf("error querying events by entity: %w", err)
	}
	defer rows.Close()

	events := make([]*models.Event, 0)

	for rows.Next() {
		var event models.Event
		var timestamp string
		var context sql.NullString

		err := rows.Scan(
			&event.ID,
			&event.DetectionID,
			&event.EntityID,
			&timestamp,
			&event.RawData,
			&context,
			&event.RiskPoints,
			&event.IsFalsePositive,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning event row: %w", err)
		}

		// Handle nullable context field
		if context.Valid {
			event.Context = context.String
		}

		// Parse timestamp
		event.Timestamp, _ = time.Parse(time.RFC3339, timestamp)

		events = append(events, &event)
	}

	return events, nil
}

// ListRiskAlerts lists all risk alerts
func (r *Repository) ListRiskAlerts() ([]*models.RiskAlert, error) {
	return r.ListRiskAlertsByStatus("")
}

// ListRiskAlertsByStatus lists risk alerts filtered by status
func (r *Repository) ListRiskAlertsByStatus(status models.AlertStatus) ([]*models.RiskAlert, error) {
	var query string
	var args []interface{}

	if status != "" {
		query = `SELECT id, entity_id, triggered_at, total_score, status, notes, owner 
                 FROM risk_alerts 
                 WHERE status = ?
                 ORDER BY triggered_at DESC`
		args = append(args, status)
	} else {
		query = `SELECT id, entity_id, triggered_at, total_score, status, notes, owner 
                 FROM risk_alerts 
                 ORDER BY triggered_at DESC`
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying risk alerts: %w", err)
	}
	defer rows.Close()

	// Initialize empty slice to avoid nil
	alerts := make([]*models.RiskAlert, 0)

	for rows.Next() {
		var alert models.RiskAlert
		var triggeredAt string
		var notes, owner sql.NullString

		err := rows.Scan(
			&alert.ID,
			&alert.EntityID,
			&triggeredAt,
			&alert.TotalScore,
			&alert.Status,
			&notes,
			&owner,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning risk alert row: %w", err)
		}

		// Handle nullable fields
		if notes.Valid {
			alert.Notes = notes.String
		}
		if owner.Valid {
			alert.Owner = owner.String
		}

		// Parse timestamp
		alert.TriggeredAt, _ = time.Parse(time.RFC3339, triggeredAt)

		alerts = append(alerts, &alert)
	}

	return alerts, nil
}

// ListRiskAlertsPaginated lists risk alerts with pagination support
func (r *Repository) ListRiskAlertsPaginated(limit, offset int, status models.AlertStatus) ([]*models.RiskAlert, int, error) {
	// Get total count first
	var countQuery string
	var countArgs []interface{}

	if status != "" {
		countQuery = `SELECT COUNT(*) FROM risk_alerts WHERE status = ?`
		countArgs = append(countArgs, status)
	} else {
		countQuery = `SELECT COUNT(*) FROM risk_alerts`
	}

	var totalCount int
	err := r.db.QueryRow(countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting risk alerts: %w", err)
	}

	// Get paginated alerts
	var query string
	var args []interface{}

	if status != "" {
		query = `SELECT id, entity_id, triggered_at, total_score, status, notes, owner 
                 FROM risk_alerts 
                 WHERE status = ?
                 ORDER BY triggered_at DESC
                 LIMIT ? OFFSET ?`
		args = append(args, status, limit, offset)
	} else {
		query = `SELECT id, entity_id, triggered_at, total_score, status, notes, owner 
                 FROM risk_alerts 
                 ORDER BY triggered_at DESC
                 LIMIT ? OFFSET ?`
		args = append(args, limit, offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying risk alerts: %w", err)
	}
	defer rows.Close()

	// Initialize empty slice to avoid nil
	alerts := make([]*models.RiskAlert, 0)

	for rows.Next() {
		var alert models.RiskAlert
		var triggeredAt string
		var notes, owner sql.NullString

		err := rows.Scan(
			&alert.ID,
			&alert.EntityID,
			&triggeredAt,
			&alert.TotalScore,
			&alert.Status,
			&notes,
			&owner,
		)

		if err != nil {
			return nil, 0, fmt.Errorf("error scanning risk alert row: %w", err)
		}

		// Handle nullable fields
		if notes.Valid {
			alert.Notes = notes.String
		}
		if owner.Valid {
			alert.Owner = owner.String
		}

		// Parse timestamp
		alert.TriggeredAt, _ = time.Parse(time.RFC3339, triggeredAt)

		alerts = append(alerts, &alert)
	}

	return alerts, totalCount, nil
}

// GetRiskAlert retrieves a risk alert by ID
func (r *Repository) GetRiskAlert(id int64) (*models.RiskAlert, error) {
	query := `SELECT id, entity_id, triggered_at, total_score, status, notes, owner 
              FROM risk_alerts 
              WHERE id = ?`

	row := r.db.QueryRow(query, id)

	var alert models.RiskAlert
	var triggeredAt string
	var notes, owner sql.NullString

	err := row.Scan(
		&alert.ID,
		&alert.EntityID,
		&triggeredAt,
		&alert.TotalScore,
		&alert.Status,
		&notes,
		&owner,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("risk alert not found: %d", id)
		}
		return nil, fmt.Errorf("error scanning risk alert: %w", err)
	}

	// Handle nullable fields
	if notes.Valid {
		alert.Notes = notes.String
	}
	if owner.Valid {
		alert.Owner = owner.String
	}

	// Parse timestamp
	alert.TriggeredAt, _ = time.Parse(time.RFC3339, triggeredAt)

	return &alert, nil
}

// UpdateRiskAlert updates a risk alert
func (r *Repository) UpdateRiskAlert(alert *models.RiskAlert) error {
	query := `UPDATE risk_alerts 
              SET status = ?, notes = ?, owner = ? 
              WHERE id = ?`

	result, err := r.db.Exec(
		query,
		alert.Status,
		alert.Notes,
		alert.Owner,
		alert.ID,
	)

	if err != nil {
		return fmt.Errorf("error updating risk alert: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("risk alert not found: %d", alert.ID)
	}

	return nil
}

// GetEventsForAlert gets events that contributed to a risk alert
func (r *Repository) GetEventsForAlert(alertID int64) ([]*models.Event, error) {
	// First get the alert to find the entity and timestamp
	var alert models.RiskAlert
	var triggeredAt string
	var notes, owner sql.NullString

	err := r.db.QueryRow(
		`SELECT id, entity_id, triggered_at, total_score, status, notes, owner FROM risk_alerts WHERE id = ?`,
		alertID,
	).Scan(&alert.ID, &alert.EntityID, &triggeredAt, &alert.TotalScore, &alert.Status, &notes, &owner)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("risk alert not found: %d", alertID)
		}
		return nil, fmt.Errorf("error scanning risk alert: %w", err)
	}

	// Handle nullable fields
	if notes.Valid {
		alert.Notes = notes.String
	}
	if owner.Valid {
		alert.Owner = owner.String
	}

	alert.TriggeredAt, _ = time.Parse(time.RFC3339, triggeredAt)

	// Get events for this entity before the alert was triggered
	// Convert alert timestamp to UTC for proper comparison with event timestamps
	alertTimeUTC := alert.TriggeredAt.UTC()

	query := `SELECT id, detection_id, entity_id, timestamp, raw_data, context, risk_points, is_false_positive 
              FROM events 
              WHERE entity_id = ? AND datetime(timestamp) <= datetime(?) AND is_false_positive = 0
              ORDER BY timestamp DESC`

	rows, err := r.db.Query(query, alert.EntityID, alertTimeUTC.Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("error querying events for alert: %w", err)
	}
	defer rows.Close()

	events := make([]*models.Event, 0)

	for rows.Next() {
		var event models.Event
		var timestamp string
		var context sql.NullString

		err := rows.Scan(
			&event.ID,
			&event.DetectionID,
			&event.EntityID,
			&timestamp,
			&event.RawData,
			&context,
			&event.RiskPoints,
			&event.IsFalsePositive,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning event row: %w", err)
		}

		// Handle nullable context field
		if context.Valid {
			event.Context = context.String
		}

		// Parse timestamp
		event.Timestamp, _ = time.Parse(time.RFC3339, timestamp)

		events = append(events, &event)
	}

	return events, nil
}

// DecayRiskScores reduces all risk scores by the decay factor
func (r *Repository) DecayRiskScores(decayFactor float64) error {
	query := `UPDATE risk_objects 
              SET current_score = CASE 
                WHEN current_score * (1.0 - ?) < 1.0 AND current_score > 0 THEN 0 
                ELSE CAST(current_score * (1.0 - ?) AS INTEGER) 
              END`

	_, err := r.db.Exec(query, decayFactor, decayFactor)
	if err != nil {
		return fmt.Errorf("error decaying risk scores: %w", err)
	}

	return nil
}
