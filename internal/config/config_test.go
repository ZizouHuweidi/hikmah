package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setValidAuthEnv(t *testing.T) {
	t.Helper()
	t.Setenv("SABEEL_OIDC_CLIENT_ID", "sabeel-web")
	t.Setenv("SABEEL_OIDC_CLIENT_SECRET", "client-secret")
	t.Setenv("SABEEL_OIDC_STATE_KEY", "0123456789abcdef0123456789abcdef")
}

func TestLoadRequiresOIDCCredentials(t *testing.T) {
	t.Setenv("SABEEL_OIDC_CLIENT_ID", "")
	t.Setenv("SABEEL_OIDC_CLIENT_SECRET", "")
	t.Setenv("SABEEL_OIDC_STATE_KEY", "0123456789abcdef0123456789abcdef")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "SABEEL_OIDC_CLIENT_ID") {
		t.Fatalf("expected missing OIDC credentials error, got %v", err)
	}
}

func TestLoadReadsOIDCCredentialsFromFiles(t *testing.T) {
	t.Setenv("SABEEL_OIDC_CLIENT_ID", "")
	t.Setenv("SABEEL_OIDC_CLIENT_SECRET", "")
	t.Setenv("SABEEL_OIDC_STATE_KEY", "0123456789abcdef0123456789abcdef")
	dir := t.TempDir()
	idPath := filepath.Join(dir, "client-id")
	secretPath := filepath.Join(dir, "client-secret")
	if err := os.WriteFile(idPath, []byte("sabeel-file-client\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(secretPath, []byte("file-secret\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SABEEL_OIDC_CLIENT_ID_FILE", idPath)
	t.Setenv("SABEEL_OIDC_CLIENT_SECRET_FILE", secretPath)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Auth.ClientID != "sabeel-file-client" || cfg.Auth.ClientSecret != "file-secret" {
		t.Fatalf("unexpected file credentials: %+v", cfg.Auth)
	}
}

func TestLoadPrefersOIDCCredentialsOverFiles(t *testing.T) {
	setValidAuthEnv(t)
	t.Setenv("SABEEL_OIDC_CLIENT_ID_FILE", filepath.Join(t.TempDir(), "missing"))
	t.Setenv("SABEEL_OIDC_CLIENT_SECRET_FILE", filepath.Join(t.TempDir(), "missing"))

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Auth.ClientID != "sabeel-web" || cfg.Auth.ClientSecret != "client-secret" {
		t.Fatalf("unexpected direct credentials: %+v", cfg.Auth)
	}
}

func TestLoadRejectsShortStateKey(t *testing.T) {
	setValidAuthEnv(t)
	t.Setenv("SABEEL_OIDC_STATE_KEY", "too-short")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "at least 32 characters") {
		t.Fatalf("expected short state key error, got %v", err)
	}
}

func TestLoadRequiresSecureProductionCookies(t *testing.T) {
	setValidAuthEnv(t)
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("SABEEL_SESSION_SECURE", "false")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "secure auth cookies") {
		t.Fatalf("expected secure cookie error, got %v", err)
	}
}

func TestLoadAcceptsProductionOIDCConfiguration(t *testing.T) {
	setValidAuthEnv(t)
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("SABEEL_SESSION_SECURE", "true")
	t.Setenv("SABEEL_OIDC_ISSUER", "https://identity.example/")
	t.Setenv("SABEEL_OIDC_INTERNAL_URL", "http://zitadel:8080/")
	t.Setenv("SABEEL_OIDC_REDIRECT_URL", "https://api.sabeel.example/auth/callback")
	t.Setenv("SABEEL_PUBLIC_URL", "https://sabeel.example")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Auth.Issuer != "https://identity.example" || cfg.Auth.InternalURL != "http://zitadel:8080" || !cfg.Auth.CookieSecure {
		t.Fatalf("unexpected auth configuration: %+v", cfg.Auth)
	}
}

func TestLoadRejectsHTTPProductionURLs(t *testing.T) {
	setValidAuthEnv(t)
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("SABEEL_SESSION_SECURE", "true")

	_, err := Load()
	if err == nil || !strings.Contains(err.Error(), "HTTPS issuer") {
		t.Fatalf("expected production HTTPS error, got %v", err)
	}
}
