# DetectionMatrix

DetectionMatrix (formerly RiskMatrix) is a lightweight, risk-based detection management platform that allows security teams to track, manage, and improve security detections through lifecycle tracking, MITRE ATT&CK mapping, false positive feedback, and risk scoring.

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
DetectionMatrix/
├── cmd/                  # Application entry points
│   ├── server/           # Main server application
│   └── import-mitre/     # MITRE ATT&CK data importer
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
├── scripts/              # Build and test scripts
│   ├── build/            # Build scripts for different platforms
│   └── test/             # Test scripts
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
   git clone https://github.com/yourusername/detectionmatrix.git
   cd DetectionMatrix
   ```

2. Build the application:

   **Important**: DetectionMatrix requires CGO to be enabled for SQLite support.

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

#### Quick Start with Docker Compose

1. Build and run using Docker Compose:
   ```
   docker-compose up -d
   ```

2. Access the application at http://localhost:8080

#### Building Docker Image Manually

1. Use the provided build script (supports ARM64/Apple Silicon):
   ```bash
   # Auto-detect platform and build
   ./docker-build.sh
   
   # Build for specific platform
   ./docker-build.sh -p linux/arm64
   
   # Build multi-platform image
   ./docker-build.sh -p multi
   
   # Set up Docker buildx for multi-platform builds
   ./docker-build.sh --setup-buildx
   ```

2. Run the container:
   ```bash
   docker run -p 8080:8080 riskmatrix:latest
   ```

## API Documentation

DetectionMatrix provides a RESTful API for interacting with the platform:

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

- Server configuration (port, address)
- Database connection (SQLite path)
- Risk engine parameters (decay interval, factor, thresholds)
- Logging levels and output
- Security settings

### Command Line Options

```bash
# Run with custom database path
./server -db /path/to/custom.db

# Run on different port
./server -addr :9000

# Combine options
./server -db /data/detections.db -addr :8090
```

## Testing

The project includes comprehensive test scripts in the `scripts/test/` directory:

### Test Scripts

```bash
# Run basic functionality test
./scripts/test/simple-test.ps1

# Run comprehensive alert workflow test
./scripts/test/comprehensive-alert-test.ps1

# Test contributing events functionality
./scripts/test/test-contributing-events.ps1

# Run Go unit tests
go test ./...
```

### Periodic Event Generator

The `periodic-events.sh` script generates test events at regular intervals to simulate real-world detection activity and test the risk scoring system:

```bash
# Usage
./scripts/test/periodic-events.sh [server_url] [interval_seconds] [events_per_interval] [max_iterations]

# Examples
# Send 5 events every 30 seconds for 100 iterations
./scripts/test/periodic-events.sh http://localhost:8080 30 5 100

# Quick test: 3 events every 10 seconds for 10 iterations
./scripts/test/periodic-events.sh http://localhost:8080 10 3 10

# Debug mode to see request details
DEBUG=1 ./scripts/test/periodic-events.sh http://localhost:8080 5 2 5
```

**Features:**
- Generates events with random severity levels (weighted towards low/medium)
- Cycles through various risk objects (users, hosts, IPs)
- Shows real-time success/failure indicators
- Displays active risk alerts during execution
- Tracks overall success rate and statistics

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Troubleshooting

### SQLite and CGO Issues

- **Error**: `Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work. This is a stub`
  - **Solution**: Make sure to build with CGO enabled as described in the installation instructions. Use the provided build scripts or set `CGO_ENABLED=1` manually.

- **Error**: `cgo: C compiler "gcc" not found`
  - **Solution**: Install a C compiler as mentioned in the prerequisites section. Alternatively, use Docker which already has all dependencies configured.

- **Windows-specific**: If you're having issues with MinGW or MSYS2, consider using Docker or WSL (Windows Subsystem for Linux) for a smoother experience.

- **Docker build stalling**: If the Docker build hangs, ensure you have sufficient resources allocated to Docker and try building with `--no-cache` flag.

- **Apple Silicon/ARM64**: Use the `./docker-build.sh` script which automatically detects and handles ARM64 architecture.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.