package models

import (
	"time"
)

// DetectionStatus represents the lifecycle stage of a detection
type DetectionStatus string

const (
	StatusIdea       DetectionStatus = "idea"
	StatusDraft      DetectionStatus = "draft"
	StatusTest       DetectionStatus = "test"
	StatusProduction DetectionStatus = "production"
	StatusRetired    DetectionStatus = "retired"
)

// Severity represents the severity level of a detection
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// RiskObjectType represents the type of risk object
type RiskObjectType string

const (
	RiskObjectIP   RiskObjectType = "IP"
	RiskObjectHost RiskObjectType = "Host"
	RiskObjectUser RiskObjectType = "User"
)

// DetectionClass represents a category for detections
type DetectionClass struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	Color        string    `json:"color,omitempty"`
	Icon         string    `json:"icon,omitempty"`
	IsSystem     bool      `json:"is_system"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Detection represents a security detection rule
type Detection struct {
	ID                       int64           `json:"id"`
	Name                     string          `json:"name"`
	Description              string          `json:"description"`
	Query                    string          `json:"query,omitempty"`
	Status                   DetectionStatus `json:"status"`
	Severity                 Severity        `json:"severity"`
	RiskPoints               int             `json:"risk_points"`
	PlaybookLink             string          `json:"playbook_link,omitempty"`
	Owner                    string          `json:"owner,omitempty"`
	RiskObject               RiskObjectType  `json:"risk_object,omitempty"`
	TestingDescription       string          `json:"testing_description,omitempty"`
	EventCountLast30Days     int             `json:"event_count_last_30_days"`
	FalsePositivesLast30Days int             `json:"false_positives_last_30_days"`
	ClassID                  *int64          `json:"class_id,omitempty"`
	CreatedAt                time.Time       `json:"created_at"`
	UpdatedAt                time.Time       `json:"updated_at"`

	// Relationships
	Class           *DetectionClass  `json:"class,omitempty"`
	MitreTechniques []MitreTechnique `json:"mitre_techniques,omitempty"`
	DataSources     []DataSource     `json:"data_sources,omitempty"`
}

// DetectionRepository defines the interface for detection data access
type DetectionRepository interface {
	// Basic CRUD operations
	GetDetection(id int64) (*Detection, error)
	ListDetections() ([]*Detection, error)
	ListDetectionsByStatus(status DetectionStatus) ([]*Detection, error)
	CreateDetection(detection *Detection) error
	UpdateDetection(detection *Detection) error
	DeleteDetection(id int64) error

	// Relationship operations
	AddMitreTechnique(detectionID int64, mitreID string) error
	RemoveMitreTechnique(detectionID int64, mitreID string) error
	AddDataSource(detectionID int64, dataSourceID int64) error
	RemoveDataSource(detectionID int64, dataSourceID int64) error

	// Detection Class operations
	GetDetectionClass(id int64) (*DetectionClass, error)
	ListDetectionClasses() ([]*DetectionClass, error)
	CreateDetectionClass(class *DetectionClass) error
	UpdateDetectionClass(class *DetectionClass) error
	DeleteDetectionClass(id int64) error
	ListDetectionsByClass(classID int64) ([]*Detection, error)

	// Analytics
	GetDetectionCount() (int, error)
	GetDetectionCountByStatus() (map[DetectionStatus]int, error)
	GetFalsePositiveRate(detectionID int64) (float64, error)
}
