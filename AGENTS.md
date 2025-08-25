# Repository Guidelines

## Project Structure & Modules
- `cmd/server`: HTTP server entrypoint. `cmd/import-mitre`: MITRE ATT&CK CSV importer.
- `internal/`: private app code (`detection/`, `risk/`, `mitre/`, `datasource/`).
- `pkg/`: reusable packages (`api/`, `database/`, `models/`, `middleware/`, `cache/`).
- `web/static/`: HTMX/Alpine/Chart.js assets. `configs/`: runtime config. `data/`: SQLite DB.
- Scripts live in `scripts/` (build/test). Default server: `localhost:8080`.

## Build, Test, and Development
- `make deps` → download modules. `make build` → compile `server` (CGO required for SQLite).
- `make run` or `make serve PORT=9000` → run locally. `make dev` → hot reload (air).
- Tests: `make test`, `make test-short`, `make test-integration`. Coverage: `make coverage`.
- Quality: `make fmt`, `make vet`, `make lint`, `make check`. Benchmarks: `make benchmark`.
- Docker: `make docker-run`, `make docker-stop`, `make docker-logs` (uses docker-compose).

## Coding Style & Naming
- Follow Go idioms; format with `gofmt`/`goimports`; keep functions focused; ≤100 char lines when reasonable.
- Document exported types and funcs. Use contextual error wrapping.
- Packages are short lowercase (e.g., `risk`); files lowercase with underscores.
- API uses stdlib (`net/http`, `html/template`), RESTful routes, and middleware for logging/auth.

## Frontend & UI Standards
- Prefer HTMX for AJAX/DOM updates, Alpine.js for state, Chart.js for charts; avoid heavy SPA frameworks.
- Progressive enhancement and responsive design. MITRE matrix: tactic columns, clear typography, accessible interactions.

## Testing Guidelines
- Co-locate tests as `*_test.go` (`TestXxx`, `BenchmarkXxx`). Use `make coverage` to generate `coverage.html`.
- Integration checks reside under `internal/` (ensure DB path exists). Scripts: see `scripts/test/`.

## Commit & Pull Requests
- Commits are imperative and concise (e.g., `fix mitre import`, `polish ui`).
- Branch naming: `feature/<slug>`, `fix/<slug>`, `chore/<slug>`.
- PRs include summary, rationale, linked issues, test steps, and UI screenshots when applicable.
- Before opening: run `make check`; update docs/samples; exclude binaries and secrets.

## Security & Configuration
- Configure via `configs/config.json`; SQLite lives in `data/riskmatrix.db` (auto-created). Never commit secrets or live data.
- If CGO toolchain is unavailable, prefer Docker. Reset local DB safely with `make reset-db`.
