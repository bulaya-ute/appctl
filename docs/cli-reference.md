# CLI Reference

All CLI commands talk to the running daemon via HTTP. The daemon address defaults to `http://localhost:7070` and can be overridden with `--host <url>` or the `APPCTL_HOST` environment variable.

## Global flags

| Flag | Default | Description |
|---|---|---|
| `--host` | `http://localhost:7070` | Daemon address (env: `APPCTL_HOST`) |

---

## appctl server

Start the appctl daemon.

```bash
appctl server [--port <port>] [--db <path>]
```

| Flag | Default | Description |
|---|---|---|
| `--port` | from db config (`7070`) | Port to listen on |
| `--db` | platform default | Path to SQLite database |

---

## appctl apps

### list

```bash
appctl apps list
```

Prints a table of all registered apps: name, type, source, domain, service name.

### add

```bash
appctl apps add
```

Interactive prompt. Registers a new app and optionally deploys immediately.

### show

```bash
appctl apps show <name>
```

Prints the full app config as JSON.

### update

```bash
appctl apps update <name> [flags]
```

| Flag | Description |
|---|---|
| `--description` | App description |
| `--branch` | Git branch |
| `--domain` | Domain for Caddy route |
| `--port` | Binding port |
| `--service` | Systemd service name |
| `--build-command` | Override default build command |
| `--run-command` | Override systemd ExecStart |
| `--publish-dir` | Override publish/dist directory |
| `--git-repo` | Git repo URL |
| `--git-token-path` | Path to file containing git token |
| `--webhook-secret` | GitHub webhook secret |

### remove

```bash
appctl apps remove <name>
```

Deregisters the app from appctl. Does **not** delete files, stop services, or remove Caddy routes.

### unit

```bash
appctl apps unit <name>
```

Writes the systemd unit file to `/etc/systemd/system/<service>.service` and enables it. Run this after registering a dotnet-api or custom app for the first time.

---

## appctl deploy

```bash
appctl deploy <name> [version]
```

Deploys an app: git pull → build → service restart → Caddy route update. Returns immediately (202); the deploy runs in the background. Use `appctl logs <name>` to follow progress.

| Argument | Description |
|---|---|
| `name` | App name |
| `version` | Git tag (e.g. `0.2.0` or `v0.2.0`). Omit or use `latest` to deploy branch HEAD. |

---

## appctl start / stop / restart

```bash
appctl start <name>
appctl stop <name>
appctl restart <name>
```

Controls the app's systemd service directly. The app must have a `service_name` configured.

---

## appctl logs

```bash
appctl logs <name> [-l]
```

Prints deployment history as a table (ID, version, status, trigger, start time, duration).

| Flag | Description |
|---|---|
| `-l`, `--log` | Also print the full log of the most recent deployment |

---

## appctl config

### show

```bash
appctl config show
```

Prints all daemon config key-value pairs.

### set

```bash
appctl config set <key> <value>
```

Sets a config value. Changes take effect immediately (no restart needed).

Built-in keys:

| Key | Default | Description |
|---|---|---|
| `appctl_port` | `7070` | Daemon listen port (restart required) |
| `caddy_admin_url` | `http://localhost:2019` | Caddy admin API address |

---

## appctl export setup

```bash
appctl export setup [app1 app2 ...] [-o <file>]
```

Generates a self-contained `setup.sh` pre-configured for the specified apps (or all registered apps if none are named). Output goes to stdout by default.

| Flag | Description |
|---|---|
| `-o`, `--out` | Write to file instead of stdout |

Example — write to a repo's deploy directory:

```bash
appctl export setup openplan-api openplan-web openplan-admin -o deploy/setup.sh
```
