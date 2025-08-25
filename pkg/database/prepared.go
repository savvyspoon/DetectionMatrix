package database

import (
	"database/sql"
	"sync"
)

// PreparedStatements holds commonly used prepared statements
type PreparedStatements struct {
	mu    sync.RWMutex
	stmts map[string]*sql.Stmt
	db    *sql.DB
}

// NewPreparedStatements creates a new prepared statements manager
func NewPreparedStatements(db *sql.DB) *PreparedStatements {
	return &PreparedStatements{
		stmts: make(map[string]*sql.Stmt),
		db:    db,
	}
}

// Get retrieves or creates a prepared statement
func (ps *PreparedStatements) Get(key, query string) (*sql.Stmt, error) {
	// Try to get existing statement
	ps.mu.RLock()
	stmt, exists := ps.stmts[key]
	ps.mu.RUnlock()

	if exists {
		return stmt, nil
	}

	// Create new prepared statement
	ps.mu.Lock()
	defer ps.mu.Unlock()

	// Double-check after acquiring write lock
	if stmt, exists := ps.stmts[key]; exists {
		return stmt, nil
	}

	// Prepare the statement
	stmt, err := ps.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	ps.stmts[key] = stmt
	return stmt, nil
}

// Close closes all prepared statements
func (ps *PreparedStatements) Close() error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	var lastErr error
	for _, stmt := range ps.stmts {
		if err := stmt.Close(); err != nil {
			lastErr = err
		}
	}

	ps.stmts = make(map[string]*sql.Stmt)
	return lastErr
}

// CommonQueries defines commonly used queries
var CommonQueries = map[string]string{
	"getDetection": `
		SELECT id, name, description, status, severity, confidence, 
		       created_at, updated_at, mitre_tactic, mitre_technique,
		       data_source, query_logic
		FROM detections
		WHERE id = ?
	`,
	"listDetections": `
		SELECT id, name, description, status, severity, confidence,
		       created_at, updated_at, mitre_tactic, mitre_technique,
		       data_source, query_logic
		FROM detections
		ORDER BY created_at DESC
	`,
	"listDetectionsByStatus": `
		SELECT id, name, description, status, severity, confidence,
		       created_at, updated_at, mitre_tactic, mitre_technique,
		       data_source, query_logic
		FROM detections
		WHERE status = ?
		ORDER BY created_at DESC
	`,
	"getRiskObject": `
		SELECT id, entity_type, entity_value, current_score, last_seen
		FROM risk_objects
		WHERE id = ?
	`,
	"getRiskObjectByEntity": `
		SELECT id, entity_type, entity_value, current_score, last_seen
		FROM risk_objects
		WHERE entity_type = ? AND entity_value = ?
	`,
	"listHighRiskObjects": `
		SELECT id, entity_type, entity_value, current_score, last_seen
		FROM risk_objects
		WHERE current_score >= ?
		ORDER BY current_score DESC
		LIMIT ?
	`,
	"getRecentEvents": `
		SELECT e.id, e.detection_id, e.entity_id, e.timestamp,
		       e.risk_points, e.raw_data, e.is_false_positive,
		       r.entity_type, r.entity_value
		FROM events e
		JOIN risk_objects r ON e.entity_id = r.id
		WHERE e.timestamp > datetime('now', '-30 days')
		ORDER BY e.timestamp DESC
		LIMIT ?
	`,
	"getMitreCoverage": `
		SELECT m.tactic, COUNT(DISTINCT dm.detection_id) as coverage_count
		FROM mitre_techniques m
		LEFT JOIN detection_mitre dm ON m.id = dm.technique_id
		GROUP BY m.tactic
		ORDER BY m.tactic
	`,
	"updateRiskScore": `
		UPDATE risk_objects
		SET current_score = ?, last_seen = CURRENT_TIMESTAMP
		WHERE id = ?
	`,
	"decayRiskScores": `
		UPDATE risk_objects
		SET current_score = CAST(current_score * ? AS INTEGER)
		WHERE current_score > 0
	`,
}

// DB extension with prepared statements
func (db *DB) PreparedStatements() *PreparedStatements {
	return NewPreparedStatements(db.DB)
}
