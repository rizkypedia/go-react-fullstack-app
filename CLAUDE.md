# Fullstack app

Fullstack app with a Go backend and a React + TypeScript frontend. Both run in Docker containers defined in `docker-compose.yml`.

## Project guidelines

### General

- Follow the existing project architecture.
- Prefer simple, maintainable solutions.
- Avoid unnecessary dependencies.
- Do not introduce unrelated changes.
- Avoid duplicating existing logic.
- If a change in the frontend might influence the backend, make the matching changes in the backend too, and vice versa (e.g. API routes, request/response shapes, auth, CORS, env vars).

### Development workflow

- Inspect existing code before making changes.
- Write tests for new functionality.
- Run the relevant tests (and `go vet` / `npm run lint`) before finishing.

## Testing

Use mocks instead of making direct requests to an API, a database or file resources.

## Commands

- Start the stack: `docker compose up --build` (backend on `127.0.0.1:8080`, frontend on `127.0.0.1:5173`)
- Backend (from `backend/`): `go vet ./...`, `go test ./...`
- Frontend (from `frontend/`): `npm run lint`, `npm test`, `npm run build` (includes type check)

## Configuration

Copy `.env.example` to `.env` and set every value; `docker compose` fails if a required one is missing. Never commit `.env` or hardcode secrets.

- `JWT_SECRET`: signing key for tokens
- `DEMO_USERNAME` / `DEMO_PASSWORD`: the single demo login accepted by the backend
- `APP_ENV` (default `development`) and `FRONTEND_ORIGIN` (default `http://localhost:5173`, the only allowed CORS origin)

## Backend

Lives in `backend/`.

- Go (latest version), module `fullstack-app/backend`
- [Air](https://github.com/air-verse/air) for hot reloading (`.air.toml`)
- Standard library `http.ServeMux` for routing (no external router)
- Secured via JWT

### Folder structure

- `cmd/api/main.go`: entry point and server startup
- `internal/api/`: server setup, config, JWT handling, and the authentication endpoints (login, logout, health, echo)
- `internal/api/handlers/`: all other handlers, e.g. `/api/customers/{id}`
- `utilities/`: reusable helpers, e.g. JSON encoding/decoding of HTTP requests (`json.go`)
- `tests/`: all tests, mirroring the source tree

### Code quality

- **Separation of concerns**: handlers go in `internal/api/handlers`; reusable helpers go in `utilities`; server security and authentication go in `internal/api`.
- **Tests** live in `tests/` and mirror the source path, e.g. `internal/api/handlers/customer.go` → `tests/internal/api/handlers/customer_test.go`. Use an external test package (`package handlers_test`) and test through exported identifiers only.

## Frontend

Lives in `frontend/`.

- React + TypeScript with Vite
- Material UI
- ESLint with typescript-eslint
- Jest + Testing Library

### Folder structure

- `src/main.tsx`: entry point, theme, and root render
- `src/components/`: React components (`.tsx`), including the top-level `App.tsx`
- `src/utilities/`: reusable helper functions (`.ts`)
- `src/styles/` (optional): additional CSS
- `tests/`: all Jest tests, mirroring `src/`; `tests/setup.ts` is the Jest setup file
- `dist/`: build output (generated, do not edit)

### Code quality

- **Separation of concerns**: UI code lives in `src/components` as `.tsx`; non-UI helpers live in `src/utilities` as `.ts`.
- **Tests** live in `tests/` and mirror the source path, e.g. `src/utilities/foo.ts` → `tests/utilities/foo.test.ts`, `src/components/App.tsx` → `tests/components/App.test.tsx`.
- **API calls**: `fetch` calls and token handling belong in `src/utilities/` (e.g. an API client module), not inline in components. Existing inline calls in `App.tsx` should move there when that code is next touched.
