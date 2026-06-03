package handlers

import (
	"embed"
	"io"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
)

// StaticHandler serves embedded frontend static files with SPA fallback
type StaticHandler struct {
	fs  fs.FS
	log *slog.Logger
}

// NewStaticHandler creates a new static file handler for embedded frontend assets
func NewStaticHandler(embeddedFS embed.FS, log *slog.Logger) *StaticHandler {
	return &StaticHandler{
		fs:  embeddedFS,
		log: log,
	}
}

// Handle serves static files or falls back to index.html for SPA routing
func (h *StaticHandler) Handle(c echo.Context) error {
	if h.fs == nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "static files not available",
		})
	}

	requestPath := c.Request().URL.Path
	// Remove leading slash for filesystem lookup
	if requestPath == "/" {
		requestPath = "index.html"
	} else if !strings.HasPrefix(requestPath, "/assets") && !strings.HasPrefix(requestPath, "/") {
		requestPath = "/" + requestPath
	}

	// For root path or asset requests, try to serve the file
	if strings.HasPrefix(requestPath, "/assets/") || requestPath == "index.html" || requestPath == "/index.html" {
		filePath := strings.TrimPrefix(requestPath, "/")
		return h.serveFile(c, filePath)
	}

	// For unknown routes (SPA routing), try to serve the requested file first
	// If it doesn't exist, fall back to index.html
	filePath := strings.TrimPrefix(requestPath, "/")
	file, err := h.fs.Open(filePath)
	if err == nil {
		file.Close()
		// File exists, serve it
		return h.serveFile(c, filePath)
	}

	// File doesn't exist, fall back to index.html for SPA routing
	return h.serveFile(c, "index.html")
}

// serveFile reads a file from the embedded filesystem and serves it with appropriate headers
func (h *StaticHandler) serveFile(c echo.Context, filePath string) error {
	file, err := h.fs.Open(filePath)
	if err != nil {
		// Not found - for SPA, return index.html instead
		if filePath != "index.html" {
			return h.serveFile(c, "index.html")
		}
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "file not found",
		})
	}
	defer file.Close()

	// Get file info for content length
	fileInfo, err := file.Stat()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "unable to read file",
		})
	}

	// Determine content type
	contentType := mime.TypeByExtension(path.Ext(filePath))
	if contentType == "" {
		contentType = "text/html; charset=utf-8"
	}

	// Set headers
	c.Response().Header().Set("Content-Type", contentType)
	c.Response().Header().Set("Content-Length", string(rune(fileInfo.Size())))

	// Set cache headers for assets vs HTML
	if strings.HasPrefix(filePath, "assets/") {
		// Assets have versioned filenames (from Vite), cache aggressively
		c.Response().Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else if filePath == "index.html" {
		// HTML should not be cached
		c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Response().Header().Set("Pragma", "no-cache")
		c.Response().Header().Set("Expires", "0")
	}

	c.Response().WriteHeader(http.StatusOK)
	_, _ = io.Copy(c.Response(), file)
	return nil
}
