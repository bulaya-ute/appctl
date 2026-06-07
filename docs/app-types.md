# App Types

Each app has a `type` that determines its default build pipeline and how it is served.

## dotnet-api

A .NET 8+ Web API project.

**Default build:**
```bash
dotnet publish <local_path> -c Release -o <publish_dir>
```

`publish_dir` defaults to `<local_path>/publish`.

**Served by:** systemd service. After the first build, run `appctl apps unit <name>` to write and enable the unit file. appctl auto-detects the DLL name by scanning the publish dir for `*.runtimeconfig.json`.

**Environment variables set in the unit file:**
- `ASPNETCORE_URLS=http://localhost:<binding_port>`
- `ASPNETCORE_ENVIRONMENT=Production`

**Example registration:**
```
Name:         openplan-api
Type:         dotnet-api
Source:       git
Git repo URL: https://github.com/you/openplan-api
Branch:       main
Local path:   /opt/apps/openplan/openplan-api
Service name: openplan-api
Binding port: 5040
Domain:       api.yourdomain.com
```

---

## react-spa

A React (or any Vite/Node) single-page application.

**Default build:**
```bash
npm ci --prefix <local_path>
npm --prefix <local_path> run build
```

Output is expected in `<local_path>/dist` (or `publish_dir` if overridden).

**Served by:** Caddy `file_server` pointing at the dist directory. No systemd service.

**Passing build-time environment variables:**

Use `build_command` override to inject variables:

```bash
appctl apps update myapp --build-command "VITE_API_URL=https://api.example.com npm --prefix /opt/apps/myapp run build"
```

**Example registration:**
```
Name:       openplan-web
Type:       react-spa
Source:     git
Git repo:   https://github.com/you/openplan-web
Local path: /opt/apps/openplan/openplan-web
Domain:     app.yourdomain.com
```

---

## static

A directory of pre-built static files. No build step.

**Served by:** Caddy `file_server`. The `local_path` (or `publish_dir`) is served directly.

Useful for generated documentation sites, pre-compiled SPAs committed to the repo, or any files that don't need a build.

---

## custom

Any app that doesn't fit the above types.

**Build:** runs `build_command` via `sh -c` from `local_path`. Nothing runs if `build_command` is empty.

**Served by:** systemd (if `service_name` is set). Requires `run_command` to be set explicitly — appctl cannot auto-detect the executable for custom apps.

**Example:** a Python FastAPI app:

```
Type:          custom
Build command: pip install -r requirements.txt
Run command:   /usr/local/bin/uvicorn app.main:app --host 127.0.0.1 --port 8000
Service name:  my-python-api
Binding port:  8000
Domain:        api.yourdomain.com
```

---

## Overriding defaults

Any field can be overridden regardless of type:

| Field | Purpose |
|---|---|
| `build_command` | Replaces the entire default build step |
| `run_command` | Replaces the auto-detected `ExecStart` in the systemd unit |
| `publish_dir` | Overrides where the build output is placed / served from |
