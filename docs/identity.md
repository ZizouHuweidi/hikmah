# Zitadel Identity

Sabeel uses the shared Zitadel installation for authentication. Sabeel does not
store passwords, Zitadel access tokens, refresh tokens, or ID tokens.

## Boundary

Zitadel owns registration, login, recovery, email verification, MFA, passkeys,
and external identity providers. Sabeel owns profiles, library data, and every
resource-level authorization decision.

Each Sabeel user is identified by the immutable `(oidc_issuer, oidc_subject)`
pair. On the first successful login, Sabeel creates a local user. A verified
Zitadel email may attach an existing password-era user so its library and notes
remain available after migration. Email is never used to authenticate later
requests.

## Browser Flow

1. `GET /auth/login` or `GET /auth/register` creates random state, nonce, and
   PKCE values in a signed, short-lived HttpOnly cookie and redirects to Zitadel.
2. `GET /auth/callback` exchanges the authorization code using the confidential
   client and PKCE S256, then verifies the ID-token signature, issuer, audience,
   expiry, not-before value, nonce, subject, and verified email.
3. Sabeel stores only a SHA-256 hash of a random opaque `sabeel_session` token
   and sends the token in an HttpOnly, SameSite cookie.
4. API requests resolve the local Sabeel user from that session. Browser code
   never receives an OAuth token.
5. `POST /auth/logout` deletes the local session, expires its cookie, and sends
   the browser through Zitadel end-session logout.

## Local Client

Start the one shared disposable identity stack used by Atlas and Trains, then
create a confidential Web application named `sabeel-web-dev` with:

| Setting | Value |
| --- | --- |
| Grant | Authorization Code |
| Client authentication | `client_secret_basic` |
| PKCE | S256 |
| Redirect URI | `http://localhost:8080/auth/callback` |
| Post-logout URI | `http://localhost:3000` |
| Scopes | `openid profile email` |

Copy `.env.example` to the ignored `.env` and set the generated client ID and
secret. The API container joins the shared `company-edge` network to reach
`zitadel-dev-proxy` while browser redirects continue to use the stable issuer
at `http://127.0.0.1:8081`.

Production must use a permanent HTTPS issuer, HTTPS callback and public URLs,
a random state key from secret storage, and `SABEEL_SESSION_SECURE=true`.
