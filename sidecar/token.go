package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v75/github"
)

// Backoff configuration for token generation retries.
const (
	// InitialBackoff is the starting backoff duration for retries.
	InitialBackoff = 1 * time.Second

	// MaxBackoff is the maximum backoff duration between retries.
	// We cap at 5 minutes to avoid excessive delays.
	MaxBackoff = 5 * time.Minute

	// BackoffMultiplier is the factor by which backoff increases on each retry.
	BackoffMultiplier = 2
)

// TokenGenerator generates GitHub App installation access tokens.
//
// It uses the ghinstallation library to handle JWT generation and signing,
// which is the de facto standard approach in Go for GitHub App authentication.
//
// Tokens are scoped to specific repositories if configured, limiting access
// to only the workspace's configured repos. This provides workspace isolation
// when multiple workspaces share the same GitHub App installation.
//
// Note: We use AppsTransport (not Transport) because CreateInstallationToken
// is an app-level endpoint that requires JWT authentication, not installation
// token authentication.
type TokenGenerator struct {
	appID          int64
	installationID int64
	transport      *ghinstallation.AppsTransport
	repos          []string
}

// NewTokenGenerator creates a TokenGenerator from the given configuration.
//
// The transport is created using ghinstallation.NewAppsTransport() which
// provides App-level JWT authentication. This is required because the
// CreateInstallationToken endpoint is an app-level API that expects JWT
// authentication, not installation token authentication.
//
// Returns an error if the transport cannot be created.
func NewTokenGenerator(cfg *Config) (*TokenGenerator, error) {
	// Create GitHub App transport for authentication.
	// We use NewAppsTransport (not New) because CreateInstallationToken
	// requires app-level JWT auth, not installation token auth.
	transport, err := ghinstallation.NewAppsTransport(
		http.DefaultTransport,
		cfg.AppID,
		cfg.PrivateKey,
	)
	if err != nil {
		return nil, fmt.Errorf("create github app transport: %w", err)
	}

	return &TokenGenerator{
		appID:          cfg.AppID,
		installationID: cfg.InstallationID,
		transport:      transport,
		repos:          cfg.Repositories,
	}, nil
}

// Generate creates a new installation access token.
//
// The token is scoped to the repositories specified in the TokenGenerator.
// If no repositories are specified, the token has access to all repositories
// in the installation.
//
// Installation tokens expire after 1 hour and cannot be refreshed - a new
// token must be generated before expiration.
//
// Returns the token string or an error if generation fails.
func (tg *TokenGenerator) Generate(ctx context.Context) (string, error) {
	client := github.NewClient(&http.Client{Transport: tg.transport})

	opts := &github.InstallationTokenOptions{}
	if len(tg.repos) > 0 {
		// Scope token to specific repositories for workspace isolation
		opts.Repositories = tg.repos
	}

	// Create installation token via GitHub API
	token, _, err := client.Apps.CreateInstallationToken(
		ctx,
		tg.installationID,
		opts,
	)
	if err != nil {
		return "", fmt.Errorf("create installation token: %w", err)
	}

	return token.GetToken(), nil
}

// Backoff implements exponential backoff with a maximum ceiling.
//
// This is used when token generation fails, to avoid hammering the GitHub API
// with rapid retry attempts. The backoff duration doubles on each failure
// until it reaches MaxBackoff.
type Backoff struct {
	current time.Duration
}

// NewBackoff creates a Backoff starting at InitialBackoff duration.
func NewBackoff() *Backoff {
	return &Backoff{current: InitialBackoff}
}

// Duration returns the current backoff duration.
func (b *Backoff) Duration() time.Duration {
	return b.current
}

// Increase doubles the backoff duration, up to MaxBackoff.
func (b *Backoff) Increase() {
	b.current *= BackoffMultiplier
	if b.current > MaxBackoff {
		b.current = MaxBackoff
	}
}

// Reset resets the backoff to InitialBackoff.
// This should be called after a successful operation.
func (b *Backoff) Reset() {
	b.current = InitialBackoff
}
