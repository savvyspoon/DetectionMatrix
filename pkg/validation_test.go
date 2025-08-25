package pkg

import (
	"strings"
	"testing"

	"riskmatrix/pkg/models"
)

// ValidationTestSuite provides comprehensive validation testing
type ValidationTestSuite struct{}

func TestDetectionValidation(t *testing.T) {
	tests := []struct {
		name      string
		detection *models.Detection
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid detection",
			detection: &models.Detection{
				Name:       "Valid Detection",
				Status:     models.StatusDraft,
				Severity:   models.SeverityMedium,
				RiskPoints: 25,
			},
			wantError: false,
		},
		{
			name: "Empty name",
			detection: &models.Detection{
				Name:       "",
				Status:     models.StatusDraft,
				Severity:   models.SeverityMedium,
				RiskPoints: 25,
			},
			wantError: true,
			errorMsg:  "name cannot be empty",
		},
		{
			name: "Name too long",
			detection: &models.Detection{
				Name:       strings.Repeat("a", 256), // Assuming 255 is max
				Status:     models.StatusDraft,
				Severity:   models.SeverityMedium,
				RiskPoints: 25,
			},
			wantError: true,
			errorMsg:  "name too long",
		},
		{
			name: "Invalid status",
			detection: &models.Detection{
				Name:       "Test Detection",
				Status:     "invalid_status",
				Severity:   models.SeverityMedium,
				RiskPoints: 25,
			},
			wantError: true,
			errorMsg:  "invalid status",
		},
		{
			name: "Invalid severity",
			detection: &models.Detection{
				Name:       "Test Detection",
				Status:     models.StatusDraft,
				Severity:   "invalid_severity",
				RiskPoints: 25,
			},
			wantError: true,
			errorMsg:  "invalid severity",
		},
		{
			name: "Negative risk points",
			detection: &models.Detection{
				Name:       "Test Detection",
				Status:     models.StatusDraft,
				Severity:   models.SeverityMedium,
				RiskPoints: -10,
			},
			wantError: true,
			errorMsg:  "risk points cannot be negative",
		},
		{
			name: "Risk points too high",
			detection: &models.Detection{
				Name:       "Test Detection",
				Status:     models.StatusDraft,
				Severity:   models.SeverityMedium,
				RiskPoints: 1000, // Assuming 500 is max
			},
			wantError: true,
			errorMsg:  "risk points too high",
		},
		{
			name: "Invalid playbook URL",
			detection: &models.Detection{
				Name:         "Test Detection",
				Status:       models.StatusDraft,
				Severity:     models.SeverityMedium,
				RiskPoints:   25,
				PlaybookLink: "not-a-valid-url",
			},
			wantError: true,
			errorMsg:  "invalid playbook URL",
		},
		{
			name: "Invalid email owner",
			detection: &models.Detection{
				Name:       "Test Detection",
				Status:     models.StatusDraft,
				Severity:   models.SeverityMedium,
				RiskPoints: 25,
				Owner:      "not-an-email",
			},
			wantError: true,
			errorMsg:  "invalid owner email",
		},
		{
			name: "Invalid risk object type",
			detection: &models.Detection{
				Name:       "Test Detection",
				Status:     models.StatusDraft,
				Severity:   models.SeverityMedium,
				RiskPoints: 25,
				RiskObject: "InvalidType",
			},
			wantError: true,
			errorMsg:  "invalid risk object type",
		},
		{
			name: "Production detection without query",
			detection: &models.Detection{
				Name:       "Production Detection",
				Status:     models.StatusProduction,
				Severity:   models.SeverityHigh,
				RiskPoints: 50,
				Query:      "", // Empty query for production detection
			},
			wantError: true,
			errorMsg:  "production detection must have query",
		},
		{
			name: "Production detection without playbook",
			detection: &models.Detection{
				Name:         "Production Detection",
				Status:       models.StatusProduction,
				Severity:     models.SeverityHigh,
				RiskPoints:   50,
				Query:        "SELECT * FROM logs WHERE suspicious = 1",
				PlaybookLink: "", // Empty playbook for production detection
			},
			wantError: true,
			errorMsg:  "production detection must have playbook",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDetection(tt.detection)

			if tt.wantError && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
			if tt.wantError && err != nil && !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("Expected error message to contain '%s', got '%v'", tt.errorMsg, err)
			}
		})
	}
}

