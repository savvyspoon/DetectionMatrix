package internal

import (
	"encoding/json"
	"testing"

	"riskmatrix/internal/detection"
	"riskmatrix/internal/datasource"
	"riskmatrix/internal/mitre"
	"riskmatrix/internal/risk"
	"riskmatrix/pkg/database"
	"riskmatrix/pkg/models"
)

// IntegrationTestSuite provides a complete testing environment
type IntegrationTestSuite struct {
	DB             *database.DB
	DetectionRepo  *detection.Repository
	MitreRepo      *mitre.Repository
	DataSourceRepo *datasource.Repository
	RiskRepo       *risk.Repository
	RiskEngine     *risk.Engine
}

// SetupIntegrationTest creates a complete test environment
func SetupIntegrationTest(t *testing.T) *IntegrationTestSuite {
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	config := risk.DefaultConfig()
	config.RiskThreshold = 75 // Lower threshold for testing

	return &IntegrationTestSuite{
		DB:             db,
		DetectionRepo:  detection.NewRepository(db),
		MitreRepo:      mitre.NewRepository(db),
		DataSourceRepo: datasource.NewRepository(db),
		RiskRepo:       risk.NewRepository(db),
		RiskEngine:     risk.NewEngine(db, config),
	}
}

func (suite *IntegrationTestSuite) Cleanup() {
	suite.DB.Close()
}

