package modules

import (
	"fmt"
	"strings"
)

// RunTestReadiness evaluates whether the project has adequate test files.
func RunTestReadiness(input ModuleInput) ModuleResult {
	result := ModuleResult{Name: "Test Readiness"}

	sourceCount := 0
	testCount := 0
	emptyTests := 0

	for _, file := range input.Files {
		base := fileBaseName(file.Path)

		if isTestFile(base, file.Path) {
			testCount++
			if file.Size < 50 { // ~50 bytes is basically an empty test shell
				emptyTests++
			}
		} else if isSourceFile(base) {
			sourceCount++
		}
	}

	// No tests at all
	if testCount == 0 && sourceCount > 0 {
		result.Findings = append(result.Findings, Finding{
			Rule:     "no-tests",
			Severity: SeverityFail,
			Message:  fmt.Sprintf("No test files found (%d source files detected)", sourceCount),
			Detail:   "Add test files to verify functionality",
		})
		return result
	}

	// Test-to-source ratio
	if sourceCount > 0 && testCount > 0 {
		ratio := float64(testCount) / float64(sourceCount)
		if ratio < 0.2 {
			result.Findings = append(result.Findings, Finding{
				Rule:     "low-test-ratio",
				Severity: SeverityWarn,
				Message:  fmt.Sprintf("Low test-to-source ratio: %d tests / %d sources (%.0f%%)", testCount, sourceCount, ratio*100),
				Detail:   "Aim for at least 1 test file per 3 source files",
			})
		} else {
			result.Findings = append(result.Findings, Finding{
				Rule:     "test-ratio",
				Severity: SeverityInfo,
				Message:  fmt.Sprintf("Test ratio: %d tests / %d sources (%.0f%%)", testCount, sourceCount, ratio*100),
			})
		}
	}

	// Empty test files
	if emptyTests > 0 {
		result.Findings = append(result.Findings, Finding{
			Rule:     "empty-tests",
			Severity: SeverityWarn,
			Message:  fmt.Sprintf("%d test files appear to be empty or stubs", emptyTests),
			Detail:   "Add meaningful test assertions",
		})
	}

	return result
}

// isTestFile returns true for common test file patterns across ecosystems.
func isTestFile(baseName, path string) bool {
	// Go: *_test.go
	if strings.HasSuffix(baseName, "_test.go") {
		return true
	}
	// JS/TS: *.test.js, *.spec.js, *.test.ts, *.spec.ts
	for _, suffix := range []string{".test.js", ".spec.js", ".test.ts", ".spec.ts", ".test.mjs", ".spec.mjs"} {
		if strings.HasSuffix(baseName, suffix) {
			return true
		}
	}
	// Python: test_*.py, *_test.py
	if strings.HasSuffix(baseName, ".py") {
		if strings.HasPrefix(baseName, "test_") || strings.HasSuffix(baseName, "_test.py") {
			return true
		}
	}
	// Java/Kotlin: *Test.java, *Test.kt
	for _, suffix := range []string{"Test.java", "Test.kt", "Tests.java"} {
		if strings.HasSuffix(baseName, suffix) {
			return true
		}
	}
	// Rust: files inside tests/ or with #[test]
	if strings.Contains(path, "tests/") && strings.HasSuffix(baseName, ".rs") {
		return true
	}
	return false
}

// isSourceFile returns true for common source code extensions.
func isSourceFile(baseName string) bool {
	sourceExts := []string{
		".go", ".py", ".js", ".ts", ".jsx", ".tsx",
		".java", ".kt", ".rs", ".c", ".cpp", ".h",
		".cs", ".rb", ".swift", ".php",
	}
	for _, ext := range sourceExts {
		if strings.HasSuffix(baseName, ext) {
			return true
		}
	}
	return false
}
