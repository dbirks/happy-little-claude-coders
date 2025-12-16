#!/bin/bash
set -e

# Configure git identity from environment variables
if [ -n "$GIT_USER_NAME" ]; then
    git config --global user.name "$GIT_USER_NAME"
fi

if [ -n "$GIT_USER_EMAIL" ]; then
    git config --global user.email "$GIT_USER_EMAIL"
fi

# Check if GitHub CLI is authenticated
if gh auth status &>/dev/null; then
    echo "✓ GitHub authenticated"

    # Setup git credential helper
    gh auth setup-git 2>/dev/null || true

    # Auto-clone repos if WORKSPACE_REPOS is set and /workspace is empty
    if [ -n "$WORKSPACE_REPOS" ] && [ ! "$(ls -A /workspace 2>/dev/null)" ]; then
        echo "Cloning repositories..."
        /usr/local/bin/clone-repos
    fi
else
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
    echo "  GitHub Authentication Required" >&2
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
    echo "" >&2
    echo "To authenticate with GitHub, run:" >&2
    echo "  $ gh auth login" >&2
    echo "" >&2
    echo "After authentication, clone repos with:" >&2
    echo "  $ clone-repos" >&2
    echo "" >&2
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
fi

# Execute the provided command or start a shell
exec "$@"