func TestFullDetectionPipeline(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.Cleanup()

	// Step 1: Create MITRE technique
	mitretech := &models.MitreTechnique{
		ID:          "T1059.001",
		Name:        "PowerShell",
		Description: "Adversaries may abuse PowerShell commands",
		Tactic:      "Execution",
		Domain:      "Enterprise",
	}
	err := suite.MitreRepo.CreateMitreTechnique(mitretech)
	if err != nil {
		t.Fatalf("Failed to create MITRE technique: %v", err)
	}

	// Step 2: Create data source
	dataSource := &models.DataSource{
		Name:        "Windows PowerShell Logs",
		Description: "PowerShell command logging",
		LogFormat:   "JSON",
	}
	err = suite.DataSourceRepo.CreateDataSource(dataSource)
	if err != nil {
		t.Fatalf("Failed to create data source: %v", err)
	}

	// Step 3: Create detection
	detection := &models.Detection{
		Name:               "Suspicious PowerShell Execution",
		Description:        "Detects potentially malicious PowerShell commands",
		Query:              "SELECT * FROM powershell_logs WHERE command_line LIKE '%Invoke-Expression%'",
		Status:             models.StatusProduction,
		Severity:           models.SeverityHigh,
		RiskPoints:         35,
		PlaybookLink:       "https://playbook.example.com/powershell",
		Owner:              "security-team@example.com",
		RiskObject:         models.RiskObjectHost,
		TestingDescription: "Test with known malicious PowerShell commands",
	}
	err = suite.DetectionRepo.CreateDetection(detection)
	if err != nil {
		t.Fatalf("Failed to create detection: %v", err)
	}

	// Step 4: Link detection to MITRE technique
	err = suite.DetectionRepo.AddMitreTechnique(detection.ID, mitretech.ID)
	if err != nil {
		t.Fatalf("Failed to link detection to MITRE technique: %v", err)
	}

	// Step 5: Link detection to data source
	err = suite.DetectionRepo.AddDataSource(detection.ID, dataSource.ID)
	if err != nil {
		t.Fatalf("Failed to link detection to data source: %v", err)
	}

	// Step 6: Process security event (risk object will be created automatically)
	event := &models.Event{
		DetectionID: detection.ID,
		RiskPoints:  35,
		RawData:     `{"command": "Invoke-Expression", "user": "suspicious-user", "timestamp": "2024-01-15T10:30:00Z"}`,
		Context:     `{"process_id": 1234, "parent_process": "cmd.exe", "source_ip": "192.168.1.100"}`,
		RiskObject: &models.RiskObject{
			EntityType:  models.EntityTypeHost,
			EntityValue: "workstation-042",
		},
	}

	err = suite.RiskEngine.ProcessEvent(event)
	if err != nil {
		t.Fatalf("Failed to process event: %v", err)
	}

	// Step 7: Verify event was stored
	if event.ID == 0 {
		t.Error("Event ID should be set after processing")
	}

	storedEvent, err := suite.RiskRepo.GetEvent(event.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve stored event: %v", err)
	}

	if storedEvent.DetectionID != detection.ID {
		t.Errorf("Expected detection ID %d, got %d", detection.ID, storedEvent.DetectionID)
	}
	if storedEvent.RiskPoints != 35 {
		t.Errorf("Expected risk points 35, got %d", storedEvent.RiskPoints)
	}

	// Step 8: Verify risk object was created and score updated
	updatedRiskObj, err := suite.RiskRepo.GetRiskObjectByEntity(models.EntityTypeHost, "workstation-042")
	if err != nil {
		t.Fatalf("Failed to get updated risk object: %v", err)
	}

	if updatedRiskObj.CurrentScore != 35 {
		t.Errorf("Expected risk score 35, got %d", updatedRiskObj.CurrentScore)
	}

	// Step 9: Process another event to trigger alert (35 + 45 = 80 > threshold 75)
	event2 := &models.Event{
		DetectionID: detection.ID,
		RiskPoints:  45,
		RawData:     `{"command": "Download-String", "user": "suspicious-user", "timestamp": "2024-01-15T10:35:00Z"}`,
		Context:     `{"process_id": 5678, "parent_process": "powershell.exe", "source_ip": "192.168.1.100"}`,
		RiskObject: &models.RiskObject{
			EntityType:  models.EntityTypeHost,
			EntityValue: "workstation-042",
		},
	}

	err = suite.RiskEngine.ProcessEvent(event2)
	if err != nil {
		t.Fatalf("Failed to process second event: %v", err)
	}

	// Step 10: Verify alert was generated
	alerts, err := suite.RiskEngine.GetRiskAlerts()
	if err != nil {
		t.Fatalf("Failed to get risk alerts: %v", err)
	}

	// Find alert for our entity
	var foundAlert *models.RiskAlert
	for _, alert := range alerts {
		if alert.EntityID == updatedRiskObj.ID {
			foundAlert = alert
			break
		}
	}

	if foundAlert == nil {
		t.Fatal("Expected risk alert to be generated")
	}

	if foundAlert.TotalScore != 80 {
		t.Errorf("Expected alert score 80, got %d", foundAlert.TotalScore)
	}
	if foundAlert.Status != models.AlertStatusNew {
		t.Errorf("Expected alert status %s, got %s", models.AlertStatusNew, foundAlert.Status)
	}

	// Step 12: Verify detection relationships were loaded properly
	fullDetection, err := suite.DetectionRepo.GetDetection(detection.ID)
	if err != nil {
		t.Fatalf("Failed to get detection with relationships: %v", err)
	}

	if len(fullDetection.MitreTechniques) != 1 {
		t.Errorf("Expected 1 MITRE technique, got %d", len(fullDetection.MitreTechniques))
	} else if fullDetection.MitreTechniques[0].ID != mitretech.ID {
		t.Errorf("Expected MITRE technique %s, got %s", mitretech.ID, fullDetection.MitreTechniques[0].ID)
	}

	if len(fullDetection.DataSources) != 1 {
		t.Errorf("Expected 1 data source, got %d", len(fullDetection.DataSources))
	} else if fullDetection.DataSources[0].ID != dataSource.ID {
		t.Errorf("Expected data source %d, got %d", dataSource.ID, fullDetection.DataSources[0].ID)
	}
}

