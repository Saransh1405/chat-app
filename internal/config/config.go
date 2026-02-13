package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server      ServerConfig
	Kafka       KafkaConfig
	Database    DatabaseConfig
	JWT         JWTConfig
	CORS        CORSConfig
	Redis       RedisConfig
	FileStorage FileStorageConfig
	Logging     LoggingConfig
	Environment string
	ImageKit    ImageKitConfig
}

type ServerConfig struct {
	Host string
	Port string
}

type DatabaseConfig struct {
	Host                  string
	Port                  string
	User                  string
	Password              string
	Name                  string
	URL                   string
	SSLMode               string
	MaxConnections        int
	MaxIdleConnections    int
	ConnectionMaxLifetime time.Duration
}

type JWTConfig struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
	Enabled  bool
}

type KafkaConfig struct {
	Brokers []string
	Topic   string
}

type FileStorageConfig struct {
	Path             string
	MaxFileSize      int64
	AllowedFileTypes []string
}

type LoggingConfig struct {
	Level  string
	Format string
}

type ImageKitConfig struct {
	PublicKey   string
	PrivateKey  string
	UrlEndpoint string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:                  getEnv("DB_HOST", "localhost"),
			Port:                  getEnv("DB_PORT", "5432"),
			User:                  getEnv("DB_USER", "chat_app_user"),
			Password:              getEnv("DB_PASSWORD", "chat_app_password"),
			URL:                   getEnv("DB_URL", ""),
			Name:                  getEnv("DB_NAME", "chat_app_db"),
			SSLMode:               getEnv("DB_SSL_MODE", "disable"),
			MaxConnections:        getEnvAsInt("DB_MAX_CONNECTIONS", 25),
			MaxIdleConnections:    getEnvAsInt("DB_MAX_IDLE_CONNECTIONS", 5),
			ConnectionMaxLifetime: getEnvAsDuration("DB_CONNECTION_MAX_LIFETIME", "300s"),
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
			AccessExpiry:  getEnvAsDuration("JWT_ACCESS_EXPIRY", "15m"),
			RefreshExpiry: getEnvAsDuration("JWT_REFRESH_EXPIRY", "7d"),
		},
		CORS: CORSConfig{
			AllowedOrigins:   getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
			AllowedMethods:   getEnvAsSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
			AllowedHeaders:   getEnvAsSlice("CORS_ALLOWED_HEADERS", []string{"Content-Type", "Authorization", "X-Requested-With"}),
			AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", true),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
			Enabled:  getEnvAsBool("REDIS_ENABLED", false),
		},
		FileStorage: FileStorageConfig{
			Path:             getEnv("FILE_STORAGE_PATH", "./uploads"),
			MaxFileSize:      getEnvAsInt64("MAX_FILE_SIZE", 10485760),
			AllowedFileTypes: getEnvAsSlice("ALLOWED_FILE_TYPES", []string{"image/jpeg", "image/png", "image/gif", "application/pdf", "text/plain"}),
		},
		Kafka: KafkaConfig{
			Brokers: getEnvAsSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
			Topic:   getEnv("KAFKA_TOPIC", "chat-app"),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "text"),
		},
		Environment: getEnv("ENVIRONMENT", "development"),
		ImageKit: ImageKitConfig{
			PublicKey:   getEnv("IMAGE_KIT_PUBLIC_KEY", ""),
			PrivateKey:  getEnv("IMAGE_KIT_PRIVATE_KEY", ""),
			UrlEndpoint: getEnv("IMAGE_KIT_ENDPOINT_URL", ""),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.JWT.Secret == "" || c.JWT.Secret == "your-super-secret-jwt-key-change-in-production" {
		if c.Environment == "production" {
			return fmt.Errorf("JWT_SECRET must be set in production")
		}
	}

	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}

	if c.Database.Name == "" {
		return fmt.Errorf("DB_NAME is required")
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsDuration(key string, defaultValue string) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		valueStr = defaultValue
	}
	duration, err := time.ParseDuration(valueStr)
	if err != nil {
		// Try to parse as seconds
		if seconds, err := strconv.Atoi(valueStr); err == nil {
			return time.Duration(seconds) * time.Second
		}
		// Fallback to default
		if defaultDuration, err := time.ParseDuration(defaultValue); err == nil {
			return defaultDuration
		}
		return 300 * time.Second // Ultimate fallback
	}
	return duration
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	var result []string
	for _, item := range splitString(valueStr, ",") {
		trimmed := trimString(item)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	if len(result) == 0 {
		return defaultValue
	}
	return result
}

func splitString(s, sep string) []string {
	var result []string
	current := ""
	for _, char := range s {
		if string(char) == sep {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func trimString(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
