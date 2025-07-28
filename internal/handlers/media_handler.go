package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"whatsapp-multi-session/pkg/logger"
)

// MediaHandler handles media file serving
type MediaHandler struct {
	logger *logger.Logger
}

// NewMediaHandler creates a new media handler
func NewMediaHandler(log *logger.Logger) *MediaHandler {
	return &MediaHandler{
		logger: log,
	}
}

// ServeTempMedia serves temporary media files with expiration and authentication check
func (h *MediaHandler) ServeTempMedia(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileName := vars["filename"]
	
	// Authentication is now handled by middleware, but let's add user context validation
	_, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	
	// Check expiration
	expiresStr := r.URL.Query().Get("expires")
	if expiresStr == "" {
		http.Error(w, "Missing expires parameter", http.StatusBadRequest)
		return
	}
	
	expires, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid expires parameter", http.StatusBadRequest)
		return
	}
	
	// Check if expired
	if time.Now().Unix() > expires {
		http.Error(w, "Media link has expired", http.StatusGone)
		return
	}
	
	// Validate filename (security check)
	if strings.Contains(fileName, "..") || strings.Contains(fileName, "/") {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}
	
	// Construct file path
	filePath := filepath.Join("./media/received", fileName)
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "Media file not found", http.StatusNotFound)
		return
	}
	
	// Determine content type based on file extension
	ext := strings.ToLower(filepath.Ext(fileName))
	var contentType string
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".pdf":
		contentType = "application/pdf"
	case ".mp4":
		contentType = "video/mp4"
	case ".ogg":
		contentType = "audio/ogg"
	case ".mp3":
		contentType = "audio/mpeg"
	default:
		contentType = "application/octet-stream"
	}
	
	// Set appropriate headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "private, max-age=3600") // Cache for 1 hour
	w.Header().Set("Content-Disposition", "inline; filename=\""+fileName+"\"")
	
	// Serve the file
	http.ServeFile(w, r, filePath)
	
	h.logger.Debug("Served temporary media file: %s", fileName)
}

// CleanupExpiredMedia removes expired media files (should be called periodically)
func (h *MediaHandler) CleanupExpiredMedia() {
	mediaDir := "./media/received"
	
	// Clean up files older than 24 hours
	cutoff := time.Now().Add(-24 * time.Hour)
	
	files, err := os.ReadDir(mediaDir)
	if err != nil {
		h.logger.Error("Failed to read media directory: %v", err)
		return
	}
	
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		
		info, err := file.Info()
		if err != nil {
			continue
		}
		
		if info.ModTime().Before(cutoff) {
			filePath := filepath.Join(mediaDir, file.Name())
			if err := os.Remove(filePath); err != nil {
				h.logger.Error("Failed to remove expired media file %s: %v", filePath, err)
			} else {
				h.logger.Info("Removed expired media file: %s", filePath)
			}
		}
	}
}