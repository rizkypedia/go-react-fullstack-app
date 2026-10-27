# Fullstack App Starter

A development starter with a React, TypeScript, Vite, and Material UI frontend plus a Go
`net/http` API. Docker Compose runs the frontend and backend as separate services; Air reloads
the Go API when source files change.

## Start locally with Docker

1. Copy `.env.example` to `.env`.
2. Set `JWT_SECRET` to a random secret of at least 32 characters, and set your own development
   username and password.
3. Run `docker compose up --build`.
4. Open [http://localhost:5173](http://localhost:5173), sign in with the configured development
   credentials, and submit a value. The API is also available at [http://localhost:8080](http://localhost:8080).
   Both ports are bound to localhost and are not exposed to the local network.

The frontend proxies `/api` requests to the backend. The development login issues an HS256 JWT
that expires after 12 hours, unless the user signs out first. Signing out calls
`POST /api/auth/logout`, which revokes the token immediately for subsequent API requests. The
token stays in frontend memory and is sent in the
`Authorization: Bearer` header for the protected echo endpoint. Postman uses the same login
endpoint and bearer-token flow; it has no authentication bypass.

## Backend API

- `GET /api/health` — liveness check.
- `POST /api/auth/login` — accepts `{"username":"…","password":"…"}` and returns an access token.
- `POST /api/auth/logout` — requires the current bearer token and immediately revokes it.
- `POST /api/echo` — requires a bearer token and accepts `{"value":"…"}`.

JSON endpoints require `Content-Type: application/json`. The backend applies an exact-origin CORS
allowlist via `FRONTEND_ORIGIN`; CORS is not used as authentication.

## Checks

```sh
cd backend && go test ./...
cd frontend && npm install && npm test && npm run lint && npm run build
```

## Security scope

The built-in credentials and login flow are for local development only. The API refuses to start
unless `APP_ENV=development`; this template does not implement a production identity provider.
Logout revocations are held in backend memory and therefore apply to the running API instance;
they are cleared if that process restarts. A production deployment should store revocations in
shared, durable storage or use an identity provider with session/token revocation support.
Before deployment, replace demo login with a trusted identity provider, provide HTTPS and managed
secrets, and add production controls such as rate limiting and operational monitoring. Do not
expose the development credentials or signing secret in a frontend bundle or commit `.env`.
