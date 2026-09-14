package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type config struct {
	apiURL         string
	instanceHost   string
	forwardedProto string
	patPath        string
	outputDir      string
	projectName    string
	appName        string
	redirectURI    string
	logoutURI      string
	devMode        bool
}

type client struct {
	baseURL        string
	instanceHost   string
	forwardedProto string
	personalToken  string
	httpClient     *http.Client
}

type project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type application struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	OIDCConfig struct {
		ClientID string `json:"clientId"`
	} `json:"oidcConfig"`
}

func main() {
	log.SetFlags(0)
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}
	if credentialsCurrent(cfg) {
		log.Print("Zitadel client credentials already exist")
		return
	}
	pat, err := os.ReadFile(cfg.patPath)
	if err != nil {
		log.Fatalf("read bootstrap PAT: %v", err)
	}
	api := &client{
		baseURL: cfg.apiURL, instanceHost: cfg.instanceHost,
		forwardedProto: cfg.forwardedProto, personalToken: strings.TrimSpace(string(pat)),
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	projectID, err := ensureProject(ctx, api, cfg.projectName)
	if err != nil {
		log.Fatalf("ensure Zitadel project: %v", err)
	}
	clientID, clientSecret, err := ensureApplication(ctx, api, projectID, cfg)
	if err != nil {
		log.Fatalf("ensure Zitadel application: %v", err)
	}
	if err := writeCredential(cfg.outputDir, "client-id", clientID); err != nil {
		log.Fatal(err)
	}
	if err := writeCredential(cfg.outputDir, "client-secret", clientSecret); err != nil {
		log.Fatal(err)
	}
	if err := writeBootstrapConfig(cfg); err != nil {
		log.Fatal(err)
	}
	log.Printf("Zitadel application %q is ready", cfg.appName)
}

func loadConfig() (config, error) {
	devMode := strings.EqualFold(env("ZITADEL_BOOTSTRAP_DEV_MODE", "false"), "true")
	cfg := config{
		apiURL:         strings.TrimSuffix(env("ZITADEL_BOOTSTRAP_API_URL", "http://zitadel-api:8080"), "/"),
		instanceHost:   env("ZITADEL_BOOTSTRAP_INSTANCE_HOST", "127.0.0.1:8081"),
		forwardedProto: env("ZITADEL_BOOTSTRAP_FORWARDED_PROTO", "http"),
		patPath:        env("ZITADEL_BOOTSTRAP_PAT_FILE", "/zitadel/bootstrap/admin.pat"),
		outputDir:      env("ZITADEL_BOOTSTRAP_OUTPUT_DIR", "/zitadel/bootstrap/sabeel"),
		projectName:    env("ZITADEL_BOOTSTRAP_PROJECT_NAME", "Sabeel"),
		appName:        env("ZITADEL_BOOTSTRAP_APP_NAME", "sabeel-web"),
		redirectURI:    strings.TrimSpace(os.Getenv("ZITADEL_BOOTSTRAP_REDIRECT_URI")),
		logoutURI:      strings.TrimSpace(os.Getenv("ZITADEL_BOOTSTRAP_LOGOUT_URI")),
		devMode:        devMode,
	}
	parsed, err := url.Parse(cfg.apiURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return config{}, errors.New("ZITADEL_BOOTSTRAP_API_URL must be an absolute URL")
	}
	if cfg.instanceHost == "" || cfg.redirectURI == "" || cfg.logoutURI == "" {
		return config{}, errors.New("instance host, redirect URI, and logout URI are required")
	}
	return cfg, nil
}

