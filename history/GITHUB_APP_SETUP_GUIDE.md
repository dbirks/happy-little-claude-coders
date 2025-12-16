# GitHub App Setup Guide for Workspace Authentication

This comprehensive guide walks you through creating, configuring, and deploying a GitHub App for workspace authentication, including Kubernetes integration.

## Table of Contents

1. [Creating the GitHub App](#1-creating-the-github-app)
2. [Configuration Details](#2-configuration-details)
3. [Permissions Setup](#3-permissions-setup)
4. [Private Key Generation](#4-private-key-generation)
5. [Installation](#5-installation)
6. [Testing](#6-testing)
7. [Kubernetes Deployment](#7-kubernetes-deployment)

---

## 1. Creating the GitHub App

### Step 1.1: Navigate to GitHub App Settings

**For Personal Account:**
1. Go to your GitHub profile → Click your profile picture (top right)
2. Select **Settings**
3. Scroll down to **Developer settings** (bottom left sidebar)
4. Click **GitHub Apps**
5. Click **New GitHub App**

**For Organization:**
1. Go to your organization's GitHub page
2. Click **Settings**
3. Scroll down to **Developer settings** (left sidebar)
4. Click **GitHub Apps**
5. Click **New GitHub App**

**Decision Guide: Personal vs Organization?**
- Use **Organization** if:
  - Multiple team members need to manage the app
  - App accesses organization repositories
  - You want centralized billing and permissions
- Use **Personal** if:
  - It's for personal projects only
  - You're the sole maintainer

### Step 1.2: URL Reference
Direct link format: `https://github.com/settings/apps/new` (personal) or `https://github.com/organizations/YOUR_ORG/settings/apps/new` (organization)

---

## 2. Configuration Details

### Step 2.1: Basic Information

**GitHub App Name**
- Maximum 34 characters
- Must be unique across GitHub
- Convention for internal apps: `company-workspace-auth` or `myorg-git-clone-app`
- Example: `acme-workspace-authenticator`

**Description** (Optional but Recommended)
- Describe the purpose clearly
- Example: "Internal workspace authentication for cloning private repositories in Kubernetes environments"

### Step 2.2: Homepage URL

**What to use for internal apps:**

Option 1: **Internal documentation URL**
```
https://wiki.company.com/github-app-setup
```

Option 2: **GitHub repository with app documentation**
```
https://github.com/yourorg/workspace-infrastructure
```

Option 3: **Placeholder (if truly internal)**
```
https://github.com/yourorg
```

**Note:** This is a required field but won't be publicly visible if your app is set to "Only on this account"

### Step 2.3: Callback URL

**Do you need it?**
- **Required for:** OAuth flows (user authorization, web-based login)
- **NOT required for:** Server-to-server authentication (cloning repos, API access)

**For workspace authentication (cloning repos):**
- You can leave this **blank** or use a placeholder
- Example placeholder: `http://localhost:3000/callback`

**If you plan to add OAuth later:**
```
https://your-app-domain.com/auth/github/callback
```

### Step 2.4: Webhook Configuration

**Webhook URL**
- **Optional** unless you need event notifications
- For basic repository cloning: **Not required** - you can uncheck "Active" under webhook settings

**If you do enable webhooks:**
```
https://your-webhook-receiver.com/webhook
```

**Webhook Secret** (Recommended if using webhooks)
- Generate a random secret: `openssl rand -hex 32`
- Store it securely alongside your private key

**For workspace authentication only:**
1. Scroll to "Webhook" section
2. **Uncheck** "Active"
3. Save

### Step 2.5: Repository Access Settings

**Where can this GitHub App be installed?**

- **Only on this account** (Recommended for internal use)
  - App only appears in your account/organization
  - Cannot be installed by others
  - Best for internal tooling

- **Any account**
  - App can be listed publicly
  - Others can request installation
  - Use for public/open-source apps

**For workspace authentication:** Choose **"Only on this account"**

---

## 3. Permissions Setup

### Step 3.1: Understanding Permission Levels

GitHub Apps use fine-grained permissions with three levels:
- **No access** - No permission granted
- **Read-only** - Can view but not modify
- **Read and write** - Full access

### Step 3.2: Required Permissions for Cloning Repositories

**Minimum Permissions:**

| Permission | Level | Required? | Purpose |
|------------|-------|-----------|---------|
| **Contents** | Read-only | **Yes** | Clone and read repository files |
| **Metadata** | Read-only | **Yes** | Automatically granted - provides access to repository metadata |

**How to set:**
1. Scroll to **Repository permissions** section
2. Find **Contents** → Select **Read-only** from dropdown
3. **Metadata** is automatically set to Read-only (cannot be changed)

### Step 3.3: Optional Permissions (Common Use Cases)

**If you need to push code:**
| Permission | Level | Purpose |
|------------|-------|---------|
| Contents | Read and write | Push commits, create branches |

**If you need to manage workflows:**
| Permission | Level | Purpose |
|------------|-------|---------|
| Actions | Read and write | Trigger GitHub Actions |
| Workflows | Read and write | Modify workflow files |

**If you need pull request access:**
| Permission | Level | Purpose |
|------------|-------|---------|
| Pull requests | Read-only | View PRs |
| Pull requests | Read and write | Create, modify, merge PRs |

### Step 3.4: Account Permissions

Usually **not needed** for repository cloning. Leave all at "No access" unless you specifically need organization-level access.

### Step 3.5: Permission Summary Screenshot

When configured, your permissions section should show:
```
Repository permissions:
  ✓ Contents: Read-only
  ✓ Metadata: Read-only (automatically granted)
```

---

## 4. Private Key Generation

### Step 4.1: Generate the Private Key

**After creating your app:**
1. You'll be redirected to your app's settings page
2. Scroll to **Private keys** section (near the bottom)
3. Click **Generate a private key**
4. A `.pem` file will automatically download

**File format:** `your-app-name.YYYY-MM-DD.private-key.pem`

**Example:**
```
acme-workspace-authenticator.2025-12-15.private-key.pem
```

### Step 4.2: Secure Storage Best Practices

**DO:**
- ✅ Store in a password manager (1Password, LastPass, Bitwarden)
- ✅ Store in a secrets management service (AWS Secrets Manager, HashiCorp Vault)
- ✅ Keep a backup in an encrypted location
- ✅ Restrict file permissions: `chmod 600 private-key.pem`

**DON'T:**
- ❌ Commit to Git repositories
- ❌ Share via email or chat
- ❌ Store in unencrypted cloud storage
- ❌ Leave in Downloads folder

### Step 4.3: Understanding the Private Key Format

The private key is in PEM format and looks like:
```
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA...
[many lines of base64-encoded data]
...
-----END RSA PRIVATE KEY-----
```

**Key facts:**
- This is an RSA private key
- Used to generate JWT tokens for authentication
- Cannot be regenerated - if lost, you must create a new one
- You can have multiple active private keys (for rotation)

### Step 4.4: Storing in Kubernetes Secrets

**Method 1: Direct kubectl command**

```bash
# Create secret from file
kubectl create secret generic github-app-credentials \
  --from-file=private-key=./your-app.private-key.pem \
  --namespace=your-namespace
```

**Method 2: Base64 encoding for YAML**

```bash
# Encode the private key
cat your-app.private-key.pem | base64 -w 0

# Copy the output and use in secret YAML
```

**secret.yaml:**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: github-app-credentials
  namespace: your-namespace
type: Opaque
data:
  private-key: <base64-encoded-private-key-here>
  app-id: "<base64-encoded-app-id>"
  installation-id: "<base64-encoded-installation-id>"
```

Apply:
```bash
kubectl apply -f secret.yaml
```

**Method 3: Using External Secrets Operator** (Recommended for production)

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: github-app-credentials
  namespace: your-namespace
spec:
  secretStoreRef:
    name: vault-backend
    kind: SecretStore
  target:
    name: github-app-credentials
  data:
    - secretKey: private-key
      remoteRef:
        key: github-app/private-key
    - secretKey: app-id
      remoteRef:
        key: github-app/app-id
    - secretKey: installation-id
      remoteRef:
        key: github-app/installation-id
```

**Method 4: Sealed Secrets** (For GitOps)

```bash
# Encrypt secret for git storage
kubectl create secret generic github-app-credentials \
  --from-file=private-key=./your-app.private-key.pem \
  --dry-run=client -o yaml | \
  kubeseal -o yaml > sealed-secret.yaml

# Commit sealed-secret.yaml to git
git add sealed-secret.yaml
```

### Step 4.5: Key Rotation

**Best practice:** Rotate keys every 6-12 months

```bash
# Generate new key from GitHub UI
# Download new-key.pem

# Update Kubernetes secret
kubectl create secret generic github-app-credentials \
  --from-file=private-key=./new-key.pem \
  --namespace=your-namespace \
  --dry-run=client -o yaml | kubectl apply -f -

# Restart pods using the secret
kubectl rollout restart deployment/your-deployment -n your-namespace

# After verification, revoke old key from GitHub UI
```

---

## 5. Installation

### Step 5.1: Install the App

**Immediately after creating the app:**
1. You'll see a green button "Install App" or "Install"
2. Click it

**From the app settings page:**
1. Go to `https://github.com/settings/apps/your-app-name`
2. Click "Install App" in the left sidebar
3. Select your account or organization

### Step 5.2: Select Repository Access

**You'll see two options:**

**Option 1: All repositories**
- Grants access to all current and future repositories
- Simpler management
- Higher security risk
- Use for: Internal tooling with broad access needs

**Option 2: Only select repositories** (Recommended)
- Choose specific repositories
- More secure (principle of least privilege)
- Must update when adding new repositories
- Use for: Production deployments, limited-scope apps

**How to select:**
1. Choose "Only select repositories"
2. Click the "Select repositories" dropdown
3. Search and select repositories
4. Click "Install"

### Step 5.3: Finding the Installation ID

**Method 1: From URL (Easiest)**

After installation, you'll be redirected to:
```
https://github.com/organizations/YOUR_ORG/settings/installations/12345678
```

The number at the end (`12345678`) is your **installation ID**.

**For personal accounts:**
```
https://github.com/settings/installations/12345678
```

**Method 2: Through Settings UI**

For organizations:
1. Go to organization page → **Settings**
2. Click **GitHub Apps** (under Integrations in left sidebar)
3. Click **Configure** next to your app
4. Check the URL - the installation ID is at the end

For personal accounts:
1. Profile picture → **Settings**
2. **Integrations** → **Applications**
3. Find your app → **Configure**
4. Check the URL

**Method 3: Using the GitHub API**

```bash
# Get JWT token first (using your App ID and private key)
# Then query installations

curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  https://api.github.com/app/installations
```

Response:
```json
[
  {
    "id": 12345678,
    "account": {
      "login": "your-org",
      "type": "Organization"
    },
    ...
  }
]
```

The `id` field is your installation ID.

### Step 5.4: Finding the App ID

**Method 1: From App Settings Page**

1. Go to `https://github.com/settings/apps/your-app-name`
2. Under "About" section (top of page)
3. You'll see:
   ```
   App ID: 123456
   Client ID: Iv1.a1b2c3d4e5f6g7h8
   ```

**The App ID** is the number (e.g., `123456`), **not** the Client ID.

**Method 2: From URL**

When viewing your app settings:
```
https://github.com/settings/apps/your-app-name
```

Or check the API URL shown on the settings page.

### Step 5.5: Document Your IDs

Create a secure note with all three values:

```yaml
GitHub App: acme-workspace-authenticator
App ID: 123456
Installation ID: 12345678
Private Key: Stored in 1Password/Vault
Organization: acme-corp
Repositories: my-repo-1, my-repo-2
Created: 2025-12-15
```

---

## 6. Testing

### Step 6.1: Install Testing Tools

**Option 1: Using GitHub CLI** (Recommended)

```bash
# Install gh CLI
# macOS
brew install gh

# Linux
sudo apt install gh

# Windows
winget install GitHub.cli
```

**Option 2: Using curl and jq**

```bash
sudo apt install jq  # or brew install jq
```

**Option 3: Using Python**

```bash
pip install pyjwt cryptography requests
```

### Step 6.2: Generate a JWT Token

GitHub Apps use JWT (JSON Web Tokens) for authentication. Here's how to generate one:

**Using Python Script (create `generate_jwt.py`):**

```python
#!/usr/bin/env python3
import jwt
import time
import sys

# Configuration
APP_ID = "123456"  # Replace with your App ID
PRIVATE_KEY_PATH = "./your-app.private-key.pem"  # Path to your private key

# Read private key
with open(PRIVATE_KEY_PATH, 'r') as key_file:
    private_key = key_file.read()

# Generate JWT
payload = {
    'iat': int(time.time()),           # Issued at time
    'exp': int(time.time()) + 600,      # JWT expiration (10 minutes)
    'iss': APP_ID                       # GitHub App's ID
}

encoded_jwt = jwt.encode(payload, private_key, algorithm='RS256')
print(encoded_jwt)
```

Run:
```bash
python3 generate_jwt.py
```

**Using Ruby:**

```ruby
require 'openssl'
require 'jwt'

APP_ID = '123456'
PRIVATE_KEY_PATH = './your-app.private-key.pem'

private_key = OpenSSL::PKey::RSA.new(File.read(PRIVATE_KEY_PATH))

payload = {
  iat: Time.now.to_i,
  exp: Time.now.to_i + (10 * 60),
  iss: APP_ID
}

jwt = JWT.encode(payload, private_key, 'RS256')
puts jwt
```

### Step 6.3: Generate an Installation Access Token

```bash
# Save JWT from previous step
JWT_TOKEN="your-jwt-token-here"
INSTALLATION_ID="12345678"

# Get installation access token
curl -X POST \
  -H "Authorization: Bearer ${JWT_TOKEN}" \
  -H "Accept: application/vnd.github+json" \
  https://api.github.com/app/installations/${INSTALLATION_ID}/access_tokens
```

**Response:**
```json
{
  "token": "ghs_1234567890abcdefghijklmnopqrstuvwxyz",
  "expires_at": "2025-12-15T12:00:00Z",
  "permissions": {
    "contents": "read",
    "metadata": "read"
  },
  "repository_selection": "selected"
}
```

**Save the token:**
```bash
GITHUB_TOKEN="ghs_1234567890abcdefghijklmnopqrstuvwxyz"
```

### Step 6.4: Test Repository Access

**Test 1: List accessible repositories**

```bash
curl -H "Authorization: token ${GITHUB_TOKEN}" \
  -H "Accept: application/vnd.github+json" \
  https://api.github.com/installation/repositories
```

**Expected response:**
```json
{
  "total_count": 2,
  "repositories": [
    {
      "id": 123456789,
      "name": "my-repo-1",
      "full_name": "acme-corp/my-repo-1",
      "private": true,
      ...
    }
  ]
}
```

**Test 2: Clone a repository**

```bash
# Clone using the installation token
git clone https://x-access-token:${GITHUB_TOKEN}@github.com/acme-corp/my-repo-1.git

# Or set git credential helper
git config --global credential.helper cache
echo "https://x-access-token:${GITHUB_TOKEN}@github.com" | git credential approve
git clone https://github.com/acme-corp/my-repo-1.git
```

**Test 3: Read file contents**

```bash
curl -H "Authorization: token ${GITHUB_TOKEN}" \
  -H "Accept: application/vnd.github+json" \
  https://api.github.com/repos/acme-corp/my-repo-1/contents/README.md
```

### Step 6.5: Verify Permissions

**Check what permissions your token has:**

```bash
curl -I -H "Authorization: token ${GITHUB_TOKEN}" \
  https://api.github.com/repos/acme-corp/my-repo-1
```

Look for the `X-OAuth-Scopes` header in the response.

**Test negative case (should fail):**

```bash
# Try to write (should fail with read-only permission)
curl -X PUT \
  -H "Authorization: token ${GITHUB_TOKEN}" \
  -H "Accept: application/vnd.github+json" \
  https://api.github.com/repos/acme-corp/my-repo-1/contents/test.txt \
  -d '{"message":"test","content":"dGVzdA=="}'
```

Expected: `403 Forbidden` (if you only have read permissions)

### Step 6.6: Automated Testing Script

**Create `test_github_app.sh`:**

```bash
#!/bin/bash
set -e

# Configuration
APP_ID="123456"
INSTALLATION_ID="12345678"
PRIVATE_KEY_PATH="./your-app.private-key.pem"
TEST_REPO="acme-corp/my-repo-1"

echo "=== GitHub App Testing Script ==="

# Generate JWT
echo "1. Generating JWT..."
JWT=$(python3 - <<EOF
import jwt
import time

with open('${PRIVATE_KEY_PATH}', 'r') as f:
    private_key = f.read()

payload = {
    'iat': int(time.time()),
    'exp': int(time.time()) + 600,
    'iss': '${APP_ID}'
}

print(jwt.encode(payload, private_key, algorithm='RS256'))
EOF
)

echo "✓ JWT generated"

# Get installation token
echo "2. Getting installation access token..."
RESPONSE=$(curl -s -X POST \
  -H "Authorization: Bearer ${JWT}" \
  -H "Accept: application/vnd.github+json" \
  https://api.github.com/app/installations/${INSTALLATION_ID}/access_tokens)

TOKEN=$(echo $RESPONSE | jq -r '.token')

if [ "$TOKEN" == "null" ]; then
  echo "✗ Failed to get token"
  echo $RESPONSE | jq
  exit 1
fi

echo "✓ Installation token obtained"

# Test repository access
echo "3. Testing repository access..."
REPO_RESPONSE=$(curl -s \
  -H "Authorization: token ${TOKEN}" \
  -H "Accept: application/vnd.github+json" \
  https://api.github.com/repos/${TEST_REPO})

REPO_NAME=$(echo $REPO_RESPONSE | jq -r '.name')

if [ "$REPO_NAME" == "null" ]; then
  echo "✗ Cannot access repository"
  echo $REPO_RESPONSE | jq
  exit 1
fi

echo "✓ Successfully accessed repository: ${REPO_NAME}"

# Test clone
echo "4. Testing git clone..."
TEST_DIR="/tmp/github-app-test-$$"
if git clone https://x-access-token:${TOKEN}@github.com/${TEST_REPO}.git ${TEST_DIR} 2>&1; then
  echo "✓ Successfully cloned repository"
  rm -rf ${TEST_DIR}
else
  echo "✗ Failed to clone repository"
  exit 1
fi

echo ""
echo "=== All tests passed! ==="
echo "Your GitHub App is working correctly."
echo ""
echo "Token (valid for 1 hour): ${TOKEN}"
```

Run:
```bash
chmod +x test_github_app.sh
./test_github_app.sh
```

---

## 7. Kubernetes Deployment

### Step 7.1: Create Kubernetes Secret

**Option A: From files**

```bash
kubectl create secret generic github-app-credentials \
  --from-literal=app-id=123456 \
  --from-literal=installation-id=12345678 \
  --from-file=private-key=./your-app.private-key.pem \
  -n your-namespace
```

**Option B: From YAML**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: github-app-credentials
  namespace: workspace
type: Opaque
stringData:
  app-id: "123456"
  installation-id: "12345678"
data:
  # Base64 encode: cat private-key.pem | base64 -w 0
  private-key: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBLi4uCg==
```

Apply:
```bash
kubectl apply -f github-app-secret.yaml
```

### Step 7.2: Use Secret in Deployment

**Example: InitContainer for git clone**

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: workspace-with-git
  namespace: workspace
spec:
  initContainers:
  - name: clone-repos
    image: bitnami/git:latest
    env:
    - name: GITHUB_APP_ID
      valueFrom:
        secretKeyRef:
          name: github-app-credentials
          key: app-id
    - name: GITHUB_INSTALLATION_ID
      valueFrom:
        secretKeyRef:
          name: github-app-credentials
          key: installation-id
    - name: GITHUB_PRIVATE_KEY
      valueFrom:
        secretKeyRef:
          name: github-app-credentials
          key: private-key
    volumeMounts:
    - name: workspace
      mountPath: /workspace
    command:
    - /bin/bash
    - -c
    - |
      set -e

      # Install dependencies
      apt-get update && apt-get install -y python3 python3-pip
      pip3 install pyjwt cryptography requests

      # Generate GitHub App token
      GITHUB_TOKEN=$(python3 <<'EOF'
      import jwt
      import time
      import os
      import requests

      app_id = os.environ['GITHUB_APP_ID']
      installation_id = os.environ['GITHUB_INSTALLATION_ID']
      private_key = os.environ['GITHUB_PRIVATE_KEY']

      # Generate JWT
      payload = {
          'iat': int(time.time()),
          'exp': int(time.time()) + 600,
          'iss': app_id
      }
      jwt_token = jwt.encode(payload, private_key, algorithm='RS256')

      # Get installation token
      headers = {
          'Authorization': f'Bearer {jwt_token}',
          'Accept': 'application/vnd.github+json'
      }
      response = requests.post(
          f'https://api.github.com/app/installations/{installation_id}/access_tokens',
          headers=headers
      )
      print(response.json()['token'])
      EOF
      )

      # Configure git
      git config --global credential.helper store
      echo "https://x-access-token:${GITHUB_TOKEN}@github.com" > ~/.git-credentials

      # Clone repositories
      cd /workspace
      git clone https://github.com/acme-corp/my-repo-1.git
      git clone https://github.com/acme-corp/my-repo-2.git

      echo "Repositories cloned successfully"

  containers:
  - name: main
    image: your-workspace-image:latest
    volumeMounts:
    - name: workspace
      mountPath: /workspace

  volumes:
  - name: workspace
    emptyDir: {}
```

### Step 7.3: Using a Sidecar for Token Refresh

**Problem:** Installation tokens expire after 1 hour.

**Solution:** Use a sidecar container to refresh tokens.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: workspace
  namespace: workspace
spec:
  replicas: 1
  selector:
    matchLabels:
      app: workspace
  template:
    metadata:
      labels:
        app: workspace
    spec:
      containers:
      # Main application container
      - name: workspace
        image: your-workspace-image:latest
        volumeMounts:
        - name: git-token
          mountPath: /var/run/secrets/github
          readOnly: true

      # Token refresh sidecar
      - name: github-token-refresher
        image: python:3.11-slim
        env:
        - name: GITHUB_APP_ID
          valueFrom:
            secretKeyRef:
              name: github-app-credentials
              key: app-id
        - name: GITHUB_INSTALLATION_ID
          valueFrom:
            secretKeyRef:
              name: github-app-credentials
              key: installation-id
        - name: GITHUB_PRIVATE_KEY
          valueFrom:
            secretKeyRef:
              name: github-app-credentials
              key: private-key
        volumeMounts:
        - name: git-token
          mountPath: /var/run/secrets/github
        command:
        - /bin/bash
        - -c
        - |
          pip install pyjwt cryptography requests

          python3 <<'EOF'
          import jwt
          import time
          import os
          import requests

          def get_installation_token():
              app_id = os.environ['GITHUB_APP_ID']
              installation_id = os.environ['GITHUB_INSTALLATION_ID']
              private_key = os.environ['GITHUB_PRIVATE_KEY']

              # Generate JWT
              payload = {
                  'iat': int(time.time()),
                  'exp': int(time.time()) + 600,
                  'iss': app_id
              }
              jwt_token = jwt.encode(payload, private_key, algorithm='RS256')

              # Get installation token
              headers = {
                  'Authorization': f'Bearer {jwt_token}',
                  'Accept': 'application/vnd.github+json'
              }
              response = requests.post(
                  f'https://api.github.com/app/installations/{installation_id}/access_tokens',
                  headers=headers
              )
              return response.json()['token']

          # Refresh token every 50 minutes
          while True:
              token = get_installation_token()
              with open('/var/run/secrets/github/token', 'w') as f:
                  f.write(token)
              print(f"Token refreshed at {time.strftime('%Y-%m-%d %H:%M:%S')}")
              time.sleep(3000)  # 50 minutes
          EOF

      volumes:
      - name: git-token
        emptyDir:
          medium: Memory
```

### Step 7.4: Using External Secrets Operator

**Install External Secrets Operator:**

```bash
helm repo add external-secrets https://charts.external-secrets.io
helm install external-secrets external-secrets/external-secrets -n external-secrets-system --create-namespace
```

**Configure GitHub App token generation:**

```yaml
apiVersion: generators.external-secrets.io/v1alpha1
kind: GCRAccessToken
metadata:
  name: github-app-token
  namespace: workspace
spec:
  auth:
    privateKey:
      secretRef:
        name: github-app-credentials
        key: private-key
  projectID: "123456"  # GitHub App ID
```

### Step 7.5: Using github-app-secret CronJob

**Deploy the github-app-secret tool:**

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: github-app-token-refresher
  namespace: workspace
spec:
  schedule: "*/50 * * * *"  # Every 50 minutes
  jobTemplate:
    spec:
      template:
        spec:
          serviceAccountName: github-app-token-refresher
          containers:
          - name: refresher
            image: ghcr.io/fluxcd-community/github-app-secret:latest
            env:
            - name: GITHUB_APP_ID
              valueFrom:
                secretKeyRef:
                  name: github-app-credentials
                  key: app-id
            - name: GITHUB_APP_INSTALLATION_ID
              valueFrom:
                secretKeyRef:
                  name: github-app-credentials
                  key: installation-id
            - name: GITHUB_APP_PRIVATE_KEY
              valueFrom:
                secretKeyRef:
                  name: github-app-credentials
                  key: private-key
            - name: SECRET_NAME
              value: github-app-token
            - name: SECRET_NAMESPACE
              value: workspace
          restartPolicy: OnFailure
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: github-app-token-refresher
  namespace: workspace
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: github-app-token-refresher
  namespace: workspace
rules:
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["get", "create", "update", "patch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: github-app-token-refresher
  namespace: workspace
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: github-app-token-refresher
subjects:
- kind: ServiceAccount
  name: github-app-token-refresher
```

---

## Troubleshooting

### Common Issues

**Issue 1: "Resource not accessible by integration"**
- **Cause:** Missing permissions
- **Fix:** Add required permissions in app settings, re-authorize

**Issue 2: "Bad credentials"**
- **Cause:** Expired or invalid token
- **Fix:** Regenerate JWT and installation token

**Issue 3: "Not Found" when accessing repository**
- **Cause:** App not installed on repository
- **Fix:** Add repository to installation

**Issue 4: JWT expired**
- **Cause:** JWT tokens expire after 10 minutes
- **Fix:** Generate fresh JWT token

**Issue 5: Installation token expired**
- **Cause:** Installation tokens expire after 1 hour
- **Fix:** Refresh token using sidecar or CronJob

### Debugging Commands

```bash
# Check secret exists
kubectl get secret github-app-credentials -n workspace

# View secret contents (base64 encoded)
kubectl get secret github-app-credentials -n workspace -o yaml

# Decode specific field
kubectl get secret github-app-credentials -n workspace -o jsonpath='{.data.app-id}' | base64 -d

# Test from within cluster
kubectl run -it --rm debug --image=python:3.11-slim --restart=Never -- bash
# Then run your token generation script
```

---

## Security Best Practices

1. **Principle of Least Privilege**
   - Only grant necessary permissions
   - Use "Only select repositories" when possible
   - Choose read-only unless write is required

2. **Key Management**
   - Rotate private keys every 6-12 months
   - Use External Secrets Operator for production
   - Never commit keys to version control
   - Use sealed secrets for GitOps

3. **Token Handling**
   - Store tokens in memory volumes (emptyDir with medium: Memory)
   - Never log tokens
   - Use short-lived tokens
   - Implement automatic refresh

4. **Access Control**
   - Use Kubernetes RBAC to restrict secret access
   - Limit which pods can access the secret
   - Use separate namespaces for isolation

5. **Monitoring**
   - Monitor app installation events
   - Set up alerts for failed authentications
   - Audit repository access patterns
   - Review app permissions regularly

---

## Additional Resources

### Official Documentation
- [GitHub Apps Documentation](https://docs.github.com/en/apps)
- [Registering a GitHub App](https://docs.github.com/en/apps/creating-github-apps/registering-a-github-app/registering-a-github-app)
- [Authenticating as a GitHub App](https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/authenticating-as-a-github-app-installation)
- [Choosing permissions for a GitHub App](https://docs.github.com/en/apps/creating-github-apps/registering-a-github-app/choosing-permissions-for-a-github-app)

### Tools and Libraries
- [External Secrets Operator](https://external-secrets.io/latest/api/generator/github/)
- [github-app-secret (FluxCD Community)](https://github.com/fluxcd-community/github-app-secret)
- [Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets)
- [PyJWT Library](https://pyjwt.readthedocs.io/)

### Tutorials
- [Building our first GitHub App (Medium)](https://medium.com/swlh/building-the-first-github-app-3ea67a76c19a)
- [How to Create a GitHub App: Step-by-Step Guide](https://www.codewalnut.com/tutorials/how-to-create-a-github-app)
- [GitHub Community Discussion: Cloning with GitHub App](https://github.com/orgs/community/discussions/24575)

### Kubernetes Secrets Management
- [Kubernetes Secrets Documentation](https://kubernetes.io/docs/concepts/configuration/secret/)
- [How to manage Kubernetes secrets securely in Git](https://dev.to/coder_society/how-to-manage-kubernetes-secrets-securely-in-git-12c5)
- [How to store Kubernetes Secrets in Git](https://learnkube.com/kubernetes-secrets-in-git)

---

## Quick Reference

### Important URLs

| Resource | URL Format |
|----------|-----------|
| Create GitHub App (Personal) | `https://github.com/settings/apps/new` |
| Create GitHub App (Org) | `https://github.com/organizations/YOUR_ORG/settings/apps/new` |
| App Settings | `https://github.com/settings/apps/YOUR_APP_NAME` |
| Installation Settings (Org) | `https://github.com/organizations/YOUR_ORG/settings/installations/INSTALL_ID` |
| Installation Settings (Personal) | `https://github.com/settings/installations/INSTALL_ID` |
| GitHub API - Get Installations | `https://api.github.com/app/installations` |
| GitHub API - Get Token | `https://api.github.com/app/installations/INSTALL_ID/access_tokens` |

### Key Concepts

| Term | Description | Example |
|------|-------------|---------|
| App ID | Unique identifier for your GitHub App | `123456` |
| Installation ID | ID for app installation in org/account | `12345678` |
| Private Key | RSA key for signing JWTs | `.pem` file |
| JWT Token | Short-lived token for API auth | Valid 10 minutes |
| Installation Token | Token for accessing resources | Valid 1 hour |
| Client ID | OAuth client identifier | `Iv1.abc123` |

### Minimum Required Information

To use a GitHub App, you need these three pieces of information:

1. **App ID**: `123456`
2. **Installation ID**: `12345678`
3. **Private Key**: `your-app.private-key.pem`

### Token Generation Flow

```
Private Key + App ID
        ↓
    Generate JWT (valid 10 min)
        ↓
JWT + Installation ID → GitHub API
        ↓
Installation Access Token (valid 1 hour)
        ↓
    Use for Git operations
```

---

## Appendix: Complete Example

### Scenario
Create a GitHub App to clone private repositories in a Kubernetes workspace.

### Step-by-Step

**1. Create the app:**
- Go to: `https://github.com/organizations/acme-corp/settings/apps/new`
- Name: `acme-workspace-git`
- Homepage: `https://github.com/acme-corp`
- Webhook: Uncheck "Active"
- Permissions: Contents = Read-only
- Install: "Only on this account"
- Click "Create GitHub App"

**2. Generate private key:**
- Click "Generate a private key"
- Save `acme-workspace-git.2025-12-15.private-key.pem`
- Note App ID: `123456`

**3. Install the app:**
- Click "Install App"
- Select "acme-corp"
- Choose "Only select repositories"
- Select: `my-repo-1`, `my-repo-2`
- Click "Install"
- Note Installation ID from URL: `12345678`

**4. Create Kubernetes secret:**
```bash
kubectl create secret generic github-app-credentials \
  --from-literal=app-id=123456 \
  --from-literal=installation-id=12345678 \
  --from-file=private-key=./acme-workspace-git.2025-12-15.private-key.pem \
  -n workspace
```

**5. Test:**
```bash
# Create test script
cat > test.py <<'EOF'
import jwt
import time
import requests
import os

# Read from files/env
app_id = "123456"
installation_id = "12345678"
with open('./acme-workspace-git.2025-12-15.private-key.pem') as f:
    private_key = f.read()

# Generate JWT
payload = {
    'iat': int(time.time()),
    'exp': int(time.time()) + 600,
    'iss': app_id
}
jwt_token = jwt.encode(payload, private_key, algorithm='RS256')

# Get installation token
headers = {
    'Authorization': f'Bearer {jwt_token}',
    'Accept': 'application/vnd.github+json'
}
response = requests.post(
    f'https://api.github.com/app/installations/{installation_id}/access_tokens',
    headers=headers
)
token = response.json()['token']
print(f"Token: {token}")

# Test clone
import subprocess
result = subprocess.run([
    'git', 'clone',
    f'https://x-access-token:{token}@github.com/acme-corp/my-repo-1.git'
])
print("Clone successful!" if result.returncode == 0 else "Clone failed!")
EOF

# Run test
pip install pyjwt cryptography requests
python test.py
```

**6. Deploy to Kubernetes:**
Use one of the deployment examples in Section 7.

---

**Document Version:** 1.0
**Last Updated:** 2025-12-15
**Author:** AI-Generated Guide
**Based on:** GitHub official documentation and community resources
