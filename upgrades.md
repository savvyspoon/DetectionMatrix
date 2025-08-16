# UI/UX Improvement Plan - RiskMatrix

This document outlines the planned improvements to the RiskMatrix user interface and user experience, specifically focusing on false positive information display and enhanced event details in the risk objects view.

## Overview

The following improvements will enhance the usability and information density of the risk-objects-detail.html page, providing analysts with better visibility into event details and false positive tracking.

## Feature 1: False Positive Information Display

### Current State
- Events marked as false positive show only a badge indicating "False Positive" status
- No additional context about who marked it as FP, when, or why
- FalsePositive model contains: `id`, `event_id`, `reason`, `analyst_name`, `timestamp`

### Planned Implementation

#### 1.1 API Enhancement
**File:** `pkg/api/risk.go`
- Modify event retrieval endpoints to include FalsePositive details when `is_false_positive = true`
- Update `ListEvents`, `GetEvent`, `ListEventsByEntity` to join with `false_positives` table
- Include FalsePositive object in Event JSON response when applicable

**Changes Required:**
```go
// In Event model JSON response, add:
type Event struct {
    // ... existing fields ...
    FalsePositiveInfo *FalsePositive `json:"false_positive_info,omitempty"`
}
```

#### 1.2 Frontend Changes
**Files:** 
- `web/static/risk-objects-detail.html` (lines 201-202)
- `web/static/events.html` (lines 86-87)
- `web/static/events-detail.html` (lines 75-76)

**Implementation:**
- Add expandable tooltip/popover on FP badge showing:
  - Analyst name who marked it
  - Date/time marked
  - Reason provided
- Use Alpine.js `x-show` and `@click` handlers for toggle behavior
- Style with compact badge design consistent with existing UI

**Template Structure:**
```html
<div class="fp-badge-container" x-data="{ showFPDetails: false }">
    <span @click="showFPDetails = !showFPDetails" 
          class="badge badge-warning clickable-badge"
          x-show="event.is_false_positive">
        False Positive ⓘ
    </span>
    <div x-show="showFPDetails" class="fp-details-popup">
        <div><strong>Analyst:</strong> <span x-text="event.false_positive_info?.analyst_name"></span></div>
        <div><strong>Date:</strong> <span x-text="formatTimestamp(event.false_positive_info?.timestamp)"></span></div>
        <div><strong>Reason:</strong> <span x-text="event.false_positive_info?.reason"></span></div>
    </div>
</div>
```

## Feature 2: Score History Table Modifications

### Current State
**File:** `web/static/risk-objects-detail.html` (lines 101-142)

Current columns: Time, Score, Change, Event, Detection, Points

### Planned Changes

#### 2.1 Remove Event Column
- Remove truncated event column (line 127-128)
- This column currently shows `raw_data.substring(0, 30) + '...'` which is not useful

#### 2.2 Add Full Event Field Column
- Add new "Event Details" column showing the full raw_data
- Implement intelligent truncation with full text available on hover/expansion
- Position this column after "Change" column

**New Table Structure:**
```html
<table class="compact-table">
    <thead>
        <tr>
            <th>Time</th>
            <th>Score</th>
            <th>Change</th>
            <th>Event Details</th>  <!-- NEW: Full event field -->
            <th>Detection</th>
            <th>Points</th>
        </tr>
    </thead>
    <!-- ... tbody ... -->
</table>
```

## Feature 3: Accordion Row Expansion

### Current State
- Score History table rows are static
- Event details require navigation to separate page

### Planned Implementation

#### 3.1 Row Click Handler
**File:** `web/static/risk-objects-detail.html`
- Add click handler to table rows in Score History section
- Implement Alpine.js reactive state for expanded rows
- Track expanded state per row using event ID or timestamp

#### 3.2 Expandable Row Content
- Create accordion-style expansion below each row
- Show detailed event information similar to `events-detail.html` format
- Display: raw_data, context, and false positive information if applicable

