# RiskMatrix Project Guidelines

## Project Overview
RiskMatrix is a Go-based application designed to help assess, visualize, and manage risks in various contexts. The project is currently in its early development stages with only the Go module definition set up.

## Project Structure
* The project is organized as a Go module named `riskmatrix`
* Go version 1.24 is used for development
* Source code will be organized following standard Go project layout:
  * `/cmd` - Main applications
  * `/pkg` - Library code that can be used by external applications
  * `/internal` - Private application and library code
  * `/api` - API definitions and documentation
  * `/web` - Web assets if applicable
  * `/configs` - Configuration file templates or default configs
  * `/docs` - Design and user documentation

## Development Guidelines
* Follow standard Go coding conventions and idioms
* Use Go modules for dependency management
* Write unit tests for all packages
* Document all exported functions, types, and constants
* Keep functions small and focused on a single responsibility

## Testing
* Write tests for all new functionality
* Run tests before submitting changes with `go test ./...`
* Aim for high test coverage, especially for core functionality

## Building
* The project can be built using standard Go tools: `go build ./...`
* Before submitting changes, ensure the project builds without errors

## Code Style
* Follow the official Go style guide
* Use `gofmt` or `goimports` to format code
* Keep line length reasonable (under 100 characters when possible)
* Use meaningful variable and function names

## Web Development Guidelines
* Use standard library packages (`net/http`, `html/template`, etc.) whenever possible
* Avoid using third-party web frameworks (like Gin, Echo, or Fiber)
* Implement proper middleware chains for common functionality (logging, authentication, etc.)
* Follow RESTful API design principles
* Use proper HTTP status codes and content types
* Implement appropriate error handling and logging for web requests
* Consider security best practices (HTTPS, input validation, CSRF protection, etc.)

## Frontend Development Guidelines
* Minimize JavaScript usage and dependencies to keep the application lightweight
* For interactive elements, prefer the following libraries:
  * HTMX for AJAX, CSS transitions, and DOM updates without writing JavaScript
  * Alpine.js for lightweight component-based interactivity when needed
  * Chart.js for data visualization and interactive charts
  * Vanilla JavaScript for simple DOM manipulations
* Avoid heavy frontend frameworks (React, Angular, Vue) unless absolutely necessary
* Use progressive enhancement to ensure functionality for users with JavaScript disabled
* Implement responsive design to support various device sizes and orientations

## Frontend Style Guide
* HTML Structure:
  * Use semantic HTML5 elements (`header`, `nav`, `main`, `section`, `article`, `footer`)
  * Maintain proper heading hierarchy (h1-h6)
  * Include appropriate ARIA attributes for accessibility
* CSS Guidelines:
  * Use a simple, consistent naming convention (BEM recommended)
  * Implement a color system with variables for consistency
  * Define a typography system with limited font families and sizes
  * Create reusable utility classes for common patterns
  * Organize CSS with a mobile-first approach
* JavaScript Style:
  * Write clean, modular JavaScript following modern ES6+ practices
  * Use event delegation for handling multiple similar elements
  * Implement proper error handling for all asynchronous operations
  * Document complex functions and components
* Performance Considerations:
  * Optimize images and assets for web delivery
  * Minimize HTTP requests by bundling assets when appropriate
  * Implement lazy loading for non-critical resources
  * Consider using CSS for animations instead of JavaScript when possible

## MITRE ATT&CK Matrix Page Guidelines

### 1. Layout & Grid System
* Grid Layout: Use a horizontal grid where each column represents a Tactic.
* Responsive Design:
  * Desktop: Full-width grid with fixed-height tactic headers and scrollable technique lists.
  * Tablet: Collapse sub-techniques under parent technique in an accordion-style UI.
  * Mobile: Convert grid to stacked accordion view with collapsible tactic sections.
* Fixed Header: Keep the matrix title and tactic headers pinned during vertical scroll.
* Scrollbars:
  * Enable horizontal scroll for full matrix view on smaller screens.
  * Preserve scroll position when returning from a detail view.

### 2. Color & Visual Hierarchy
* Tactic Headers:
  * Background: Dark gray or navy blue
  * Text: White, bold, small-caps or uppercase
* Technique Blocks:
  * Background: White or very light gray
  * Border: 1px solid muted gray
  * Hover State: Light blue highlight or box-shadow
  * Clicked/Active State: Bold border or shaded background
* Sub-techniques:
  * Indented or nested beneath parent technique
  * Visual cue (e.g., vertical line or tree node)
* Theme:
  * Default to light mode
  * Optional: Dark mode toggle for accessibility

