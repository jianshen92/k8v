package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Logger wraps log functionality with file output
type Logger struct {
	file   *os.File
	logger *log.Logger
}

// NewLogger creates a new logger that writes to both stdout and a file
func NewLogger() (*Logger, error) {
	// Create logs directory if it doesn't exist
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Use single log file
	logPath := filepath.Join(logsDir, "k8v.log")

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Create multi-writer to write to both stdout and file
	multiWriter := io.MultiWriter(os.Stdout, file)
	logger := log.New(multiWriter, "", log.LstdFlags)

	logger.Printf("=== K8V Server Started (%s) ===", time.Now().Format("2006-01-02 15:04:05"))

	return &Logger{
		file:   file,
		logger: logger,
	}, nil
}

// Close closes the log file
func (l *Logger) Close() error {
	if l.file != nil {
		l.logger.Printf("=== K8V Server Stopped ===")
		return l.file.Close()
	}
	return nil
}

// Printf logs a formatted message
func (l *Logger) Printf(format string, v ...interface{}) {
	l.logger.Printf(format, v...)
}

// LoggingMiddleware returns an HTTP middleware that logs all requests
func (l *Logger) LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the ResponseWriter to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the actual handler
		next.ServeHTTP(wrapped, r)

		// Log the request
		duration := time.Since(start)
		l.logger.Printf(
			"%s %s %s - %d - %v",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration,
		)
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Hijack implements http.Hijacker interface for WebSocket support
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("responsewriter does not implement http.Hijacker")
	}
	return hijacker.Hijack()
}
