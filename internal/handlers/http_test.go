package handlers

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveIndexPathPrefersWorkingDirectory(t *testing.T) {
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() {
		_ = os.Chdir(oldWD)
	}()

	tmp := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmp, "web"), 0o755); err != nil {
		t.Fatalf("mkdir web: %v", err)
	}

	expected := filepath.Join(tmp, "web", "index.html")
	if err := os.WriteFile(expected, []byte("<html>ok</html>"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	got, err := resolveIndexPath()
	if err != nil {
		t.Fatalf("resolveIndexPath returned error: %v", err)
	}

	if filepath.Clean(got) != filepath.Clean(expected) {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}
