# AGENTS.md — Billedapparat

Compact cheat-sheet for OpenCode. Omit anything an agent could guess from filenames.

---

## Repo layout

- **Go backend**: `backend/` (module `github.com/potibm/billedapparat`)
- **React frontend**: `frontend/` (Vite + React Admin + Tailwind v4)
- **Infra**: `docker-compose.yaml` (OpenObserve + OTel collector)
- **Root `go.work`**: points only to `./backend`. Run Go commands from `backend/`.

Frontend apps (three SPAs routed by React Router):

- `/` → Splash page
- `/admin/*` → React Admin dashboard
- `/beamer/:id?` → Bigscreen slideshow viewer

Path aliases (vite + tsconfig): `@core`, `@splash`, `@admin`.

---

## Everyday commands (all via `mise`)

Install mise tools once: `mise install`  
Full setup (deps + infra): `mise run setup`  
Dev (hot-reload both): `mise run dev` (uses Overmind / Procfile)

Backend only: `mise run be:dev` (Air, port 3101)  
Frontend only: `mise run fe:dev` (Vite, port 3100, HTTPS, proxies `/api|media|style` → :3101)

Test everything: `mise run test`  
Backend tests: `mise run be:test`  
Frontend tests: `mise run fe:test`

Lint everything: `mise run lint`  
Lint with auto-fix: `mise run lint --fix`

Docker image check: `mise run docker:build` (uses a custom `billedapparat-builder` buildx builder)

---

## Critical gotchas

### `cmd/assets` must exist before any Go build

`backend/cmd/serve.go` embeds `//go:embed assets`. If the directory is missing, **compilation fails**.

- `mise run be:setup` creates it (plus a dummy `index.html`), copies `.env.example` → `.env`, and creates `data/` subdirectories.
- `mise run be:lint` also creates a dummy file for this reason.
- `mise run be:test` depends on `be:setup`, so tests via `mise` are safe; running `go test ./...` directly without `be:setup` first will fail.
- Dockerfile copies the real frontend build into `backend/cmd/assets`.

### Config loading order

1. `backend/config/config.yaml` (committed defaults)
2. `backend/config/config.local.yaml` (gitignored overrides)
3. `.env` (loaded by `godotenv`)
4. Environment variables (`APP_LOG_LEVEL` maps to `app.log_level`)
5. CLI flags (`--log-level`, `--port`, etc.)

Use `config/config.local.yaml` for local secrets; do not edit `config.yaml`.

### Frontend dev proxy

Vite proxies `/api`, `/media`, and `/style` to `http://127.0.0.1:3101`.  
The backend dev server (`air`) runs on **3101**, not 3100.

### Frontend tests need `--no-webstorage`

The `fe:test` task sets `NODE_OPTIONS="--no-webstorage"`. Running `npm run test` directly may behave differently.

---

## Lint / typecheck / test pipeline

CI order: `deps:install` → `lint` → `test` → `docker:build`.

- **Backend lint**: `golangci-lint` (config in `backend/.golangci.yaml`). Uses `gofumpt` + `golines` (120 cols).
- **Frontend lint**: ESLint + Prettier + `tsc --noEmit` + `dotenv-linter`.
- **Repo lint**: Prettier over `*`, `.github`, and `backend/**/*.{json,yml,yaml,md}`.

---

## Testing

- **Backend**: `go test -v -coverprofile=coverage.out ./...` (run from `backend/`).
- **Frontend**: `vitest`, jsdom, globals enabled, coverage via v8 (`frontend/coverage/lcov.info`).
- SonarCloud ingests `backend/coverage.out` and `frontend/coverage/lcov.info`.

---

## Database / data directories

SQLite file lives in `backend/data/db/`.  
`ensureAppInfrastructure()` creates these subdirs under `./data/` at runtime:
`db`, `media`, `style`, `import`, `seed_cache`.

---

## Releasing

- Semantic Release with Angular commit preset.
- Docker release triggers on tags matching `[0-9]+.[0-9]+.[0-9]+`.
- Image is signed with Cosign and attested with SBOM.
- PR titles are validated by `amannn/action-semantic-pull-request`.

---

## Infra (local)

`docker compose up -d` starts:

- OpenObserve UI: http://localhost:3105 (admin@example.com / password123)
- OTel gRPC: localhost:3117

Backend dev (`air`) defaults to `--otel-endpoint=localhost:3117`.

---

## Style notes

- Go: `slog` only (depguard blocks `logrus`). Use snake_case for slog keys.
- Go: avoid `math/rand` (use `math/rand/v2`).
- TS: `no-console` and `no-alert` are errors. `_` prefix ignores unused vars.
- TS: new `.js` files are forbidden by ESLint (use `.ts`/`.tsx`).
