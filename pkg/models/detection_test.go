package models

import (
	"testing"
	"time"
)

func TestDetectionStatus(t *testing.T) {
	tests := []struct {
		status DetectionStatus
		valid  bool
	}{
		{StatusIdea, true},
		{StatusDraft, true},
		{StatusTest, true},
		{StatusProduction, true},
		{StatusRetired, true},
		{"unknown", false},
	}

	for _, test := range tests {
		valid := isValidStatus(test.status)
		if valid != test.valid {
			t.Errorf("Status %s validation: expected %v, got %v", test.status, test.valid, valid)
		}
	}
}

func isValidStatus(status DetectionStatus) bool {
	switch status {
	case StatusIdea, StatusDraft, StatusTest, StatusProduction, StatusRetired:
		return true
	default:
		return false
	}
}

func TestSeverity(t *testing.T) {
	tests := []struct {
		severity Severity
		valid    bool
	}{
		{SeverityLow, true},
		{SeverityMedium, true},
		{SeverityHigh, true},
		{SeverityCritical, true},
		{"unknown", false},
	}

	for _, test := range tests {
		valid := isValidSeverity(test.severity)
		if valid != test.valid {
			t.Errorf("Severity %s validation: expected %v, got %v", test.severity, test.valid, valid)
		}
	}
}

func isValidSeverity(severity Severity) bool {
	switch severity {
	case SeverityLow, SeverityMedium, SeverityHigh, SeverityCritical:
		return true
	default:
		return false
	}
}

func TestDetection(t *testing.T) {
	// Create a detection
	detection := Detection{
		ID:           1,
		Name:         "Test Detection",
		Description:  "This is a test detection",
		Status:       StatusDraft,
		Severity:     SeverityMedium,
		RiskPoints:   25,
		PlaybookLink: "https://example.com/playbooks/test",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Test detection fields
	if detection.ID != 1 {
		t.Errorf("Expected ID 1, got %d", detection.ID)
	}

	if detection.Name != "Test Detection" {
		t.Errorf("Expected name 'Test Detection', got '%s'", detection.Name)
	}

	if detection.Status != StatusDraft {
		t.Errorf("Expected status '%s', got '%s'", StatusDraft, detection.Status)
	}

	if detection.Severity != SeverityMedium {
		t.Errorf("Expected severity '%s', got '%s'", SeverityMedium, detection.Severity)
	}

	if detection.RiskPoints != 25 {
		t.Errorf("Expected risk points 25, got %d", detection.RiskPoints)
	}
}