func TestEventValidation(t *testing.T) {
	tests := []struct {
		name      string
		event     *models.Event
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid event",
			event: &models.Event{
				DetectionID: 1,
				EntityID:    1,
				RiskPoints:  25,
				RawData:     `{"valid": "json"}`,
				Context:     `{"context": "data"}`,
			},
			wantError: false,
		},
		{
			name: "Zero detection ID",
			event: &models.Event{
				DetectionID: 0,
				EntityID:    1,
				RiskPoints:  25,
			},
			wantError: true,
			errorMsg:  "detection ID is required",
		},
		{
			name: "Zero entity ID",
			event: &models.Event{
				DetectionID: 1,
				EntityID:    0,
				RiskPoints:  25,
			},
			wantError: true,
			errorMsg:  "entity ID is required",
		},
		{
			name: "Negative risk points",
			event: &models.Event{
				DetectionID: 1,
				EntityID:    1,
				RiskPoints:  -5,
			},
			wantError: true,
			errorMsg:  "risk points cannot be negative",
		},
		{
			name: "Invalid JSON in raw data",
			event: &models.Event{
				DetectionID: 1,
				EntityID:    1,
				RiskPoints:  25,
				RawData:     `{"invalid": json}`, // Missing quotes around json
			},
			wantError: true,
			errorMsg:  "invalid JSON in raw data",
		},
		{
			name: "Invalid JSON in context",
			event: &models.Event{
				DetectionID: 1,
				EntityID:    1,
				RiskPoints:  25,
				RawData:     `{"valid": "json"}`,
				Context:     `{"invalid": context}`, // Missing quotes around context
			},
			wantError: true,
			errorMsg:  "invalid JSON in context",
		},
		{
			name: "Raw data too large",
			event: &models.Event{
				DetectionID: 1,
				EntityID:    1,
				RiskPoints:  25,
				RawData:     strings.Repeat("a", 10001), // Assuming 10KB limit
			},
			wantError: true,
			errorMsg:  "raw data too large",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEvent(tt.event)

			if tt.wantError && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
			if tt.wantError && err != nil && !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("Expected error message to contain '%s', got '%v'", tt.errorMsg, err)
			}
		})
	}
}

func TestRiskObjectValidation(t *testing.T) {
	tests := []struct {
		name       string
		riskObject *models.RiskObject
		wantError  bool
		errorMsg   string
	}{
		{
			name: "Valid risk object",
			riskObject: &models.RiskObject{
				EntityType:   models.EntityTypeUser,
				EntityValue:  "user@example.com",
				CurrentScore: 50,
			},
			wantError: false,
		},
		{
			name: "Invalid entity type",
			riskObject: &models.RiskObject{
				EntityType:   "invalid_type",
				EntityValue:  "value",
				CurrentScore: 50,
			},
			wantError: true,
			errorMsg:  "invalid entity type",
		},
		{
			name: "Empty entity value",
			riskObject: &models.RiskObject{
				EntityType:   models.EntityTypeHost,
				EntityValue:  "",
				CurrentScore: 50,
			},
			wantError: true,
			errorMsg:  "entity value cannot be empty",
		},
		{
			name: "Negative current score",
			riskObject: &models.RiskObject{
				EntityType:   models.EntityTypeIP,
				EntityValue:  "192.168.1.1",
				CurrentScore: -10,
			},
			wantError: true,
			errorMsg:  "current score cannot be negative",
		},
		{
			name: "Invalid email for user entity",
			riskObject: &models.RiskObject{
				EntityType:   models.EntityTypeUser,
				EntityValue:  "not-an-email",
				CurrentScore: 50,
			},
			wantError: true,
			errorMsg:  "invalid email format for user entity",
		},
		{
			name: "Invalid IP address for IP entity",
			riskObject: &models.RiskObject{
				EntityType:   models.EntityTypeIP,
				EntityValue:  "not-an-ip",
				CurrentScore: 50,
			},
			wantError: true,
			errorMsg:  "invalid IP address format",
		},
		{
			name: "Invalid hostname for host entity",
			riskObject: &models.RiskObject{
				EntityType:   models.EntityTypeHost,
				EntityValue:  "host@with@invalid@chars",
				CurrentScore: 50,
			},
			wantError: true,
			errorMsg:  "invalid hostname format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRiskObject(tt.riskObject)

			if tt.wantError && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
			if tt.wantError && err != nil && !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("Expected error message to contain '%s', got '%v'", tt.errorMsg, err)
			}
		})
	}
}