**Implementation Structure:**
```html
<tbody x-data="{ expandedRows: new Set() }">
    <template x-for="historyItem in calculateRiskScoreHistory()" :key="historyItem.timestamp.getTime()">
        <template>
            <!-- Main Row -->
            <tr class="clickable-row" 
                @click="toggleRowExpansion(historyItem.timestamp.getTime())">
                <!-- existing columns -->
            </tr>
            
            <!-- Expanded Details Row -->
            <tr x-show="expandedRows.has(historyItem.timestamp.getTime())" 
                class="expanded-row">
                <td colspan="6" class="expanded-content">
                    <div class="event-details-expanded">
                        <div class="detail-section">
                            <h5>Raw Data</h5>
                            <pre class="json-display" x-text="formatJSON(historyItem.event?.raw_data)"></pre>
                        </div>
                        <div class="detail-section">
                            <h5>Context</h5>
                            <pre class="json-display" x-text="formatJSON(historyItem.event?.context)"></pre>
                        </div>
                        <div class="detail-section" x-show="historyItem.event?.is_false_positive">
                            <h5>False Positive Information</h5>
                            <div class="fp-info">
                                <p><strong>Analyst:</strong> <span x-text="historyItem.event?.false_positive_info?.analyst_name"></span></p>
                                <p><strong>Reason:</strong> <span x-text="historyItem.event?.false_positive_info?.reason"></span></p>
                                <p><strong>Marked:</strong> <span x-text="formatTimestamp(historyItem.event?.false_positive_info?.timestamp)"></span></p>
                            </div>
                        </div>
                    </div>
                </td>
            </tr>
        </template>
    </template>
</tbody>
```

#### 3.3 JavaScript Functions
**File:** `web/static/js/risk-objects-detail.js`

**New Functions Required:**
```javascript
// Row expansion management
toggleRowExpansion(rowId) {
    if (this.expandedRows.has(rowId)) {
        this.expandedRows.delete(rowId);
    } else {
        this.expandedRows.add(rowId);
    }
},

// JSON formatting for display
formatJSON(jsonString) {
    try {
        return JSON.stringify(JSON.parse(jsonString), null, 2);
    } catch (e) {
        return jsonString;
    }
}
```

#### 3.4 CSS Styling
**File:** `web/static/risk-objects-detail.html` (style section)

**New Styles:**
```css
/* Clickable row styling */
.clickable-row {
    cursor: pointer;
    transition: background-color 0.2s ease;
}

.clickable-row:hover {
    background-color: #e8f4f8 !important;
}

/* Expanded row content */
.expanded-row {
    background-color: #f8f9fa !important;
}

.expanded-content {
    padding: 1rem !important;
    border-top: 1px solid #dee2e6;
}

.event-details-expanded {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
}

.detail-section {
    background: white;
    padding: 0.75rem;
    border-radius: 4px;
    border: 1px solid #e9ecef;
}

.detail-section h5 {
    margin: 0 0 0.5rem 0;
    font-size: 0.9rem;
    color: var(--primary-color);
    font-weight: 600;
}

.json-display {
    background-color: #f8f9fa;
    border: 1px solid #e9ecef;
    border-radius: 3px;
    padding: 0.5rem;
    font-family: 'Courier New', monospace;
    font-size: 0.8rem;
    white-space: pre-wrap;
    word-break: break-all;
    margin: 0;
    max-height: 200px;
    overflow-y: auto;
}

.fp-info p {
    margin: 0.25rem 0;
    font-size: 0.85rem;
}

/* False positive badge enhancements */
.clickable-badge {
    cursor: pointer;
    position: relative;
}

.fp-details-popup {
    position: absolute;
    top: 100%;
    left: 0;
    background: white;
    border: 1px solid #ddd;
    border-radius: 4px;
    padding: 0.5rem;
    box-shadow: 0 2px 8px rgba(0,0,0,0.1);
    z-index: 1000;
    min-width: 200px;
    font-size: 0.8rem;
}

@media (max-width: 768px) {
    .event-details-expanded {
        grid-template-columns: 1fr;
    }
}
```

## Implementation Timeline

### Phase 1: Backend API Changes (1-2 days)
1. Update Event model to include FalsePositive relationship
2. Modify repository methods to fetch FP data
3. Update API endpoints to return FP information
4. Test API changes

