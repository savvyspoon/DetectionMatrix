package pkg

import (
	"encoding/json"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strings"

	"riskmatrix/pkg/models"
)

const (
	MaxDetectionNameLength   = 255
	MaxDescriptionLength     = 1000
	MaxRiskPoints           = 500
	MaxRawDataSize          = 10000 // 10KB
	MaxNotesLength          = 2000
	MaxReasonLength         = 1000
	MaxDataSourceNameLength = 255
)

var (
	// MITRE technique ID pattern: T followed by 4 digits, optionally .XXX for sub-techniques
	mitreIDPattern = regexp.MustCompile(`^T\d{4}(\.\d{3})?$`)
	
	// Valid MITRE domains
	validDomains = map[string]bool{
		"Enterprise": true,
		"Mobile":     true,
		"ICS":        true,
	}
	
	// Valid log formats
	validLogFormats = map[string]bool{
		"JSON":    true,
		"XML":     true,
		"CEF":     true,
		"LEEF":    true,
		"Syslog":  true,
		"CSV":     true,
		"Raw":     true,
	}
)

// ValidateDetection validates a detection model
func ValidateDetection(detection *models.Detection) error {
	if detection.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	
	if len(detection.Name) > MaxDetectionNameLength {
		return fmt.Errorf("name too long (max %d characters)", MaxDetectionNameLength)
	}
	
	if len(detection.Description) > MaxDescriptionLength {
		return fmt.Errorf("description too long (max %d characters)", MaxDescriptionLength)
	}
	
	// Validate status
	if !isValidDetectionStatus(detection.Status) {
		return fmt.Errorf("invalid status: %s", detection.Status)
	}
	
	// Validate severity
	if !isValidSeverity(detection.Severity) {
		return fmt.Errorf("invalid severity: %s", detection.Severity)
	}
	
	// Validate risk points
	if detection.RiskPoints < 0 {
		return fmt.Errorf("risk points cannot be negative")
	}
	
	if detection.RiskPoints > MaxRiskPoints {
		return fmt.Errorf("risk points too high (max %d)", MaxRiskPoints)
	}
	
	// Validate playbook URL if provided
	if detection.PlaybookLink != "" {
		if _, err := url.ParseRequestURI(detection.PlaybookLink); err != nil {
			return fmt.Errorf("invalid playbook URL: %v", err)
		}
	}
	
	// Validate owner email if provided
	if detection.Owner != "" {
		if !isValidEmail(detection.Owner) {
			return fmt.Errorf("invalid owner email: %s", detection.Owner)
		}
	}
	
	// Validate risk object type if provided
	if detection.RiskObject != "" && !isValidRiskObjectType(detection.RiskObject) {
		return fmt.Errorf("invalid risk object type: %s", detection.RiskObject)
	}
	
	// Additional validation for production detections
	if detection.Status == models.StatusProduction {
		if detection.Query == "" {
			return fmt.Errorf("production detection must have query")
		}
		if detection.PlaybookLink == "" {
			return fmt.Errorf("production detection must have playbook")
		}
	}
	
	return nil
}

// ValidateEvent validates an event model
func ValidateEvent(event *models.Event) error {
	if event.DetectionID <= 0 {
		return fmt.Errorf("detection ID is required")
	}
	
	if event.EntityID <= 0 {
		return fmt.Errorf("entity ID is required")
	}
	
	if event.RiskPoints < 0 {
		return fmt.Errorf("risk points cannot be negative")
	}
	
	// Validate raw data size
	if len(event.RawData) > MaxRawDataSize {
		return fmt.Errorf("raw data too large (max %d bytes)", MaxRawDataSize)
	}
	
	// Validate JSON format in raw data if provided
	if event.RawData != "" {
		if !isValidJSON(event.RawData) {
			return fmt.Errorf("invalid JSON in raw data")
		}
	}
	
	// Validate JSON format in context if provided
	if event.Context != "" {
		if !isValidJSON(event.Context) {
			return fmt.Errorf("invalid JSON in context")
		}
	}
	
	return nil
}

// ValidateRiskObject validates a risk object model
func ValidateRiskObject(obj *models.RiskObject) error {
	// Validate entity type
	if !isValidEntityType(obj.EntityType) {
		return fmt.Errorf("invalid entity type: %s", obj.EntityType)
	}
	
	// Validate entity value
	if obj.EntityValue == "" {
		return fmt.Errorf("entity value cannot be empty")
	}
	
	// Type-specific validation
	switch obj.EntityType {
	case models.EntityTypeUser:
		if !isValidEmail(obj.EntityValue) {
			return fmt.Errorf("invalid email format for user entity: %s", obj.EntityValue)
		}
	case models.EntityTypeIP:
		if !isValidIPAddress(obj.EntityValue) {
			return fmt.Errorf("invalid IP address format: %s", obj.EntityValue)
		}
	case models.EntityTypeHost:
		if !isValidHostname(obj.EntityValue) {
			return fmt.Errorf("invalid hostname format: %s", obj.EntityValue)
		}
	}
	
	// Validate current score
	if obj.CurrentScore < 0 {
		return fmt.Errorf("current score cannot be negative")
	}
	
	return nil
}

