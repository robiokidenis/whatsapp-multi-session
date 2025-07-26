package utils

import (
	"net/http"
	"os"
	"path/filepath"
)

// SPAHandler serves the Single Page Application with fallback to index.html
func SPAHandler(staticPath string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle root path
		if r.URL.Path == "/" {
			http.ServeFile(w, r, filepath.Join(staticPath, "index.html"))
			return
		}
		
		// Build the full file path
		filePath := filepath.Join(staticPath, r.URL.Path)
		
		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// File doesn't exist, serve index.html for SPA routing
			http.ServeFile(w, r, filepath.Join(staticPath, "index.html"))
			return
		}
		
		// File exists, serve it
		http.ServeFile(w, r, filePath)
	})
}