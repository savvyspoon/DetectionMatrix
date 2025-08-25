package models

import (
	"time"
)

// EntityType represents the type of entity that can accumulate risk
type EntityType string

const (
	EntityTypeUser EntityType = "user"
	EntityTypeHost EntityType = "host"
	EntityTypeIP   EntityType = "ip"
)

// AlertStatus represents the status of a risk alert
type AlertStatus string

const (
	AlertStatusNew           AlertStatus = "New"
	AlertStatusTriage        AlertStatus = "Triage"
	AlertStatusInvestigation AlertStatus = "Investigation"
	AlertStatusOnHold        AlertStatus = "On Hold"
	AlertStatusIncident      AlertStatus = "Incident"
	AlertStatusClosed        AlertStatus = "Closed"
)

// RiskObject represents an entity that accumulates risk
type RiskObject struct {
	ID           int64      `json:"id"`
	EntityType   EntityType `json:"entity_type"` // user, host, IP
	EntityValue  string     `json:"entity_value"`
	CurrentScore int        `json:"current_score"`
	LastSeen     time.Time  `json:"last_seen"`
}

// Event represents a detection trigger
type Event struct {
	ID              int64     `json:"id"`
	DetectionID     int64     `json:"detection_id"`
	EntityID        int64     `json:"entity_id"`
	Timestamp       time.Time `json:"timestamp"`
	RawData         string    `json:"raw_data,omitempty"`
	Context         string    `json:"context,omitempty"` // JSON field for detection context information
	RiskPoints      int       `json:"risk_points"`
	IsFalsePositive bool      `json:"is_false_positive"`

	// Relationships (for convenience)
	Detection  *Detection  `json:"detection,omitempty"`
	RiskObject *RiskObject `json:"risk_object,omitempty"`
}

// RiskAlert represents a high-level alert generated when risk threshold is exceeded
type RiskAlert struct {
	ID          int64       `json:"id"`
	EntityID    int64       `json:"entity_id"`
	TriggeredAt time.Time   `json:"triggered_at"`
	TotalScore  int         `json:"total_score"`
	Status      AlertStatus `json:"status"`
	Notes       string      `json:"notes,omitempty"`
	Owner       string      `json:"owner,omitempty"`

	// Relationships (for convenience)
	RiskObject *RiskObject `json:"risk_object,omitempty"`
	Events     []*Event    `json:"events,omitempty"` // Contributing events
}

// FalsePositive represents an analyst-logged false positive
type FalsePositive struct {
	ID          int64     `json:"id"`
	EventID     int64     `json:"event_id"`
	Reason      string    `json:"reason,omitempty"`
	AnalystName string    `json:"analyst_name"`
	Timestamp   time.Time `json:"timestamp"`

	// Relationships (for convenience)
	Event *Event `json:"event,omitempty"`
}

// RiskRepository defines the interface for risk data access
type RiskRepository interface {
	// RiskObject operations
	GetRiskObject(id int64) (*RiskObject, error)
	GetRiskObjectByEntity(entityType EntityType, entityValue string) (*RiskObject, error)
	ListRiskObjects() ([]*RiskObject, error)
	ListHighRiskObjects(threshold int) ([]*RiskObject, error)
	CreateRiskObject(obj *RiskObject) error
	UpdateRiskObject(obj *RiskObject) error
	DeleteRiskObject(id int64) error

	// Event operations
	GetEvent(id int64) (*Event, error)
	ListEvents() ([]*Event, error)
	ListEventsPaginated(limit, offset int) ([]*Event, int, error)
	ListEventsByEntity(entityID int64) ([]*Event, error)
	ListEventsByDetection(detectionID int64) ([]*Event, error)
	CreateEvent(event *Event) error
	MarkEventAsFalsePositive(eventID int64, fpInfo *FalsePositive) error

	// RiskAlert operations
	GetRiskAlert(id int64) (*RiskAlert, error)
	ListRiskAlerts() ([]*RiskAlert, error)
	ListRiskAlertsByEntity(entityID int64) ([]*RiskAlert, error)
	CreateRiskAlert(alert *RiskAlert) error
	UpdateRiskAlert(alert *RiskAlert) error
	GetEventsForAlert(alertID int64) ([]*Event, error)

	// FalsePositive operations
	GetFalsePositive(id int64) (*FalsePositive, error)
	ListFalsePositives() ([]*FalsePositive, error)
	ListFalsePositivesByDetection(detectionID int64) ([]*FalsePositive, error)

	// Risk scoring
	UpdateEntityRiskScore(entityID int64, points int) error
	DecayRiskScores(decayFactor float64) error // Periodically reduce risk scores
}
