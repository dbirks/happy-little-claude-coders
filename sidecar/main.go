package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v66/github"
)

const (
	tokenPath         = "/var/run/github/token"
	appIDPath         = "/var/run/secrets/github-app/app-id"
	installationIDPath = "/var/run/secrets/github-app/installation-id"
	privateKeyPath    = "/var/run/secrets/github-app/private-key"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("GitHub App token refresh sidecar starting...")

	// Read configuration
	appID, err := readInt64File(appIDPath)
	if err != nil {
		log.Fatalf("Failed to read app-id: %v", err)
	}

	installationID, err := readInt64File(installationIDPath)
	if err != nil {
		log.Fatalf("Failed to read installation-id: %v", err)
	}

	privateKey, err := os.ReadFile(privateKeyPath)
	if err != nil {
		log.Fatalf("Failed to read private-key: %v", err)
	}

	// Get workspace repos from environment
	reposEnv := os.Getenv("WORKSPACE_REPOS")
	if reposEnv == "" {
		log.Println("WARNING: WORKSPACE_REPOS not set, token will have access to all repos")
	}

	// Parse repository names from URLs
	repos := parseRepoNames(reposEnv)
	log.Printf("Repository scoping: %v", repos)

	// Get refresh interval
	refreshInterval := getRefreshInterval()
	log.Printf("Token refresh interval: %v", refreshInterval)

	// Create GitHub App transport
	itr, err := ghinstallation.New(http.DefaultTransport, appID, installationID, privateKey)
	if err != nil {
		log.Fatalf("Failed to create GitHub App transport: %v", err)
	}

	// Main refresh loop with retry logic
	backoff := time.Second
	maxBackoff := 5 * time.Minute

	for {
		token, err := generateToken(itr, repos)
		if err != nil {
			log.Printf("ERROR: Failed to generate token: %v", err)
			log.Printf("Retrying in %v...", backoff)
			time.Sleep(backoff)
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		// Reset backoff on success
		backoff = time.Second

		// Write token to shared volume
		if err := writeToken(token); err != nil {
			log.Printf("ERROR: Failed to write token: %v", err)
			continue
		}

		log.Printf("✓ Token refreshed successfully (expires in ~60 minutes)")

		// Sleep until next refresh
		time.Sleep(refreshInterval)
	}
}

// generateToken creates a repository-scoped installation token
func generateToken(itr *ghinstallation.AppsTransport, repos []string) (string, error) {
	ctx := context.Background()
	client := github.NewClient(&http.Client{Transport: itr})

	opts := &github.InstallationTokenOptions{}
	if len(repos) > 0 {
		opts.Repositories = repos
	}

	token, _, err := client.Apps.CreateInstallationToken(ctx, itr.InstallationID, opts)
	if err != nil {
		return "", fmt.Errorf("create installation token: %w", err)
	}

	return token.GetToken(), nil
}

// writeToken writes the token to the shared volume atomically
func writeToken(token string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(tokenPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create token directory: %w", err)
	}

	// Write to temp file first, then rename (atomic)
	tmpPath := tokenPath + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(token), 0600); err != nil {
		return fmt.Errorf("write temp token: %w", err)
	}

	if err := os.Rename(tmpPath, tokenPath); err != nil {
		return fmt.Errorf("rename token file: %w", err)
	}

	return nil
}

// parseRepoNames extracts repository names from git URLs
// Example: "https://github.com/owner/repo.git" -> "repo"
func parseRepoNames(reposEnv string) []string {
	if reposEnv == "" {
		return nil
	}

	urls := strings.Fields(reposEnv)
	repos := make([]string, 0, len(urls))

	for _, url := range urls {
		// Extract repo name from URL
		// https://github.com/owner/repo.git -> repo
		parts := strings.Split(url, "/")
		if len(parts) < 2 {
			continue
		}
		repoName := parts[len(parts)-1]
		repoName = strings.TrimSuffix(repoName, ".git")
		if repoName != "" {
			repos = append(repos, repoName)
		}
	}

	return repos
}

// readInt64File reads a file and parses it as int64
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

// getRefreshInterval returns the refresh interval from env var or default
func getRefreshInterval() time.Duration {
	intervalStr := os.Getenv("REFRESH_INTERVAL_MINUTES")
	if intervalStr == "" {
		return 45 * time.Minute
	}

	minutes, err := strconv.Atoi(intervalStr)
	if err != nil {
		log.Printf("WARNING: Invalid REFRESH_INTERVAL_MINUTES, using default 45")
		return 45 * time.Minute
	}

	return time.Duration(minutes) * time.Minute
}