func TestMITRECoverageAnalysis(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.Cleanup()

	// Create multiple MITRE techniques across different tactics
	techniques := []*models.MitreTechnique{
		{ID: "T1059", Name: "Command and Scripting Interpreter", Tactic: "Execution", Domain: "Enterprise"},
		{ID: "T1055", Name: "Process Injection", Tactic: "Defense Evasion", Domain: "Enterprise"},
		{ID: "T1083", Name: "File and Directory Discovery", Tactic: "Discovery", Domain: "Enterprise"},
		{ID: "T1070", Name: "Indicator Removal on Host", Tactic: "Defense Evasion", Domain: "Enterprise"},
	}

	for _, tech := range techniques {
		err := suite.MitreRepo.CreateMitreTechnique(tech)
		if err != nil {
			t.Fatalf("Failed to create MITRE technique %s: %v", tech.ID, err)
		}
	}

	// Create detections that cover some techniques
	detections := []*models.Detection{
		{
			Name:       "PowerShell Detection",
			Status:     models.StatusProduction,
			Severity:   models.SeverityMedium,
			RiskPoints: 25,
			RiskObject: models.RiskObjectHost,
		},
		{
			Name:       "Process Injection Detection",
			Status:     models.StatusProduction,
			Severity:   models.SeverityHigh,
			RiskPoints: 40,
			RiskObject: models.RiskObjectHost,
		},
		{
			Name:       "File Discovery Detection",
			Status:     models.StatusDraft,
			Severity:   models.SeverityLow,
			RiskPoints: 15,
			RiskObject: models.RiskObjectUser,
		},
	}

	for _, det := range detections {
		err := suite.DetectionRepo.CreateDetection(det)
		if err != nil {
			t.Fatalf("Failed to create detection %s: %v", det.Name, err)
		}
	}

	// Link detections to techniques
	linkages := []struct {
		detectionIdx int
		techniqueIdx int
	}{
		{0, 0}, // PowerShell -> T1059
		{1, 1}, // Process Injection -> T1055
		{2, 2}, // File Discovery -> T1083
		{0, 1}, // PowerShell also covers T1055 (multi-mapping)
	}

	for _, link := range linkages {
		err := suite.DetectionRepo.AddMitreTechnique(
			detections[link.detectionIdx].ID,
			techniques[link.techniqueIdx].ID,
		)
		if err != nil {
			t.Fatalf("Failed to link detection to technique: %v", err)
		}
	}

	// Analyze coverage by tactic
	coverage, err := suite.MitreRepo.GetCoverageByTactic()
	if err != nil {
		t.Fatalf("Failed to get coverage by tactic: %v", err)
	}

	// Since GetCoverageByTactic returns coverage percentages, 
	// we just verify that tactics have some coverage
	expectedTactics := []string{"Execution", "Defense Evasion", "Discovery"}

	for _, tactic := range expectedTactics {
		if coverage[tactic] == 0 {
			t.Errorf("Expected some coverage for tactic %s, got 0", tactic)
		}
		t.Logf("Coverage for tactic %s: %.1f%%", tactic, coverage[tactic])
	}

	// Test detection retrieval by technique
	detectionsForT1055, err := suite.MitreRepo.GetDetectionsByTechnique("T1055")
	if err != nil {
		t.Fatalf("Failed to get detections for T1055: %v", err)
	}

	if len(detectionsForT1055) != 2 {
		t.Errorf("Expected 2 detections for T1055, got %d", len(detectionsForT1055))
	}
}

