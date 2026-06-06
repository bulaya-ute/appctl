#!/usr/bin/env bash
# =============================================================================
# appctl installer for Linux
# https://github.com/bulaya-ute/appctl
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/bulaya-ute/appctl/main/install.sh | bash
#   curl -fsSL https://raw.githubusercontent.com/bulaya-ute/appctl/main/install.sh | bash -s -- --version v0.1.1
# =============================================================================
set -euo pipefail

REPO="bulaya-ute/appctl"
INSTALL_BIN="/usr/local/bin/appctl"
SERVICE_FILE="/etc/systemd/system/appctl.service"
DATA_DIR="/var/lib/appctl"
VERSION=""

# ── Parse args ────────────────────────────────────────────────────────────────
while [[ $# -gt 0 ]]; do
  case "$1" in
    --version) VERSION="$2"; shift 2 ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

# ── Colour helpers ────────────────────────────────────────────────────────────
GREEN='\033[0;32m'; CYAN='\033[0;36m'; RED='\033[0;31m'; NC='\033[0m'
ok()  { printf "${GREEN}  ✓  %s${NC}\n" "$*"; }
info(){ printf "${CYAN}  →  %s${NC}\n" "$*"; }
die() { printf "${RED}  ✗  %s${NC}\n" "$*" >&2; exit 1; }

# ── Detect architecture ───────────────────────────────────────────────────────
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)          ARCH_TAG="amd64" ;;
  aarch64|arm64)   ARCH_TAG="arm64" ;;
  *) die "Unsupported architecture: $ARCH" ;;
esac

# ── Resolve version ───────────────────────────────────────────────────────────
if [[ -z "$VERSION" ]]; then
  info "Fetching latest release tag..."
  VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
    | grep '"tag_name"' | head -1 | cut -d'"' -f4)
  [[ -n "$VERSION" ]] || die "Could not determine latest version"
fi
info "Installing appctl $VERSION (linux/$ARCH_TAG)"

# ── Download binary ───────────────────────────────────────────────────────────
BINARY="appctl-linux-$ARCH_TAG"
URL="https://github.com/$REPO/releases/download/$VERSION/$BINARY"
info "Downloading $URL"
curl -fsSL "$URL" -o /tmp/appctl || die "Download failed — does release $VERSION have pre-built binaries?"
chmod +x /tmp/appctl

# ── Install binary ────────────────────────────────────────────────────────────
if [[ "$EUID" -ne 0 ]]; then
  info "Not running as root — using sudo to install binary and service"
  SUDO="sudo"
else
  SUDO=""
fi
$SUDO mv /tmp/appctl "$INSTALL_BIN"
ok "Binary installed: $INSTALL_BIN"

# ── Create data directory ─────────────────────────────────────────────────────
$SUDO mkdir -p "$DATA_DIR"
ok "Data directory: $DATA_DIR"

# ── Install systemd service ───────────────────────────────────────────────────
$SUDO tee "$SERVICE_FILE" > /dev/null <<'EOF'
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

$SUDO systemctl daemon-reload
$SUDO systemctl enable appctl
$SUDO systemctl start appctl
ok "Service started: appctl"

# ── Done ──────────────────────────────────────────────────────────────────────
printf "\n${GREEN}  appctl $VERSION installed successfully.${NC}\n\n"
printf "  Test:   curl http://localhost:7070/health\n"
printf "  Manage: appctl apps list\n"
printf "  Docs:   https://github.com/$REPO\n\n"
