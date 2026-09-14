# Independent Deployment Runbook

This runbook deploys Sabeel and its dedicated Zitadel instance from this
repository. No other application repository, container, network, client, or
secret is part of the runtime contract.

## Host And DNS

The Compose production path is suitable for a single host and an initial V1
deployment. Provide a Linux host with Podman Compose or Docker Compose v2, at
least 2 GB of memory for the identity stack, and inbound TCP 80/443 plus UDP
443. Create two DNS records pointing to the host:

- the application domain, such as `sabeel.example.com`;
- the identity domain, such as `auth.sabeel.example.com`.

Caddy obtains and renews certificates. Both names must resolve correctly before
the first production start.

## Secrets

Create the ignored production environment file and restrict it:

```sh
cp .env.production.example .env.production
chmod 0600 .env.production
```

Generate independent random values. The Zitadel master key must contain exactly
32 characters and is not rotatable after initialization without re-encrypting
or losing encrypted identity data. Store an offline copy before first startup.
The OIDC state key must contain at least 32 random characters.

Do not reuse development, database, SMTP, or identity secrets in production.
Do not commit `.env.production`.

## Validate And Start

```sh
just prod-config
just prod-up
just prod-ps
```

The one-shot `migrate` and `zitadel-bootstrap` services must exit successfully.
The API, frontend, identity API, Login V2, databases, proxies, and gateway must
be healthy or running. Follow startup with:

```sh
just prod-logs
```

Open both public origins and complete the verification checklist in
[`identity.md`](identity.md). The first administrator password is required to
change on first login. Create a named operator account with MFA, verify the
break-glass machine PAT, and then stop using the bootstrap human account for
routine administration.

## Updates And Rollback

1. Back up both PostgreSQL databases and the `zitadel_bootstrap` volume.
2. Build or pull immutable Sabeel image tags and pin the desired Zitadel/Login
   versions in a reviewed change.
3. Run `just prod-config`, then `just prod-up`.
4. Verify migrations, health, sign-in, and one representative read/write flow.
5. For an application rollback, restore the previous image tags only when its
   code supports the migrated schema. Use a tested down migration or database
   restore when it does not.

Never roll back the Zitadel image across incompatible database migrations.
Follow the Zitadel release notes and test the exact upgrade against a restored
staging copy first.

## Backup Set

Back up these resources on the same schedule and retain encrypted off-host
copies:

| Resource | Purpose |
| --- | --- |
| `sabeel-production_postgres_data` | Profiles, sessions, libraries, notes, reviews, and collections. |
| `sabeel-production_zitadel_postgres_data` | Users, credentials, identity configuration, and OIDC clients. |
| `sabeel-production_zitadel_bootstrap` | Login and bootstrap PATs plus Sabeel's client credentials. |
| `.env.production` secret store | Master key and runtime secrets needed to decrypt and reconnect restored state. |
| `sabeel-production_caddy_data` | TLS account and certificate state; replaceable, but useful for recovery. |

Use PostgreSQL-native logical or physical backups with consistency guarantees;
copying a live database volume is not a database backup. Encrypt backups and
test restoration at least quarterly.

## Recovery

Restore the Sabeel database, Zitadel database, bootstrap volume, and original
Zitadel master key together. Start the databases and identity services first,
then bootstrap, migrations, API, frontend, and gateway. Verify the issuer in
existing Sabeel user rows exactly matches the restored public identity URL.

If interactive identity administration is unavailable, use the offline
`sabeel-bootstrap` IAM owner PAT against the Zitadel API to repair Login V2 or
administrator access. Rotate the PAT after use and update the protected backup.

If only the generated Sabeel OIDC client credential files are lost, restart
`zitadel-bootstrap`; it finds the existing application, rotates its client
secret, and writes a replacement pair before the API starts.

## Production Evolution

The single-host Compose deployment is intentionally self-contained. As load and
availability requirements grow, move Zitadel to its official production-like
init/setup/start split or Helm deployment, and move both databases to managed or
independently replicated PostgreSQL. Preserve the same two public origins,
issuer, OIDC client contract, and Sabeel ownership boundary during that move.
