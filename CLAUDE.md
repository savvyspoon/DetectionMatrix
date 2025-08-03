# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Guidelines & Standards

### Code Style & Quality
- Follow standard Go coding conventions and idioms
- Use `gofmt` or `goimports` to format code
- Write unit tests for all packages - run with `go test ./...`
- Document all exported functions, types, and constants
- Keep functions small and focused on single responsibility
- Use meaningful variable and function names
- Keep line length reasonable (under 100 characters when possible)

### Web Development Standards
- Use standard library packages (`net/http`, `html/template`) whenever possible
- Avoid third-party web frameworks (Gin, Echo, Fiber)
- Implement proper middleware chains for logging, authentication
- Follow RESTful API design principles with proper HTTP status codes
- Implement appropriate error handling and logging for web requests
- Consider security best practices (HTTPS, input validation, CSRF protection)

### Frontend Development Standards
- Minimize JavaScript usage to keep application lightweight
- Preferred libraries for interactivity:
  - **HTMX** for AJAX, CSS transitions, DOM updates without JavaScript
  - **Alpine.js** for lightweight component-based interactivity
  - **Chart.js** for data visualization and interactive charts
  - **Vanilla JavaScript** for simple DOM manipulations
- Avoid heavy frameworks (React, Angular, Vue) unless absolutely necessary
- Use progressive enhancement for JavaScript-disabled users
- Implement responsive design for various device sizes

### MITRE ATT&CK Matrix UI Guidelines
- **Layout**: Horizontal grid with tactic columns, responsive design
- **Colors**: Dark tactic headers, white/light gray technique blocks
- **Typography**: System fonts, 13-16px sizes, proper hierarchy
- **Interactions**: Hover tooltips, keyboard navigation, WCAG 2.1 AA compliance
- **Navigation**: Persistent top nav, breadcrumbs, back-to-matrix links

## Common Development Commands

### Building the Application
- **Windows**: `.\scripts\build\build.ps1` or `$env:CGO_ENABLED=1; go build -o server.exe ./cmd/server`
- **Linux/macOS**: `./build.sh` or `CGO_ENABLED=1 go build -o server ./cmd/server`
- **Docker**: `docker-compose up -d`

**Important**: CGO must be enabled for SQLite support. All build commands ensure `CGO_ENABLED=1`.

### Running Tests
Test scripts are located in `scripts/test/` directory:
- `scripts/test/simple-test.ps1` - Basic functionality test
- `scripts/test/comprehensive-alert-test.ps1` - Complete alert workflow test
- `scripts/test/test-contributing-events.ps1` - Contributing events functionality test

### Running the Server
```bash
# Windows
.\server.exe

# Linux/macOS  
./server

# With custom database path
./server -db /path/to/custom.db -addr :9000
```

## Architecture Overview

RiskMatrix is a Go-based security detection management platform with the following architecture:

### Core Components
- **cmd/server/**: Main application entry point
- **pkg/api/**: HTTP API handlers and server setup
- **pkg/database/**: SQLite database connection and schema management
- **pkg/models/**: Domain models (Detection, Risk, DataSource, MITRE)
- **internal/**: Private application logic organized by domain:
  - `detection/`: Detection lifecycle management
  - `risk/`: Risk scoring engine and analytics
  - `mitre/`: MITRE ATT&CK integration
  - `datasource/`: Data source inventory management

### Key Architecture Patterns
- Repository pattern for data access (each domain has its own repository)
- Dependency injection through constructor functions
- HTTP handlers organized by domain with dedicated handler structs
- SQLite with foreign key constraints and proper indexing
- Risk engine runs as background goroutine for score decay

### Database Schema
The application uses SQLite with the following core entities:
- `detections`: Security detection rules with lifecycle tracking
- `mitre_techniques`: MITRE ATT&CK technique mappings
- `risk_objects`: Entities that accumulate risk (users, hosts, IPs)
- `events`: Detection triggers linked to risk objects
- `risk_alerts`: Generated when risk thresholds are exceeded
- `false_positives`: False positive tracking for detection tuning

### API Structure
RESTful API organized by domain:
- `/api/detections/*` - Detection management
- `/api/mitre/*` - MITRE ATT&CK techniques and coverage
- `/api/datasources/*` - Data source inventory
- `/api/events/*` - Event processing and false positive marking
- `/api/risk/*` - Risk objects and alerts

### Frontend
- Static HTML with HTMX for dynamic interactions
- Alpine.js for component state management  
- Chart.js for visualizations
- No build process required - served directly from `web/static/`

## Key Configuration
- Server config: `configs/config.json`
- Database: SQLite at `data/riskmatrix.db` (auto-created)
- Web assets: `web/static/` directory
- Default server: `localhost:8080`

## Development Notes
- Go 1.24+ required
- C compiler required for SQLite CGO support
- Use Docker if C compiler setup is problematic
- Risk engine includes automatic score decay mechanism
- All repositories include comprehensive test coverage

## Project Context & Business Logic

### Detection Efficacy Model
Based on Elastic DEBMM framework:
- **Coverage Score**: Maps to relevant ATT&CK technique
- **Accuracy Score**: False positive rate over time
- **Enrichment Score**: Provides context (host, user, process)
- **Actionability Score**: Links to playbook

### Detection Lifecycle Stages
Following Detection Development Lifecycle (DDL):
- **Idea**: Detection requested by hunter or during exercise
- **Draft**: Detection logic authored
- **Test**: Deployed to test environment, results reviewed
- **Production**: Active in SIEM or alert pipeline
- **Retired**: Disabled or superseded by new detection

### Risk Scoring System
- Events aggregate by entity (user, host, IP) with configurable risk points
- Risk alerts generated when entity score exceeds threshold
- Automatic score decay over time (configurable interval/factor)
- False positive feedback loop for detection tuning

### Core Entity Relationships
- Detections map to multiple MITRE techniques and data sources
- Events link detections to risk objects with timestamps
- Risk alerts generated per risk object when thresholds exceeded
- False positives tracked with analyst feedback and reasons