### Phase 2: False Positive Display (2-3 days)
1. Implement FP information tooltips/popovers
2. Update all pages that display event status
3. Style and test FP detail display
4. Cross-browser testing

### Phase 3: Table Modifications (1 day)
1. Remove truncated Event column
2. Add full Event Details column
3. Update column headers and styling
4. Test table layout responsiveness

### Phase 4: Accordion Implementation (3-4 days)
1. Implement row click handlers
2. Create expandable row content structure
3. Add JSON formatting and display
4. Implement row expansion state management
5. Style expanded content sections
6. Mobile responsiveness testing

### Phase 5: Integration Testing (1-2 days)
1. End-to-end testing of all features
2. Performance testing with large datasets
3. Accessibility testing
4. Bug fixes and refinements

## Dependencies

- **Alpine.js**: Already included, used for reactive state management
- **Existing CSS**: Build upon current compact table styling
- **API**: Requires backend modifications to include FP data in responses

## Considerations

### Performance
- FP data adds to response size; consider lazy loading for large event lists
- Row expansion state should be memory-efficient for large tables

### Accessibility
- Ensure keyboard navigation works for expanded rows
- Add ARIA labels for screen readers
- Maintain focus management during row expansion

### Mobile Responsiveness
- Expanded content should stack vertically on mobile
- Consider collapsing some columns on smaller screens
- Touch-friendly click targets

## Testing Strategy

1. **Unit Tests**: Frontend state management functions
2. **Integration Tests**: API responses include FP data correctly
3. **UI Tests**: Row expansion/collapse behavior
4. **Cross-browser**: Chrome, Firefox, Safari, Edge
5. **Mobile Testing**: iOS Safari, Android Chrome
6. **Accessibility**: Screen reader compatibility, keyboard navigation

## Success Metrics

- Reduced time to view event details (no navigation required)
- Improved false positive tracking visibility
- Enhanced analyst workflow efficiency
- Positive user feedback on information density improvements

---

# Additional Features - Configuration Management & Enhanced Statistics

## Feature 4: Application Configuration Page

### Current State
- Configuration values are stored in `configs/config.json`
- No UI to view current configuration
- Configuration includes: server settings, database config, risk engine parameters, logging, and security settings

### Planned Implementation

#### 4.1 Backend API Development
**File:** `pkg/api/config.go` (new file)

**New Configuration API Endpoints:**
```go
type ConfigResponse struct {
    Server     ServerConfig     `json:"server"`
    Database   DatabaseConfig   `json:"database"`
    RiskEngine RiskEngineConfig `json:"risk_engine"`
    Logging    LoggingConfig    `json:"logging"`
    Security   SecurityConfig   `json:"security"`
    Runtime    RuntimeConfig    `json:"runtime"`
}

type RuntimeConfig struct {
    Version        string    `json:"version"`
    BuildTime      string    `json:"build_time"`
    Uptime         string    `json:"uptime"`
    StartTime      time.Time `json:"start_time"`
    GoVersion      string    `json:"go_version"`
    DatabasePath   string    `json:"database_path"`
    ConfigPath     string    `json:"config_path"`
}
```

**API Endpoints:**
- `GET /api/config` - Retrieve current configuration (sanitized)
- `GET /api/config/health` - System health information
- `GET /api/config/status` - Runtime status and metrics

#### 4.2 Frontend Configuration Page
**File:** `web/static/config.html` (new file)

**Page Structure:**
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <title>Configuration - RiskMatrix</title>
    <!-- existing CSS includes -->
