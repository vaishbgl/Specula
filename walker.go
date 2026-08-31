package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// WalkOptions controls the file walker behavior.
type WalkOptions struct {
	MaxSourceSize   int64    // max file size for source files (bytes), default 512KB
	MaxManifestSize int64    // max file size for manifest/lockfiles (bytes), default 10MB
	ExcludeDirs     []string // additional dirs to skip (on top of builtins)
}

// DefaultWalkOptions returns sensible defaults.
func DefaultWalkOptions() WalkOptions {
	return WalkOptions{
		MaxSourceSize:   512 * 1024,       // 512 KB
		MaxManifestSize: 10 * 1024 * 1024, // 10 MB
	}
}

// ScannedFile represents a single file discovered by the walker.
type ScannedFile struct {
	Path    string // relative path from the walk root
	AbsPath string // absolute path on disk
	Size    int64
	IsDir   bool
}

// ScanResult holds the output of a full walk.
type ScanResult struct {
	Root     string
	Files    []ScannedFile
	Warnings []string // permission errors, skipped files, etc.
	Stats    ScanStats
}

// ScanStats tracks scanning metrics.
type ScanStats struct {
	FilesSeen    int
	FilesScanned int
	FilesSkipped int
	DirsSkipped  int
	BytesTotal   int64
	Duration     time.Duration
}

// builtinIgnoreDirs are always skipped.
var builtinIgnoreDirs = map[string]bool{
	".git":          true,
	"node_modules":  true,
	"vendor":        true,
	"dist":          true,
	"build":         true,
	".next":         true,
	".cache":        true,
	"coverage":      true,
	"__pycache__":   true,
	"target":        true,
	".venv":         true,
	".tox":          true,
	".mypy_cache":   true,
	".pytest_cache": true,
}

// manifestNames are files that get the larger size limit.
var manifestNames = map[string]bool{
	"package-lock.json": true,
	"yarn.lock":         true,
	"pnpm-lock.yaml":    true,
	"go.sum":            true,
	"Cargo.lock":        true,
	"Pipfile.lock":      true,
	"poetry.lock":       true,
}

// Walk recursively walks root and returns all scannable files.
// It skips built-in ignore dirs, binary files, oversized files,
// and symlinked directories. Results are sorted for determinism.
func Walk(root string, opts WalkOptions) (ScanResult, error) {
	start := time.Now()

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return ScanResult{}, fmt.Errorf("walk: resolve path %q: %w", root, err)
	}

	info, err := os.Stat(absRoot)
	if err != nil {
		return ScanResult{}, fmt.Errorf("walk: stat %q: %w", root, err)
	}
	if !info.IsDir() {
		return ScanResult{}, fmt.Errorf("walk: %q is not a directory", root)
	}

	// Build the full ignore set
	ignoreDirs := make(map[string]bool, len(builtinIgnoreDirs)+len(opts.ExcludeDirs))
	for k, v := range builtinIgnoreDirs {
		ignoreDirs[k] = v
	}
	for _, d := range opts.ExcludeDirs {
		ignoreDirs[d] = true
	}

	result := ScanResult{Root: absRoot}

	walkErr := filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Permission error — skip, don't crash
			if os.IsPermission(err) {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("permission denied: %s", path))
				if d != nil && d.IsDir() {
					result.Stats.DirsSkipped++
					return filepath.SkipDir
				}
				result.Stats.FilesSkipped++
				return nil
			}
			return err
		}

		name := d.Name()

		// --- Directory handling ---
		if d.IsDir() {
			// Skip the root itself (we still walk into it)
			if path == absRoot {
				return nil
			}

			// Skip ignored directories
			if ignoreDirs[name] {
				result.Stats.DirsSkipped++
				return filepath.SkipDir
			}

			// Skip hidden directories except .github: workflow files are source
			// configuration and must be visible to the Project Setup module.
			if strings.HasPrefix(name, ".") && name != ".github" {
				result.Stats.DirsSkipped++
				return filepath.SkipDir
			}

			// Don't follow symlinked directories
			if d.Type()&fs.ModeSymlink != 0 {
				result.Stats.DirsSkipped++
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("skipped symlinked directory: %s", path))
				return filepath.SkipDir
			}

			return nil
		}

		// --- File handling ---
		result.Stats.FilesSeen++

		// Skip non-regular files (symlinks, devices, pipes)
		if !d.Type().IsRegular() {
			result.Stats.FilesSkipped++
			return nil
		}

		// Get file info for size
		fi, err := d.Info()
		if err != nil {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("cannot stat %s: %v", path, err))
			result.Stats.FilesSkipped++
			return nil
		}

		size := fi.Size()

		// Size limits
		maxSize := opts.MaxSourceSize
		if manifestNames[name] {
			maxSize = opts.MaxManifestSize
		}
		if size > maxSize {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("skipped oversized file (%d bytes): %s", size, path))
			result.Stats.FilesSkipped++
			return nil
		}

		// Binary detection: check first 512 bytes for null bytes
		if size > 0 {
			isBin, binErr := isBinaryFile(path)
			if binErr != nil {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("cannot read %s: %v", path, binErr))
				result.Stats.FilesSkipped++
				return nil
			}
			if isBin {
				result.Stats.FilesSkipped++
				return nil
			}
		}

		// Compute relative path
		relPath, err := filepath.Rel(absRoot, path)
		if err != nil {
			relPath = path // fallback
		}

		result.Files = append(result.Files, ScannedFile{
			Path:    relPath,
			AbsPath: path,
			Size:    size,
		})
		result.Stats.FilesScanned++
		result.Stats.BytesTotal += size

		return nil
	})

	if walkErr != nil {
		return result, fmt.Errorf("walk: %w", walkErr)
	}

	// Sort files for deterministic output
	sort.Slice(result.Files, func(i, j int) bool {
		return result.Files[i].Path < result.Files[j].Path
	})

	result.Stats.Duration = time.Since(start)
	return result, nil
}

// isBinaryFile checks if a file is binary by reading the first 512 bytes
// and looking for null bytes (0x00).
func isBinaryFile(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && n == 0 {
		return false, err
	}

	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true, nil
		}
	}
	return false, nil
}
