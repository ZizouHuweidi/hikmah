package auth

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/labstack/echo/v5"
)

type fakeRepository struct {
	user        *User
	identity    OIDCIdentity
	createdHash []byte
	deletedHash []byte
}

func (r *fakeRepository) ProvisionOIDCUser(_ context.Context, identity OIDCIdentity) (*User, error) {
	r.identity = identity
	if r.user == nil {
		return nil, errors.New("missing user")
	}
	return r.user, nil
}
func (r *fakeRepository) GetUserByID(context.Context, uuid.UUID) (*User, error) { return r.user, nil }
func (r *fakeRepository) CreateSession(_ context.Context, _ uuid.UUID, hash []byte, _ time.Time) error {
	r.createdHash = append([]byte(nil), hash...)
	return nil
}
func (r *fakeRepository) UserForSession(context.Context, []byte) (*User, error) {
	if r.user == nil {
		return nil, ErrInvalidSession
	}
	return r.user, nil
}
func (r *fakeRepository) DeleteSession(_ context.Context, hash []byte) error {
	r.deletedHash = append([]byte(nil), hash...)
	return nil
}

type fakeOIDCFlow struct {
	identity OIDCIdentity
	err      error
	state    string
	nonce    string
	verifier string
	create   bool
}

func (f *fakeOIDCFlow) AuthorizationURL(state, nonce, verifier string, create bool) string {
	f.state, f.nonce, f.verifier, f.create = state, nonce, verifier, create
	return "https://identity.example/authorize?state=" + url.QueryEscape(state)
}
func (f *fakeOIDCFlow) Authenticate(_ context.Context, _ string, verifier, nonce string) (OIDCIdentity, error) {
	if verifier != f.verifier || nonce != f.nonce {
		return OIDCIdentity{}, errors.New("unexpected oidc parameters")
	}
	return f.identity, f.err
}

func testHandler(repository Repository, flow OIDCFlow) (*echo.Echo, *Handler) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewHandler(repository, flow, HandlerConfig{
		SessionDuration: time.Hour,
		TrustedOrigins:  []string{"https://sabeel.example"},
		StateKey:        "0123456789abcdef0123456789abcdef",
		PublicURL:       "https://sabeel.example",
		OIDCIssuer:      "https://identity.example",
		OIDCClientID:    "sabeel-web",
	}, logger)
	e := echo.New()
	handler.RegisterRoutes(e)
	return e, handler
}

func TestLoginUsesSignedStateAndSafeReturnPath(t *testing.T) {
	flow := &fakeOIDCFlow{}
	e, _ := testHandler(&fakeRepository{}, flow)
	request := httptest.NewRequest(http.MethodGet, "/auth/login?return_to=https://attacker.example", nil)
	response := httptest.NewRecorder()
	e.ServeHTTP(response, request)
	if response.Code != http.StatusFound || !strings.Contains(response.Header().Get("Location"), "identity.example") {
		t.Fatalf("unexpected login response: %d %s", response.Code, response.Header().Get("Location"))
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != OAuthStateCookieName || !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteLaxMode {
		t.Fatalf("unexpected oauth state cookie: %+v", cookies)
	}
	state, err := verifyOAuthState([]byte("0123456789abcdef0123456789abcdef"), cookies[0].Value, time.Now())
	if err != nil || state.ReturnTo != "/" || state.State != flow.state {
		t.Fatalf("expected safe signed state, got %+v, %v", state, err)
	}
}

func TestRegistrationRequestsZitadelCreatePrompt(t *testing.T) {
	flow := &fakeOIDCFlow{}
	e, _ := testHandler(&fakeRepository{}, flow)
	e.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/auth/register", nil))
	if !flow.create {
		t.Fatal("expected registration to request account creation")
	}
}

func TestCallbackCreatesOpaqueSession(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	identity := OIDCIdentity{Issuer: "https://identity.example", Subject: "subject-1", Email: "reader@example.test", EmailVerified: true}
	flow := &fakeOIDCFlow{identity: identity}
	repository := &fakeRepository{user: &User{ID: userID, Email: identity.Email}}
	e, _ := testHandler(repository, flow)

	login := httptest.NewRecorder()
	e.ServeHTTP(login, httptest.NewRequest(http.MethodGet, "/auth/login?return_to=/dashboard", nil))
	callback := httptest.NewRequest(http.MethodGet, "/auth/callback?code=code&state="+url.QueryEscape(flow.state), nil)
	callback.AddCookie(login.Result().Cookies()[0])
	response := httptest.NewRecorder()
	e.ServeHTTP(response, callback)
	if response.Code != http.StatusFound || response.Header().Get("Location") != "https://sabeel.example/dashboard" {
		t.Fatalf("unexpected callback response: %d %s", response.Code, response.Header().Get("Location"))
	}
	var session *http.Cookie
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == SessionCookieName {
			session = cookie
		}
	}
	if session == nil || !session.HttpOnly || len(repository.createdHash) != 32 {
		t.Fatalf("unexpected session cookie or hash: %+v", session)
	}
	if string(repository.createdHash) != string(HashSessionToken(session.Value)) || repository.identity.Subject != identity.Subject {
		t.Fatal("expected an opaque session and immutable oidc identity")
	}
}

func TestLogoutRevokesSessionAndReturnsProviderURL(t *testing.T) {
	repository := &fakeRepository{}
	e, _ := testHandler(repository, &fakeOIDCFlow{})
	request := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	request.Header.Set("Origin", "https://sabeel.example")
	request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: "session-token"})
	response := httptest.NewRecorder()
	e.ServeHTTP(response, request)
	if response.Code != http.StatusOK || len(repository.deletedHash) != 32 || !strings.Contains(response.Body.String(), "/oidc/v1/end_session") {
		t.Fatalf("unexpected logout response: %d %s", response.Code, response.Body.String())
	}
}

func TestOAuthStateExpires(t *testing.T) {
	value, err := signOAuthState([]byte("0123456789abcdef0123456789abcdef"), oauthState{State: "state", Nonce: "nonce", Verifier: "verifier", ReturnTo: "/", ExpiresAt: time.Now().Add(-time.Second)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := verifyOAuthState([]byte("0123456789abcdef0123456789abcdef"), value, time.Now()); err == nil {
		t.Fatal("expected expired oauth state rejection")
	}
}