func TestDataSourceUtilizationAnalysis(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.Cleanup()

	// Create data sources
	dataSources := []*models.DataSource{
		{Name: "Windows Event Logs", Description: "Windows security events"},
		{Name: "Sysmon", Description: "System Monitor events"},
		{Name: "Network Logs", Description: "Network traffic logs"},
		{Name: "Cloud Audit Logs", Description: "Cloud service audit logs"},
	}

	for _, ds := range dataSources {
		err := suite.DataSourceRepo.CreateDataSource(ds)
		if err != nil {
			t.Fatalf("Failed to create data source %s: %v", ds.Name, err)
		}
	}

	// Create detections
	detections := []*models.Detection{
		{Name: "Logon Detection", Status: models.StatusProduction, Severity: models.SeverityMedium, RiskPoints: 20, RiskObject: models.RiskObjectUser},
		{Name: "Process Creation", Status: models.StatusProduction, Severity: models.SeverityLow, RiskPoints: 10, RiskObject: models.RiskObjectHost},
		{Name: "Network Connection", Status: models.StatusTest, Severity: models.SeverityMedium, RiskPoints: 15, RiskObject: models.RiskObjectIP},
		{Name: "Cloud Access", Status: models.StatusDraft, Severity: models.SeverityHigh, RiskPoints: 30, RiskObject: models.RiskObjectUser},
	}

	for _, det := range detections {
		err := suite.DetectionRepo.CreateDetection(det)
		if err != nil {
			t.Fatalf("Failed to create detection %s: %v", det.Name, err)
		}
	}

	// Link detections to data sources
	linkages := []struct {
		detectionIdx   int
		dataSourceIdx  int
	}{
		{0, 0}, // Logon -> Windows Event Logs
		{0, 1}, // Logon -> Sysmon (multi-source detection)
		{1, 1}, // Process Creation -> Sysmon
		{2, 2}, // Network Connection -> Network Logs
		{3, 3}, // Cloud Access -> Cloud Audit Logs
	}

	for _, link := range linkages {
		err := suite.DetectionRepo.AddDataSource(
			detections[link.detectionIdx].ID,
			dataSources[link.dataSourceIdx].ID,
		)
		if err != nil {
			t.Fatalf("Failed to link detection to data source: %v", err)
		}
	}

	// Analyze data source utilization
	utilization, err := suite.DataSourceRepo.GetDataSourceUtilization()
	if err != nil {
		t.Fatalf("Failed to get data source utilization: %v", err)
	}

	expectedUtilization := map[string]int{
		"Windows Event Logs": 1,
		"Sysmon":            2,
		"Network Logs":      1,
		"Cloud Audit Logs":  1,
	}

	if len(utilization) != len(expectedUtilization) {
		t.Errorf("Expected %d data sources in utilization, got %d", 
			len(expectedUtilization), len(utilization))
	}

	for dataSourceName, expectedCount := range expectedUtilization {
		actualCount, exists := utilization[dataSourceName]
		if !exists {
			t.Errorf("Expected data source %s not found in utilization", dataSourceName)
			continue
		}
		if actualCount != expectedCount {
			t.Errorf("Expected %d detections for %s, got %d", 
				expectedCount, dataSourceName, actualCount)
		}
	}
}

