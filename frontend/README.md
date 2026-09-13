# Sabeel Frontend

React Router framework mode app configured as a static SPA with `ssr: false`.

## Commands

- `npm run dev`: start the React Router dev server.
- `npm run typecheck`: generate route types and run TypeScript.
- `npm run build`: build static assets into `build/client`.
- `npm run start`: serve `build/client` with SPA fallback.

## Architecture

- React Router framework mode for route organization.
- React Query for backend/server state, caching, mutations, and invalidation.
- Zustand for client auth state backed by an HttpOnly Sabeel session cookie.
- shadcn/Tailwind components under `app/components/ui`.
- Go backend API URL comes from `VITE_MAKTABA_API_URL`, defaulting to `http://localhost:8080`.

## Deployment

The production artifact is static SPA content in `build/client`.

Serve it from any static host, CDN, Nginx/Caddy, the provided `server.mjs`, or the Go backend. Configure SPA fallback so unknown paths return `index.html`.
