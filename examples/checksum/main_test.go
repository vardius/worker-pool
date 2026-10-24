package main

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestChecksums(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{"empty": "", "hello": "hello", "name": "worker-pool\n"}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	paths := []string{
		filepath.Join(dir, "hello"),
		filepath.Join(dir, "missing"),
		filepath.Join(dir, "empty"),
		filepath.Join(dir, "name"),
	}

	sums, err := checksums(t.Context(), paths, 2)
	if err != nil {
		t.Fatalf("checksums() error = %v", err)
	}

	// From shasum -a 256.
	want := []string{
		"2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		"",
		"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		"3d15482619565c6e8c3a27b9429107bf5108684a665b6cf0c3ce000b4b774051",
	}
	for i, s := range sums {
		if s.hex != want[i] {
			t.Errorf("sum of %s = %q, want %q", paths[i], s.hex, want[i])
		}
	}
	if !errors.Is(sums[1].err, fs.ErrNotExist) {
		t.Errorf("error for %s = %v, want %v", paths[1], sums[1].err, fs.ErrNotExist)
	}
}

func TestChecksumsStopsQueueingWhenCtxIsDone(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	if _, err := checksums(ctx, []string{"go.mod"}, 1); !errors.Is(err, context.Canceled) {
		t.Errorf("checksums() error = %v, want %v", err, context.Canceled)
	}
}