func TestFalsePositiveFeedbackLoop(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.Cleanup()

	// Create detection
	detection := &models.Detection{
		Name:       "Test Detection for FP",
		Status:     models.StatusProduction,
		Severity:   models.SeverityMedium,
		RiskPoints: 20,
		RiskObject: models.RiskObjectUser,
	}
	err := suite.DetectionRepo.CreateDetection(detection)
	if err != nil {
		t.Fatalf("Failed to create detection: %v", err)
	}

	// Process multiple events (risk object will be created automatically)
	riskObject := &models.RiskObject{
		EntityType:  models.EntityTypeUser,
		EntityValue: "fptest@example.com",
	}

	events := []*models.Event{
		{DetectionID: detection.ID, RiskPoints: 20, RawData: `{"event": 1}`, RiskObject: riskObject},
		{DetectionID: detection.ID, RiskPoints: 20, RawData: `{"event": 2}`, RiskObject: riskObject},
		{DetectionID: detection.ID, RiskPoints: 20, RawData: `{"event": 3}`, RiskObject: riskObject},
		{DetectionID: detection.ID, RiskPoints: 20, RawData: `{"event": 4}`, RiskObject: riskObject},
		{DetectionID: detection.ID, RiskPoints: 20, RawData: `{"event": 5}`, RiskObject: riskObject},
	}

	for i, event := range events {
		err = suite.RiskEngine.ProcessEvent(event)
		if err != nil {
			t.Fatalf("Failed to process event %d: %v", i+1, err)
		}
	}

	// Verify risk score accumulated
	updatedRiskObj, err := suite.RiskRepo.GetRiskObjectByEntity(models.EntityTypeUser, "fptest@example.com")
	if err != nil {
		t.Fatalf("Failed to get updated risk object: %v", err)
	}

	expectedScore := 20 * 5 // 100
	if updatedRiskObj.CurrentScore != expectedScore {
		t.Errorf("Expected risk score %d, got %d", expectedScore, updatedRiskObj.CurrentScore)
	}

	// Mark some events as false positives
	falsePositives := []struct {
		eventIdx int
		reason   string
		analyst  string
	}{
		{1, "Legitimate admin activity", "analyst1@example.com"},
		{3, "Scheduled maintenance task", "analyst2@example.com"},
	}

	for _, fp := range falsePositives {
		fpInfo := &models.FalsePositive{
			EventID:     events[fp.eventIdx].ID,
			Reason:      fp.reason,
			AnalystName: fp.analyst,
		}

		err = suite.RiskEngine.MarkEventAsFalsePositive(events[fp.eventIdx].ID, fpInfo)
		if err != nil {
			t.Fatalf("Failed to mark event as false positive: %v", err)
		}
	}

	// Verify events are marked as false positives
	for _, fp := range falsePositives {
		updatedEvent, err := suite.RiskRepo.GetEvent(events[fp.eventIdx].ID)
		if err != nil {
			t.Fatalf("Failed to get updated event: %v", err)
		}

		if !updatedEvent.IsFalsePositive {
			t.Errorf("Event %d should be marked as false positive", fp.eventIdx)
		}
	}

	// Calculate false positive rate
	fpRate, err := suite.DetectionRepo.GetFalsePositiveRate(detection.ID)
	if err != nil {
		t.Fatalf("Failed to get false positive rate: %v", err)
	}

	expectedFpRate := 2.0 / 5.0 // 2 FPs out of 5 total events = 0.4 (40%)
	if fpRate != expectedFpRate {
		t.Errorf("Expected FP rate %f, got %f", expectedFpRate, fpRate)
	}

	// Verify false positive count by checking events marked as false positive
	fpCount := 0
	for _, event := range events {
		updatedEvent, err := suite.RiskRepo.GetEvent(event.ID)
		if err != nil {
			continue
		}
		if updatedEvent.IsFalsePositive {
			fpCount++
		}
	}

	if fpCount != 2 {
		t.Errorf("Expected 2 false positives, got %d", fpCount)
	}
}

