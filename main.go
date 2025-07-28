package main

import (
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
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	log := logger.New(cfg.EnableLogging, cfg.LogLevel)
	log.Info("Starting WhatsApp Multi-Session Manager")
	log.Info("Configuration loaded - Port: %s, Log Level: %s", cfg.Port, cfg.LogLevel)

	// Initialize database
	db, err := repository.NewDatabase(cfg.DatabasePath)
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

	// Initialize services
	userService := services.NewUserService(userRepo, cfg.JWTSecret, log)
	whatsappService, err := services.NewWhatsAppService(cfg.WhatsAppDBPath, sessionRepo, log)
	if err != nil {
		log.Fatalf("Failed to initialize WhatsApp service: %v", err)
	}
	defer whatsappService.Close()

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

	// Setup routes
	router := setupRoutes(authHandler, sessionHandler, adminHandler, mediaHandler, cfg)

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
	cfg *config.Config,
) *mux.Router {
	router := mux.NewRouter()

	// API routes (register these first)
	api := router.PathPrefix("/api").Subrouter()

	// Auth routes (no authentication required)
	auth := api.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/login", authHandler.Login).Methods("POST")
	auth.HandleFunc("/register", authHandler.Register).Methods("POST")
	
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

	// User management routes (admin only)
	admin := protected.PathPrefix("/admin").Subrouter()
	admin.Use(middleware.RequireRole("admin"))
	admin.HandleFunc("/users", adminHandler.GetUsers).Methods("GET")
	admin.HandleFunc("/users", adminHandler.CreateUser).Methods("POST")
	admin.HandleFunc("/users/{id}", adminHandler.GetUser).Methods("GET")
	admin.HandleFunc("/users/{id}", adminHandler.UpdateUser).Methods("PUT")
	admin.HandleFunc("/users/{id}", adminHandler.DeleteUser).Methods("DELETE")

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
