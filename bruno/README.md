# Sabeel API Bruno Collection

Open this `bruno/` directory in Bruno.

Use the `local` environment for local development. Start with `Auth/Register` or `Auth/Login`, then copy the returned `tokens.access_token` into the `access_token` environment variable for protected requests.

The refresh endpoint uses the `bh_refresh_token` HttpOnly cookie returned by register/login.
