# AGENTS.md

## Cursor Cloud specific instructions

### Services overview

This is a single Go REST API server (`cmd/server`) with PostgreSQL as the only hard dependency. See `CLAUDE.md` for full domain model, API endpoints, architecture, and build commands.

### Running the server

PostgreSQL must be running and a database must exist before starting the server. Start PostgreSQL with `sudo pg_ctlcluster 16 main start`. The server requires `DATABASE_MASTER_URL`:

```bash
export DATABASE_MASTER_URL="postgres://fleet:fleet@localhost:5432/fleet?sslmode=disable"
go run ./cmd/server        # or: make run
```

Migrations run automatically on startup. The server listens on `:8080` by default.

### Running tests

Tests use mocks and do **not** require PostgreSQL. However, `TestAppWiring` in `cmd/server` binds to port `:8080`, so stop any running server before running `go test ./...`.

### Lint and static analysis

- `make lint` — requires `golangci-lint` (install: `curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sudo sh -s -- -b /usr/local/bin`)
- `make vet` / `go vet ./...`

### Known issue: FX route registration

The HTTP handler FX annotations in `internal/adapters/in/http/fx.go` use `fx.ResultTags` with `group:"routes"` on concrete handler types, but `NewServer` expects `[]RouteRegistrar` (an interface). Without `fx.As(new(RouteRegistrar))` on each handler annotation, the FX value group delivers an empty slice, causing all API routes (except `/health`) to return 404. Unit tests pass because they wire handlers directly without FX. This is a pre-existing code issue, not an environment problem.
