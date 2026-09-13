package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type OIDCIdentity struct {
	Issuer            string
	Subject           string
	Email             string
	EmailVerified     bool
	DisplayName       string
	PreferredUsername string
}

type OIDCFlow interface {
	AuthorizationURL(state, nonce, verifier string, createAccount bool) string
	Authenticate(ctx context.Context, code, verifier, nonce string) (OIDCIdentity, error)
}

type OIDCConfig struct {
	Issuer       string
	InternalURL  string
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type oidcFlow struct {
	issuer   string
	oauth    oauth2.Config
	verifier *oidc.IDTokenVerifier
}

func NewOIDCFlow(ctx context.Context, config OIDCConfig) (OIDCFlow, error) {
	issuer := strings.TrimSuffix(config.Issuer, "/")
	internalURL := strings.TrimSuffix(config.InternalURL, "/")
	if issuer == "" || internalURL == "" || config.ClientID == "" || config.ClientSecret == "" || config.RedirectURL == "" {
		return nil, errors.New("incomplete oidc configuration")
	}
	keySet := oidc.NewRemoteKeySet(ctx, internalURL+"/oauth/v2/keys")
	return &oidcFlow{
		issuer: issuer,
		oauth: oauth2.Config{
			ClientID: config.ClientID, ClientSecret: config.ClientSecret, RedirectURL: config.RedirectURL,
			Endpoint: oauth2.Endpoint{AuthURL: issuer + "/oauth/v2/authorize", TokenURL: internalURL + "/oauth/v2/token"},
			Scopes:   []string{oidc.ScopeOpenID, "profile", "email"},
		},
		verifier: oidc.NewVerifier(issuer, keySet, &oidc.Config{ClientID: config.ClientID}),
	}, nil
}

func (f *oidcFlow) AuthorizationURL(state, nonce, verifier string, createAccount bool) string {
	options := []oauth2.AuthCodeOption{oauth2.S256ChallengeOption(verifier), oauth2.SetAuthURLParam("nonce", nonce)}
	if createAccount {
		options = append(options, oauth2.SetAuthURLParam("prompt", "create"))
	}
	return f.oauth.AuthCodeURL(state, options...)
}

func (f *oidcFlow) Authenticate(ctx context.Context, code, verifier, nonce string) (OIDCIdentity, error) {
	token, err := f.oauth.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return OIDCIdentity{}, fmt.Errorf("exchange authorization code: %w", err)
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return OIDCIdentity{}, errors.New("id token is missing")
	}
	idToken, err := f.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return OIDCIdentity{}, fmt.Errorf("verify id token: %w", err)
	}
	var claims struct {
		Issuer            string `json:"iss"`
		Subject           string `json:"sub"`
		Email             string `json:"email"`
		EmailVerified     bool   `json:"email_verified"`
		Name              string `json:"name"`
		PreferredUsername string `json:"preferred_username"`
		Nonce             string `json:"nonce"`
		NotBefore         int64  `json:"nbf"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return OIDCIdentity{}, fmt.Errorf("decode id token claims: %w", err)
	}
	if claims.Issuer != f.issuer || claims.Subject == "" || claims.NotBefore > time.Now().Unix() || !hmac.Equal([]byte(claims.Nonce), []byte(nonce)) {
		return OIDCIdentity{}, errors.New("id token claims do not match login state")
	}
	displayName := strings.TrimSpace(claims.Name)
	if displayName == "" {
		displayName = strings.TrimSpace(claims.PreferredUsername)
	}
	return OIDCIdentity{
		Issuer: claims.Issuer, Subject: claims.Subject, Email: claims.Email,
		EmailVerified: claims.EmailVerified, DisplayName: displayName,
		PreferredUsername: claims.PreferredUsername,
	}, nil
}

type oauthState struct {
	State     string    `json:"state"`
	Nonce     string    `json:"nonce"`
	Verifier  string    `json:"verifier"`
	ReturnTo  string    `json:"returnTo"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func signOAuthState(key []byte, value oauthState) (string, error) {
	if len(key) < 32 {
		return "", errors.New("oauth state key is too short")
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(encoded))
	return encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func verifyOAuthState(key []byte, raw string, now time.Time) (oauthState, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 2 {
		return oauthState{}, errors.New("invalid oauth state format")
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return oauthState{}, errors.New("invalid oauth state signature")
	}
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(parts[0]))
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return oauthState{}, errors.New("oauth state signature does not match")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return oauthState{}, errors.New("invalid oauth state payload")
	}
	var value oauthState
	if err := json.Unmarshal(payload, &value); err != nil || value.State == "" || value.Nonce == "" || value.Verifier == "" || !value.ExpiresAt.After(now) {
		return oauthState{}, errors.New("oauth state is invalid or expired")
	}
	value.ReturnTo = safeReturnTo(value.ReturnTo)
	return value, nil
}

func safeReturnTo(value string) string {
	if !strings.HasPrefix(value, "/") || strings.HasPrefix(value, "//") || strings.Contains(value, "\\") {
		return "/"
	}
	return value
}
