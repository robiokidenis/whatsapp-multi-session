package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
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
	//templateRepo := repository.NewTemplateRepository(db.DB()) // Temporarily disabled
	autoReplyRepo := repository.NewAutoReplyRepository(db.DB())
	analyticsRepo := repository.NewAnalyticsRepository(db.DB())
	messageRepo := repository.NewMessageRepository(db.DB())
	
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
	whatsappService, err := services.NewWhatsAppService(cfg.WhatsAppDBPath, sessionRepo, log)
	if err != nil {
		log.Fatalf("Failed to initialize WhatsApp service: %v", err)
	}
	defer whatsappService.Close()

	// Initialize CRM services
	contactDetectionService := services.NewContactDetectionService(*log)
	bulkMessagingService := services.NewBulkMessagingService(whatsappService, *log)
	analyticsService := services.NewAnalyticsService(analyticsRepo, userRepo, log)

	// Ensure default admin user exists
	if err := userService.EnsureDefaultAdmin(cfg.AdminUsername, cfg.AdminPassword); err != nil {
		log.Fatalf("Failed to ensure default admin: %v", err)
	}

	// Initialize rate limiter
	rateLimiter := ratelimiter.NewLoginRateLimiter()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userService, rateLimiter, log)
	sessionHandler := handlers.NewSessionHandler(whatsappService, messageRepo, cfg.JWTSecret, log, cfg.CORSAllowedOrigins)
	adminHandler := handlers.NewAdminHandler(userService, log)
	mediaHandler := handlers.NewMediaHandler(log)

	// Initialize CRM handlers
	contactHandler := handlers.NewContactHandler(contactRepo, contactGroupRepo, contactDetectionService, log)
	contactGroupHandler := handlers.NewContactGroupHandler(contactGroupRepo, log)
	//templateHandler := handlers.NewTemplateHandler(templateRepo, contactRepo, log)
	bulkMessagingHandler := handlers.NewBulkMessagingHandler(bulkMessagingService, log)
	autoReplyHandler := handlers.NewAutoReplyHandler(autoReplyRepo, log)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService, log)

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
		nil, // templateHandler temporarily disabled
		bulkMessagingHandler,
		autoReplyHandler,
		analyticsHandler,
		userService,
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
	templateHandler interface{},
	bulkMessagingHandler *handlers.BulkMessagingHandler,
	autoReplyHandler *handlers.AutoReplyHandler,
	analyticsHandler *handlers.AnalyticsHandler,
	userService *services.UserService,
	cfg *config.Config,
) *mux.Router {
	router := mux.NewRouter()

	// API routes (register these first)
	api := router.PathPrefix("/api").Subrouter()

	// Auth routes (no authentication required)
	auth := api.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/login", authHandler.Login).Methods("POST")

	// Authenticated auth routes (for password change and API key management)
	authProtected := api.PathPrefix("/auth").Subrouter()
	authProtected.Use(middleware.FlexibleAuthMiddleware(cfg.JWTSecret, userService))
	authProtected.HandleFunc("/change-password", authHandler.ChangePassword).Methods("POST")
	authProtected.HandleFunc("/api-key", authHandler.GenerateAPIKey).Methods("POST")
	authProtected.HandleFunc("/api-key", authHandler.RevokeAPIKey).Methods("DELETE")
	authProtected.HandleFunc("/api-key", authHandler.GetAPIKeyInfo).Methods("GET")
	
	// Health check
	api.HandleFunc("/health", healthHandler).Methods("GET")

	// Media routes (authentication required for security)
	media := api.PathPrefix("/media").Subrouter()
	media.Use(middleware.FlexibleAuthMiddleware(cfg.JWTSecret, userService))
	media.HandleFunc("/temp/{filename}", mediaHandler.ServeTempMedia).Methods("GET")

	// Protected routes (authentication required)
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.FlexibleAuthMiddleware(cfg.JWTSecret, userService))

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
	sessions.HandleFunc("/{sessionId}/auto-reply", sessionHandler.UpdateSessionAutoReply).Methods("PUT")
	sessions.HandleFunc("/{sessionId}/proxy", sessionHandler.UpdateSessionProxy).Methods("PUT")

	// Message routes
	sessions.HandleFunc("/{sessionId}/send", sessionHandler.SendMessage).Methods("POST")
	sessions.HandleFunc("/{sessionId}/send-location", sessionHandler.SendLocation).Methods("POST")
	sessions.HandleFunc("/{sessionId}/send-attachment", sessionHandler.SendAttachment).Methods("POST")
	sessions.HandleFunc("/{sessionId}/send-image", sessionHandler.SendImage).Methods("POST")
	sessions.HandleFunc("/{sessionId}/send-file-url", sessionHandler.SendFileFromURL).Methods("POST")
	sessions.HandleFunc("/{sessionId}/check-number", sessionHandler.CheckNumber).Methods("POST")
	sessions.HandleFunc("/{sessionId}/typing", sessionHandler.SendTyping).Methods("POST")
	sessions.HandleFunc("/{sessionId}/stop-typing", sessionHandler.StopTyping).Methods("POST")
	sessions.HandleFunc("/{sessionId}/presence", sessionHandler.SetPresence).Methods("POST")
	sessions.HandleFunc("/{sessionId}/groups", sessionHandler.GetGroups).Methods("GET")
	sessions.HandleFunc("/{sessionId}/conversations", sessionHandler.GetConversations).Methods("GET")

	// General send endpoint for compatibility with original API
	protected.HandleFunc("/send", sessionHandler.SendMessageGeneral).Methods("POST")

	// WebSocket endpoint with token-based authentication (no middleware needed as it authenticates via query params)
	api.HandleFunc("/ws/{sessionId}", sessionHandler.WebSocketHandler).Methods("GET")

	// Proxy testing route (no authentication required for testing)
	api.HandleFunc("/proxy/test", sessionHandler.TestProxy).Methods("POST")

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

	// Message templates management - temporarily disabled

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

	// Analytics routes
	protected.HandleFunc("/analytics", analyticsHandler.GetAnalytics).Methods("GET")
	protected.HandleFunc("/analytics/messages", analyticsHandler.GetMessageStats).Methods("GET")
	protected.HandleFunc("/analytics/sessions", analyticsHandler.GetSessionStats).Methods("GET")

	// User management routes (admin only)
	admin := protected.PathPrefix("/admin").Subrouter()
	admin.Use(middleware.RequireRole("admin"))
	admin.HandleFunc("/users", adminHandler.GetUsers).Methods("GET")
	admin.HandleFunc("/users", adminHandler.CreateUser).Methods("POST")
	admin.HandleFunc("/users/{id}", adminHandler.GetUser).Methods("GET")
	admin.HandleFunc("/users/{id}", adminHandler.UpdateUser).Methods("PUT")
	admin.HandleFunc("/users/{id}", adminHandler.DeleteUser).Methods("DELETE")
	
	// Admin API key management
	admin.HandleFunc("/users/{userId}/api-key", authHandler.AdminGenerateAPIKey).Methods("POST")
	admin.HandleFunc("/users/{userId}/api-key", authHandler.AdminRevokeAPIKey).Methods("DELETE")
	
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
	auth_admin.Use(middleware.FlexibleAuthMiddleware(cfg.JWTSecret, userService))
	auth_admin.Use(middleware.RequireRole("admin"))
	auth_admin.HandleFunc("/register", authHandler.Register).Methods("POST")

	// Static files (frontend) - register last to avoid conflicts
	if cfg.EnableFrontend {
		router.PathPrefix("/").Handler(SPAHandler("./frontend/dist/"))
	} else {
		// Serve frontend disabled message for all non-API routes
		router.PathPrefix("/").HandlerFunc(frontendDisabledHandler)
	}

	return router
}

