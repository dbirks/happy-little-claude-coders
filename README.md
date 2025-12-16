# Happy Little Claude Coders

Containerized AI coding workspaces with Claude Code CLI and Happy CLI, designed for Kubernetes deployment.

## Features

- **Claude Code CLI**: AI-assisted development with Claude
- **Happy CLI**: Mobile/web session streaming and monitoring
- **Multi-repo support**: Clone multiple repositories per workspace
- **Interactive GitHub auth**: Device flow with persistent credentials
- **Shared Claude subscription**: OAuth token shared across all pods
- **Automated releases**: Release Please for semantic versioning

## Quick Start

### Prerequisites

- Kubernetes cluster
- Helm 3+
- Claude subscription (Pro/Max plan)
- GitHub account

### Creating Secrets

Before installing the Helm chart, you need to create Kubernetes secrets for authentication.

#### 1. Claude OAuth Token Secret (Required)

The Claude OAuth token allows the workspace to use your Claude subscription. Generate this token on a machine where you're already authenticated with Claude:

```bash
# Generate the OAuth token (requires authenticated Claude CLI)
claude setup-token
```

This will output a token that looks like: `sk_ant_oauth_...`

**Create the secret in Kubernetes:**

```bash
kubectl create secret generic claude-oauth \
  --from-literal=token=sk_ant_oauth_YOUR_TOKEN_HERE
```

**Alternative: YAML manifest**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: claude-oauth
type: Opaque
stringData:
  token: sk_ant_oauth_YOUR_TOKEN_HERE
```

Apply with:
```bash
kubectl apply -f claude-oauth-secret.yaml
```

#### 2. GitHub Token Secret (Optional)

If you want to automatically clone private repositories without interactive authentication, you can create a GitHub token secret. Otherwise, you'll use the interactive device flow authentication (recommended for most users).

**Generate a GitHub Personal Access Token:**
1. Go to https://github.com/settings/tokens
2. Click "Generate new token (classic)"
3. Select scopes: `repo` (full repository access)
4. Generate and copy the token

**Create the secret in Kubernetes:**

```bash
kubectl create secret generic github-token \
  --from-literal=token=ghp_YOUR_GITHUB_TOKEN_HERE
```

**Alternative: YAML manifest**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: github-token
type: Opaque
stringData:
  token: ghp_YOUR_GITHUB_TOKEN_HERE
```

Apply with:
```bash
kubectl apply -f github-token-secret.yaml
```

**Note:** The GitHub token is optional. If not provided, you'll authenticate interactively using `gh auth login` (device flow) after the workspace starts.

#### Best Practices for Secret Management

- **Never commit secrets to version control**: Always use `.gitignore` for secret YAML files
- **Use namespace isolation**: Create secrets in the same namespace as your workspaces
- **Rotate tokens regularly**: Update secrets periodically for security
- **Limit token permissions**: Use minimal required scopes for GitHub tokens
- **Verify secret creation**:
  ```bash
  # Check if secret exists
  kubectl get secret claude-oauth

  # Verify secret content (decode base64)
  kubectl get secret claude-oauth -o jsonpath='{.data.token}' | base64 -d
  ```
- **Use RBAC**: Restrict access to secrets using Kubernetes RBAC policies
- **Consider external secret management**: For production, use tools like:
  - HashiCorp Vault
  - AWS Secrets Manager
  - Azure Key Vault
  - Google Secret Manager
  - Sealed Secrets

### Installation

