package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

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

type Fields map[string]interface{}

func Initialize(level string, format string) {
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

	logger = log.New(os.Stdout, "", 0)
}

func shouldLog(level LogLevel) bool {
	return level >= currentLevel
}

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

func getCallerInfo() (string, int) {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		return "unknown", 0
	}

	parts := strings.Split(file, "/")
	filename := parts[len(parts)-1]

	return filename, line
}

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

func WithFields(fields Fields) *Entry {
	return &Entry{fields: fields}
}

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
