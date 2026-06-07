# appctl

Self-hosted deployment manager for Linux servers. Manages git pull, build, systemd services, and Caddy reverse proxy routes — all from one binary.

**License:** MIT · **Status:** Early development · **Platform:** Linux

---

## Quick Install (Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/bulaya-ute/appctl/main/install.sh | bash
```

The script detects your architecture (amd64/arm64), downloads the latest binary, installs it to `/usr/local/bin`, and registers appctl as a systemd service.

---

## Quick Start

```bash
# Build
go build -o appctl ./cmd/appctl

# Start the daemon (requires root for systemd + caddy)
sudo ./appctl server

# Register an app (interactive)
appctl apps add

# Deploy
appctl deploy my-api 0.2.0

# View deployment history
appctl logs my-api --log
```

## Installation as a service

```bash
sudo cp appctl /usr/local/bin/appctl
sudo appctl server --port 7070 &   # or set up a systemd unit
```

---

## App types

| Type | Build | Served by |
|---|---|---|
| `dotnet-api` | `dotnet publish -c Release` | systemd service |
| `react-spa` | `npm ci && npm run build` | Caddy `file_server` |
| `static` | none | Caddy `file_server` |
| `custom` | user-defined `build_command` | optional systemd service |

---

## CLI reference

```
appctl server                        start the daemon

appctl apps list                     list registered apps
appctl apps add                      register a new app (interactive)
appctl apps show <name>              show app details (JSON)
appctl apps update <name> [flags]    update app config
appctl apps remove <name>            deregister an app
appctl apps unit <name>              write + enable the systemd unit file

appctl deploy <name> [version]       deploy (omit version for branch HEAD)
appctl start <name>                  start service
appctl stop <name>                   stop service
appctl restart <name>                restart service
appctl logs <name> [-l]              deployment history (-l prints full log)

appctl config show                   print config key-value pairs
appctl config set <key> <value>      set a config value

appctl export setup [app1 app2 ...]  generate a setup.sh for the named apps
```

---

## Configuration

Config is stored in the SQLite database (key-value table). Defaults:

| Key | Default | Description |
|---|---|---|
| `appctl_port` | `7070` | Port the daemon listens on |
| `caddy_admin_url` | `http://localhost:2019` | Caddy admin API address |

Update with: `appctl config set caddy_admin_url http://localhost:2019`

---

## GitHub webhooks

Each app can receive `release.published` events from GitHub. When a new release is tagged, appctl automatically deploys it.

1. Set a webhook secret on the app: `appctl apps update my-api --webhook-secret <secret>`
2. Register the webhook in your GitHub repo settings:
   - Payload URL: `http://your-server:7070/webhooks/github/my-api`
   - Content type: `application/json`
   - Secret: the value from step 1
   - Events: **Releases** only

---

## Caddy integration

appctl calls the Caddy admin API (`localhost:2019` by default) to add or update routes when an app is deployed. Routes are tagged with `@id: "appctl-<name>"` for idempotent updates.

Caddy must be installed and its admin API must be enabled (it is by default).

---

## Tech stack

- Go 1.22 — single static binary, no runtime dependencies
- [chi](https://github.com/go-chi/chi) — HTTP router
- [cobra](https://github.com/spf13/cobra) — CLI
- [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) — pure-Go SQLite (no cgo)

---

## Documentation

| Document | Description |
|---|---|
| [CLI Reference](docs/cli-reference.md) | All commands, flags, and examples |
| [App Types](docs/app-types.md) | Build pipelines for dotnet-api, react-spa, static, custom |
| [Setup & Self-Hosting](docs/setup.md) | Manual install, building from source, daemon config |
| [Caddy Integration](docs/caddy.md) | How appctl manages Caddy routes |
| [GitHub Webhooks](docs/webhooks.md) | Auto-deploy on release.published |
| [Script Export](docs/export.md) | Generating setup.sh for a project's GitHub repo |

---

## License

[MIT](LICENSE)
