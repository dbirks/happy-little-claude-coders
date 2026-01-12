# Local Kind Testing

This directory contains configuration and scripts for local testing using [Kind](https://kind.sigs.k8s.io/) (Kubernetes in Docker).

## Prerequisites

- Docker
- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [Task](https://taskfile.dev/installation/) (task runner)
- kubectl
- helm

## Quick Start

### Create cluster and install chart

```bash
task test
```

This will:
1. Create a Kind cluster named `happy-test`
2. Install the happy-little-claude-coders Helm chart
3. Wait for pods to stabilize
4. Verify deployment is working
5. Show logs with authentication instructions

### View status

```bash
task status
```

### View logs

```bash
task logs
```

### Exec into pod

```bash
task exec
```

### Clean up

```bash
task down
```

## Available Tasks

Run `task` to see all available tasks:

```bash
$ task

task: Available tasks for this project:
* auth:             Run happy CLI authentication interactively
* default:          Show available tasks
* down:             Delete Kind cluster
* exec:             Exec into test-workspace pod
* logs:             Show logs from test-workspace pod
* rebuild:          Rebuild Docker image and update deployment
* status:           Show cluster and deployment status
* test:             Run full test cycle (up, verify, down)
* up:               Create Kind cluster and install happy-little-claude-coders
* verify:           Verify deployment is working
* chart:install:    Install happy-little-claude-coders Helm chart
* chart:uninstall:  Uninstall happy-little-claude-coders Helm chart
* cluster:create:   Create Kind cluster
```

## Testing Workflow

### 1. Initial Setup

```bash
# Create cluster and install chart
task up

# Check status
task status
```

### 2. View Logs

```bash
# Follow logs from test-workspace
task logs
```

You should see:
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Happy CLI Authentication Required
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

To authenticate with Happy CLI, run:
  $ kubectl exec -it deployment/happy-little-claude-coders-WORKSPACE -- bash
  $ happy --no-qr

Then:
  1. Select option 1 (Mobile App)
  2. Scan the pairing code in your Happy mobile app
  3. Exit the shell (credentials persist)

After authentication, Happy CLI will auto-start in the background.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Waiting for Happy CLI authentication...
```

### 3. Manual Authentication (Optional)

If you want to test the authentication flow:

```bash
# Exec into pod
task exec

# Inside pod, run happy CLI
happy --no-qr

# Follow prompts to authenticate
# Exit when done
```

The background watcher will detect the credentials and auto-start the happy daemon.

### 4. Testing Local Changes

If you make changes to the code:

```bash
# Rebuild and reload
task rebuild
```

This will:
1. Build a new Docker image with tag `local`
2. Load it into the Kind cluster
3. Update the Helm release to use the local image
4. Restart pods

### 5. Clean Up

```bash
# Delete the cluster
task down
```

## Configuration

### values.yaml

The local test configuration is in `local-test/values.yaml`:

```yaml
# Simplified for local testing
workspaces:
  - name: test-workspace
    repos:
      - "https://github.com/dbirks/happy-little-claude-coders.git"
    storageSize: 2Gi
    persistent: false  # Uses emptyDir for speed

# GitHub App and Claude OAuth disabled for local testing
githubApp:
  enabled: false
```

You can customize this file for your testing needs.

## Troubleshooting

### Pods not starting

```bash
# Check pod status
kubectl get pods -n default

# Describe pod for events
kubectl describe pod -l app.kubernetes.io/name=happy-little-claude-coders

# Check logs
task logs
```

### Image pull errors

If using a local image:

```bash
# Rebuild and ensure image is loaded into Kind
task rebuild
```

### Cluster won't create

```bash
# Clean up any existing cluster first
task down

# Try again
task up
```

### Check if Kind cluster exists

```bash
kind get clusters
```

### Access Kubernetes directly

```bash
# Set context
kubectl config use-context kind-happy-test

# List all resources
kubectl get all -n default
```

## Testing Different Scenarios

### Test without repos

Edit `local-test/values.yaml`:

```yaml
workspaces:
  - name: test-workspace
    repos: []  # No repos
```

Then:

```bash
task down
task up
```

### Test with persistent storage

Edit `local-test/values.yaml`:

```yaml
workspaces:
  - name: test-workspace
    persistent: true  # Use PVC instead of emptyDir
    storageSize: 5Gi
```

### Test with multiple workspaces

Edit `local-test/values.yaml`:

```yaml
workspaces:
  - name: workspace-1
    repos:
      - "https://github.com/example/repo1.git"
  - name: workspace-2
    repos:
      - "https://github.com/example/repo2.git"
```

## CI/CD Integration

You can use these same tasks in CI:

```yaml
# GitHub Actions example
- name: Test with Kind
  run: |
    task test
    # Additional test commands here
    task down
```

## Notes

- The Kind cluster uses the latest Kubernetes version by default
- Storage is ephemeral (emptyDir) by default for fast testing
- GitHub App and Claude OAuth are disabled in local testing
- The test workspace clones the happy-little-claude-coders repo by default
