package api

import (
	"log"
	"net/http"
	"os"
	"time"

	"riskmatrix/internal/datasource"
	"riskmatrix/internal/detection"
	"riskmatrix/internal/mitre"
	"riskmatrix/internal/risk"
	"riskmatrix/pkg/cache"
	"riskmatrix/pkg/database"
	"riskmatrix/pkg/middleware"
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
	handler        http.Handler
	cache          *cache.Cache
	preparedStmts  *database.PreparedStatements
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

	// Create cache with 5 minute TTL
	apiCache := cache.New(5 * time.Minute)

	// Create prepared statements
	preparedStmts := db.PreparedStatements()

	// Create server
	server := &Server{
		db:             db,
		detectionRepo:  detectionRepo,
		mitreRepo:      mitreRepo,
		dataSourceRepo: dataSourceRepo,
		riskRepo:       riskRepo,
		riskEngine:     riskEngine,
		router:         http.NewServeMux(),
		cache:          apiCache,
		preparedStmts:  preparedStmts,
	}

	// Set up routes
	server.setupRoutes()

	// Set up middleware chain
	server.setupMiddleware()

	return server
}

// setupRoutes sets up the API routes
func (s *Server) setupRoutes() {
	// Create handlers
	detectionHandler := NewDetectionHandler(s.detectionRepo)
	detectionClassHandler := NewDetectionClassHandler(s.detectionRepo)
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
	
	// API routes - Detection Classes
	s.router.HandleFunc("GET /api/detection-classes", detectionClassHandler.ListDetectionClasses)
	s.router.HandleFunc("POST /api/detection-classes", detectionClassHandler.CreateDetectionClass)
	s.router.HandleFunc("GET /api/detection-classes/{id}", detectionClassHandler.GetDetectionClass)
	s.router.HandleFunc("PUT /api/detection-classes/{id}", detectionClassHandler.UpdateDetectionClass)
	s.router.HandleFunc("DELETE /api/detection-classes/{id}", detectionClassHandler.DeleteDetectionClass)
	s.router.HandleFunc("GET /api/detection-classes/{id}/detections", detectionClassHandler.ListDetectionsByClass)
	
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

// setupMiddleware sets up the middleware chain
func (s *Server) setupMiddleware() {
	// Get auth credentials from environment or use defaults
	authEnabled := os.Getenv("AUTH_ENABLED") == "true"
	authUser := os.Getenv("AUTH_USER")
	authPass := os.Getenv("AUTH_PASSWORD")
	if authUser == "" {
		authUser = "admin"
	}
	if authPass == "" {
		authPass = "changeme"
	}

	// Create middleware instances
	authMiddleware := middleware.NewAuthMiddleware(middleware.AuthConfig{
		Username:        authUser,
		Password:        authPass,
		SessionDuration: 24 * time.Hour,
		Enabled:         authEnabled,
	})

	csrfMiddleware := middleware.NewCSRFMiddleware(middleware.CSRFConfig{
		Enabled: authEnabled, // Enable CSRF only if auth is enabled
	})

	rateLimiter := middleware.NewRateLimiter(middleware.RateLimitConfig{
		RequestsPerWindow: 1000, // Increased from 100 to 1000 for development
		WindowDuration:    time.Minute,
		Enabled:           true,
	})

	bodyLimiter := middleware.NewBodyLimitMiddleware(middleware.BodyLimitConfig{
		MaxBodySize: 10 * 1024 * 1024, // 10MB
		Enabled:     true,
	})

	// Build middleware chain
	handler := http.Handler(s.router)
	handler = csrfMiddleware.Middleware(handler)
	handler = authMiddleware.Middleware(handler)
	handler = rateLimiter.Middleware(handler)
	handler = bodyLimiter.Middleware(handler)

	s.handler = handler
}

// ServeHTTP implements the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

// Start starts the API server with proper timeout configuration
func (s *Server) Start(addr string) error {
	log.Printf("Starting server on %s", addr)
	
	// Configure server with timeouts to prevent resource exhaustion
	srv := &http.Server{
		Addr:         addr,
		Handler:      s,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		// Limit request header size to 1MB
		MaxHeaderBytes: 1 << 20,
	}
	
	return srv.ListenAndServe()
}

// StartRiskDecayProcess starts the background process to decay risk scores
func (s *Server) StartRiskDecayProcess() chan struct{} {
	stop := make(chan struct{})
	go s.riskEngine.StartDecayProcess(stop)
	return stop
}