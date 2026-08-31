package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWalk_BasicFiles(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", "package main")
	writeFile(t, dir, "util.go", "package main")
	writeFile(t, dir, "README.md", "# Hello")

	result, err := Walk(dir, DefaultWalkOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Stats.FilesScanned != 3 {
		t.Errorf("got %d files scanned, want 3", result.Stats.FilesScanned)
	}
}

func TestWalk_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	result, err := Walk(dir, DefaultWalkOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Stats.FilesScanned != 0 {
		t.Errorf("got %d files scanned, want 0", result.Stats.FilesScanned)
	}
	if result.Stats.FilesSeen != 0 {
		t.Errorf("got %d files seen, want 0", result.Stats.FilesSeen)
	}
}

func TestWalk_SkipsIgnoredDirs(t *testing.T) {
	dir := t.TempDir()

	// Create files in ignored dirs
	for _, ignored := range []string{"node_modules", ".git", "vendor", "__pycache__"} {
		subdir := filepath.Join(dir, ignored)
		os.MkdirAll(subdir, 0755)
		writeFile(t, subdir, "should_be_ignored.js", "var x = 1;")
	}

	// Create a real file
	writeFile(t, dir, "app.go", "package main")

	result, err := Walk(dir, DefaultWalkOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Stats.FilesScanned != 1 {
		t.Errorf("got %d files scanned, want 1 (only app.go)", result.Stats.FilesScanned)
	}
	if result.Stats.DirsSkipped < 4 {
		t.Errorf("got %d dirs skipped, want >= 4", result.Stats.DirsSkipped)
	}
}

func TestWalk_SkipsHiddenDirs(t *testing.T) {
	dir := t.TempDir()

	hiddenDir := filepath.Join(dir, ".hidden")
	os.MkdirAll(hiddenDir, 0755)
	writeFile(t, hiddenDir, "secret.txt", "hidden content")
	writeFile(t, dir, "visible.txt", "visible content")

	result, err := Walk(dir, DefaultWalkOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Stats.FilesScanned != 1 {
		t.Errorf("got %d files scanned, want 1", result.Stats.FilesScanned)
	}
}

func TestWalk_SkipsBinaryFiles(t *testing.T) {
	dir := t.TempDir()

	// Write a text file
	writeFile(t, dir, "main.go", "package main\nfunc main() {}\n")

	// Write a binary file (contains null bytes)
	binPath := filepath.Join(dir, "image.png")
	binData := []byte{0x89, 0x50, 0x4E, 0x47, 0x00, 0x00, 0x01, 0x02}
	if err := os.WriteFile(binPath, binData, 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Walk(dir, DefaultWalkOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Stats.FilesScanned != 1 {
		t.Errorf("got %d files scanned, want 1 (only main.go)", result.Stats.FilesScanned)
	}
	if result.Stats.FilesSkipped != 1 {
		t.Errorf("got %d files skipped, want 1 (binary)", result.Stats.FilesSkipped)
	}
}

func TestWalk_SkipsOversizedFiles(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "small.go", "package main")

	// Write a file larger than the 512KB limit
	bigPath := filepath.Join(dir, "huge.go")
	bigData := make([]byte, 600*1024) // 600KB of 'x'
	for i := range bigData {
		bigData[i] = 'x'
	}
	if err := os.WriteFile(bigPath, bigData, 0644); err != nil {
		t.Fatal(err)
	}

	opts := DefaultWalkOptions()
	result, err := Walk(dir, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Stats.FilesScanned != 1 {
		t.Errorf("got %d files scanned, want 1", result.Stats.FilesScanned)
	}
	if result.Stats.FilesSkipped != 1 {
		t.Errorf("got %d files skipped, want 1 (oversized)", result.Stats.FilesSkipped)
	}
}

func TestWalk_ManifestGetsLargerLimit(t *testing.T) {
	dir := t.TempDir()

	// A 1MB package-lock.json should NOT be skipped (limit is 10MB)
	lockData := make([]byte, 1*1024*1024)
	for i := range lockData {
		lockData[i] = '{'
	}
	lockPath := filepath.Join(dir, "package-lock.json")
	if err := os.WriteFile(lockPath, lockData, 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Walk(dir, DefaultWalkOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Stats.FilesScanned != 1 {
		t.Errorf("got %d files scanned, want 1 (package-lock.json)", result.Stats.FilesScanned)
	}
}

func TestWalk_DeterministicOrder(t *testing.T) {
	dir := t.TempDir()

	// Create files in reverse alphabetical order
	for _, name := range []string{"z.go", "m.go", "a.go"} {
		writeFile(t, dir, name, "package main")
	}

	result, err := Walk(dir, DefaultWalkOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Files) != 3 {
		t.Fatalf("got %d files, want 3", len(result.Files))
	}
	if result.Files[0].Path != "a.go" {
		t.Errorf("first file is %q, want 'a.go'", result.Files[0].Path)
	}
	if result.Files[1].Path != "m.go" {
		t.Errorf("second file is %q, want 'm.go'", result.Files[1].Path)
	}
	if result.Files[2].Path != "z.go" {
		t.Errorf("third file is %q, want 'z.go'", result.Files[2].Path)
	}
}

func TestWalk_NonexistentDir(t *testing.T) {
	_, err := Walk("/nonexistent/path/that/does/not/exist", DefaultWalkOptions())
	if err == nil {
		t.Fatal("expected error for nonexistent directory, got nil")
	}
}

func TestWalk_FileNotDir(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "notadir.txt")
	writeFile(t, dir, "notadir.txt", "hello")

	_, err := Walk(filePath, DefaultWalkOptions())
	if err == nil {
		t.Fatal("expected error when walking a file (not dir), got nil")
	}
}

func TestWalk_NestedSubdirectories(t *testing.T) {
	dir := t.TempDir()

	// Create nested structure: src/pkg/util/helper.go
	nested := filepath.Join(dir, "src", "pkg", "util")
	os.MkdirAll(nested, 0755)
	writeFile(t, nested, "helper.go", "package util")
	writeFile(t, dir, "main.go", "package main")

	result, err := Walk(dir, DefaultWalkOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Stats.FilesScanned != 2 {
		t.Errorf("got %d files scanned, want 2", result.Stats.FilesScanned)
	}
}

func TestWalk_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "empty.go", "")

	result, err := Walk(dir, DefaultWalkOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Empty files should be scanned (0 bytes, no null bytes)
	if result.Stats.FilesScanned != 1 {
		t.Errorf("got %d files scanned, want 1", result.Stats.FilesScanned)
	}
}

func TestWalk_CustomExcludeDirs(t *testing.T) {
	dir := t.TempDir()

	customDir := filepath.Join(dir, "myoutput")
	os.MkdirAll(customDir, 0755)
	writeFile(t, customDir, "generated.go", "package gen")
	writeFile(t, dir, "main.go", "package main")

	opts := DefaultWalkOptions()
	opts.ExcludeDirs = []string{"myoutput"}

	result, err := Walk(dir, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Stats.FilesScanned != 1 {
		t.Errorf("got %d files scanned, want 1", result.Stats.FilesScanned)
	}
}

func TestWalk_StatsArePopulated(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", "package main\n")

	result, err := Walk(dir, DefaultWalkOptions())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Stats.Duration == 0 {
		t.Error("expected non-zero duration")
	}
	if result.Stats.BytesTotal == 0 {
		t.Error("expected non-zero bytes total")
	}
	if result.Root == "" {
		t.Error("expected non-empty root")
	}
}

// --- helper ---

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
