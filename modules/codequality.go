package modules

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// RunCodeQuality scans for code quality signals: TODO markers, long files,
// duplicate files, and AI context rot.
func RunCodeQuality(input ModuleInput) ModuleResult {
	result := ModuleResult{Name: "Code & Context Lint"}

	todoCount := 0
	todoFiles := 0
	longFiles := 0
	hashes := make(map[string][]string) // hash -> list of file paths

	for _, file := range input.Files {
		base := fileBaseName(file.Path)

		// Check for AI context rot files
		if base == ".cursorrules" || base == "CLAUDE.md" || base == ".claude" {
			findings := checkContextRot(file, input.RootDir)
			result.Findings = append(result.Findings, findings...)
		}

		// Only scan source files for TODOs and line counts
		if !isSourceFile(base) {
			continue
		}

		// Scan for TODOs and line count
		todos, lineCount := scanCodeFile(file)
		if todos > 0 {
			todoCount += todos
			todoFiles++
		}

		if lineCount > 500 {
			longFiles++
			result.Findings = append(result.Findings, Finding{
				Rule:     "long-file",
				Severity: SeverityInfo,
				Message:  fmt.Sprintf("File has %d lines (threshold: 500)", lineCount),
				File:     file.Path,
				Detail:   "Consider splitting into smaller modules",
			})
		}

		// Hash file content for duplicate detection
		hash, err := hashFileContent(file.AbsPath)
		if err == nil && file.Size > 100 { // only hash non-trivial files
			hashes[hash] = append(hashes[hash], file.Path)
		}
	}

	// Aggregate TODO finding
	if todoCount > 0 {
		severity := SeverityInfo
		if todoCount > 20 {
			severity = SeverityWarn
		}
		result.Findings = append(result.Findings, Finding{
			Rule:     "todo-markers",
			Severity: severity,
			Message:  fmt.Sprintf("%d TODO/FIXME/HACK/XXX markers across %d files", todoCount, todoFiles),
			Detail:   "Resolve or track TODOs before shipping",
		})
	}

	// Duplicate file findings
	for _, paths := range hashes {
		if len(paths) > 1 {
			result.Findings = append(result.Findings, Finding{
				Rule:     "duplicate-files",
				Severity: SeverityWarn,
				Message:  fmt.Sprintf("%d files have identical content", len(paths)),
				Detail:   fmt.Sprintf("Files: %s", strings.Join(paths, ", ")),
			})
		}
	}

	return result
}

// todoPattern matches TODO, FIXME, HACK, XXX comments.
var todoPattern = regexp.MustCompile(`(?i)\b(TODO|FIXME|HACK|XXX)\b`)

// scanCodeFile scans a source file for TODO markers and counts lines.
func scanCodeFile(file FileEntry) (todos int, lineCount int) {
	f, err := os.Open(file.AbsPath)
	if err != nil {
		return 0, 0
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lineCount++
		if todoPattern.MatchString(scanner.Text()) {
			todos++
		}
	}
	return
}

// hashFileContent returns the SHA-256 hex digest of a file's content.
func hashFileContent(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// contextRotPathPattern matches relative file paths in AI context files.
var contextRotPathPattern = regexp.MustCompile(`(?:^|\s)([./][^\s:]+\.[a-zA-Z]{1,5})`)

// checkContextRot scans AI context files (.cursorrules, CLAUDE.md) for
// broken file references.
func checkContextRot(file FileEntry, rootDir string) []Finding {
	var findings []Finding

	f, err := os.Open(file.AbsPath)
	if err != nil {
		return findings
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNum := 0
	brokenPaths := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		matches := contextRotPathPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}
			refPath := match[1]
			// Resolve relative to project root
			absRef := filepath.Join(rootDir, refPath)
			if _, err := os.Stat(absRef); os.IsNotExist(err) {
				brokenPaths++
			}
		}
	}

	if brokenPaths > 0 {
		findings = append(findings, Finding{
			Rule:     "ai-context-rot",
			Severity: SeverityInfo,
			Message:  fmt.Sprintf("AI context file %q has %d broken file references", file.Path, brokenPaths),
			File:     file.Path,
			Detail:   "Stale paths cause AI agents to hallucinate. Update or remove them.",
		})
	}

	return findings
}
