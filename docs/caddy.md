# Caddy Integration

appctl manages Caddy reverse proxy and file server routes via the [Caddy admin API](https://caddyserver.com/docs/api).

## How it works

When an app with a `domain` is deployed, appctl:

1. Builds a route object tagged with `@id: "appctl-<name>"`
2. Tries `PATCH /id/appctl-<name>` to update an existing route
3. If the route doesn't exist, appends it via `POST /config/apps/http/servers/srv0/routes/...`

The `@id` tag makes subsequent deploys idempotent — the same route is updated in place rather than duplicated.

## Route types

| App type | Route |
|---|---|
| `dotnet-api`, `custom` | `reverse_proxy` → `localhost:<binding_port>` |
| `react-spa`, `static` | `file_server` from `publish_dir` (defaults to `<local_path>/dist`) |

## Requirements

- Caddy must be installed and running
- Caddy's admin API must be enabled on `localhost:2019` (this is the default — no extra config needed)
- appctl must have network access to the Caddy admin API

## Configuration

The Caddy admin URL is stored in appctl's config table:

```bash
appctl config show         # see current value
appctl config set caddy_admin_url http://localhost:2019
```

## Removing a route

When an app is deregistered (`appctl apps remove`), its Caddy route is **not** automatically removed (the app removal is non-destructive). To remove the route manually:

```bash
curl -X DELETE http://localhost:2019/id/appctl-<name>
```

## TLS / HTTPS

appctl does not configure TLS directly. Caddy handles TLS automatically via Let's Encrypt when a domain is set — no additional configuration is needed on appctl's side. Caddy must be reachable from the internet on port 80 for the ACME challenge.

## Troubleshooting

**Route not appearing after deploy:**

- Check the deploy log: `appctl logs <name> -l`
- Confirm Caddy is running: `systemctl status caddy`
- Test the admin API directly: `curl http://localhost:2019/config/`
- Check if `domain` is set on the app: `appctl apps show <name>`
