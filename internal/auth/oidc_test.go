package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

func TestEnrichIdentityClaimsFromUserInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer access-token" {
			t.Fatalf("expected bearer token, got %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sub":"subject-1","email":"reader@example.test","email_verified":true,"name":"Sabeel Reader","preferred_username":"reader"}`))
	}))
	defer server.Close()

	flow := &oidcFlow{userInfoURL: server.URL}
	claims := identityClaims{Subject: "subject-1"}
	if err := flow.enrichIdentityClaims(context.Background(), &oauth2.Token{AccessToken: "access-token", TokenType: "Bearer"}, &claims); err != nil {
		t.Fatalf("enrich identity claims: %v", err)
	}
	if claims.Email != "reader@example.test" || !claims.EmailVerified || claims.Name != "Sabeel Reader" || claims.PreferredUsername != "reader" {
		t.Fatalf("unexpected enriched claims: %+v", claims)
	}
}

func TestEnrichIdentityClaimsRejectsSubjectMismatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sub":"subject-2","email":"reader@example.test","email_verified":true}`))
	}))
	defer server.Close()

	flow := &oidcFlow{userInfoURL: server.URL}
	claims := identityClaims{Subject: "subject-1"}
	err := flow.enrichIdentityClaims(context.Background(), &oauth2.Token{AccessToken: "access-token", TokenType: "Bearer"}, &claims)
	if err == nil || !strings.Contains(err.Error(), "subject does not match") {
		t.Fatalf("expected subject mismatch, got %v", err)
	}
}
