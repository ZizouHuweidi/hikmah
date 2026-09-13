package config

import (
	"fmt"
	"net/url"
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
	Issuer          string
	InternalURL     string
	ClientID        string
	ClientSecret    string
	RedirectURL     string
	StateKey        string
	PublicURL       string
	SessionLifetime time.Duration
	CookieSecure    bool
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
	sessionLifetime, err := getDurationEnv("SABEEL_SESSION_LIFETIME", 12*time.Hour)
	if err != nil {
		return nil, err
	}
	cookieSecure, err := getBoolEnv("SABEEL_SESSION_SECURE", false)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
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
			Issuer:          strings.TrimSuffix(getEnv("SABEEL_OIDC_ISSUER", "http://127.0.0.1:8081"), "/"),
			InternalURL:     strings.TrimSuffix(getEnv("SABEEL_OIDC_INTERNAL_URL", "http://zitadel-dev-proxy:8080"), "/"),
			ClientID:        strings.TrimSpace(os.Getenv("SABEEL_OIDC_CLIENT_ID")),
			ClientSecret:    strings.TrimSpace(os.Getenv("SABEEL_OIDC_CLIENT_SECRET")),
			RedirectURL:     getEnv("SABEEL_OIDC_REDIRECT_URL", "http://localhost:8080/auth/callback"),
			StateKey:        strings.TrimSpace(os.Getenv("SABEEL_OIDC_STATE_KEY")),
			PublicURL:       strings.TrimSuffix(getEnv("SABEEL_PUBLIC_URL", "http://localhost:3000"), "/"),
			SessionLifetime: sessionLifetime,
			CookieSecure:    cookieSecure,
		},
	}
	if !absoluteHTTPURL(cfg.Auth.Issuer) || !absoluteHTTPURL(cfg.Auth.InternalURL) || !absoluteHTTPURL(cfg.Auth.RedirectURL) || !absoluteHTTPURL(cfg.Auth.PublicURL) {
		return nil, fmt.Errorf("Sabeel OIDC and public URLs must be absolute http or https URLs")
	}
	if cfg.Auth.ClientID == "" || cfg.Auth.ClientSecret == "" {
		return nil, fmt.Errorf("SABEEL_OIDC_CLIENT_ID and SABEEL_OIDC_CLIENT_SECRET are required")
	}
	if len(cfg.Auth.StateKey) < 32 {
		return nil, fmt.Errorf("SABEEL_OIDC_STATE_KEY must contain at least 32 characters")
	}
	if cfg.Auth.SessionLifetime <= 0 {
		return nil, fmt.Errorf("SABEEL_SESSION_LIFETIME must be positive")
	}
	if cfg.Environment == "production" && !cfg.Auth.CookieSecure {
		return nil, fmt.Errorf("production requires secure auth cookies")
	}
	if cfg.Environment == "production" && (!httpsURL(cfg.Auth.Issuer) || !httpsURL(cfg.Auth.RedirectURL) || !httpsURL(cfg.Auth.PublicURL)) {
		return nil, fmt.Errorf("production requires HTTPS issuer, redirect, and public URLs")
	}
	return cfg, nil
}

func absoluteHTTPURL(raw string) bool {
	parsed, err := url.Parse(raw)
	return err == nil && parsed.Host != "" && (parsed.Scheme == "http" || parsed.Scheme == "https")
}

func httpsURL(raw string) bool {
	parsed, err := url.Parse(raw)
	return err == nil && parsed.Host != "" && parsed.Scheme == "https"
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
