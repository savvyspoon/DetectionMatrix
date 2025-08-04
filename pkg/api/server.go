package api

import (
	"log"
	"net/http"

	"riskmatrix/internal/datasource"
	"riskmatrix/internal/detection"
	"riskmatrix/internal/mitre"
	"riskmatrix/internal/risk"
	"riskmatrix/pkg/database"
)

// Server represents the API server
type Server struct {
	db             *database.DB
	detectionRepo  *detection.Repository
	mitreRepo      *mitre.Repository
	dataSourceRepo *datasource.Repository
	riskRepo       *risk.Repository
	riskEngine     *risk.Engine
	router         *http.ServeMux
}

// NewServer creates a new API server
func NewServer(db *database.DB) *Server {
	// Create repositories
	detectionRepo := detection.NewRepository(db)
	mitreRepo := mitre.NewRepository(db)
	dataSourceRepo := datasource.NewRepository(db)
	riskRepo := risk.NewRepository(db)

	// Create risk engine
	riskEngine := risk.NewEngine(db, risk.DefaultConfig())

	// Create server
	server := &Server{
		db:             db,
		detectionRepo:  detectionRepo,
		mitreRepo:      mitreRepo,
		dataSourceRepo: dataSourceRepo,
		riskRepo:       riskRepo,
		riskEngine:     riskEngine,
		router:         http.NewServeMux(),
	}

	// Set up routes
	server.setupRoutes()

	return server
}