1. **Create secrets** (see [Creating Secrets](#creating-secrets) section above)

2. **Install Helm chart**:
   ```bash
   helm install my-workspace oci://ghcr.io/yourorg/charts/happy-claude-coders \
     --set claude.secretName=claude-oauth
   ```

### First-Time Setup

3. **Attach to the workspace pod**:
   ```bash
   kubectl exec -it deployment/my-workspace -- bash
   ```

4. **Authenticate with GitHub** (device flow):
   ```bash
   gh auth login
   ```
   Follow the prompts to complete authentication in your browser.

5. **Clone repositories**:
   ```bash
   clone-repos
   ```

6. **Start coding**!
   Your repositories are now cloned in `/workspace` and GitHub credentials are persisted.

## Configuration

### values.yaml Options

```yaml
workspace:
  # Repositories to clone (default: happy-cli)
  repos:
    - "https://github.com/slopus/happy-cli.git"
    - "https://github.com/yourorg/your-repo.git"

  # Workspace storage size
  storageSize: 10Gi

  # Use persistent storage (default: false = emptyDir)
  persistent: false

git:
  # Git identity
  userName: "Your Bot Name"
  userEmail: "bot@example.com"

github:
  # PVC size for GitHub CLI credentials
  configPvcSize: 100Mi

claude:
  # Secret containing CLAUDE_CODE_OAUTH_TOKEN
  secretName: "claude-oauth"

resources:
  limits:
    cpu: 2000m
    memory: 4Gi
  requests:
    cpu: 500m
    memory: 1Gi
```

## Architecture

### Authentication Strategy

**GitHub (Per-Workspace)**:
- Interactive `gh auth login` via device flow
- Credentials persist on PVC at `/root/.config/gh/`
- Each workspace has its own GitHub identity

**Claude (Shared)**:
- `CLAUDE_CODE_OAUTH_TOKEN` from K8s Secret
- Single subscription shared across all pods
- Weekly usage caps apply (Pro: 40-80hrs, Max: 140-280hrs)

### Volume Strategy

| Mount Point | Type | Purpose |
|------------|------|---------|
| `/workspace` | emptyDir | Code (ephemeral, cloned fresh) |
| `/root/.config/gh` | PVC | GitHub credentials (persistent) |

### Repo Cloning

**Single repo**: Cloned directly into `/workspace`
```bash
git clone https://github.com/owner/repo.git /workspace
```

**Multiple repos**: Each in its own subdirectory
```bash
/workspace/
  ├── repo1/
  ├── repo2/
  └── repo3/
```

## Development

### Building Locally

```bash
# Build Docker image
docker build -t happy-claude-coders:dev .

# Test locally
docker run -it \
  -e GIT_USER_NAME="Test User" \
  -e GIT_USER_EMAIL="test@example.com" \
  -e WORKSPACE_REPOS="https://github.com/slopus/happy-cli.git" \
  happy-claude-coders:dev
```

### Release Process

This project uses **Release Please** for automated semantic versioning.

**Commit message format**:
- `feat: ...` - New feature (minor version bump)
- `fix: ...` - Bug fix (patch version bump)
- `feat!: ...` or `fix!: ...` - Breaking change (major version bump)

**Workflow**:
1. Merge PRs to `main` with conventional commits
2. Release Please creates a release PR
3. Review and merge the release PR
4. GitHub Actions automatically:
   - Tags the release
   - Builds and pushes Docker image to GHCR
   - Packages and pushes Helm chart to GHCR

## Files and Structure

```
happy-little-claude-coders/
├── Dockerfile                          # Container image
├── scripts/
│   ├── entrypoint.sh                  # Container entrypoint
│   └── clone-repos.sh                 # Repo cloning helper
├── chart/happy-claude-coders/         # Helm chart
│   ├── Chart.yaml
│   ├── values.yaml
│   └── templates/
│       ├── deployment.yaml
│       ├── configmap.yaml
│       ├── pvc.yaml
│       ├── serviceaccount.yaml
│       └── _helpers.tpl
├── .github/workflows/                  # CI/CD
│   ├── release-please.yml
│   ├── docker-build.yml
│   └── helm-release.yml
├── .release-please-config.json        # Release Please config
├── .release-please-manifest.json      # Version manifest
└── package.json                        # Version tracking
```

## Troubleshooting

### GitHub authentication fails

Check if credentials are persisted:
```bash
kubectl exec deployment/my-workspace -- gh auth status
```

Re-authenticate if needed:
```bash
kubectl exec -it deployment/my-workspace -- gh auth login
```

### Claude Code not working

Verify the OAuth token secret:
```bash
kubectl get secret claude-oauth -o jsonpath='{.data.token}' | base64 -d
```

### Repos not cloning

Check the workspace repos configuration:
```bash
kubectl exec deployment/my-workspace -- env | grep WORKSPACE_REPOS
```

Manually trigger cloning:
```bash
kubectl exec deployment/my-workspace -- clone-repos
```

## Contributing

Contributions welcome! Please use conventional commits for all pull requests.

## License

MIT

## Credits

- [Claude Code](https://claude.ai/code) by Anthropic
- [Happy CLI](https://github.com/slopus/happy-cli) by slopus
