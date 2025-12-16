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
    echo "Cloning single repository into /workspace..."
    git clone "$WORKSPACE_REPOS" /workspace
    echo "✓ Repository cloned successfully"
else
    # Multiple repos: clone each into its own subdirectory
    echo "Cloning $REPO_COUNT repositories..."
    for repo in $WORKSPACE_REPOS; do
        repo_name=$(basename -s .git "$repo")
        echo "  → $repo_name"
        git clone "$repo" "/workspace/$repo_name"
    done
    echo "✓ All repositories cloned successfully"
fi
