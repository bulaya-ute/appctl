# appctl — CLAUDE.md

Self-hosted deployment manager. Single binary: `appctl server` runs the daemon; all other subcommands are CLI clients that talk to it over HTTP.

## Structure

```
cmd/appctl/main.go         — entry point (calls cli.Execute)
internal/
  db/                      — SQLite layer (Open, models, CRUD)
  deploy/                  — deploy pipeline (git, build, systemd, caddy)
  api/                     — HTTP daemon (chi router + handlers)
  cli/                     — cobra CLI (one file per command group)
  export/                  — setup.sh generator
```

## Build & Run

Requires Go 1.22+. On first clone, run `go mod tidy` to fetch dependencies and generate `go.sum`.

```bash
go mod tidy
go build -o appctl ./cmd/appctl

# Start the daemon (needs root or sudo for systemd + caddy)
sudo ./appctl server

# Use the CLI (in another terminal or after the daemon is started as a service)
./appctl apps list
./appctl apps add
./appctl deploy <name> [version]
./appctl logs <name> --log
```

## Database

SQLite at `/var/lib/appctl/appctl.db` (root) or `~/.local/share/appctl/appctl.db` (user).
Override with `--db` flag on `appctl server` or `APPCTL_DB` env var.

Schema is applied automatically on first run via `db.Open`.

## Daemon address

CLI reads daemon address from `--host` flag or `APPCTL_HOST` env var (default: `http://localhost:7070`).
Port is stored in the `config` table (`appctl_port` key) and read at daemon start.

## App types

| Type         | Build default                          | Service |
|---|---|---|
| `dotnet-api` | `dotnet publish -c Release -o publish/` | systemd |
| `react-spa`  | `npm ci && npm run build`               | Caddy file_server |
| `static`     | none (files already there)              | Caddy file_server |
| `custom`     | user-provided `build_command`           | optional systemd |

## Key design decisions

- Deploy runs in a background goroutine; handler returns 202 immediately
- Caddy routes are tagged with `@id = "appctl-<name>"` so updates are idempotent
- Git token is stored as a file path (never inline); read at deploy time
- `run_command` in app config overrides the auto-detected `ExecStart` for systemd units
- For dotnet apps, the DLL is found by scanning the publish dir for `*.runtimeconfig.json`