func ensureProject(ctx context.Context, api *client, name string) (string, error) {
	var listed struct {
		Result []project `json:"result"`
	}
	if err := api.callWithRetry(ctx, http.MethodPost, "/management/v1/projects/_search", map[string]any{"query": map[string]any{"limit": 100}}, &listed); err != nil {
		return "", err
	}
	for _, candidate := range listed.Result {
		if candidate.Name == name {
			return candidate.ID, nil
		}
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := api.call(ctx, http.MethodPost, "/management/v1/projects", map[string]any{
		"name":                   name,
		"privateLabelingSetting": "PRIVATE_LABELING_SETTING_ENFORCE_PROJECT_RESOURCE_OWNER_POLICY",
	}, &created); err != nil {
		return "", err
	}
	if created.ID == "" {
		return "", errors.New("create project returned no ID")
	}
	return created.ID, nil
}

func ensureApplication(ctx context.Context, api *client, projectID string, cfg config) (string, string, error) {
	path := "/management/v1/projects/" + url.PathEscape(projectID) + "/apps/_search"
	var listed struct {
		Result []application `json:"result"`
	}
	if err := api.call(ctx, http.MethodPost, path, map[string]any{"query": map[string]any{"limit": 100}}, &listed); err != nil {
		return "", "", err
	}
	for _, candidate := range listed.Result {
		if candidate.Name != cfg.appName {
			continue
		}
		if candidate.OIDCConfig.ClientID == "" {
			return "", "", fmt.Errorf("application %q exists but is not an OIDC application", cfg.appName)
		}
		configPath := "/management/v1/projects/" + url.PathEscape(projectID) + "/apps/" + url.PathEscape(candidate.ID) + "/oidc_config"
		if err := api.call(ctx, http.MethodPut, configPath, oidcConfiguration(cfg), nil); err != nil {
			return "", "", err
		}
		var generated struct {
			ClientSecret string `json:"clientSecret"`
		}
		secretPath := "/management/v1/projects/" + url.PathEscape(projectID) + "/apps/" + url.PathEscape(candidate.ID) + "/oidc_config/_generate_client_secret"
		if err := api.call(ctx, http.MethodPost, secretPath, nil, &generated); err != nil {
			return "", "", err
		}
		return candidate.OIDCConfig.ClientID, generated.ClientSecret, nil
	}

	var created struct {
		ClientID     string `json:"clientId"`
		ClientSecret string `json:"clientSecret"`
	}
	createPath := "/management/v1/projects/" + url.PathEscape(projectID) + "/apps/oidc"
	body := oidcConfiguration(cfg)
	body["name"] = cfg.appName
	body["version"] = "OIDC_VERSION_1_0"
	if err := api.call(ctx, http.MethodPost, createPath, body, &created); err != nil {
		return "", "", err
	}
	if created.ClientID == "" || created.ClientSecret == "" {
		return "", "", errors.New("create OIDC application returned incomplete credentials")
	}
	return created.ClientID, created.ClientSecret, nil
}

func oidcConfiguration(cfg config) map[string]any {
	return map[string]any{
		"redirectUris":           []string{cfg.redirectURI},
		"responseTypes":          []string{"OIDC_RESPONSE_TYPE_CODE"},
		"grantTypes":             []string{"OIDC_GRANT_TYPE_AUTHORIZATION_CODE"},
		"appType":                "OIDC_APP_TYPE_WEB",
		"authMethodType":         "OIDC_AUTH_METHOD_TYPE_BASIC",
		"postLogoutRedirectUris": []string{cfg.logoutURI},
		"devMode":                cfg.devMode,
	}
}

func (c *client) callWithRetry(ctx context.Context, method, path string, body, target any) error {
	var lastErr error
	for {
		if err := c.call(ctx, method, path, body, target); err == nil {
			return nil
		} else {
			lastErr = err
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("Zitadel did not become ready: %w", lastErr)
		case <-time.After(2 * time.Second):
		}
	}
}

func (c *client) call(ctx context.Context, method, path string, body, target any) error {
	var requestBody io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		requestBody = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, requestBody)
	if err != nil {
		return err
	}
	request.Host = c.instanceHost
	request.Header.Set("Authorization", "Bearer "+c.personalToken)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Forwarded-Host", c.instanceHost)
	request.Header.Set("X-Forwarded-Proto", c.forwardedProto)
	response, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	payload, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("%s %s returned %s: %s", method, path, response.Status, strings.TrimSpace(string(payload)))
	}
	if target != nil && len(payload) > 0 {
		if err := json.Unmarshal(payload, target); err != nil {
			return fmt.Errorf("decode %s: %w", path, err)
		}
	}
	return nil
}

func credentialsCurrent(cfg config) bool {
	for _, name := range []string{"client-id", "client-secret"} {
		value, err := os.ReadFile(filepath.Join(cfg.outputDir, name))
		if err != nil || strings.TrimSpace(string(value)) == "" {
			return false
		}
	}
	value, err := os.ReadFile(filepath.Join(cfg.outputDir, "configuration.json"))
	if err != nil {
		return false
	}
	var recorded struct {
		RedirectURI string `json:"redirectUri"`
		LogoutURI   string `json:"logoutUri"`
		DevMode     bool   `json:"devMode"`
	}
	if json.Unmarshal(value, &recorded) != nil {
		return false
	}
	return recorded.RedirectURI == cfg.redirectURI && recorded.LogoutURI == cfg.logoutURI && recorded.DevMode == cfg.devMode
}

func writeBootstrapConfig(cfg config) error {
	value, err := json.MarshalIndent(map[string]any{
		"redirectUri": cfg.redirectURI,
		"logoutUri":   cfg.logoutURI,
		"devMode":     cfg.devMode,
	}, "", "  ")
	if err != nil {
		return err
	}
	return writeCredential(cfg.outputDir, "configuration.json", string(value))
}

func writeCredential(dir, name, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("refusing to write empty %s", name)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create credential directory: %w", err)
	}
	temporary := filepath.Join(dir, "."+name+".tmp")
	if err := os.WriteFile(temporary, []byte(value+"\n"), 0o600); err != nil {
		return fmt.Errorf("write %s: %w", name, err)
	}
	if err := os.Rename(temporary, filepath.Join(dir, name)); err != nil {
		return fmt.Errorf("publish %s: %w", name, err)
	}
	return nil
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
