package auth

import (
	"strings"
	"testing"
)

func TestOIDCUsernameIsStableAndValid(t *testing.T) {
	first := oidcUsername("Reader.Name@example.test", "reader@example.test", "https://identity.example", "subject")
	second := oidcUsername("Reader.Name@example.test", "reader@example.test", "https://identity.example", "subject")
	if first != second || !strings.HasPrefix(first, "reader-name-example-") || len(first) > 32 {
		t.Fatalf("unexpected oidc username: %q", first)
	}
}
