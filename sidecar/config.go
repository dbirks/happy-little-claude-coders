package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Default values for configuration.
const (
	// DefaultRefreshInterval is the default time between token refreshes.
	// GitHub App installation tokens expire after 1 hour, so we refresh
	// at 45 minutes to ensure continuous availability.
	DefaultRefreshInterval = 45 * time.Minute

	// DefaultTokenPath is the default location where the token is written.
	// This should be a shared volume (tmpfs) mounted in both the sidecar
	// and main container.
	DefaultTokenPath = "/var/run/github/token"

	// DefaultSecretsPath is the default directory containing GitHub App credentials.
	DefaultSecretsPath = "/var/run/secrets/github-app"
)

// Config holds the configuration for the GitHub App token refresh sidecar.
type Config struct {
	// GitHub App ID (numeric identifier for the app)
	AppID int64

	// Installation ID (specific installation of the app on an org/user)
	InstallationID int64

	// Private key in PEM format for signing JWTs
	PrivateKey []byte

	// Repositories to scope the token to (optional).
	// If empty, token has access to all repos in the installation.
	Repositories []string

	// RefreshInterval determines how often to refresh the token.
	RefreshInterval time.Duration

	// TokenPath is where the generated token will be written.
	TokenPath string
}

// LoadConfig reads configuration from environment variables and mounted secrets.
//
// Environment variables:
//   - WORKSPACE_REPOS: Space-separated list of repository URLs to scope tokens to
//   - REFRESH_INTERVAL_MINUTES: Token refresh interval in minutes (default: 45)
//
// Mounted secrets (from Kubernetes secret):
//   - /var/run/secrets/github-app/app-id: GitHub App ID
//   - /var/run/secrets/github-app/installation-id: Installation ID
//   - /var/run/secrets/github-app/private-key: Private key in PEM format
//
// Returns an error if any required configuration is missing or invalid.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		RefreshInterval: DefaultRefreshInterval,
		TokenPath:       DefaultTokenPath,
	}

	// Read GitHub App credentials from mounted secrets
	appID, err := readInt64File(DefaultSecretsPath + "/app-id")
	if err != nil {
		return nil, fmt.Errorf("read app-id: %w", err)
	}
	cfg.AppID = appID

	installationID, err := readInt64File(DefaultSecretsPath + "/installation-id")
	if err != nil {
		return nil, fmt.Errorf("read installation-id: %w", err)
	}
	cfg.InstallationID = installationID

	privateKey, err := os.ReadFile(DefaultSecretsPath + "/private-key")
	if err != nil {
		return nil, fmt.Errorf("read private-key: %w", err)
	}
	cfg.PrivateKey = privateKey

	// Parse workspace repositories from environment
	if reposEnv := os.Getenv("WORKSPACE_REPOS"); reposEnv != "" {
		cfg.Repositories = parseRepositoryNames(reposEnv)
	}

	// Get refresh interval from environment
	if intervalStr := os.Getenv("REFRESH_INTERVAL_MINUTES"); intervalStr != "" {
		minutes, err := strconv.Atoi(intervalStr)
		if err != nil {
			return nil, fmt.Errorf("invalid REFRESH_INTERVAL_MINUTES: %w", err)
		}
		if minutes <= 0 {
			return nil, fmt.Errorf("REFRESH_INTERVAL_MINUTES must be positive, got %d", minutes)
		}
		cfg.RefreshInterval = time.Duration(minutes) * time.Minute
	}

	return cfg, nil
}

// parseRepositoryNames extracts repository names from git URLs.
//
// Given a space-separated string of repository URLs like:
//   "https://github.com/owner/repo1.git https://github.com/owner/repo2.git"
//
// Returns repository names:
//   ["repo1", "repo2"]
//
// This is used for GitHub App token repository scoping. The GitHub API
// expects repository names (not URLs), so we extract the final path component
// and remove the .git suffix if present.
func parseRepositoryNames(reposEnv string) []string {
	if reposEnv == "" {
		return nil
	}

	urls := strings.Fields(reposEnv)
	repos := make([]string, 0, len(urls))

	for _, url := range urls {
		// Extract repo name from URL path
		// Example: "https://github.com/owner/repo.git" -> "repo"
		parts := strings.Split(url, "/")
		if len(parts) < 2 {
			continue // Skip malformed URLs
		}

		repoName := parts[len(parts)-1]
		repoName = strings.TrimSuffix(repoName, ".git")

		if repoName != "" {
			repos = append(repos, repoName)
		}
	}

	return repos
}

// readInt64File reads a file and parses its contents as an int64.
// The file is expected to contain a single integer, possibly with whitespace.
func readInt64File(path string) (int64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	value, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", path, err)
	}

	return value, nil
}
