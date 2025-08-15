// +build test

package database

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"sync"
	"time"
)

// MockDB is a mock database for testing
type MockDB struct {
	*sql.DB
	data   map[string][]map[string]interface{}
	mu     sync.RWMutex
	closed bool
}

// NewMockDB creates a new mock database
func NewMockDB() (*MockDB, error) {
	// Register mock driver
	sql.Register("mock", &mockDriver{})
	
	db, err := sql.Open("mock", "mock://test")
	if err != nil {
		return nil, err
	}
	
	return &MockDB{
		DB:   db,
		data: make(map[string][]map[string]interface{}),
	}, nil
}

// SetMockData sets mock data for a table
func (m *MockDB) SetMockData(table string, data []map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[table] = data
}

// GetMockData gets mock data for a table
func (m *MockDB) GetMockData(table string) []map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.data[table]
}

// Close closes the mock database
func (m *MockDB) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return m.DB.Close()
}

// mockDriver implements the database/sql/driver interfaces
type mockDriver struct{}

func (d *mockDriver) Open(name string) (driver.Conn, error) {
	return &mockConn{}, nil
}

type mockConn struct {
	closed bool
	mu     sync.Mutex
}

func (c *mockConn) Prepare(query string) (driver.Stmt, error) {
	return &mockStmt{query: query}, nil
}

func (c *mockConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	return nil
}

func (c *mockConn) Begin() (driver.Tx, error) {
	return &mockTx{}, nil
}

type mockStmt struct {
	query string
}

func (s *mockStmt) Close() error {
	return nil
}

func (s *mockStmt) NumInput() int {
	return -1
}

func (s *mockStmt) Exec(args []driver.Value) (driver.Result, error) {
	return &mockResult{lastInsertId: 1, rowsAffected: 1}, nil
}

func (s *mockStmt) Query(args []driver.Value) (driver.Rows, error) {
	return &mockRows{}, nil
}

type mockTx struct{}

func (t *mockTx) Commit() error {
	return nil
}

func (t *mockTx) Rollback() error {
	return nil
}

type mockResult struct {
	lastInsertId int64
	rowsAffected int64
}

func (r *mockResult) LastInsertId() (int64, error) {
	return r.lastInsertId, nil
}

func (r *mockResult) RowsAffected() (int64, error) {
	return r.rowsAffected, nil
}

type mockRows struct {
	columns []string
	data    [][]driver.Value
	pos     int
}

func (r *mockRows) Columns() []string {
	return r.columns
}

func (r *mockRows) Close() error {
	return nil
}

func (r *mockRows) Next(dest []driver.Value) error {
	if r.pos >= len(r.data) {
		return fmt.Errorf("no more rows")
	}
	copy(dest, r.data[r.pos])
	r.pos++
	return nil
}

// MockRepository provides mock repository functions for testing
type MockRepository struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

// NewMockRepository creates a new mock repository
func NewMockRepository() *MockRepository {
	return &MockRepository{
		data: make(map[string]interface{}),
	}
}

// Set stores a value in the mock repository
func (r *MockRepository) Set(key string, value interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[key] = value
}

// Get retrieves a value from the mock repository
func (r *MockRepository) Get(key string) (interface{}, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	val, ok := r.data[key]
	return val, ok
}

// TestHelpers provides helper functions for tests
type TestHelpers struct {
	DB *MockDB
}

// NewTestHelpers creates new test helpers
func NewTestHelpers() (*TestHelpers, error) {
	db, err := NewMockDB()
	if err != nil {
		return nil, err
	}
	
	return &TestHelpers{DB: db}, nil
}

// CreateTestDetection creates a test detection
func (h *TestHelpers) CreateTestDetection(name string) map[string]interface{} {
	return map[string]interface{}{
		"id":          1,
		"name":        name,
		"description": "Test detection",
		"status":      "draft",
		"severity":    "medium",
		"confidence":  75,
		"created_at":  time.Now(),
		"updated_at":  time.Now(),
	}
}

// CreateTestRiskObject creates a test risk object
func (h *TestHelpers) CreateTestRiskObject(entityType, entityValue string, score int) map[string]interface{} {
	return map[string]interface{}{
		"id":            1,
		"entity_type":   entityType,
		"entity_value":  entityValue,
		"current_score": score,
		"last_seen":     time.Now(),
	}
}

// CreateTestEvent creates a test event
func (h *TestHelpers) CreateTestEvent(detectionID, entityID int64, riskPoints int) map[string]interface{} {
	return map[string]interface{}{
		"id":               1,
		"detection_id":     detectionID,
		"entity_id":        entityID,
		"timestamp":        time.Now(),
		"risk_points":      riskPoints,
		"is_false_positive": false,
		"raw_data":         `{"test": "data"}`,
	}
}

// Cleanup cleans up test data
func (h *TestHelpers) Cleanup() error {
	return h.DB.Close()
}