func TestRiskDecayAndMaintenance(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.Cleanup()

	// Create risk objects with different risk levels by inserting directly into database
	riskObjects := []struct {
		entityType  models.EntityType
		entityValue string
		score       int
	}{
		{models.EntityTypeUser, "lowrisk@example.com", 30},
		{models.EntityTypeHost, "mediumrisk-host", 75},
		{models.EntityTypeIP, "10.0.0.100", 150},
	}

	var riskObjIDs []int64
	for _, obj := range riskObjects {
		query := `INSERT INTO risk_objects (entity_type, entity_value, current_score) VALUES (?, ?, ?)`
		result, err := suite.DB.Exec(query, obj.entityType, obj.entityValue, obj.score)
		if err != nil {
			t.Fatalf("Failed to create risk object: %v", err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("Failed to get risk object ID: %v", err)
		}
		riskObjIDs = append(riskObjIDs, id)
	}

	// Get initial high-risk entities (threshold 75)
	highRiskBefore, err := suite.RiskRepo.ListHighRiskObjects(75)
	if err != nil {
		t.Fatalf("Failed to get high-risk objects: %v", err)
	}

	if len(highRiskBefore) != 2 {
		t.Errorf("Expected 2 high-risk objects initially, got %d", len(highRiskBefore))
	}

	// Apply decay (factor 0.1 means 10% reduction)
	err = suite.RiskEngine.DecayRiskScores()
	if err != nil {
		t.Fatalf("Failed to decay risk scores: %v", err)
	}

	// Verify scores were decayed
	expectedScores := []int{27, 68, 135} // 30*0.9, 75*0.9, 150*0.9 (rounded)

	for i := range riskObjects {
		updatedObj, err := suite.RiskRepo.GetRiskObject(riskObjIDs[i])
		if err != nil {
			t.Fatalf("Failed to get updated risk object: %v", err)
		}

		// Allow for small rounding differences
		if abs(updatedObj.CurrentScore-expectedScores[i]) > 1 {
			t.Errorf("Expected score around %d for object %d, got %d", 
				expectedScores[i], riskObjIDs[i], updatedObj.CurrentScore)
		}
	}

	// Verify high-risk count changed after decay
	highRiskAfter, err := suite.RiskRepo.ListHighRiskObjects(75)
	if err != nil {
		t.Fatalf("Failed to get high-risk objects after decay: %v", err)
	}

	if len(highRiskAfter) != 1 {
		t.Errorf("Expected 1 high-risk object after decay, got %d", len(highRiskAfter))
	}
}

// Helper function for absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func TestComplexEventContext(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.Cleanup()

	// Create detection
	detection := &models.Detection{
		Name:       "Complex Context Detection",
		Status:     models.StatusProduction,
		Severity:   models.SeverityHigh,
		RiskPoints: 40,
		RiskObject: models.RiskObjectHost,
	}
	err := suite.DetectionRepo.CreateDetection(detection)
	if err != nil {
		t.Fatalf("Failed to create detection: %v", err)
	}

	// Create event with complex JSON context
	contextData := map[string]interface{}{
		"source_ip":      "192.168.1.50",
		"destination_ip": "8.8.8.8",
		"port":           443,
		"protocol":       "HTTPS",
		"user_agent":     "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
		"request_headers": map[string]string{
			"Accept":     "application/json",
			"Authorization": "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...",
		},
		"response_code":   200,
		"bytes_sent":      1024,
		"bytes_received":  4096,
		"threat_intel": map[string]interface{}{
			"malicious_domains": []string{"evil.com", "malware.net"},
			"reputation_score": 85,
			"categories":       []string{"malware", "botnet"},
		},
	}

	contextJSON, err := json.Marshal(contextData)
	if err != nil {
		t.Fatalf("Failed to marshal context data: %v", err)
	}

	event := &models.Event{
		DetectionID: detection.ID,
		RiskPoints:  40,
		RawData:     `{"activity": "suspicious_network_connection", "timestamp": "2024-01-15T14:30:00Z"}`,
		Context:     string(contextJSON),
		RiskObject: &models.RiskObject{
			EntityType:  models.EntityTypeHost,
			EntityValue: "compromised-workstation",
		},
	}

	err = suite.RiskEngine.ProcessEvent(event)
	if err != nil {
		t.Fatalf("Failed to process event with complex context: %v", err)
	}

	// Retrieve and verify the stored event
	storedEvent, err := suite.RiskRepo.GetEvent(event.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve stored event: %v", err)
	}

	// Verify context was stored correctly
	var storedContext map[string]interface{}
	err = json.Unmarshal([]byte(storedEvent.Context), &storedContext)
	if err != nil {
		t.Fatalf("Failed to unmarshal stored context: %v", err)
	}

	// Verify some key context fields
	if storedContext["source_ip"] != "192.168.1.50" {
		t.Errorf("Expected source_ip '192.168.1.50', got %v", storedContext["source_ip"])
	}

	if storedContext["port"].(float64) != 443 {
		t.Errorf("Expected port 443, got %v", storedContext["port"])
	}

	threatIntel, ok := storedContext["threat_intel"].(map[string]interface{})
	if !ok {
		t.Error("threat_intel should be a nested object")
	} else {
		if threatIntel["reputation_score"].(float64) != 85 {
			t.Errorf("Expected reputation_score 85, got %v", threatIntel["reputation_score"])
		}
	}
}