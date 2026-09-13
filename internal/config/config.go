package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the service
type Config struct {
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	Auth        AuthConfig
}

type ServerConfig struct {
	Port               string
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	CORSAllowedOrigins []string
}

type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type AuthConfig struct {
	Issuer               string
	Audience             string
	Ed25519PrivateKey    string
	AccessTokenLifetime  time.Duration
	RefreshTokenLifetime time.Duration
	CookieSecure         bool
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	readTimeout, err := getDurationEnv("SERVER_READ_TIMEOUT", 30*time.Second)
	if err != nil {
		return nil, err
	}
	writeTimeout, err := getDurationEnv("SERVER_WRITE_TIMEOUT", 30*time.Second)
	if err != nil {
		return nil, err
	}
	maxOpenConns, err := getIntEnv("DB_MAX_OPEN_CONNS", 25)
	if err != nil {
		return nil, err
	}
	maxIdleConns, err := getIntEnv("DB_MAX_IDLE_CONNS", 5)
	if err != nil {
		return nil, err
	}
	connMaxLifetime, err := getDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute)
	if err != nil {
		return nil, err
	}
	accessTokenLifetime, err := getDurationEnv("AUTH_ACCESS_TOKEN_LIFETIME", 15*time.Minute)
	if err != nil {
		return nil, err
	}
	refreshTokenLifetime, err := getDurationEnv("AUTH_REFRESH_TOKEN_LIFETIME", 720*time.Hour)
	if err != nil {
		return nil, err
	}
	cookieSecure, err := getBoolEnv("AUTH_COOKIE_SECURE", false)
	if err != nil {
		return nil, err
	}

	return &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Server: ServerConfig{
			Port:               getEnv("PORT", "8080"),
			ReadTimeout:        readTimeout,
			WriteTimeout:       writeTimeout,
			CORSAllowedOrigins: getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://127.0.0.1:5173", "http://localhost:5173"}),
		},
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", "postgres://maktaba:maktaba@localhost:5432/maktaba?sslmode=disable"),
			MaxOpenConns:    maxOpenConns,
			MaxIdleConns:    maxIdleConns,
			ConnMaxLifetime: connMaxLifetime,
		},
		Auth: AuthConfig{
			Issuer:               getEnv("AUTH_ISSUER", "sabeel"),
			Audience:             getEnv("AUTH_AUDIENCE", "sabeel-api"),
			Ed25519PrivateKey:    getEnv("AUTH_ED25519_PRIVATE_KEY", ""),
			AccessTokenLifetime:  accessTokenLifetime,
			RefreshTokenLifetime: refreshTokenLifetime,
			CookieSecure:         cookieSecure,
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) (int, error) {
	if value := os.Getenv(key); value != "" {
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return 0, fmt.Errorf("invalid %s: %w", key, err)
		}
		return intVal, nil
	}
	return defaultValue, nil
}

func getDurationEnv(key string, defaultValue time.Duration) (time.Duration, error) {
	if value := os.Getenv(key); value != "" {
		d, err := time.ParseDuration(value)
		if err != nil {
			return 0, fmt.Errorf("invalid %s: %w", key, err)
		}
		return d, nil
	}
	return defaultValue, nil
}

func getBoolEnv(key string, defaultValue bool) (bool, error) {
	if value := os.Getenv(key); value != "" {
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return false, fmt.Errorf("invalid %s: %w", key, err)
		}
		return boolVal, nil
	}
	return defaultValue, nil
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		parts := strings.Split(value, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				result = append(result, part)
			}
		}
		return result
	}
	return defaultValue
}
