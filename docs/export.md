# Script Export

`appctl export setup` generates a self-contained `setup.sh` pre-configured for a set of registered apps. The generated script can be committed to a GitHub repo and used as a one-liner Quick Install.

## Usage

```bash
# Print to stdout
appctl export setup

# Filter to specific apps
appctl export setup openplan-api openplan-web openplan-admin

# Write to a file
appctl export setup -o deploy/setup.sh
```

## What the generated script does

1. **Prerequisites** — installs .NET SDK (if any dotnet-api apps), Node.js (if any react-spa/static apps), and git
2. **Per-app setup** — for each app:
   - Clones the git repo if not already present, otherwise pulls latest
   - Builds (using the app's type-specific default or `build_command` override)
   - Writes and enables the systemd unit file (for dotnet-api and custom apps)
3. **Summary** — prints all app directories, domains, and service names

## Workflow: embedding a setup script in a GitHub repo

1. Register all your apps in appctl:
   ```bash
   appctl apps add   # repeat for each app
   ```

2. Export the script:
   ```bash
   appctl export setup openplan-api openplan-web openplan-admin \
     -o /path/to/openplan-api/deploy/setup.sh
   ```

3. Commit and push:
   ```bash
   cd /path/to/openplan-api
   git add deploy/setup.sh
   git commit -m "chore: update generated setup script"
   git push
   ```

4. Document the one-liner in the repo's README:
   ```bash
   curl -fsSL https://raw.githubusercontent.com/you/openplan-api/main/deploy/setup.sh \
     -o setup.sh && bash setup.sh
   ```

## What still requires manual input

The generated script contains all structural configuration (paths, ports, service names, build commands). It still prompts for secrets at install time — DB passwords, JWT secrets, API tokens — since those must never be committed.

## Limitations

- The generated script targets Ubuntu/Debian (uses `apt-get`)
- Build-time environment variables (e.g. `VITE_API_URL`) are baked in from the app's `build_command` override — if `build_command` is empty, the generated script uses the type default which may not have the right env vars set
- The script does not configure Caddy (Caddyfile management is out of scope); a warning is printed at the end if any apps have a domain set
