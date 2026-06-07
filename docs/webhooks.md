# GitHub Webhooks

appctl can receive `release.published` events from GitHub and automatically deploy the new version.

## Setup

### 1. Set a webhook secret on the app

```bash
appctl apps update my-api --webhook-secret "$(openssl rand -hex 32)"
```

Note the secret — you'll enter it in the GitHub UI.

### 2. Register the webhook in GitHub

Go to your repo → **Settings** → **Webhooks** → **Add webhook**:

| Field | Value |
|---|---|
| Payload URL | `http://your-server:7070/webhooks/github/<app-name>` |
| Content type | `application/json` |
| Secret | the value from step 1 |
| Events | Select **"Let me select individual events"** → check **Releases** only |

### 3. Publish a release on GitHub

Tag a commit (`git tag v0.2.0 && git push origin v0.2.0`), then create a GitHub release from that tag. appctl will receive the webhook and deploy the tagged version automatically.

## What happens on a webhook

1. GitHub POSTs the release payload to `/webhooks/github/<app-name>`
2. appctl verifies the HMAC-SHA256 signature using the app's `webhook_secret`
3. If the event is `release.published`, appctl extracts the tag name and starts a deploy
4. The deploy runs in the background; the webhook returns 202 immediately

## Security

- Requests with an invalid or missing signature are rejected with 401
- If `webhook_secret` is empty, **all requests are accepted without verification** — always set a secret in production
- The webhook endpoint is unauthenticated by design (GitHub needs to reach it without credentials); the HMAC signature is the authentication mechanism

## Viewing webhook-triggered deployments

```bash
appctl logs my-api
```

The `TRIGGER` column shows `webhook` for auto-deployments.

## Webhook URL format

```
POST /webhooks/github/{app-name}
```

One URL per app. The app name in the URL must exactly match the registered name in appctl.
