package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var (
	currentLevel LogLevel = INFO
	logger       *log.Logger
)

// Fields represents structured log fields
type Fields map[string]interface{}

// Initialize sets up the logger with the specified level
func Initialize(level string, format string) {
	// Parse log level
	switch strings.ToUpper(level) {
	case "DEBUG":
		currentLevel = DEBUG
	case "INFO":
		currentLevel = INFO
	case "WARN", "WARNING":
		currentLevel = WARN
	case "ERROR":
		currentLevel = ERROR
	case "FATAL":
		currentLevel = FATAL
	default:
		currentLevel = INFO
	}

	// Set up logger output
	logger = log.New(os.Stdout, "", 0)
}

// shouldLog checks if the given level should be logged
func shouldLog(level LogLevel) bool {
	return level >= currentLevel
}

// formatMessage formats the log message with fields
func formatMessage(level string, msg string, fields Fields) string {
	timestamp := time.Now().UTC().Format(time.RFC3339)

	var parts []string
	parts = append(parts, fmt.Sprintf("[%s]", timestamp))
	parts = append(parts, fmt.Sprintf("[%s]", level))
	parts = append(parts, msg)

	if len(fields) > 0 {
		var fieldParts []string
		for k, v := range fields {
			fieldParts = append(fieldParts, fmt.Sprintf("%s=%v", k, v))
		}
		parts = append(parts, strings.Join(fieldParts, " "))
	}

	return strings.Join(parts, " ")
}

// getCallerInfo gets the file and line number of the caller
func getCallerInfo() (string, int) {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return "unknown", 0
	}

	// Extract just the filename
	parts := strings.Split(file, "/")
	filename := parts[len(parts)-1]

	return filename, line
}

// Debug logs a debug message
func Debug(msg string, fields ...Fields) {
	if !shouldLog(DEBUG) {
		return
	}

	var f Fields
	if len(fields) > 0 {
		f = fields[0]
	} else {
		f = make(Fields)
	}

	file, line := getCallerInfo()
	f["file"] = fmt.Sprintf("%s:%d", file, line)

	logger.Println(formatMessage("DEBUG", msg, f))
}

// Info logs an info message
func Info(msg string, fields ...Fields) {
	if !shouldLog(INFO) {
		return
	}

	var f Fields
	if len(fields) > 0 {
		f = fields[0]
	} else {
		f = make(Fields)
	}

	file, line := getCallerInfo()
	f["file"] = fmt.Sprintf("%s:%d", file, line)

	logger.Println(formatMessage("INFO", msg, f))
}

// Warn logs a warning message
func Warn(msg string, fields ...Fields) {
	if !shouldLog(WARN) {
		return
	}

	var f Fields
	if len(fields) > 0 {
		f = fields[0]
	} else {
		f = make(Fields)
	}

	file, line := getCallerInfo()
	f["file"] = fmt.Sprintf("%s:%d", file, line)

	logger.Println(formatMessage("WARN", msg, f))
}

// Error logs an error message
func Error(msg string, err error, fields ...Fields) {
	if !shouldLog(ERROR) {
		return
	}

	var f Fields
	if len(fields) > 0 {
		f = fields[0]
	} else {
		f = make(Fields)
	}

	if err != nil {
		f["error"] = err.Error()
	}

	file, line := getCallerInfo()
	f["file"] = fmt.Sprintf("%s:%d", file, line)

	logger.Println(formatMessage("ERROR", msg, f))
}

// Fatal logs a fatal message and exits
func Fatal(msg string, err error, fields ...Fields) {
	var f Fields
	if len(fields) > 0 {
		f = fields[0]
	} else {
		f = make(Fields)
	}

	if err != nil {
		f["error"] = err.Error()
	}

	file, line := getCallerInfo()
	f["file"] = fmt.Sprintf("%s:%d", file, line)

	logger.Println(formatMessage("FATAL", msg, f))
	os.Exit(1)
}

// WithFields creates a new logger instance with fields
func WithFields(fields Fields) *Entry {
	return &Entry{fields: fields}
}

// Entry represents a logger entry with fields
type Entry struct {
	fields Fields
}

func (e *Entry) Debug(msg string) {
	Debug(msg, e.fields)
}

func (e *Entry) Info(msg string) {
	Info(msg, e.fields)
}

func (e *Entry) Warn(msg string) {
	Warn(msg, e.fields)
}

func (e *Entry) Error(msg string, err error) {
	Error(msg, err, e.fields)
}

func (e *Entry) Fatal(msg string, err error) {
	Fatal(msg, err, e.fields)
}

// Helper functions for common logging patterns

// LogRequest logs an HTTP request
func LogRequest(method, path, clientIP string, statusCode int, latency time.Duration, fields ...Fields) {
	var f Fields
	if len(fields) > 0 {
		f = fields[0]
	} else {
		f = make(Fields)
	}

	f["method"] = method
	f["path"] = path
	f["client_ip"] = clientIP
	f["status_code"] = statusCode
	f["latency_ms"] = latency.Milliseconds()

	if statusCode >= 500 {
		Error("HTTP request completed with server error", nil, f)
	} else if statusCode >= 400 {
		Warn("HTTP request completed with client error", f)
	} else {
		Info("HTTP request completed", f)
	}
}

// LogDatabaseOperation logs a database operation
func LogDatabaseOperation(operation, table string, err error, duration time.Duration, fields ...Fields) {
	var f Fields
	if len(fields) > 0 {
		f = fields[0]
	} else {
		f = make(Fields)
	}

	f["operation"] = operation
	f["table"] = table
	f["duration_ms"] = duration.Milliseconds()

	if err != nil {
		Error(fmt.Sprintf("Database operation failed: %s on %s", operation, table), err, f)
	} else {
		Debug(fmt.Sprintf("Database operation completed: %s on %s", operation, table), f)
	}
}

// LogWebSocketEvent logs a WebSocket event
func LogWebSocketEvent(eventType, userID, roomID string, fields ...Fields) {
	var f Fields
	if len(fields) > 0 {
		f = fields[0]
	} else {
		f = make(Fields)
	}

	f["event_type"] = eventType
	f["user_id"] = userID
	if roomID != "" {
		f["room_id"] = roomID
	}

	Info(fmt.Sprintf("WebSocket event: %s", eventType), f)
}

// LogAuthEvent logs an authentication event
func LogAuthEvent(eventType, userID string, success bool, fields ...Fields) {
	var f Fields
	if len(fields) > 0 {
		f = fields[0]
	} else {
		f = make(Fields)
	}

	f["event_type"] = eventType
	f["user_id"] = userID
	f["success"] = success

	if success {
		Info(fmt.Sprintf("Authentication event: %s", eventType), f)
	} else {
		Warn(fmt.Sprintf("Authentication event failed: %s", eventType), f)
	}
}
