package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// File permissions for token files.
const (
	// TokenFileMode restricts token file access to owner read/write only (0600).
	// This prevents other users on the system from reading the token.
	TokenFileMode = 0600

	// TokenDirMode sets directory permissions to owner rwx (0755).
	TokenDirMode = 0755
)

// TokenWriter handles atomic writes of GitHub App tokens to the filesystem.
//
// Tokens are written atomically using the write-to-temp-then-rename pattern,
// which ensures that readers never see a partially-written token file.
//
// The token file is written with restrictive permissions (0600) to prevent
// unauthorized access.
type TokenWriter struct {
	path string
}

// NewTokenWriter creates a TokenWriter that writes to the specified path.
func NewTokenWriter(path string) *TokenWriter {
	return &TokenWriter{path: path}
}

// Write atomically writes the token to the configured path.
//
// The write is atomic: readers will either see the old token or the new token,
// never a partial write. This is achieved by:
//  1. Writing to a temporary file in the same directory
//  2. Setting restrictive permissions (0600) on the temp file
//  3. Atomically renaming the temp file to the target path
//
// On Unix systems, rename() is atomic when source and destination are on
// the same filesystem, which is guaranteed by creating the temp file in
// the same directory.
//
// Returns an error if directory creation, file writing, or renaming fails.
func (tw *TokenWriter) Write(token string) error {
	// Ensure the parent directory exists
	dir := filepath.Dir(tw.path)
	if err := os.MkdirAll(dir, TokenDirMode); err != nil {
		return fmt.Errorf("create token directory: %w", err)
	}

	// Write to temporary file in the same directory.
	// This ensures the temp file is on the same filesystem as the target,
	// which guarantees atomic rename.
	tmpPath := tw.path + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(token), TokenFileMode); err != nil {
		return fmt.Errorf("write temp token file: %w", err)
	}

	// Atomically replace the target file with the temp file.
	// On Unix, rename() is atomic when source and dest are on the same filesystem.
	if err := os.Rename(tmpPath, tw.path); err != nil {
		// Clean up temp file on failure
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename token file: %w", err)
	}

	return nil
}

// Path returns the path where tokens are written.
func (tw *TokenWriter) Path() string {
	return tw.path
}