func TestMitreTechniqueValidation(t *testing.T) {
	tests := []struct {
		name      string
		technique *models.MitreTechnique
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid MITRE technique",
			technique: &models.MitreTechnique{
				ID:          "T1059.001",
				Name:        "PowerShell",
				Description: "Adversaries may abuse PowerShell commands",
				Tactic:      "Execution",
				Domain:      "Enterprise",
			},
			wantError: false,
		},
		{
			name: "Invalid MITRE ID format",
			technique: &models.MitreTechnique{
				ID:     "INVALID_ID",
				Name:   "Test Technique",
				Tactic: "Execution",
				Domain: "Enterprise",
			},
			wantError: true,
			errorMsg:  "invalid MITRE technique ID format",
		},
		{
			name: "Empty name",
			technique: &models.MitreTechnique{
				ID:     "T1059",
				Name:   "",
				Tactic: "Execution",
				Domain: "Enterprise",
			},
			wantError: true,
			errorMsg:  "technique name cannot be empty",
		},
		{
			name: "Empty tactic",
			technique: &models.MitreTechnique{
				ID:     "T1059",
				Name:   "Test Technique",
				Tactic: "",
				Domain: "Enterprise",
			},
			wantError: true,
			errorMsg:  "tactic cannot be empty",
		},
		{
			name: "Invalid domain",
			technique: &models.MitreTechnique{
				ID:     "T1059",
				Name:   "Test Technique",
				Tactic: "Execution",
				Domain: "InvalidDomain",
			},
			wantError: true,
			errorMsg:  "invalid domain",
		},
		{
			name: "Sub-technique without parent",
			technique: &models.MitreTechnique{
				ID:             "T1059.001",
				Name:           "PowerShell",
				Tactic:         "Execution",
				Domain:         "Enterprise",
				IsSubTechnique: true,
				SubTechniqueOf: "", // Missing parent
			},
			wantError: true,
			errorMsg:  "sub-technique must have parent technique ID",
		},
		{
			name: "Parent technique with sub-technique flag",
			technique: &models.MitreTechnique{
				ID:             "T1059",
				Name:           "Command and Scripting Interpreter",
				Tactic:         "Execution",
				Domain:         "Enterprise",
				IsSubTechnique: true, // Should be false for parent
				SubTechniqueOf: "",
			},
			wantError: true,
			errorMsg:  "parent technique cannot be marked as sub-technique",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMitreTechnique(tt.technique)

			if tt.wantError && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
			if tt.wantError && err != nil && !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("Expected error message to contain '%s', got '%v'", tt.errorMsg, err)
			}
		})
	}
}

func TestDataSourceValidation(t *testing.T) {
	tests := []struct {
		name       string
		dataSource *models.DataSource
		wantError  bool
		errorMsg   string
	}{
		{
			name: "Valid data source",
			dataSource: &models.DataSource{
				Name:        "Windows Event Logs",
				Description: "Windows security event logs",
				LogFormat:   "JSON",
			},
			wantError: false,
		},
		{
			name: "Empty name",
			dataSource: &models.DataSource{
				Name:        "",
				Description: "Description",
				LogFormat:   "JSON",
			},
			wantError: true,
			errorMsg:  "data source name cannot be empty",
		},
		{
			name: "Name too long",
			dataSource: &models.DataSource{
				Name:        strings.Repeat("a", 256),
				Description: "Description",
				LogFormat:   "JSON",
			},
			wantError: true,
			errorMsg:  "data source name too long",
		},
		{
			name: "Invalid log format",
			dataSource: &models.DataSource{
				Name:        "Test Source",
				Description: "Description",
				LogFormat:   "INVALID_FORMAT",
			},
			wantError: true,
			errorMsg:  "invalid log format",
		},
		{
			name: "Description too long",
			dataSource: &models.DataSource{
				Name:        "Test Source",
				Description: strings.Repeat("a", 1001), // Assuming 1000 char limit
				LogFormat:   "JSON",
			},
			wantError: true,
			errorMsg:  "description too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDataSource(tt.dataSource)

			if tt.wantError && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
			if tt.wantError && err != nil && !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("Expected error message to contain '%s', got '%v'", tt.errorMsg, err)
			}
		})
	}
}

