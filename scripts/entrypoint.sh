#!/bin/bash
set -e

# Configure git identity from environment variables
if [ -n "$GIT_USER_NAME" ]; then
    git config --global user.name "$GIT_USER_NAME"
fi

if [ -n "$GIT_USER_EMAIL" ]; then
    git config --global user.email "$GIT_USER_EMAIL"
fi

# Function to check if repos should be cloned
should_clone_repos() {
    if [ -z "$WORKSPACE_REPOS" ]; then
        return 1  # No repos configured
    fi

    REPO_COUNT=$(echo "$WORKSPACE_REPOS" | wc -w)
    if [ "$REPO_COUNT" -eq 1 ]; then
        # Single repo: check if /workspace/.git exists
        [ ! -d "/workspace/.git" ]
    else
        # Multiple repos: check if at least one is missing
        for repo in $WORKSPACE_REPOS; do
            repo_name=$(basename -s .git "$repo")
            if [ ! -d "/workspace/$repo_name" ]; then
                return 0  # At least one repo missing
            fi
        done
        return 1  # All repos exist
    fi
}

# Check for GitHub App token (from sidecar)
if [ -f /var/run/github/token ]; then
    echo "Using GitHub App authentication..."

    # Authenticate gh CLI with token
    if ! gh auth status &>/dev/null; then
        cat /var/run/github/token | gh auth login --with-token
        echo "✓ GitHub App authenticated"
    else
        echo "✓ GitHub authenticated"
    fi

    # Setup git credential helper
    gh auth setup-git 2>/dev/null || true

    # Auto-clone repos if needed
    if should_clone_repos; then
        echo "Cloning repositories..."
        /usr/local/bin/clone-repos
    else
        echo "✓ Repositories already present"
    fi
elif gh auth status &>/dev/null; then
    echo "✓ GitHub authenticated"

    # Setup git credential helper
    gh auth setup-git 2>/dev/null || true

    # Auto-clone repos if needed
    if should_clone_repos; then
        echo "Cloning repositories..."
        /usr/local/bin/clone-repos
    else
        echo "✓ Repositories already present"
    fi
else
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
    echo "  GitHub Authentication Required" >&2
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
    echo "" >&2
    echo "To authenticate with GitHub, run:" >&2
    echo "  $ gh auth login" >&2
    echo "" >&2
    echo "After authentication, repos will auto-clone in the background." >&2
    echo "Or manually clone with:" >&2
    echo "  $ clone-repos" >&2
    echo "" >&2
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2

    # Start background watcher that auto-clones after authentication
    if [ -n "$WORKSPACE_REPOS" ]; then
        (
            echo "Waiting for GitHub authentication..." >&2
            while ! gh auth status &>/dev/null; do
                sleep 5
            done
            echo "" >&2
            echo "✓ GitHub authenticated - auto-cloning repositories..." >&2
            gh auth setup-git 2>/dev/null || true
            /usr/local/bin/clone-repos
        ) &
    fi
fi

# Execute the provided command or start a shell
exec "$@"