</head>
<body>
    <main class="container" x-data="configPageData()" x-init="init()">
        <div class="card">
            <div class="card-header">
                <h2>Application Configuration</h2>
                <button @click="refreshConfig()" class="btn btn-secondary">Refresh</button>
            </div>
            
            <!-- Server Configuration Section -->
            <div class="config-section">
                <h3><i class="icon-server"></i> Server Configuration</h3>
                <div class="config-grid">
                    <div class="config-item">
                        <label>Host:</label>
                        <span class="config-value" x-text="config.server?.host"></span>
                    </div>
                    <div class="config-item">
                        <label>Port:</label>
                        <span class="config-value" x-text="config.server?.port"></span>
                    </div>
                    <!-- Additional server config items -->
                </div>
            </div>
            
            <!-- Database Configuration Section -->
            <div class="config-section">
                <h3><i class="icon-database"></i> Database Configuration</h3>
                <!-- Database config items -->
            </div>
            
            <!-- Risk Engine Configuration Section -->
            <div class="config-section">
                <h3><i class="icon-engine"></i> Risk Engine Configuration</h3>
                <!-- Risk engine config items -->
            </div>
            
            <!-- Runtime Information Section -->
            <div class="config-section">
                <h3><i class="icon-info"></i> Runtime Information</h3>
                <!-- Runtime info items -->
            </div>
        </div>
    </main>
</body>
</html>
```

#### 4.3 JavaScript Controller
**File:** `web/static/js/config.js` (new file)

```javascript
class ConfigAPI {
    static async fetchConfig() {
        return await APIUtils.fetchAPI('/api/config');
    }
    
    static async fetchHealthStatus() {
        return await APIUtils.fetchAPI('/api/config/health');
    }
}

function configPageData() {
    return {
        config: {},
        healthStatus: {},
        loading: true,
        lastUpdated: null,
        
        async init() {
            await this.loadConfigData();
        },
        
        async loadConfigData() {
            try {
                this.loading = true;
                const [configData, healthData] = await Promise.all([
                    ConfigAPI.fetchConfig(),
                    ConfigAPI.fetchHealthStatus()
                ]);
                
                this.config = configData;
                this.healthStatus = healthData;
                this.lastUpdated = new Date();
            } catch (error) {
                console.error('Error loading config:', error);
                UIUtils.showAlert('Failed to load configuration', 'error');
            } finally {
                this.loading = false;
            }
        },
        
        async refreshConfig() {
            await this.loadConfigData();
            UIUtils.showAlert('Configuration refreshed', 'success');
        }
    };
}
```

#### 4.4 Navigation Integration
**Files:** All navigation menus (header sections)

Add configuration link to navigation:
```html
<li><a href="config.html">Configuration</a></li>
```

#### 4.5 Security Considerations
- Sanitize sensitive values (passwords, tokens, keys)
- Read-only display (no editing capability)
- Admin-only access consideration for future enhancement

## Feature 5: Enhanced Risk Alert Statistics

### Current State
Dashboard shows:
- Total detections, active detections by status
- MITRE coverage statistics
- Basic risk metrics: high risk entities, active alerts, events today, false positives

### Planned Enhancements

#### 5.1 New Risk Alert Statistics
**Enhanced Dashboard Metrics:**

1. **Alert Status Breakdown**
   - New alerts (last 24h)
   - Alerts in triage
   - Alerts under investigation
   - Alerts on hold
   - Critical incidents
   - Closed alerts (last 7 days)

2. **Alert Trends**
   - Alert volume trend (7-day chart)
   - Average time to triage
   - Average time to resolution
   - Alert escalation rate

3. **Risk Distribution**
   - Alerts by risk score ranges (0-25, 26-50, 51-75, 76-100)
   - Top 5 entities by alert count
   - Alert types by detection category

4. **Performance Metrics**
   - Alert response time SLA compliance
   - False positive rate in alerts
   - Analyst workload distribution

#### 5.2 Backend API Enhancements
**File:** `pkg/api/statistics.go` (new file)

**New API Endpoints:**
```go
// GET /api/statistics/alerts
type AlertStatistics struct {
    StatusBreakdown    map[string]int    `json:"status_breakdown"`
    NewAlertsLast24h   int               `json:"new_alerts_last_24h"`
    AlertTrend7Days    []TrendPoint      `json:"alert_trend_7_days"`
    AvgTimeToTriage    float64           `json:"avg_time_to_triage_hours"`
    AvgTimeToResolution float64          `json:"avg_time_to_resolution_hours"`
    RiskScoreDistribution map[string]int `json:"risk_score_distribution"`
    TopEntitiesByAlerts []EntityAlert    `json:"top_entities_by_alerts"`
}

