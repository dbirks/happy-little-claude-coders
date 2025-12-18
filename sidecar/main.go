// Package main implements a GitHub App installation token refresh sidecar.
//
// This sidecar runs alongside a main application container in Kubernetes,
// continuously refreshing GitHub App installation tokens and writing them
// to a shared volume. The main container can then use these tokens for
// GitHub API access.
//
// # Architecture
//
// The sidecar follows this pattern:
//  1. Load configuration from mounted secrets and environment variables
//  2. Create a GitHub App transport using ghinstallation library
//  3. Generate repository-scoped installation tokens
//  4. Write tokens atomically to shared tmpfs volume
//  5. Refresh tokens every 45 minutes (before 1-hour expiration)
//  6. Handle errors with exponential backoff retry logic
//
// # Token Scoping
//
// Tokens can be scoped to specific repositories via the WORKSPACE_REPOS
// environment variable. This provides workspace isolation when multiple
// workspaces share the same GitHub App installation.
//
// Without repository scoping, tokens have access to all repositories
// in the installation.
//
// # Security
//
// - Runs as non-root user (UID 65532 in distroless image)
// - Private keys are mounted from Kubernetes secrets
// - Tokens are written to tmpfs (memory-backed, never persisted to disk)
// - Token files have 0600 permissions (owner read/write only)
// - Atomic writes prevent partial token exposure
//
// # Configuration
//
// Environment variables:
//   - WORKSPACE_REPOS: Space-separated repository URLs to scope tokens to
//   - REFRESH_INTERVAL_MINUTES: Token refresh interval in minutes (default: 45)
//
// Mounted secrets (from Kubernetes secret):
//   - /var/run/secrets/github-app/app-id: GitHub App ID
//   - /var/run/secrets/github-app/installation-id: Installation ID
//   - /var/run/secrets/github-app/private-key: Private key in PEM format
//
// # Graceful Shutdown
//
// The sidecar handles SIGTERM and SIGINT signals for graceful shutdown,
// completing any in-flight token generation before exiting.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Configure logging with timestamps and source file locations
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("GitHub App token refresh sidecar starting...")

	// Load configuration from environment and mounted secrets
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Log configuration (excluding sensitive data)
	log.Printf("App ID: %d", cfg.AppID)
	log.Printf("Installation ID: %d", cfg.InstallationID)
	log.Printf("Token path: %s", cfg.TokenPath)
	log.Printf("Refresh interval: %v", cfg.RefreshInterval)

	if len(cfg.Repositories) > 0 {
		log.Printf("Repository scoping enabled: %v", cfg.Repositories)
	} else {
		log.Println("WARNING: No repository scoping configured - token will have access to all repos in installation")
	}

	// Create token generator
	generator, err := NewTokenGenerator(cfg)
	if err != nil {
		log.Fatalf("Failed to create token generator: %v", err)
	}

	// Create token writer
	writer := NewTokenWriter(cfg.TokenPath)

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, shutting down gracefully...", sig)
		cancel()
	}()

	// Run the token refresh loop
	if err := run(ctx, generator, writer, cfg.RefreshInterval); err != nil {
		log.Fatalf("Token refresh loop failed: %v", err)
	}

	log.Println("Sidecar shut down successfully")
}

// run executes the main token refresh loop with exponential backoff retry logic.
//
// The loop runs until the context is cancelled (via SIGTERM/SIGINT) or a fatal
// error occurs. On each iteration:
//  1. Generate a new installation token
//  2. Write it atomically to the shared volume
//  3. Sleep until the next refresh interval
//
// If token generation or writing fails, the loop retries with exponential
// backoff up to a maximum of 5 minutes between attempts.
func run(ctx context.Context, generator *TokenGenerator, writer *TokenWriter, refreshInterval time.Duration) error {
	backoff := NewBackoff()

	for {
		select {
		case <-ctx.Done():
			// Context cancelled, exit gracefully
			return nil
		default:
		}

		// Generate token with context
		token, err := generator.Generate(ctx)
		if err != nil {
			log.Printf("ERROR: Failed to generate token: %v", err)
			log.Printf("Retrying in %v...", backoff.Duration())

			// Sleep with context awareness for cancellation
			select {
			case <-time.After(backoff.Duration()):
				backoff.Increase()
				continue
			case <-ctx.Done():
				return nil
			}
		}

		// Token generation succeeded, reset backoff
		backoff.Reset()

		// Write token to shared volume
		if err := writer.Write(token); err != nil {
			log.Printf("ERROR: Failed to write token: %v", err)
			// Don't retry immediately, wait for next refresh cycle
			continue
		}

		log.Printf("✓ Token refreshed successfully (expires in ~60 minutes)")

		// Sleep until next refresh, with context awareness
		select {
		case <-time.After(refreshInterval):
			// Time to refresh
		case <-ctx.Done():
			// Shutdown requested
			return nil
		}
	}
}
