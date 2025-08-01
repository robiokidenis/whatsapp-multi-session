package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/gorilla/mux"

	"whatsapp-multi-session/internal/config"
	"whatsapp-multi-session/internal/handlers"
	"whatsapp-multi-session/internal/middleware"
	"whatsapp-multi-session/internal/repository"
	"whatsapp-multi-session/internal/services"
	"whatsapp-multi-session/pkg/logger"
	"whatsapp-multi-session/pkg/ratelimiter"
)

func main() {
	// Load configuration (will use environment variables if set, otherwise defaults)
	cfg := config.Load()

	// Initialize logger
	log := logger.New(cfg.EnableLogging, cfg.LogLevel)
	log.Info("Starting WhatsApp Multi-Session Manager")
	log.Info("Configuration loaded - Port: %s, Log Level: %s, Database Logging: %t", cfg.Port, cfg.LogLevel, cfg.EnableDatabaseLog)

	// Get absolute path for WhatsApp DB
	whatsappDBPath, err := filepath.Abs(cfg.WhatsAppDBPath)
	if err != nil {
		log.Fatalf("Failed to get absolute path for WhatsApp DB: %v", err)
	}

	// Initialize database
	dbConfig := repository.DatabaseConfig{
		Type:     cfg.DatabaseType,
		Path:     cfg.DatabasePath,
		Host:     cfg.MySQLHost,
		Port:     cfg.MySQLPort,
		User:     cfg.MySQLUser,
		Password: cfg.MySQLPassword,
		Database: cfg.MySQLDatabase,
	}
	
	db, err := repository.NewDatabase(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize database tables
	if err := db.InitTables(); err != nil {
		log.Fatalf("Failed to initialize database tables: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.DB())
	sessionRepo := repository.NewSessionRepository(db.DB())
	contactRepo := repository.NewContactRepository(db.DB())
	contactGroupRepo := repository.NewContactGroupRepository(db.DB())
	templateRepo := repository.NewTemplateRepository(db.DB())
	autoReplyRepo := repository.NewAutoReplyRepository(db.DB())
	userSettingsRepo := repository.NewUserSettingsRepository(db.DB())
	jobQueueRepo := repository.NewJobQueueRepository(db.DB())
	
	var logRepo *repository.LogRepository
	// Setup database logging only if enabled
	if cfg.EnableDatabaseLog {
		logRepo = repository.NewLogRepository(db)
		dbWriter := logger.NewDatabaseWriter(logRepo)
		log.AddWriter(dbWriter)
		log.Info("Database logging enabled - logs will be stored in database")
	} else {
		log.Info("Database logging disabled - logs will only appear in console")
	}

	// Initialize services
	userService := services.NewUserService(userRepo, cfg.JWTSecret, log)
	whatsappService, err := services.NewWhatsAppService(whatsappDBPath, sessionRepo, log)
	if err != nil {
		log.Fatalf("Failed to initialize WhatsApp service: %v", err)
	}
	defer whatsappService.Close()

	// Initialize CRM services
	contactDetectionService := services.NewContactDetectionService(*log)
	bulkMessagingService := services.NewBulkMessagingService(whatsappService, contactRepo, templateRepo, *log)
	
	// Initialize job queue service
	jobQueueService := services.NewJobQueueService(jobQueueRepo, bulkMessagingService, log)
	
	// Start job queue service
	if err := jobQueueService.Start(); err != nil {
		log.Fatalf("Failed to start job queue service: %v", err)
	}
	defer jobQueueService.Stop()

	// Ensure default admin user exists
	if err := userService.EnsureDefaultAdmin(cfg.AdminUsername, cfg.AdminPassword); err != nil {
		log.Fatalf("Failed to ensure default admin: %v", err)
	}

	// Initialize rate limiter
	rateLimiter := ratelimiter.NewLoginRateLimiter()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userService, rateLimiter, log)
	sessionHandler := handlers.NewSessionHandler(whatsappService, cfg.JWTSecret, log, cfg.CORSAllowedOrigins)
	adminHandler := handlers.NewAdminHandler(userService, log)
	mediaHandler := handlers.NewMediaHandler(log)

	// Initialize CRM handlers
	contactHandler := handlers.NewContactHandler(contactRepo, contactGroupRepo, contactDetectionService, log)
	contactGroupHandler := handlers.NewContactGroupHandler(contactGroupRepo, log)
	templateHandler := handlers.NewTemplateHandler(templateRepo, contactRepo, log)
	bulkMessagingHandler := handlers.NewBulkMessagingHandler(jobQueueService, log)
	autoReplyHandler := handlers.NewAutoReplyHandler(autoReplyRepo, log)
	userSettingsHandler := handlers.NewUserSettingsHandler(userSettingsRepo, log)
	jobQueueHandler := handlers.NewJobQueueHandler(jobQueueService, log)

	var logHandler *handlers.LogHandler
	if cfg.EnableDatabaseLog && logRepo != nil {
		logHandler = handlers.NewLogHandler(logRepo, log)
	}

	// Setup routes
	router := setupRoutes(
		authHandler, 
		sessionHandler, 
		adminHandler, 
		mediaHandler, 
		logHandler, 
		contactHandler,
		contactGroupHandler,
		templateHandler,
		bulkMessagingHandler,
		autoReplyHandler,
		userSettingsHandler,
		jobQueueHandler,
		cfg,
	)

	// Setup CORS
	corsHandler := middleware.NewCORS(cfg.CORSAllowedOrigins)
	handler := corsHandler.Handler(router)

	// Start server
	address := ":" + cfg.Port
	log.Info("Server starting on %s", address)

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	server := &http.Server{
		Addr:    address,
		Handler: handler,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	log.Info("Server is ready to handle requests at %s", address)

	// Wait for shutdown signal
	<-c
	log.Info("Shutting down server...")

	// Perform cleanup here if needed
	log.Info("Server shutdown complete")
}

// setupRoutes configures all HTTP routes
func setupRoutes(
	authHandler *handlers.AuthHandler,
	sessionHandler *handlers.SessionHandler,
	adminHandler *handlers.AdminHandler,
	mediaHandler *handlers.MediaHandler,
	logHandler *handlers.LogHandler,
	contactHandler *handlers.ContactHandler,
	contactGroupHandler *handlers.ContactGroupHandler,
	templateHandler *handlers.TemplateHandler,
	bulkMessagingHandler *handlers.BulkMessagingHandler,
	autoReplyHandler *handlers.AutoReplyHandler,
	userSettingsHandler *handlers.UserSettingsHandler,
	jobQueueHandler *handlers.JobQueueHandler,
	cfg *config.Config,
) *mux.Router {
	router := mux.NewRouter()

	// API routes (register these first)
	api := router.PathPrefix("/api").Subrouter()

	// Auth routes (no authentication required)
	auth := api.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/login", authHandler.Login).Methods("POST")
	
	// Health check
	api.HandleFunc("/health", healthHandler).Methods("GET")

	// Media routes (authentication required for security)
	media := api.PathPrefix("/media").Subrouter()
	media.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	media.HandleFunc("/temp/{filename}", mediaHandler.ServeTempMedia).Methods("GET")

	// Protected routes (authentication required)
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))

	// Session routes
	sessions := protected.PathPrefix("/sessions").Subrouter()
	sessions.HandleFunc("", sessionHandler.GetSessions).Methods("GET")
	sessions.HandleFunc("", sessionHandler.CreateSession).Methods("POST")
	sessions.HandleFunc("/{sessionId}", sessionHandler.GetSession).Methods("GET")
	sessions.HandleFunc("/{sessionId}", sessionHandler.UpdateSession).Methods("PUT")
	sessions.HandleFunc("/{sessionId}", sessionHandler.DeleteSession).Methods("DELETE")

	// Connection management
	sessions.HandleFunc("/{sessionId}/connect", sessionHandler.ConnectSession).Methods("POST")
	sessions.HandleFunc("/{sessionId}/disconnect", sessionHandler.DisconnectSession).Methods("POST")
	sessions.HandleFunc("/{sessionId}/login", sessionHandler.LoginSession).Methods("POST")
	sessions.HandleFunc("/{sessionId}/logout", sessionHandler.LogoutSession).Methods("POST")

	// QR code and WebSocket
	sessions.HandleFunc("/{sessionId}/qr", sessionHandler.GetQRCode).Methods("GET")
	sessions.HandleFunc("/{sessionId}/ws", sessionHandler.WebSocketHandler).Methods("GET")

	// Session metadata updates
	sessions.HandleFunc("/{sessionId}/webhook", sessionHandler.UpdateSessionWebhook).Methods("PUT")
	sessions.HandleFunc("/{sessionId}/name", sessionHandler.UpdateSessionName).Methods("PUT")

	// Message routes
	sessions.HandleFunc("/{sessionId}/send", sessionHandler.SendMessage).Methods("POST")
	sessions.HandleFunc("/{sessionId}/send-location", sessionHandler.SendLocation).Methods("POST")
	sessions.HandleFunc("/{sessionId}/send-attachment", sessionHandler.SendAttachment).Methods("POST")
	sessions.HandleFunc("/{sessionId}/send-image", sessionHandler.SendImage).Methods("POST")
	sessions.HandleFunc("/{sessionId}/send-file-url", sessionHandler.SendFileFromURL).Methods("POST")
	sessions.HandleFunc("/{sessionId}/check-number", sessionHandler.CheckNumber).Methods("POST")
	sessions.HandleFunc("/{sessionId}/typing", sessionHandler.SendTyping).Methods("POST")
	sessions.HandleFunc("/{sessionId}/stop-typing", sessionHandler.StopTyping).Methods("POST")
	sessions.HandleFunc("/{sessionId}/groups", sessionHandler.GetGroups).Methods("GET")

	// General send endpoint for compatibility with original API
	protected.HandleFunc("/send", sessionHandler.SendMessageGeneral).Methods("POST")

	// WebSocket endpoint with token-based authentication (no middleware needed as it authenticates via query params)
	api.HandleFunc("/ws/{sessionId}", sessionHandler.WebSocketHandler).Methods("GET")

	// CRM routes (authentication required)
	// Contact management
	protected.HandleFunc("/contacts", contactHandler.GetContacts).Methods("GET")
	protected.HandleFunc("/contacts", contactHandler.CreateContact).Methods("POST")
	protected.HandleFunc("/contacts/{id}", contactHandler.UpdateContact).Methods("PUT")
	protected.HandleFunc("/contacts/{id}", contactHandler.DeleteContact).Methods("DELETE")
	protected.HandleFunc("/contacts/bulk", contactHandler.BulkActions).Methods("POST")
	protected.HandleFunc("/contacts/detect", contactHandler.DetectContacts).Methods("POST")
	protected.HandleFunc("/contacts/import", contactHandler.ImportContacts).Methods("POST")

	// Contact groups management
	protected.HandleFunc("/contact-groups", contactGroupHandler.GetContactGroups).Methods("GET")
	protected.HandleFunc("/contact-groups", contactGroupHandler.CreateContactGroup).Methods("POST")
	protected.HandleFunc("/contact-groups/{id}", contactGroupHandler.UpdateContactGroup).Methods("PUT")
	protected.HandleFunc("/contact-groups/{id}", contactGroupHandler.DeleteContactGroup).Methods("DELETE")

	// Message templates management
	if templateHandler != nil {
		protected.HandleFunc("/message-templates", templateHandler.GetMessageTemplates).Methods("GET")
		protected.HandleFunc("/message-templates", templateHandler.CreateMessageTemplate).Methods("POST")
		protected.HandleFunc("/message-templates/{id}", templateHandler.UpdateMessageTemplate).Methods("PUT")
		protected.HandleFunc("/message-templates/{id}", templateHandler.DeleteMessageTemplate).Methods("DELETE")
		protected.HandleFunc("/message-templates/categories", templateHandler.GetMessageTemplateCategories).Methods("GET")
		protected.HandleFunc("/message-templates/preview", templateHandler.PreviewMessageTemplate).Methods("POST")
	}

	// Bulk messaging
	protected.HandleFunc("/bulk-messages", bulkMessagingHandler.GetBulkMessagingJobs).Methods("GET")
	protected.HandleFunc("/bulk-messages", bulkMessagingHandler.StartBulkMessaging).Methods("POST")
	protected.HandleFunc("/bulk-messages/{jobId}", bulkMessagingHandler.GetBulkMessagingJob).Methods("GET")
	protected.HandleFunc("/bulk-messages/{jobId}", bulkMessagingHandler.CancelBulkMessagingJob).Methods("DELETE")

	// Auto-reply management
	protected.HandleFunc("/auto-replies", autoReplyHandler.GetAutoReplies).Methods("GET")
	protected.HandleFunc("/auto-replies", autoReplyHandler.CreateAutoReply).Methods("POST")
	protected.HandleFunc("/auto-replies/{id}", autoReplyHandler.UpdateAutoReply).Methods("PUT")
	protected.HandleFunc("/auto-replies/{id}", autoReplyHandler.DeleteAutoReply).Methods("DELETE")

	// User settings management
	protected.HandleFunc("/user/settings", userSettingsHandler.GetUserSettings).Methods("GET")
	protected.HandleFunc("/user/settings", userSettingsHandler.UpdateUserSettings).Methods("PUT")
	protected.HandleFunc("/user/settings", userSettingsHandler.PatchUserSettings).Methods("PATCH")
	protected.HandleFunc("/user/settings", userSettingsHandler.DeleteUserSettings).Methods("DELETE")

	// Job queue management - specific routes first, then general ones
	protected.HandleFunc("/job-queue/statistics", jobQueueHandler.GetStatistics).Methods("GET")
	protected.HandleFunc("/job-queue/cleanup", jobQueueHandler.CleanupJobs).Methods("POST")
	protected.HandleFunc("/job-queue/bulk-message", jobQueueHandler.CreateBulkMessageJob).Methods("POST")
	protected.HandleFunc("/job-queue/scheduled-message", jobQueueHandler.CreateScheduledMessageJob).Methods("POST")
	protected.HandleFunc("/job-queue", jobQueueHandler.GetJobs).Methods("GET")
	protected.HandleFunc("/job-queue", jobQueueHandler.CreateJob).Methods("POST")
	protected.HandleFunc("/job-queue/{jobId}", jobQueueHandler.GetJob).Methods("GET")
	protected.HandleFunc("/job-queue/{jobId}", jobQueueHandler.CancelJob).Methods("DELETE")
	protected.HandleFunc("/job-queue/{jobId}/retry", jobQueueHandler.RetryJob).Methods("POST")
	
	// Debug endpoint
	protected.HandleFunc("/job-queue-test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message":"job-queue handler is working","handler_exists":%t}`, jobQueueHandler != nil)
	}).Methods("GET")

	// User management routes (admin only)
	admin := protected.PathPrefix("/admin").Subrouter()
	admin.Use(middleware.RequireRole("admin"))
	admin.HandleFunc("/users", adminHandler.GetUsers).Methods("GET")
	admin.HandleFunc("/users", adminHandler.CreateUser).Methods("POST")
	admin.HandleFunc("/users/{id}", adminHandler.GetUser).Methods("GET")
	admin.HandleFunc("/users/{id}", adminHandler.UpdateUser).Methods("PUT")
	admin.HandleFunc("/users/{id}", adminHandler.DeleteUser).Methods("DELETE")
	
	// Admin user settings management
	admin.HandleFunc("/user-settings", userSettingsHandler.GetUserSettingsByID).Methods("GET")
	
	// Log status endpoint (always available)
	admin.HandleFunc("/logs/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		status := map[string]interface{}{
			"database_logging_enabled": cfg.EnableDatabaseLog,
			"console_logging_enabled":  cfg.EnableLogging,
			"log_level":               cfg.LogLevel,
		}
		json.NewEncoder(w).Encode(status)
	}).Methods("GET")

	// Log management routes (admin only) - only if database logging is enabled
	if logHandler != nil {
		admin.HandleFunc("/logs", logHandler.GetLogs).Methods("GET")
		admin.HandleFunc("/logs/levels", logHandler.GetLogLevels).Methods("GET")
		admin.HandleFunc("/logs/components", logHandler.GetLogComponents).Methods("GET")
		admin.HandleFunc("/logs/cleanup/{days}", logHandler.DeleteOldLogs).Methods("DELETE")
		admin.HandleFunc("/logs/clear", logHandler.ClearAllLogs).Methods("DELETE")
	}
	
	// User registration (admin only)
	auth_admin := api.PathPrefix("/auth").Subrouter()
	auth_admin.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	auth_admin.Use(middleware.RequireRole("admin"))
	auth_admin.HandleFunc("/register", authHandler.Register).Methods("POST")

	// Static files (frontend) - register last to avoid conflicts
	router.PathPrefix("/").Handler(SPAHandler("./frontend/dist/"))

	return router
}

// healthHandler provides a simple health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok","service":"whatsapp-multi-session"}`)
}

// SPAHandler implements the http.Handler interface, serving static files from the filesystem
// and falling back to index.html for routes that don't exist (for single-page applications)
func SPAHandler(staticPath string) http.Handler {
	fileServer := http.FileServer(http.Dir(staticPath))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the requested file exists
		path := staticPath + r.URL.Path
		if _, err := os.Stat(path); os.IsNotExist(err) {
			// File does not exist, serve index.html
			http.ServeFile(w, r, staticPath+"/index.html")
			return
		}

		// File exists, serve it normally
		fileServer.ServeHTTP(w, r)
	})
}
