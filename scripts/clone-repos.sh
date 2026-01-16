#!/bin/bash
set -e

# Clone repositories from WORKSPACE_REPOS environment variable
# Format: Space-separated list of git URLs

if [ -z "$WORKSPACE_REPOS" ]; then
    echo "No repositories configured (WORKSPACE_REPOS is empty)"
    echo "Set WORKSPACE_REPOS to a space-separated list of git URLs"
    exit 1
fi

# Count repos
REPO_COUNT=$(echo "$WORKSPACE_REPOS" | wc -w)

if [ "$REPO_COUNT" -eq 1 ]; then
    # Single repo: clone directly into /workspace
    if [ ! -d "/workspace/.git" ]; then
        # Check if /workspace has content (excluding lost+found which is common in empty PVCs)
        if [ -n "$(ls -A /workspace 2>/dev/null | grep -v '^lost+found$')" ]; then
            echo "⚠ Warning: /workspace exists with content but is not a git repository"
            echo "   This usually means a previous clone failed. Cleaning up..."
            # Preserve lost+found if it exists (required by some storage systems)
            find /workspace -mindepth 1 -maxdepth 1 ! -name 'lost+found' -exec rm -rf {} +
            echo "   Cleaned non-git content from /workspace"
        fi
        echo "Cloning single repository into /workspace..."
        git clone "$WORKSPACE_REPOS" /workspace
        echo "✓ Repository cloned successfully"
    else
        echo "✓ Repository already exists in /workspace"
    fi
else
    # Multiple repos: clone each into its own subdirectory
    echo "Cloning $REPO_COUNT repositories..."
    for repo in $WORKSPACE_REPOS; do
        repo_name=$(basename -s .git "$repo")
        if [ ! -d "/workspace/$repo_name" ]; then
            echo "  → $repo_name (cloning)"
            git clone "$repo" "/workspace/$repo_name"
        else
            echo "  → $repo_name (already exists)"
        fi
    done
    echo "✓ Repository check complete"
fi
