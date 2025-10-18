// Copyright (c) 2025 by Developer Network.
//
// For more details, see the LICENSE file in the root directory of this
// source code repository or contact Developer Network at info@devnw.com.

// CANARY: REQ=CBIN-146; FEATURE="TestInfrastructure"; ASPECT=Storage; STATUS=IMPL; UPDATED=2025-10-18
package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// TempDir creates a temporary directory for testing
// Returns the path and a cleanup function
func TempDir(t *testing.T) (string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "canary-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

// TempHomeDir creates a temporary HOME directory for testing
// Returns the original HOME value and a cleanup function
func TempHomeDir(t *testing.T) (string, func()) {
	t.Helper()

	originalHome := os.Getenv("HOME")
	tmpHome, err := os.MkdirTemp("", "canary-home-*")
	if err != nil {
		t.Fatalf("failed to create temp home dir: %v", err)
	}

	os.Setenv("HOME", tmpHome)

	cleanup := func() {
		os.RemoveAll(tmpHome)
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		} else {
			os.Unsetenv("HOME")
		}
	}

	return tmpHome, cleanup
}

// SetupProjectDir creates a project directory with .canary subdirectory
func SetupProjectDir(t *testing.T, dir string) error {
	t.Helper()

	canaryDir := filepath.Join(dir, ".canary")
	if err := os.MkdirAll(canaryDir, 0755); err != nil {
		return err
	}

	return nil
}

// Chdir changes to a directory and returns a cleanup function to restore
func Chdir(t *testing.T, dir string) func() {
	t.Helper()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change directory to %s: %v", dir, err)
	}

	return func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("failed to restore directory: %v", err)
		}
	}
}

// FileExists checks if a file exists at the given path
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// TB is an interface that covers both *testing.T and *testing.B
type TB interface {
	Helper()
	Fatalf(format string, args ...interface{})
}

// TempHomeDirB creates a temporary HOME directory for benchmarking
// Returns the original HOME value and a cleanup function
func TempHomeDirB(b *testing.B) (string, func()) {
	b.Helper()

	originalHome := os.Getenv("HOME")
	tmpHome, err := os.MkdirTemp("", "canary-home-*")
	if err != nil {
		b.Fatalf("failed to create temp home dir: %v", err)
	}

	os.Setenv("HOME", tmpHome)

	cleanup := func() {
		os.RemoveAll(tmpHome)
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		} else {
			os.Unsetenv("HOME")
		}
	}

	return tmpHome, cleanup
}

// TempDirB creates a temporary directory for benchmarking
// Returns the path and a cleanup function
func TempDirB(b *testing.B) (string, func()) {
	b.Helper()

	tmpDir, err := os.MkdirTemp("", "canary-test-*")
	if err != nil {
		b.Fatalf("failed to create temp dir: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}