func TestRiskAlertValidation(t *testing.T) {
	tests := []struct {
		name      string
		alert     *models.RiskAlert
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid risk alert",
			alert: &models.RiskAlert{
				EntityID:   1,
				TotalScore: 100,
				Status:     models.AlertStatusNew,
			},
			wantError: false,
		},
		{
			name: "Zero entity ID",
			alert: &models.RiskAlert{
				EntityID:   0,
				TotalScore: 100,
				Status:     models.AlertStatusNew,
			},
			wantError: true,
			errorMsg:  "entity ID is required",
		},
		{
			name: "Negative total score",
			alert: &models.RiskAlert{
				EntityID:   1,
				TotalScore: -50,
				Status:     models.AlertStatusNew,
			},
			wantError: true,
			errorMsg:  "total score cannot be negative",
		},
		{
			name: "Invalid status",
			alert: &models.RiskAlert{
				EntityID:   1,
				TotalScore: 100,
				Status:     "InvalidStatus",
			},
			wantError: true,
			errorMsg:  "invalid alert status",
		},
		{
			name: "Invalid owner email",
			alert: &models.RiskAlert{
				EntityID:   1,
				TotalScore: 100,
				Status:     models.AlertStatusNew,
				Owner:      "not-an-email",
			},
			wantError: true,
			errorMsg:  "invalid owner email format",
		},
		{
			name: "Notes too long",
			alert: &models.RiskAlert{
				EntityID:   1,
				TotalScore: 100,
				Status:     models.AlertStatusNew,
				Notes:      strings.Repeat("a", 2001), // Assuming 2000 char limit
			},
			wantError: true,
			errorMsg:  "notes too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRiskAlert(tt.alert)

			if tt.wantError && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
			if tt.wantError && err != nil && !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("Expected error message to contain '%s', got '%v'", tt.errorMsg, err)
			}
		})
	}
}

func TestFalsePositiveValidation(t *testing.T) {
	tests := []struct {
		name          string
		falsePositive *models.FalsePositive
		wantError     bool
		errorMsg      string
	}{
		{
			name: "Valid false positive",
			falsePositive: &models.FalsePositive{
				EventID:     1,
				Reason:      "Legitimate admin activity",
				AnalystName: "analyst@example.com",
			},
			wantError: false,
		},
		{
			name: "Zero event ID",
			falsePositive: &models.FalsePositive{
				EventID:     0,
				Reason:      "Test reason",
				AnalystName: "analyst@example.com",
			},
			wantError: true,
			errorMsg:  "event ID is required",
		},
		{
			name: "Empty analyst name",
			falsePositive: &models.FalsePositive{
				EventID:     1,
				Reason:      "Test reason",
				AnalystName: "",
			},
			wantError: true,
			errorMsg:  "analyst name cannot be empty",
		},
		{
			name: "Invalid analyst email",
			falsePositive: &models.FalsePositive{
				EventID:     1,
				Reason:      "Test reason",
				AnalystName: "not-an-email",
			},
			wantError: true,
			errorMsg:  "invalid analyst email format",
		},
		{
			name: "Reason too long",
			falsePositive: &models.FalsePositive{
				EventID:     1,
				Reason:      strings.Repeat("a", 1001), // Assuming 1000 char limit
				AnalystName: "analyst@example.com",
			},
			wantError: true,
			errorMsg:  "reason too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFalsePositive(tt.falsePositive)

			if tt.wantError && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
			if tt.wantError && err != nil && !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("Expected error message to contain '%s', got '%v'", tt.errorMsg, err)
			}
		})
	}
}
