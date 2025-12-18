# GitHub App Token Refresh Sidecar

This sidecar container automatically refreshes GitHub App installation tokens for workspace authentication.

## Features

- **Automatic token refresh**: Refreshes tokens every 45 minutes (configurable)
- **Repository scoping**: Limits token access to only configured workspace repos
- **Official libraries**: Uses `bradleyfalzon/ghinstallation` for token generation
- **Secure storage**: Writes tokens to tmpfs (in-memory) volume
- **Retry logic**: Exponential backoff on failures
- **Minimal footprint**: Distroless base image, runs as non-root (UID 65532)

## Building

```bash
docker build -t ghcr.io/yourorg/github-token-sidecar:latest .
docker push ghcr.io/yourorg/github-token-sidecar:latest
```

## Usage in Helm Chart

The sidecar is automatically deployed when `githubApp.enabled=true`:

```yaml
githubApp:
  enabled: true
  secretName: my-github-app-secret
  refreshIntervalMinutes: 45
  sidecarImage: ghcr.io/yourorg/github-token-sidecar:latest
```

## Required Secret

Create a Kubernetes secret with GitHub App credentials:

```bash
kubectl create secret generic my-github-app-secret \
  --from-literal=app-id=123456 \
  --from-literal=installation-id=789012 \
  --from-file=private-key=path/to/private-key.pem
```

## Architecture

1. **Init Container** (`generate-initial-token`): Generates first token before main container starts
2. **Clone Repos Init Container**: Uses token to clone repositories
3. **Main Container**: Workspace with `gh` CLI authenticated using token
4. **Sidecar Container** (`github-token-sidecar`): Continuously refreshes token every 45 minutes

## Token Flow

```
┌─────────────────────────────────────┐
│ generate-initial-token (init)       │
│ - Reads GitHub App credentials      │
│ - Generates first token             │
│ - Writes to /var/run/github/token   │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│ clone-repos (init)                  │
│ - Reads token from volume           │
│ - Clones workspace repositories     │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│ workspace (main container)          │
│ - Authenticates gh CLI with token   │
│ - Runs Claude Code and Happy CLI    │
└─────────────────────────────────────┘
               │
               │ (parallel)
               ▼
┌─────────────────────────────────────┐
│ github-token-sidecar                │
│ - Refreshes token every 45 min      │
│ - Overwrites /var/run/github/token  │
└─────────────────────────────────────┘
```

## Environment Variables

- `WORKSPACE_REPOS`: Space-separated list of repository URLs (for scoping)
- `REFRESH_INTERVAL_MINUTES`: Token refresh interval (default: 45)

## Volume Mounts

- `/var/run/github`: Shared emptyDir (tmpfs) for token storage
- `/var/run/secrets/github-app`: GitHub App credentials from K8s secret

## Security

- Runs as non-root user (UID 65532)
- No privilege escalation
- All capabilities dropped
- Token stored in tmpfs (memory-backed, never persisted to disk)
- Read-only secret mount
- Minimal attack surface (distroless base)

## See Also

- [GitHub App Setup Guide](../history/GITHUB_APP_SETUP_GUIDE.md)
- [Helm Chart Documentation](../chart/happy-little-claude-coders/README.md)
