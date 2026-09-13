package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/zizouhuweidi/maktaba/internal/auth"
	"github.com/zizouhuweidi/maktaba/internal/collections"
	"github.com/zizouhuweidi/maktaba/internal/config"
	"github.com/zizouhuweidi/maktaba/internal/db"
	"github.com/zizouhuweidi/maktaba/internal/echox"
	"github.com/zizouhuweidi/maktaba/internal/health"
	"github.com/zizouhuweidi/maktaba/internal/library"
	"github.com/zizouhuweidi/maktaba/internal/notes"
	"github.com/zizouhuweidi/maktaba/internal/profiles"
	"github.com/zizouhuweidi/maktaba/internal/reviews"
	"github.com/zizouhuweidi/maktaba/internal/sources"
)

func New(cfg *config.Config, database *db.DB, logger *slog.Logger) (*http.Server, error) {
	oidcFlow, err := auth.NewOIDCFlow(context.Background(), auth.OIDCConfig{
		Issuer: cfg.Auth.Issuer, InternalURL: cfg.Auth.InternalURL,
		ClientID: cfg.Auth.ClientID, ClientSecret: cfg.Auth.ClientSecret,
		RedirectURL: cfg.Auth.RedirectURL,
	})
	if err != nil {
		return nil, err
	}

	authRepo := auth.NewPostgresRepository(database)
	collectionRepo := collections.NewPostgresRepository(database)
	libraryRepo := library.NewPostgresRepository(database)
	sourceRepo := sources.NewPostgresRepository(database)
	noteRepo := notes.NewPostgresRepository(database)
	profileRepo := profiles.NewPostgresRepository(database)
	reviewRepo := reviews.NewPostgresRepository(database)

	collectionSvc := collections.NewService(collectionRepo, logger)
	librarySvc := library.NewService(libraryRepo, logger)
	sourceSvc := sources.NewService(sourceRepo, logger)
	noteSvc := notes.NewService(noteRepo, logger)
	profileSvc := profiles.NewService(profileRepo, logger)
	reviewSvc := reviews.NewService(reviewRepo, logger)

	authHndlr := auth.NewHandler(authRepo, oidcFlow, auth.HandlerConfig{
		SessionDuration: cfg.Auth.SessionLifetime,
		CookieSecure:    cfg.Auth.CookieSecure,
		TrustedOrigins:  cfg.Server.CORSAllowedOrigins,
		StateKey:        cfg.Auth.StateKey,
		PublicURL:       cfg.Auth.PublicURL,
		OIDCIssuer:      cfg.Auth.Issuer,
		OIDCClientID:    cfg.Auth.ClientID,
	}, logger)
	collectionHndlr := collections.NewHandler(collectionSvc, logger)
	libraryHndlr := library.NewHandler(librarySvc, logger)
	sourceHndlr := sources.NewHandler(sourceSvc, logger)
	noteHndlr := notes.NewHandler(noteSvc, logger)
	profileHndlr := profiles.NewHandler(profileSvc, logger)
	reviewHndlr := reviews.NewHandler(reviewSvc, logger)

	e := echo.New()
	e.Logger = logger
	e.Validator = newRequestValidator()
	e.IPExtractor = echo.LegacyIPExtractor()
	e.HTTPErrorHandler = echoErrorHandler(logger)
	e.Use(middleware.RequestID())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     cfg.Server.CORSAllowedOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{DisableStackAll: true}))
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogLatency:      true,
		LogMethod:       true,
		LogURIPath:      true,
		LogRoutePath:    true,
		LogRequestID:    true,
		LogStatus:       true,
		LogResponseSize: true,
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			logger.Info("request completed", "method", v.Method, "path", v.URIPath, "route", v.RoutePath, "status", v.Status, "bytes", v.ResponseSize, "duration", v.Latency, "request_id", v.RequestID)
			return nil
		},
	}))

	e.GET("/health", health.Handler)
	e.GET("/ready", health.ReadyHandler(database))
	authHndlr.RegisterRoutes(e)
	collectionHndlr.RegisterPublicRoutes(e)
	libraryHndlr.RegisterPublicRoutes(e)
	sourceHndlr.RegisterPublicRoutes(e)
	noteHndlr.RegisterPublicRoutes(e)
	profileHndlr.RegisterPublicRoutes(e)
	reviewHndlr.RegisterPublicRoutes(e)

	protected := e.Group("/api")
	protected.Use(authHndlr.Middleware)
	authHndlr.RegisterProtectedRoutes(protected)
	collectionHndlr.RegisterProtectedRoutes(protected)
	libraryHndlr.RegisterProtectedRoutes(protected)
	sourceHndlr.RegisterProtectedRoutes(protected)
	noteHndlr.RegisterProtectedRoutes(protected)
	profileHndlr.RegisterProtectedRoutes(protected)
	reviewHndlr.RegisterProtectedRoutes(protected)

	return &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      e,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}, nil
}

func echoErrorHandler(logger *slog.Logger) echo.HTTPErrorHandler {
	return func(c *echo.Context, err error) {
		if response, ok := c.Response().(*echo.Response); ok && response.Committed {
			return
		}

		status := http.StatusInternalServerError
		message := "internal server error"

		var httpError *echo.HTTPError
		if errors.As(err, &httpError) {
			status = httpError.Code
			message = httpError.Message
		}
		if status >= http.StatusInternalServerError {
			logger.Error("request failed", "error", err)
		}

		if jsonErr := c.JSON(status, echox.ErrorResponse{Error: message}); jsonErr != nil {
			logger.Error("failed to write error response", "error", jsonErr)
		}
	}
}
