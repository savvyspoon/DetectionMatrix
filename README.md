# RiskMatrix

RiskMatrix is a lightweight, risk-based detection management platform that allows security teams to track, manage, and improve security detections through lifecycle tracking, MITRE ATT&CK mapping, false positive feedback, and risk scoring.

## Features

- **Detection Knowledge Base**: Store detection metadata, logic, severity, data sources, etc.
- **MITRE ATT&CK Coverage Mapping**: Map detections to ATT&CK techniques/tactics with visual matrix view
- **Data Source Inventory**: Track which detections depend on which data sources
- **Risk-Based Analytics Engine**: Aggregate events by entity and generate risk alerts when thresholds are exceeded
- **Detection Lifecycle Management**: Track detections through their lifecycle stages
- **False Positive Tracking**: Flag false positives with reasons and compute false positive rates
- **Visual Analytics**: View MITRE matrix, detection efficacy scoring, log source contribution, and entity risk score trends

## Technology Stack

- **Backend**: Go
- **Database**: SQLite
- **Frontend**: HTML with HTMX for interactivity, Alpine.js for components, and Chart.js for visualizations

## Project Structure

```
riskmatrix/
├── cmd/                  # Application entry points
│   └── server/           # Main server application
├── pkg/                  # Public packages
│   ├── api/              # API handlers
│   ├── database/         # Database connection and schema
│   └── models/           # Domain models
├── internal/             # Private application code
│   ├── detection/        # Detection management
│   ├── mitre/            # MITRE ATT&CK integration
│   ├── datasource/       # Data source management
│   └── risk/             # Risk scoring engine
├── web/                  # Web assets
│   ├── static/           # Static files (CSS, JS)
│   └── templates/        # HTML templates
├── configs/              # Configuration files
├── data/                 # Data storage (SQLite database)
└── docs/                 # Documentation
```

## Getting Started

### Prerequisites

- Go 1.24 or higher
- SQLite
- C compiler:
  - **Windows**: MinGW-w64 or MSYS2 with GCC
  - **macOS**: Xcode Command Line Tools (install with `xcode-select --install`)
  - **Linux**: GCC (install with your distribution's package manager, e.g., `apt install build-essential` on Debian/Ubuntu)

### Installation

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/riskmatrix.git
   cd riskmatrix
   ```

2. Build the application:

   **Important**: RiskMatrix requires CGO to be enabled for SQLite support.

   **Windows**:
   ```
   .\scripts\build\build.ps1
   ```

   **Linux/macOS**:
   ```
   chmod +x build.sh
   ./build.sh
   ```

   **Manual build**:
   ```
   # Windows PowerShell
   $env:CGO_ENABLED=1; go build -o server.exe ./cmd/server

   # Linux/macOS
   CGO_ENABLED=1 go build -o server ./cmd/server
   ```

3. Run the application:
   ```
   # Windows
   .\server.exe

   # Linux/macOS
   ./server
   ```

### Docker Deployment

**Note**: Using Docker is recommended if you don't want to install a C compiler or if you're having issues with CGO. The Dockerfile already has all necessary dependencies configured.

1. Build and run using Docker Compose:
   ```
   docker-compose up -d
   ```

2. Access the application at http://localhost:8080

## API Documentation

RiskMatrix provides a RESTful API for interacting with the platform:

### Detections

- `GET /api/detections` - List all detections
- `GET /api/detections/{id}` - Get a specific detection
- `POST /api/detections` - Create a new detection
- `PUT /api/detections/{id}` - Update a detection
- `DELETE /api/detections/{id}` - Delete a detection

### MITRE ATT&CK

- `GET /api/mitre/techniques` - List all MITRE techniques
- `GET /api/mitre/techniques/{id}` - Get a specific MITRE technique
- `GET /api/mitre/coverage` - Get coverage statistics by tactic

### Data Sources

- `GET /api/datasources` - List all data sources
- `GET /api/datasources/{id}` - Get a specific data source
- `GET /api/datasources/utilization` - Get data source utilization metrics

### Risk Management

- `POST /api/events` - Process a security event
- `GET /api/risk/objects` - List risk objects
- `GET /api/risk/alerts` - List risk alerts
- `POST /api/events/{id}/false-positive` - Mark an event as a false positive

## Configuration

Configuration is stored in `configs/config.json` and includes settings for:

- Server configuration
- Database connection
- Risk engine parameters
- Logging
- Security settings

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Troubleshooting

### SQLite and CGO Issues

- **Error**: `Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work. This is a stub`
  - **Solution**: Make sure to build with CGO enabled as described in the installation instructions. Use the provided build scripts or set `CGO_ENABLED=1` manually.

- **Error**: `cgo: C compiler "gcc" not found`
  - **Solution**: Install a C compiler as mentioned in the prerequisites section. Alternatively, use Docker which already has all dependencies configured.

- **Windows-specific**: If you're having issues with MinGW or MSYS2, consider using Docker or WSL (Windows Subsystem for Linux) for a smoother experience.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.