type TrendPoint struct {
    Date  string `json:"date"`
    Count int    `json:"count"`
}

type EntityAlert struct {
    EntityType  string `json:"entity_type"`
    EntityValue string `json:"entity_value"`
    AlertCount  int    `json:"alert_count"`
}
```

#### 5.3 Frontend Dashboard Enhancements
**File:** `web/static/index.html`

**New Statistics Section:**
```html
<!-- Enhanced Risk Alert Statistics -->
<div class="card" x-show="!loading">
    <div class="card-header">
        <h2>Risk Alert Analytics</h2>
        <button @click="refreshAlertStats()" class="btn btn-sm btn-secondary">Refresh</button>
    </div>
    
    <!-- Alert Status Grid -->
    <div class="alert-stats-grid">
        <div class="stat-card">
            <h4>New Alerts (24h)</h4>
            <span class="stat-value large" x-text="alertStats.newAlertsLast24h">0</span>
        </div>
        <div class="stat-card">
            <h4>In Triage</h4>
            <span class="stat-value" x-text="alertStats.statusBreakdown?.Triage || 0">0</span>
        </div>
        <div class="stat-card">
            <h4>Under Investigation</h4>
            <span class="stat-value" x-text="alertStats.statusBreakdown?.Investigation || 0">0</span>
        </div>
        <div class="stat-card critical">
            <h4>Critical Incidents</h4>
            <span class="stat-value" x-text="alertStats.statusBreakdown?.Incident || 0">0</span>
        </div>
    </div>
    
    <!-- Alert Trend Chart -->
    <div class="chart-section">
        <h4>7-Day Alert Trend</h4>
        <canvas id="alertTrendChart"></canvas>
    </div>
    
    <!-- Performance Metrics -->
    <div class="performance-metrics">
        <div class="metric">
            <label>Avg. Time to Triage:</label>
            <span x-text="formatHours(alertStats.avgTimeToTriage)"></span>
        </div>
        <div class="metric">
            <label>Avg. Time to Resolution:</label>
            <span x-text="formatHours(alertStats.avgTimeToResolution)"></span>
        </div>
    </div>
    
    <!-- Top Entities by Alert Count -->
    <div class="top-entities">
        <h4>Top Entities by Alert Count</h4>
        <table class="compact-table">
            <template x-for="entity in alertStats.topEntitiesByAlerts?.slice(0,5)" :key="entity.entity_value">
                <tr>
                    <td x-text="entity.entity_type"></td>
                    <td x-text="entity.entity_value"></td>
                    <td x-text="entity.alert_count"></td>
                </tr>
            </template>
        </table>
    </div>
</div>
```

#### 5.4 JavaScript Enhancements
**File:** `web/static/js/dashboard.js`

**Enhanced Dashboard API:**
```javascript
class DashboardAPI {
    // ... existing methods ...
    
    static async fetchAlertStatistics() {
        return await APIUtils.fetchAPI('/api/statistics/alerts');
    }
}

