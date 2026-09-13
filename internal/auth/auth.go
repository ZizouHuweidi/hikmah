package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/labstack/echo/v5"
)

const userIDContextKey = "user_id"

const (
	SessionCookieName    = "sabeel_session"
	OAuthStateCookieName = "sabeel_oauth_state"
)

type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Username  string    `json:"username" db:"username"`
	Issuer    string    `json:"-" db:"oidc_issuer"`
	Subject   string    `json:"-" db:"oidc_subject"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func NewSessionToken() (string, []byte, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", nil, err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	return token, HashSessionToken(token), nil
}

func HashSessionToken(token string) []byte {
	hash := sha256.Sum256([]byte(token))
	return hash[:]
}

func SetUserID(c *echo.Context, userID uuid.UUID) {
	c.Set(userIDContextKey, userID)
}

func UserID(c *echo.Context) (uuid.UUID, bool) {
	userID, ok := c.Get(userIDContextKey).(uuid.UUID)
	return userID, ok
}