// setupRoutes sets up the API routes
func (s *Server) setupRoutes() {
	// Create handlers
	detectionHandler := NewDetectionHandler(s.detectionRepo)
	mitreHandler := NewMitreHandler(s.mitreRepo)
	dataSourceHandler := NewDataSourceHandler(s.dataSourceRepo)
	riskHandler := NewRiskHandler(s.riskEngine, s.riskRepo)

	// Static files
	s.router.Handle("/", http.FileServer(http.Dir("web/static")))
	
	// Convenience routes for common pages
	s.router.HandleFunc("GET /alerts", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/static/risk-alerts.html")
	})
	
	// API routes - Detections
	s.router.HandleFunc("GET /api/detections", detectionHandler.ListDetections)
	s.router.HandleFunc("POST /api/detections", detectionHandler.CreateDetection)
	s.router.HandleFunc("GET /api/detections/count", detectionHandler.GetDetectionCount)
	s.router.HandleFunc("GET /api/detections/count/status", detectionHandler.GetDetectionCountByStatus)
	s.router.HandleFunc("GET /api/detections/{id}", detectionHandler.GetDetection)
	s.router.HandleFunc("PUT /api/detections/{id}", detectionHandler.UpdateDetection)
	s.router.HandleFunc("DELETE /api/detections/{id}", detectionHandler.DeleteDetection)
	s.router.HandleFunc("GET /api/detections/{id}/fp-rate", detectionHandler.GetFalsePositiveRate)
	s.router.HandleFunc("GET /api/detections/{id}/events/count/30days", detectionHandler.GetEventCountLast30Days)
	s.router.HandleFunc("GET /api/detections/{id}/false-positives/count/30days", detectionHandler.GetFalsePositivesLast30Days)
	s.router.HandleFunc("POST /api/detections/{id}/mitre/{technique_id}", detectionHandler.AddMitreTechnique)
	s.router.HandleFunc("DELETE /api/detections/{id}/mitre/{technique_id}", detectionHandler.RemoveMitreTechnique)
	s.router.HandleFunc("POST /api/detections/{id}/datasource/{datasource_id}", detectionHandler.AddDataSource)
	s.router.HandleFunc("DELETE /api/detections/{id}/datasource/{datasource_id}", detectionHandler.RemoveDataSource)
	
	// API routes - MITRE
	s.router.HandleFunc("GET /api/mitre/techniques", mitreHandler.ListMitreTechniques)
	s.router.HandleFunc("POST /api/mitre/techniques", mitreHandler.CreateMitreTechnique)
	s.router.HandleFunc("GET /api/mitre/techniques/{id}", mitreHandler.GetMitreTechnique)
	s.router.HandleFunc("PUT /api/mitre/techniques/{id}", mitreHandler.UpdateMitreTechnique)
	s.router.HandleFunc("DELETE /api/mitre/techniques/{id}", mitreHandler.DeleteMitreTechnique)
	s.router.HandleFunc("GET /api/mitre/techniques/{id}/detections", mitreHandler.GetDetectionsByTechnique)
	s.router.HandleFunc("GET /api/mitre/coverage", mitreHandler.GetCoverageByTactic)
	s.router.HandleFunc("GET /api/mitre/coverage/summary", mitreHandler.GetCoverageSummary)
	
	// API routes - Data Sources
	s.router.HandleFunc("GET /api/datasources", dataSourceHandler.ListDataSources)
	s.router.HandleFunc("POST /api/datasources", dataSourceHandler.CreateDataSource)
	s.router.HandleFunc("GET /api/datasources/utilization", dataSourceHandler.GetDataSourceUtilization)
	s.router.HandleFunc("GET /api/datasources/by-name/{name}", dataSourceHandler.GetDataSourceByName)
	s.router.HandleFunc("GET /api/datasources/{id}", dataSourceHandler.GetDataSource)
	s.router.HandleFunc("PUT /api/datasources/{id}", dataSourceHandler.UpdateDataSource)
	s.router.HandleFunc("DELETE /api/datasources/{id}", dataSourceHandler.DeleteDataSource)
	s.router.HandleFunc("GET /api/datasources/id/{id}/detections", dataSourceHandler.GetDetectionsByDataSource)
	s.router.HandleFunc("GET /api/datasources/id/{id}/techniques", dataSourceHandler.GetMitreTechniquesByDataSource)
	
	// API routes - Risk
	s.router.HandleFunc("POST /api/events", riskHandler.ProcessEvent)
	s.router.HandleFunc("POST /api/events/batch", riskHandler.ProcessEvents)
	s.router.HandleFunc("GET /api/events", riskHandler.ListEvents)
	s.router.HandleFunc("GET /api/events/{id}", riskHandler.GetEvent)
	s.router.HandleFunc("GET /api/events/entity/{id}", riskHandler.ListEventsByEntity)
	s.router.HandleFunc("POST /api/events/{id}/false-positive", riskHandler.MarkEventAsFalsePositive)
	s.router.HandleFunc("DELETE /api/events/{id}/false-positive", riskHandler.UnmarkEventAsFalsePositive)
	s.router.HandleFunc("GET /api/risk/objects", riskHandler.ListRiskObjects)
	s.router.HandleFunc("GET /api/risk/objects/{id}", riskHandler.GetRiskObject)
	s.router.HandleFunc("GET /api/risk/objects/entity", riskHandler.GetRiskObjectByEntity)
	s.router.HandleFunc("GET /api/risk/alerts", riskHandler.ListRiskAlerts)
	s.router.HandleFunc("GET /api/risk/alerts/{id}", riskHandler.GetRiskAlert)
	s.router.HandleFunc("PUT /api/risk/alerts/{id}", riskHandler.UpdateRiskAlert)
	s.router.HandleFunc("GET /api/risk/alerts/{id}/events", riskHandler.GetEventsForAlert)
	s.router.HandleFunc("POST /api/risk/decay", riskHandler.DecayRiskScores)
	s.router.HandleFunc("GET /api/risk/high", riskHandler.GetHighRiskEntities)
}

// ServeHTTP implements the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// Start starts the API server
func (s *Server) Start(addr string) error {
	log.Printf("Starting server on %s", addr)
	return http.ListenAndServe(addr, s)
}

// StartRiskDecayProcess starts the background process to decay risk scores
func (s *Server) StartRiskDecayProcess() chan struct{} {
	stop := make(chan struct{})
	go s.riskEngine.StartDecayProcess(stop)
	return stop
}