### 3. Typography
* Font Family: System stack or clean sans-serif (e.g., Segoe UI, Roboto, Helvetica Neue)
* Font Sizes:
  * Tactic Headers: 14–16px, uppercase, bold
  * Techniques: 13–14px, regular weight
  * Sub-techniques: 12–13px, slightly indented
* Line Height: 1.4–1.6 for readability

### 4. Interaction & Behavior
* Hover: Highlight technique with tooltip showing short description
* Click: Opens detailed technique page in the same tab
* Accessibility:
  * All interactive elements should be keyboard navigable
  * ARIA labels for screen readers
  * Ensure high contrast and WCAG 2.1 AA compliance
* Tooltip:
  * Appears on hover or focus
  * Delay: ~250ms
  * Includes technique name and short description
* Expandable Sub-techniques:
  * Use chevron or plus/minus icon
  * Expand on click or Enter key

### 5. Navigation
* Primary Navigation:
  * Persistent top nav bar with links: Matrices, Techniques, Mitigations, Groups, Software, etc.
* Breadcrumbs:
  * Display below nav bar when viewing technique details
  * Format: Home > Enterprise Matrix > Execution > Command and Scripting Interpreter (T1059)
* Back to Matrix:
  * Include a fixed-position "Return to Matrix" link or button when in detail views

### 6. Detail Pages (Technique View)
* Layout:
  * Left-aligned content with TOC-style sidebar (optional)
  * Use consistent section headers: Description, Examples, Mitigations, Detections, Data Sources, etc.
* Highlight ID and Name:
  * Technique ID: Upper left, monospaced or bold (e.g., T1059)
  * Technique Name: Large, bold, descriptive
* Tag Styling:
  * Tags like tactic, platform, permissions required should be pill-style with muted borders

### 7. Visual Consistency
* Icons:
  * Use consistent vector icons (e.g., chevrons, info, expand/collapse)
  * Ensure clarity at small sizes
* Whitespace:
  * Apply ample spacing between tactic columns and between techniques
  * Avoid visual clutter in large matrix views
* Content Alignment:
  * Align all tactic headers and technique entries to a common vertical baseline

## 🛠️ Detection Management Platform – Software Planning Document

### 1. Project Overview

**Purpose:**
To build a lightweight, risk-based detection management platform that allows SOC analysts, threat hunters, and detection engineers to track, manage, and improve security detections through lifecycle tracking, ATT&CK mapping, false positive feedback, and risk scoring.

**Primary Functions:**

* Visualize detection coverage across MITRE ATT&CK
* Track relationships between detections, data sources, MITRE techniques, events, and risk objects
* Score and aggregate risk events to form higher fidelity alerts
* Manage detection lifecycle and efficacy
* Provide analytics for playbooks, log source utility, and detection health

### 2. Target Users & Use Cases

| Role               | Use Case                                                                |
| ------------------ | ----------------------------------------------------------------------- |
| SOC Analyst        | Investigate alerts, view detection context, mark false positives        |
| Threat Hunter      | Identify detection gaps, MITRE coverage, log source visibility analysis |
| Detection Engineer | Manage detection development lifecycle, analyze false positives         |
| SOC Manager        | View metrics on detection coverage, false positives, and rule health    |

### 3. Core Features

#### ✅ Detection Knowledge Base

* Store detection metadata, logic, severity, data sources, etc.
* Supports DML model: *Detection Source → Detection Logic → Alert Metadata*

#### ✅ MITRE ATT&CK Coverage Mapping

* Map detections to ATT&CK techniques/tactics
* Visual ATT&CK matrix view (similar to ATT&CK Navigator)
* Link MITRE → detection → data sources

#### ✅ Data Source Inventory

* Track which detections depend on which data sources
* View log source utility metrics
* Visual coverage view per log source (e.g., sysmon, firewall, cloudtrail)

#### ✅ Risk-Based Analytics Engine

* Aggregate events by entity (user, host, IP)
* Track entity risk score over time
* Generate risk alerts when thresholds exceeded
* Batch ingestion of events from SIEMs

#### ✅ Detection Lifecycle Management

* Lifecycle stages: `Idea → Draft → Test → Production → Retired`
* Tracks authorship, versioning, tuning notes

#### ✅ False Positive Tracking

* Analysts can flag false positives with reasons
* False positive rate computed per detection
* Feedback loop for tuning/removal

#### ✅ Visual Analytics

* MITRE matrix
* Detection efficacy scoring (based on DEBMM and FP rate)
* Log source contribution views
* Entity risk score trends

### 4. Technical Architecture

#### 📦 Backend

* **Language:** Go
* **Database:** SQLite (support for multi-table relational schemas)
* **Risk Engine:** Go batch processor (can run on cron)
* **Log Ingestion:** Accept JSON payloads from SIEM via REST endpoint or local file drop