// healthHandler provides a simple health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok","service":"whatsapp-multi-session"}`)
}

// frontendDisabledHandler serves a message when the frontend is disabled
func frontendDisabledHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>WhatsApp Multi-Session - Frontend Disabled</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen', 'Ubuntu', 'Cantarell', sans-serif;
            margin: 0;
            padding: 0;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .container {
            background: white;
            border-radius: 12px;
            padding: 40px;
            text-align: center;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            max-width: 500px;
            margin: 20px;
        }
        .icon {
            font-size: 64px;
            margin-bottom: 20px;
            color: #667eea;
        }
        h1 {
            color: #333;
            margin: 0 0 16px 0;
            font-size: 28px;
            font-weight: 600;
        }
        p {
            color: #666;
            font-size: 16px;
            line-height: 1.6;
            margin: 0 0 24px 0;
        }
        .api-info {
            background: #f8f9ff;
            border: 1px solid #e1e5f7;
            border-radius: 8px;
            padding: 20px;
            margin: 20px 0;
        }
        .api-info h3 {
            color: #333;
            margin: 0 0 12px 0;
            font-size: 16px;
            font-weight: 600;
        }
        .api-info code {
            background: #667eea;
            color: white;
            padding: 2px 6px;
            border-radius: 4px;
            font-size: 14px;
        }
        .footer {
            margin-top: 30px;
            padding-top: 20px;
            border-top: 1px solid #eee;
            color: #999;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">🚫</div>
        <h1>Frontend Disabled</h1>
        <p>The web interface for this WhatsApp Multi-Session service has been disabled by the administrator.</p>
        
        <div class="api-info">
            <h3>📡 API Access Available</h3>
            <p>You can still access all functionality through the REST API endpoints:</p>
            <p><code>/api/sessions</code> - Session management</p>
            <p><code>/api/send</code> - Send messages</p>
            <p><code>/api/auth</code> - Authentication</p>
        </div>
        
        <p>To enable the web interface, set the <strong>ENABLE_FRONTEND=true</strong> environment variable and restart the service.</p>
        
        <div class="footer">
            WhatsApp Multi-Session API Server
        </div>
    </div>
</body>
</html>`
	
	fmt.Fprint(w, html)
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
