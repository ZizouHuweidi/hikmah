package auth

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
)

type HandlerConfig struct {
	SessionDuration time.Duration
	CookieSecure    bool
	TrustedOrigins  []string
	StateKey        string
	PublicURL       string
	OIDCIssuer      string
	OIDCClientID    string
}

type Handler struct {
	repository      Repository
	oidc            OIDCFlow
	logger          *slog.Logger
	sessionDuration time.Duration
	cookieSecure    bool
	trustedOrigins  map[string]struct{}
	stateKey        []byte
	publicURL       string
	logoutURL       string
}

func NewHandler(repository Repository, oidc OIDCFlow, cfg HandlerConfig, logger *slog.Logger) *Handler {
	origins := make(map[string]struct{}, len(cfg.TrustedOrigins))
	for _, origin := range cfg.TrustedOrigins {
		origins[strings.TrimSuffix(strings.TrimSpace(origin), "/")] = struct{}{}
	}
	query := url.Values{"client_id": {cfg.OIDCClientID}, "post_logout_redirect_uri": {strings.TrimSuffix(cfg.PublicURL, "/")}}
	return &Handler{
		repository: repository, oidc: oidc, logger: logger, sessionDuration: cfg.SessionDuration,
		cookieSecure: cfg.CookieSecure, trustedOrigins: origins, stateKey: []byte(cfg.StateKey),
		publicURL: strings.TrimSuffix(cfg.PublicURL, "/"),
		logoutURL: strings.TrimSuffix(cfg.OIDCIssuer, "/") + "/oidc/v1/end_session?" + query.Encode(),
	}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.GET("/auth/login", h.Login)
	e.GET("/auth/register", h.Register)
	e.GET("/auth/callback", h.Callback)
	e.POST("/auth/logout", h.Logout)
}

func (h *Handler) RegisterProtectedRoutes(g *echo.Group) { g.GET("/me", h.Me) }

func (h *Handler) Login(c *echo.Context) error    { return h.start(c, false) }
func (h *Handler) Register(c *echo.Context) error { return h.start(c, true) }

func (h *Handler) start(c *echo.Context, createAccount bool) error {
	state, _, err := NewSessionToken()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "login could not be started")
	}
	nonce, _, err := NewSessionToken()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "login could not be started")
	}
	verifier, _, err := NewSessionToken()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "login could not be started")
	}
	expiresAt := time.Now().Add(10 * time.Minute)
	value, err := signOAuthState(h.stateKey, oauthState{State: state, Nonce: nonce, Verifier: verifier, ReturnTo: safeReturnTo(c.QueryParam("return_to")), ExpiresAt: expiresAt})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "login could not be started")
	}
	h.setOAuthStateCookie(c, value, expiresAt)
	return c.Redirect(http.StatusFound, h.oidc.AuthorizationURL(state, nonce, verifier, createAccount))
}

func (h *Handler) Callback(c *echo.Context) error {
	h.clearOAuthStateCookie(c)
	cookie, err := c.Cookie(OAuthStateCookieName)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "login state is missing or expired")
	}
	state, err := verifyOAuthState(h.stateKey, cookie.Value, time.Now())
	if err != nil || subtle.ConstantTimeCompare([]byte(state.State), []byte(c.QueryParam("state"))) != 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "login state is invalid or expired")
	}
	if c.QueryParam("error") != "" || c.QueryParam("code") == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "identity provider did not complete login")
	}
	identity, err := h.oidc.Authenticate(c.Request().Context(), c.QueryParam("code"), state.Verifier, state.Nonce)
	if err != nil {
		h.logger.Warn("oidc callback rejected", "error", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "identity provider token validation failed")
	}
	user, err := h.repository.ProvisionOIDCUser(c.Request().Context(), identity)
	if err != nil {
		h.logger.Warn("oidc principal unavailable", "error", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "a verified email address is required")
	}
	token, tokenHash, err := NewSessionToken()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "login could not be completed")
	}
	expiresAt := time.Now().Add(h.sessionDuration)
	if err := h.repository.CreateSession(c.Request().Context(), user.ID, tokenHash, expiresAt); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "login could not be completed")
	}
	h.setSessionCookie(c, token, expiresAt)
	return c.Redirect(http.StatusFound, h.publicURL+state.ReturnTo)
}

func (h *Handler) Logout(c *echo.Context) error {
	if !h.originAllowed(c.Request()) {
		return echo.NewHTTPError(http.StatusForbidden, "request origin is not allowed")
	}
	if cookie, err := c.Cookie(SessionCookieName); err == nil && cookie.Value != "" {
		_ = h.repository.DeleteSession(c.Request().Context(), HashSessionToken(cookie.Value))
	}
	h.clearSessionCookie(c)
	return c.JSON(http.StatusOK, map[string]string{"logout_url": h.logoutURL})
}

func (h *Handler) Me(c *echo.Context) error {
	userID, ok := UserID(c)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	user, err := h.repository.GetUserByID(c.Request().Context(), userID)
	if err != nil || user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	return c.JSON(http.StatusOK, user)
}

func (h *Handler) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		cookie, err := c.Cookie(SessionCookieName)
		if err != nil || cookie.Value == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
		}
		user, err := h.repository.UserForSession(c.Request().Context(), HashSessionToken(cookie.Value))
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
		}
		if isUnsafeMethod(c.Request().Method) && !h.originAllowed(c.Request()) {
			return echo.NewHTTPError(http.StatusForbidden, "request origin is not allowed")
		}
		SetUserID(c, user.ID)
		return next(c)
	}
}

func (h *Handler) originAllowed(request *http.Request) bool {
	origin := strings.TrimSuffix(strings.TrimSpace(request.Header.Get("Origin")), "/")
	if origin == "" {
		return true
	}
	_, ok := h.trustedOrigins[origin]
	return ok
}

func isUnsafeMethod(method string) bool {
	return method != http.MethodGet && method != http.MethodHead && method != http.MethodOptions
}

func (h *Handler) setSessionCookie(c *echo.Context, value string, expiresAt time.Time) {
	c.SetCookie(&http.Cookie{Name: SessionCookieName, Value: value, Path: "/", HttpOnly: true, Secure: h.cookieSecure, SameSite: http.SameSiteLaxMode, Expires: expiresAt})
}

func (h *Handler) clearSessionCookie(c *echo.Context) {
	c.SetCookie(&http.Cookie{Name: SessionCookieName, Path: "/", HttpOnly: true, Secure: h.cookieSecure, SameSite: http.SameSiteLaxMode, MaxAge: -1, Expires: time.Unix(1, 0)})
}

func (h *Handler) setOAuthStateCookie(c *echo.Context, value string, expiresAt time.Time) {
	c.SetCookie(&http.Cookie{Name: OAuthStateCookieName, Value: value, Path: "/auth", HttpOnly: true, Secure: h.cookieSecure, SameSite: http.SameSiteLaxMode, Expires: expiresAt})
}

func (h *Handler) clearOAuthStateCookie(c *echo.Context) {
	c.SetCookie(&http.Cookie{Name: OAuthStateCookieName, Path: "/auth", HttpOnly: true, Secure: h.cookieSecure, SameSite: http.SameSiteLaxMode, MaxAge: -1, Expires: time.Unix(1, 0)})
}