#### 🌐 Frontend

* **Framework:** Pure HTML with [HTMX](https://htmx.org/) for interactivity
* **Charts:** Chart.js or D3.js for MITRE & risk dashboards
* **No SPA frameworks** (e.g., React, Vue)

#### 🔗 API Layer

* RESTful JSON-based endpoints for all entities (detections, events, risks)
* Designed for use by SIEMs or scripts (e.g., pushing events)

### 5. Key Components

| Component          | Description                                          |
| ------------------ | ---------------------------------------------------- |
| `detections`       | Detection metadata, status, MITRE mappings, FP count |
| `mitre_techniques` | Static list of techniques and tactics from ATT&CK   |
| `data_sources`     | List of log sources (EDR, sysmon, firewall, etc.)    |
| `events`           | Detection events with entity, timestamp, risk points |
| `risk_objects`     | Aggregated risk scores by entity (user/host/IP)      |
| `risk_alerts`      | Generated when object risk score exceeds threshold   |
| `false_positives`  | Analyst-logged FPs against events                    |

### 6. Detection Efficacy Model

Based on [Elastic DEBMM](https://www.elastic.co/security-labs/elastic-releases-debmm):

* Coverage Score: Does it map to a relevant ATT&CK technique?
* Accuracy Score: FP rate over time
* Enrichment Score: Does detection provide context (host, user, process)?
* Actionability Score: Does it link to a playbook?

Efficacy displayed visually on detection dashboard using a compact bar or radar chart.

### 7. Lifecycle Mapping

Detection development follows the [Detection Development Lifecycle (DDL)](https://medium.com/snowflake/detection-development-lifecycle-af166fffb3bc):

| Stage      | Description                                              |
| ---------- | -------------------------------------------------------- |
| Idea       | Detection requested (e.g., by hunter or during exercise) |
| Draft      | Detection logic authored                                 |
| Test       | Deployed to test environment, results reviewed           |
| Production | Active in SIEM or alert pipeline                         |
| Retired    | Disabled or superseded by new detection                  |

Tracked in the UI with state-change timestamps.

### 8. Metrics and Dashboards

* **MITRE Heatmap:** Technique coverage by detection count or data source
* **Log Source Utility:** Number of detections per log source
* **Detection Health:** Efficacy + FP Rate + Last Updated
* **Risk Trends:** Entity risk score graphs over time
* **Development Pipeline:** Detections per lifecycle stage

### 9. Entity Relationship Diagram (ERD)

```plaintext
+------------------+
|   detections     |
+------------------+
| id (PK)          |
| name             |
| description      |
| status           | -- e.g. idea, draft, test, production, retired
| severity         |
| risk_points      |
| playbook_link    |
| created_at       |
| updated_at       |
+------------------+

+---------------------------+
|  detection_mitre_map      |
+---------------------------+
| detection_id (FK)         |
| mitre_id (FK)             |
+---------------------------+

+----------------------+
|   mitre_techniques   |
+----------------------+
| id (PK)              | -- e.g. T1059.001
| tactic               | -- e.g. Execution
| name                 |
| description          |
+----------------------+

+-------------------------+
|   detection_datasource  |
+-------------------------+
| detection_id (FK)       |
| datasource_id (FK)      |
+-------------------------+

+-------------------+
|   data_sources    |
+-------------------+
| id (PK)           |
| name              | -- e.g. sysmon, cloudtrail
| description       |
| log_format        |
+-------------------+

+------------------+
|   events         |
+------------------+
| id (PK)          |
| detection_id (FK)|
| entity_id (FK)   |
| timestamp        |
| raw_data         | -- optional short blob or reference
| risk_points      |
| is_false_positive|
+------------------+

+-----------------+
|  risk_objects   |
+-----------------+
| id (PK)         |
| entity_type     | -- user, host, IP
| entity_value    |
| current_score   |
| last_seen       |
+-----------------+

+------------------+
|  risk_alerts     |
+------------------+
| id (PK)          |
| entity_id (FK)   |
| triggered_at     |
| total_score      |
+------------------+

+------------------------+
|  false_positives       |
+------------------------+
| id (PK)                |
| event_id (FK)          |
| reason                 |
| analyst_name           |
| timestamp              |
+------------------------+
```

#### Relationships Summary

* A **detection** can map to multiple **MITRE techniques** and **data sources**
* A **detection** generates multiple **events**
* Each **event** is linked to a **risk object** (user, host, IP) and possibly marked as a **false positive**
* **Risk alerts** are generated per **risk object** once cumulative score exceeds a threshold