// ValidateMitreTechnique validates a MITRE technique model
func ValidateMitreTechnique(technique *models.MitreTechnique) error {
	// Validate ID format
	if !mitreIDPattern.MatchString(technique.ID) {
		return fmt.Errorf("invalid MITRE technique ID format: %s", technique.ID)
	}
	
	// Validate name
	if technique.Name == "" {
		return fmt.Errorf("technique name cannot be empty")
	}
	
	// Validate tactic
	if technique.Tactic == "" {
		return fmt.Errorf("tactic cannot be empty")
	}
	
	// Validate domain
	if technique.Domain != "" && !validDomains[technique.Domain] {
		return fmt.Errorf("invalid domain: %s", technique.Domain)
	}
	
	// Validate sub-technique logic
	if technique.IsSubTechnique {
		// Check if this looks like a parent technique ID (no dot)
		if !strings.Contains(technique.ID, ".") {
			return fmt.Errorf("parent technique cannot be marked as sub-technique")
		}
		// Sub-techniques must have parent
		if technique.SubTechniqueOf == "" {
			return fmt.Errorf("sub-technique must have parent technique ID")
		}
	} else {
		// Parent techniques should not have SubTechniqueOf set
		if technique.SubTechniqueOf != "" {
			return fmt.Errorf("parent technique cannot have SubTechniqueOf field set")
		}
	}
	
	return nil
}

// ValidateDataSource validates a data source model
func ValidateDataSource(dataSource *models.DataSource) error {
	// Validate name
	if dataSource.Name == "" {
		return fmt.Errorf("data source name cannot be empty")
	}
	
	if len(dataSource.Name) > MaxDataSourceNameLength {
		return fmt.Errorf("data source name too long (max %d characters)", MaxDataSourceNameLength)
	}
	
	// Validate description length
	if len(dataSource.Description) > MaxDescriptionLength {
		return fmt.Errorf("description too long (max %d characters)", MaxDescriptionLength)
	}
	
	// Validate log format if provided
	if dataSource.LogFormat != "" && !validLogFormats[dataSource.LogFormat] {
		return fmt.Errorf("invalid log format: %s", dataSource.LogFormat)
	}
	
	return nil
}

// ValidateRiskAlert validates a risk alert model
func ValidateRiskAlert(alert *models.RiskAlert) error {
	if alert.EntityID <= 0 {
		return fmt.Errorf("entity ID is required")
	}
	
	if alert.TotalScore < 0 {
		return fmt.Errorf("total score cannot be negative")
	}
	
	// Validate status
	if !isValidAlertStatus(alert.Status) {
		return fmt.Errorf("invalid alert status: %s", alert.Status)
	}
	
	// Validate owner email if provided
	if alert.Owner != "" && !isValidEmail(alert.Owner) {
		return fmt.Errorf("invalid owner email format: %s", alert.Owner)
	}
	
	// Validate notes length
	if len(alert.Notes) > MaxNotesLength {
		return fmt.Errorf("notes too long (max %d characters)", MaxNotesLength)
	}
	
	return nil
}

// ValidateFalsePositive validates a false positive model
func ValidateFalsePositive(fp *models.FalsePositive) error {
	if fp.EventID <= 0 {
		return fmt.Errorf("event ID is required")
	}
	
	if fp.AnalystName == "" {
		return fmt.Errorf("analyst name cannot be empty")
	}
	
	// Validate analyst email format
	if !isValidEmail(fp.AnalystName) {
		return fmt.Errorf("invalid analyst email format: %s", fp.AnalystName)
	}
	
	// Validate reason length
	if len(fp.Reason) > MaxReasonLength {
		return fmt.Errorf("reason too long (max %d characters)", MaxReasonLength)
	}
	
	return nil
}

// Helper functions

func isValidDetectionStatus(status models.DetectionStatus) bool {
	switch status {
	case models.StatusIdea, models.StatusDraft, models.StatusTest, models.StatusProduction, models.StatusRetired:
		return true
	default:
		return false
	}
}

func isValidSeverity(severity models.Severity) bool {
	switch severity {
	case models.SeverityLow, models.SeverityMedium, models.SeverityHigh, models.SeverityCritical:
		return true
	default:
		return false
	}
}

func isValidRiskObjectType(riskType models.RiskObjectType) bool {
	switch riskType {
	case models.RiskObjectIP, models.RiskObjectHost, models.RiskObjectUser:
		return true
	default:
		return false
	}
}

func isValidEntityType(entityType models.EntityType) bool {
	switch entityType {
	case models.EntityTypeUser, models.EntityTypeHost, models.EntityTypeIP:
		return true
	default:
		return false
	}
}

func isValidAlertStatus(status models.AlertStatus) bool {
	switch status {
	case models.AlertStatusNew, models.AlertStatusTriage, models.AlertStatusInvestigation,
		 models.AlertStatusOnHold, models.AlertStatusIncident, models.AlertStatusClosed:
		return true
	default:
		return false
	}
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func isValidJSON(jsonStr string) bool {
	// Use json.Valid for efficient validation without allocation
	return json.Valid([]byte(jsonStr))
}

func isValidIPAddress(ip string) bool {
	// Use standard library net.ParseIP for proper IP validation
	return net.ParseIP(ip) != nil
}

func isValidHostname(hostname string) bool {
	// Basic hostname validation
	if len(hostname) == 0 || len(hostname) > 253 {
		return false
	}
	
	// Check for invalid characters
	for _, char := range hostname {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '.') {
			return false
		}
	}
	
	// Hostname cannot start or end with hyphen
	if strings.HasPrefix(hostname, "-") || strings.HasSuffix(hostname, "-") {
		return false
	}
	
	return true
}