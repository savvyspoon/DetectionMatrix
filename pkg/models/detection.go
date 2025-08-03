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
	SeverityLow    Severity = "low"
	SeverityMedium Severity = "medium"
	SeverityHigh   Severity = "high"
	SeverityCritical Severity = "critical"
)

// RiskObjectType represents the type of risk object
type RiskObjectType string

const (
	RiskObjectIP   RiskObjectType = "IP"
	RiskObjectHost RiskObjectType = "Host"
	RiskObjectUser RiskObjectType = "User"
)

// Detection represents a security detection rule
type Detection struct {
	ID                         int64           `json:"id"`
	Name                       string          `json:"name"`
	Description                string          `json:"description"`
	Query                      string          `json:"query,omitempty"`
	Status                     DetectionStatus `json:"status"`
	Severity                   Severity        `json:"severity"`
	RiskPoints                 int             `json:"risk_points"`
	PlaybookLink               string          `json:"playbook_link,omitempty"`
	Owner                      string          `json:"owner,omitempty"`
	RiskObject                 RiskObjectType  `json:"risk_object,omitempty"`
	TestingDescription         string          `json:"testing_description,omitempty"`
	CreatedAt                  time.Time       `json:"created_at"`
	UpdatedAt                  time.Time       `json:"updated_at"`
	
	// Relationships
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
	
	// Analytics
	GetDetectionCount() (int, error)
	GetDetectionCountByStatus() (map[DetectionStatus]int, error)
	GetFalsePositiveRate(detectionID int64) (float64, error)
}