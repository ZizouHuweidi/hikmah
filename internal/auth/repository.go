package auth

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/zizouhuweidi/maktaba/internal/db"
)

var ErrInvalidSession = errors.New("invalid session")

type Repository interface {
	ProvisionOIDCUser(ctx context.Context, identity OIDCIdentity) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	CreateSession(ctx context.Context, userID uuid.UUID, tokenHash []byte, expiresAt time.Time) error
	UserForSession(ctx context.Context, tokenHash []byte) (*User, error)
	DeleteSession(ctx context.Context, tokenHash []byte) error
}

type postgresRepository struct{ db *db.DB }

func NewPostgresRepository(database *db.DB) Repository { return &postgresRepository{db: database} }

func (r *postgresRepository) ProvisionOIDCUser(ctx context.Context, identity OIDCIdentity) (*User, error) {
	issuer := strings.TrimSuffix(strings.TrimSpace(identity.Issuer), "/")
	subject := strings.TrimSpace(identity.Subject)
	email := strings.ToLower(strings.TrimSpace(identity.Email))
	if issuer == "" || subject == "" || email == "" || !identity.EmailVerified {
		return nil, errors.New("verified oidc identity is required")
	}

	if user, err := scanUser(r.db.Pool.QueryRow(ctx, `
		SELECT id, email, username, oidc_issuer, oidc_subject, created_at, updated_at
		FROM users WHERE oidc_issuer = $1 AND oidc_subject = $2`, issuer, subject)); err == nil {
		_, err = r.db.Pool.Exec(ctx, `UPDATE users SET email = $2, updated_at = NOW() WHERE id = $1`, user.ID, email)
		user.Email = email
		return user, err
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	userID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	username := oidcUsername(identity.PreferredUsername, email, issuer, subject)
	user, err := scanUser(r.db.Pool.QueryRow(ctx, `
		INSERT INTO users (id, email, username, password_hash, oidc_issuer, oidc_subject)
		VALUES ($1, $2, $3, NULL, $4, $5)
		ON CONFLICT (email) DO UPDATE
		SET oidc_issuer = EXCLUDED.oidc_issuer,
		    oidc_subject = EXCLUDED.oidc_subject,
		    updated_at = NOW()
		WHERE users.oidc_issuer IS NULL AND users.oidc_subject IS NULL
		RETURNING id, email, username, oidc_issuer, oidc_subject, created_at, updated_at`,
		userID, email, username, issuer, subject))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("email belongs to another identity")
	}
	if err != nil {
		return nil, fmt.Errorf("provision oidc user: %w", err)
	}
	return user, nil
}

func (r *postgresRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	user, err := scanUser(r.db.Pool.QueryRow(ctx, `
		SELECT id, email, username, oidc_issuer, oidc_subject, created_at, updated_at
		FROM users WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return user, err
}

func (r *postgresRepository) CreateSession(ctx context.Context, userID uuid.UUID, tokenHash []byte, expiresAt time.Time) error {
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO auth_sessions (token_hash, user_id, expires_at) VALUES ($1, $2, $3)`, tokenHash, userID, expiresAt)
	return err
}

func (r *postgresRepository) UserForSession(ctx context.Context, tokenHash []byte) (*User, error) {
	user, err := scanUser(r.db.Pool.QueryRow(ctx, `
		SELECT u.id, u.email, u.username, u.oidc_issuer, u.oidc_subject, u.created_at, u.updated_at
		FROM auth_sessions s JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1 AND s.expires_at > NOW()`, tokenHash))
	if errors.Is(err, pgx.ErrNoRows) {
		_, _ = r.db.Pool.Exec(ctx, `DELETE FROM auth_sessions WHERE token_hash = $1`, tokenHash)
		return nil, ErrInvalidSession
	}
	return user, err
}

func (r *postgresRepository) DeleteSession(ctx context.Context, tokenHash []byte) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM auth_sessions WHERE token_hash = $1`, tokenHash)
	return err
}

type rowScanner interface{ Scan(...any) error }

func scanUser(row rowScanner) (*User, error) {
	var user User
	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.Issuer, &user.Subject, &user.CreatedAt, &user.UpdatedAt)
	return &user, err
}

var usernameCharacters = regexp.MustCompile(`[^a-z0-9_-]+`)

func oidcUsername(preferred, email, issuer, subject string) string {
	base := strings.ToLower(strings.TrimSpace(preferred))
	if base == "" {
		base = strings.SplitN(email, "@", 2)[0]
	}
	base = strings.Trim(usernameCharacters.ReplaceAllString(base, "-"), "-_")
	if len(base) < 3 {
		base = "user"
	}
	if len(base) > 20 {
		base = base[:20]
	}
	suffix := fmt.Sprintf("%x", sha256.Sum256([]byte(issuer+"\x00"+subject)))[:10]
	return base + "-" + suffix
}
