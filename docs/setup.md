# Setup & Self-Hosting

## Quick Install (Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/bulaya-ute/appctl/main/install.sh | bash
```

Installs the latest binary, creates `/var/lib/appctl/`, and registers `appctl` as a systemd service.

To pin a specific version:

```bash
curl -fsSL https://raw.githubusercontent.com/bulaya-ute/appctl/main/install.sh | bash -s -- --version v0.1.1
```

## Building from source

Requires Go 1.22+.

```bash
git clone https://github.com/bulaya-ute/appctl.git
cd appctl
go mod tidy
go build -o appctl ./cmd/appctl
sudo cp appctl /usr/local/bin/appctl
```

## Manual service setup

```bash
sudo mkdir -p /var/lib/appctl

sudo tee /etc/systemd/system/appctl.service > /dev/null <<'EOF'
[Unit]
Description=appctl deployment daemon
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/appctl server
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable appctl
sudo systemctl start appctl
```

## Verifying the daemon is running

```bash
curl http://localhost:7070/health
# → {"status":"ok"}
```

## Database

SQLite database location (auto-created on first run):

| Condition | Path |
|---|---|
| Running as root | `/var/lib/appctl/appctl.db` |
| Running as user | `~/.local/share/appctl/appctl.db` |

Override with `--db /path/to/custom.db` on `appctl server`.

## Daemon port

Default: `7070`. Change it permanently:

```bash
appctl config set appctl_port 8080
sudo systemctl restart appctl
```

Or override at startup: `appctl server --port 8080`

## Permissions

appctl needs root (or equivalent) to:
- Write systemd unit files to `/etc/systemd/system/`
- Run `systemctl` commands
- Call the Caddy admin API (if Caddy runs as root)

Running `appctl server` as root is the simplest setup. For tighter permissions, a sudoers rule scoped to `systemctl` is an alternative.