// Enhanced dashboard data
function dashboardData() {
    return {
        // ... existing properties ...
        alertStats: {
            statusBreakdown: {},
            newAlertsLast24h: 0,
            alertTrend7Days: [],
            avgTimeToTriage: 0,
            avgTimeToResolution: 0,
            riskScoreDistribution: {},
            topEntitiesByAlerts: []
        },
        alertTrendChart: null,
        
        // ... existing methods ...
        
        async loadAllData() {
            try {
                // ... existing data loading ...
                
                // Load enhanced alert statistics
                this.alertStats = await DashboardAPI.fetchAlertStatistics();
                this.renderAlertTrendChart();
                
                this.lastUpdated = new Date();
            } catch (error) {
                console.error('Error loading dashboard data:', error);
            }
        },
        
        renderAlertTrendChart() {
            const ctx = document.getElementById('alertTrendChart');
            if (!ctx || !this.alertStats.alertTrend7Days) return;
            
            if (this.alertTrendChart) {
                this.alertTrendChart.destroy();
            }
            
            this.alertTrendChart = new Chart(ctx, {
                type: 'line',
                data: {
                    labels: this.alertStats.alertTrend7Days.map(point => point.date),
                    datasets: [{
                        label: 'Alerts',
                        data: this.alertStats.alertTrend7Days.map(point => point.count),
                        borderColor: '#dc3545',
                        backgroundColor: 'rgba(220, 53, 69, 0.1)',
                        tension: 0.4
                    }]
                },
                options: {
                    responsive: true,
                    scales: {
                        y: {
                            beginAtZero: true,
                            precision: 0
                        }
                    }
                }
            });
        },
        
        formatHours(hours) {
            if (!hours) return 'N/A';
            if (hours < 1) return `${Math.round(hours * 60)}min`;
            return `${hours.toFixed(1)}h`;
        }
    };
}
```

#### 5.5 CSS Styling Enhancements
**File:** `web/static/css/main.css`

```css
/* Alert Statistics Grid */
.alert-stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1rem;
    margin-bottom: 1.5rem;
}

.stat-card {
    background: #f8f9fa;
    border-radius: 8px;
    padding: 1rem;
    text-align: center;
    border: 1px solid #e9ecef;
}

.stat-card.critical {
    background: #fff5f5;
    border-color: #fed7d7;
}

.stat-card h4 {
    margin: 0 0 0.5rem 0;
    font-size: 0.9rem;
    color: #666;
    font-weight: 600;
}

.stat-value.large {
    font-size: 2rem;
    font-weight: bold;
    color: #dc3545;
}

/* Configuration Page Styles */
.config-section {
    background: #f8f9fa;
    border-radius: 8px;
    padding: 1.5rem;
    margin-bottom: 1.5rem;
}

.config-section h3 {
    margin: 0 0 1rem 0;
    color: var(--primary-color);
    font-size: 1.1rem;
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.config-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 1rem;
}

.config-item {
    background: white;
    padding: 0.75rem;
    border-radius: 4px;
    border: 1px solid #e9ecef;
    display: flex;
    justify-content: space-between;
    align-items: center;
}

.config-item label {
    font-weight: 600;
    color: #333;
    margin: 0;
}

.config-value {
    font-family: monospace;
    background: #f1f3f4;
    padding: 0.25rem 0.5rem;
    border-radius: 3px;
    font-size: 0.9rem;
}

/* Performance metrics */
.performance-metrics {
    display: flex;
    gap: 2rem;
    margin: 1rem 0;
}

.metric {
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.metric label {
    font-weight: 600;
    color: #666;
}

/* Top entities table */
.top-entities {
    margin-top: 1.5rem;
}

.top-entities h4 {
    margin-bottom: 0.5rem;
    color: var(--primary-color);
}
```

## Implementation Timeline - Additional Features

### Phase 6: Configuration Page (3-4 days)
1. Create configuration API endpoints
2. Develop configuration page HTML/CSS
3. Implement JavaScript controller
4. Add navigation integration
5. Security sanitization implementation
6. Testing and documentation

### Phase 7: Enhanced Alert Statistics (4-5 days)
1. Design and implement statistics API endpoints
2. Database queries for new metrics
3. Frontend dashboard enhancements
4. Chart.js integration for trend visualization
5. Performance optimization for statistics queries
6. Mobile responsiveness for new components

### Phase 8: Integration & Testing (2-3 days)
1. End-to-end testing of new features
2. Performance testing with statistics queries
3. Mobile and cross-browser testing
4. Documentation updates
5. User acceptance testing

## Dependencies - Additional Features

- **Chart.js**: Already included for trend visualizations
- **Database Indexes**: May need optimization for statistics queries
- **Caching**: Consider implementing response caching for statistics endpoints

## Security Considerations - Additional Features

### Configuration Page
- Sanitize sensitive configuration values (passwords, keys)
- Consider role-based access for sensitive configuration data
- Audit logging for configuration access

### Statistics
- Rate limiting for statistics endpoints
- Data aggregation to prevent information leakage
- Performance monitoring to prevent DoS via expensive queries