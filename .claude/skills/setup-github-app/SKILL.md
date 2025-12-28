---
name: setup-github-app
description: Guide users through creating and configuring a GitHub App for workspace authentication. Use when setting up GitHub App authentication for happy-little-claude-coders, creating github-app-credentials secret, or configuring automatic token refresh.
allowed-tools: Read, Bash(gh:*), Bash(curl:*), Bash(cat:*), Bash(base64:*), Bash(kubectl get:*), Bash(kubectl describe:*), WebFetch
---

# GitHub App Setup for happy-little-claude-coders

This skill guides you through creating a GitHub App and configuring it for workspace authentication.

## Overview

The happy-little-claude-coders chart uses GitHub Apps for secure, scoped repository access. Each workspace gets automatically refreshed tokens via the sidecar container.

## What You'll Need

- GitHub account with permission to create Apps (personal or organization)
- `gh` CLI installed and authenticated
- Access to your Kubernetes cluster (for secret creation)
- The private key PEM file (downloaded during app creation)

## Step 1: Create the GitHub App

### Option A: Using GitHub Manifest Flow (Recommended)

The manifest flow creates an app from a JSON configuration. You'll need to:

1. Visit one of these URLs to start app creation:
   - **Personal account**: `https://github.com/settings/apps/new`
   - **Organization**: `https://github.com/organizations/YOUR_ORG/settings/apps/new`

2. Fill in:
   - **Name**: `hlcc-workspace-auth` (or similar, must be unique on GitHub)
   - **Homepage URL**: Your repo URL or `https://github.com/dbirks/happy-little-claude-coders`
   - **Webhook**: Uncheck "Active" (not needed)
   - **Repository permissions**:
     - Contents: **Read-only** (required for cloning)
     - Metadata: Read-only (auto-granted)
   - **Where can this app be installed?**: "Only on this account"

3. Click "Create GitHub App"

### Option B: Using gh CLI

```bash
# Create app manifest
cat > /tmp/github-app-manifest.json << 'EOF'
{
  "name": "hlcc-workspace-auth",
  "url": "https://github.com/dbirks/happy-little-claude-coders",
  "hook_attributes": {
    "active": false
  },
  "public": false,
  "default_permissions": {
    "contents": "read",
    "metadata": "read"
  }
}
EOF

# Open browser to create app from manifest (you'll need to complete in browser)
echo "Visit: https://github.com/settings/apps/new?manifest=$(cat /tmp/github-app-manifest.json | jq -c | jq -sRr @uri)"
```

## Step 2: Generate Private Key

After creating the app:

1. You'll be redirected to the app's settings page
2. Scroll to **Private keys** section
3. Click **Generate a private key**
4. A `.pem` file will download automatically

**Save this file securely!** You cannot regenerate it.

## Step 3: Install the App

1. From the app settings page, click **Install App** (left sidebar)
2. Select your account/organization
3. Choose **Only select repositories**
4. Select the repositories your workspaces need access to
5. Click **Install**

## Step 4: Gather Required Information

You need three pieces of information:

### App ID
Found on the app settings page under "About" section:
```
https://github.com/settings/apps/YOUR_APP_NAME
```
Look for: `App ID: 123456`

### Installation ID
After installing, check the URL:
```
https://github.com/settings/installations/12345678
                                          ^^^^^^^^
                                          This is your Installation ID
```

Or use gh CLI:
```bash
# Get your App ID first, then:
gh api /user/installations --jq '.installations[] | select(.app_slug == "YOUR_APP_NAME") | .id'
```

### Private Key
The `.pem` file you downloaded in Step 2.

## Step 5: Create Kubernetes Secret

### Preview the command (dry-run)

By default, this shows what WOULD be created without actually creating it:

```bash
kubectl create secret generic github-app-credentials \
  --from-literal=app-id=YOUR_APP_ID \
  --from-literal=installation-id=YOUR_INSTALLATION_ID \
  --from-file=private-key=/path/to/your-app.private-key.pem \
  --namespace=default \
  --dry-run=client -o yaml
```

This outputs the YAML that would be created. Review it carefully.

### Actually create the secret

When you're ready to create the secret for real, remove `--dry-run=client -o yaml`:

```bash
kubectl create secret generic github-app-credentials \
  --from-literal=app-id=YOUR_APP_ID \
  --from-literal=installation-id=YOUR_INSTALLATION_ID \
  --from-file=private-key=/path/to/your-app.private-key.pem \
  --namespace=default
```

### Verify the secret

```bash
kubectl get secret github-app-credentials -o yaml
kubectl get secret github-app-credentials -o jsonpath='{.data.app-id}' | base64 -d
```

## Step 6: Enable GitHub App in HelmRelease

Update your HelmRelease values:

```yaml
values:
  githubApp:
    enabled: true
    secretName: github-app-credentials
    refreshIntervalMinutes: 50  # Tokens expire after 1 hour
```

## Verification

After deployment, check the sidecar logs:

```bash
# Find your workspace pod
kubectl get pods -l app.kubernetes.io/name=happy-little-claude-coders

# Check sidecar logs for token refresh
kubectl logs <pod-name> -c github-token-refresh
```

You should see messages like:
```
Token refreshed successfully for repositories: [your-org/repo1, your-org/repo2]
```

## Troubleshooting

### "Resource not accessible by integration"
- App doesn't have access to the repository
- Fix: Add repository to app installation

### "Bad credentials"
- JWT or installation token expired
- Fix: Sidecar should auto-refresh; check sidecar logs

### "Not Found" when accessing repository
- App not installed on that repository
- Fix: Go to app installation settings, add the repository

### Secret not found
- `github-app-credentials` secret doesn't exist in the namespace
- Fix: Create the secret using Step 5

## Reference

For detailed information, see the comprehensive guide:
- [GITHUB_APP_SETUP_GUIDE.md](../../../history/GITHUB_APP_SETUP_GUIDE.md)

## Quick Reference

| Item | Where to Find |
|------|---------------|
| App ID | App settings page > About section |
| Installation ID | URL after installing the app |
| Private Key | Downloaded `.pem` file |
| Create App (Personal) | `https://github.com/settings/apps/new` |
| Create App (Org) | `https://github.com/organizations/ORG/settings/apps/new` |
