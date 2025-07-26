package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.mau.fi/whatsmeow/proto/waCompanionReg"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
	_ "github.com/mattn/go-sqlite3"

	"whatsapp-multi-session/internal/auth"
	"whatsapp-multi-session/internal/config"
	"whatsapp-multi-session/internal/database"
	"whatsapp-multi-session/internal/handlers"
	"whatsapp-multi-session/internal/session"
	"whatsapp-multi-session/internal/types"
)

// Global variables
var (
	sessionManager *types.SessionManager
	cfg            *types.Config
	logger         *types.Logger
	loginLimiter   *types.LoginRateLimiter
)

func init() {
	// Set up logging
	waLog.Stdout("Main", "INFO", true)

	// Set device properties to avoid client outdated error
	store.DeviceProps.PlatformType = waCompanionReg.DeviceProps_CHROME.Enum()
	store.DeviceProps.Os = proto.String("Windows")
	store.DeviceProps.RequireFullSync = proto.Bool(false)
}

func main() {
	// Load configuration
	cfg = config.LoadConfig()

	// Initialize logger
	logger = types.NewLogger(cfg.EnableLogging, cfg.LogLevel)
	logger.Info("Starting WhatsApp Multi-Session Manager with sqlite database")

	// Initialize login rate limiter
	loginLimiter = auth.NewLoginRateLimiter()
	logger.Info("Login rate limiter initialized")

	// Initialize WhatsApp database (always SQLite for whatsmeow)
	dbLog := waLog.Stdout("Database", "INFO", true)
	ctx := context.Background()
	// Ensure database directory exists
	if err := os.MkdirAll("database", 0755); err != nil {
		logger.Fatal("Failed to create database directory:", err)
	}

	container, err := sqlstore.New(ctx, "sqlite3", "file:database/sessions.db?_foreign_keys=on", dbLog)
	if err != nil {
		logger.Fatal("Failed to initialize WhatsApp database:", err)
	}

	// Initialize metadata database (SQLite)
	metadataDB, err := database.SetupDatabase(cfg)
	if err != nil {
		if logger != nil {
			logger.Fatal("Failed to setup metadata database:", err)
		} else {
			log.Fatal("Failed to setup metadata database:", err)
		}
	}

	// Initialize database tables
	logger.Info("Initializing database tables...")
	if err := database.InitSessionsTable(metadataDB); err != nil {
		logger.Fatal("Failed to initialize sessions table:", err)
	}
	logger.Info("Sessions table initialized successfully")

	if err := database.InitUsersTable(metadataDB, logger); err != nil {
		logger.Fatal("Failed to initialize users table:", err)
	}
	logger.Info("Users table initialized successfully")

	// Initialize session manager
	sessionManager = &types.SessionManager{
		Sessions:   make(map[string]*types.Session),
		Store:      container,
		MetadataDB: metadataDB,
	}

	// Restore existing sessions
	logger.Info("Loading existing sessions...")
	sessionMetadata, err := database.LoadSessionMetadata(sessionManager)
	if err != nil {
		logger.Error("Error loading session metadata: %v", err)
	} else {
		for _, metadata := range sessionMetadata {
			restoredSession := session.RestoreSession(metadata, container, sessionManager, logger)
			if restoredSession != nil {
				sessionManager.Sessions[restoredSession.ID] = restoredSession
				logger.Info("Restored session %s (%s)", restoredSession.ID, restoredSession.Name)
			}
		}
		logger.Info("Restored %d sessions", len(sessionManager.Sessions))
	}

	// Set up routes
	router := mux.NewRouter()

	// Create auth middleware
	authMiddleware := auth.Middleware(cfg)

	// Debug endpoint
	router.HandleFunc("/debug", makeSimpleHandler("debug")).Methods("GET")
	
	// Public API routes (no authentication required)
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/login", handlers.LoginHandler(sessionManager, loginLimiter, cfg, logger)).Methods("POST")
	api.HandleFunc("/register", handlers.RegisterHandler(sessionManager, logger)).Methods("POST")

	// Protected API routes (authentication required)
	api.HandleFunc("/sessions", authMiddleware(handlers.ListSessionsHandler(sessionManager, logger))).Methods("GET")
	api.HandleFunc("/sessions", authMiddleware(handlers.CreateSessionHandler(sessionManager, logger))).Methods("POST")

	// Serve static files first
	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("./frontend/dist/assets/"))))
	
	// Serve React frontend for all other routes
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip API routes
		if r.URL.Path != "/" && !strings.HasPrefix(r.URL.Path, "/api/") {
			// Check if it's a static file request
			filePath := "./frontend/dist" + r.URL.Path
			if _, err := os.Stat(filePath); err == nil {
				http.ServeFile(w, r, filePath)
				return
			}
		}
		// Serve index.html for SPA routing
		http.ServeFile(w, r, "./frontend/dist/index.html")
	})

	// Enable CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("API available at http://localhost:8080/api")
	fmt.Println("Frontend available at http://localhost:8080")

	logger.Fatal(http.ListenAndServe(":8080", handler))
}

// Simple handler for testing
func makeSimpleHandler(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message": "Handler %s working", "status": "success"}`, name)
	}
}