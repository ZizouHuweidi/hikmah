# Sabeel Identity

Sabeel runs its own Zitadel installation. The development and production
deployments in this repository do not join an external network, read another
repository, reuse another application's client, or require another project's
containers to be running.

Sabeel does not store passwords, Zitadel access tokens, refresh tokens, or ID
tokens.

## Ownership Boundary

Zitadel owns registration, login, recovery, email verification, MFA, passkeys,
and external identity providers. Sabeel owns profiles, library data, and every
resource-level authorization decision.

Each Sabeel user is identified by the immutable `(oidc_issuer, oidc_subject)`
pair. On the first successful login, Sabeel creates a local user. A verified
Zitadel email may attach an existing password-era user so their library and
notes remain available after migration. Email is never used to authenticate
later requests.

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

## Automatic Client Bootstrap

On the first start, Zitadel creates two narrowly scoped machine identities:

- `sabeel-login-client` receives `IAM_LOGIN_CLIENT`; Login V2 uses its PAT.
- `sabeel-bootstrap` receives `IAM_OWNER`; the one-shot bootstrap process uses
  its PAT to create the `Sabeel` project and confidential OIDC web application.

The bootstrap process writes the generated client ID and secret into the
`zitadel_bootstrap` volume. The Sabeel API reads those values through
`SABEEL_OIDC_CLIENT_ID_FILE` and `SABEEL_OIDC_CLIENT_SECRET_FILE`. Credentials
are not rendered into Compose output, frontend assets, or the checked-in env
files.

Bootstrap is idempotent. If the credential files exist and the callback
configuration matches, it does nothing. If the callback configuration changed,
it reconciles the application and rotates the secret. If the project/application
exist but the files were lost, it generates a new client secret and restores
the files. Keep the bootstrap volume and Zitadel database in the same backup
and restore set.

## Development

`compose.yaml` starts the complete local environment:

- Sabeel API, frontend, migrations, and PostgreSQL;
- Zitadel API, Login V2, PostgreSQL, and identity proxy;
- client bootstrap; and
- Mailpit for verification and recovery email.

```sh
cp .env.example .env
just up
```

No manual Zitadel project or application setup is required. Open:

| Service | URL |
| --- | --- |
| Sabeel | `http://localhost:3000` |
| Sabeel API | `http://localhost:8080` |
| Sabeel identity | `http://127.0.0.1:8081` |
| Mailpit | `http://127.0.0.1:8025` |

Registration intentionally requires email verification. In development, open
Mailpit, select the verification message, and follow its link. This exercises
the same product contract as production instead of silently treating an
unverified address as trusted.

The initial development administrator is `zitadel-admin` with password
`Password1!`. It is for local use only.

To inspect initialization:

```sh
podman compose logs zitadel-api zitadel-login zitadel-bootstrap app
podman compose ps
```

`podman compose down` preserves all data. `podman compose down -v` deletes both
application and identity state and should be used only when a completely clean
local identity is intended.

## Production

Production uses `compose.production.yaml`, the Sabeel-owned Caddy gateway, and a
dedicated `.env.production` file. See [`deployment.md`](deployment.md) for the
full deployment, backup, and recovery runbook.

The production contract requires:

- separate HTTPS origins for Sabeel and its identity endpoint;
- secure cookies and a random OIDC state-signing key;
- a durable, exactly 32-character Zitadel master key;
- distinct application and identity database passwords;
- a real SMTP relay for verification, recovery, and security messages;
- persistent and backed-up Sabeel, Zitadel, bootstrap, and Caddy volumes; and
- an offline copy of the bootstrap `IAM_OWNER` PAT for break-glass recovery.

`ZITADEL_FIRSTINSTANCE_*` and `ZITADEL_DEFAULTINSTANCE_*` values apply only when
Zitadel creates a new instance. Later policy, branding, and administrator
changes must be made through Zitadel's APIs or Console and included in the
operational change record.

## Verification Checklist

Before a release, exercise all of these through the public origins:

1. Register a new address and receive exactly one verification message.
2. Verify the address and complete the first Sabeel session.
3. Sign out of both Sabeel and Zitadel, then sign in again.
4. Request password recovery and complete it from the delivered message.
5. Confirm an unverified identity cannot provision a local Sabeel account.
6. Confirm callback state, nonce, issuer, audience, and PKCE failures are rejected.
7. Confirm private API responses are not cached by the gateway or browser PWA.
8. Restore both databases and the bootstrap volume into a staging environment
   and repeat sign-in.
