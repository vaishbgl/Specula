package modules

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// RunDepsAudit inspects dependency manifests for bloat, unpinned versions,
// and zero-dep replacement opportunities.
func RunDepsAudit(input ModuleInput) ModuleResult {
	result := ModuleResult{Name: "Dependency Audit"}

	for _, file := range input.Files {
		base := fileBaseName(file.Path)
		switch base {
		case "go.mod":
			findings := auditGoMod(file)
			result.Findings = append(result.Findings, findings...)
		case "package.json":
			findings := auditPackageJSON(file)
			result.Findings = append(result.Findings, findings...)
		case "requirements.txt":
			findings := auditRequirementsTxt(file)
			result.Findings = append(result.Findings, findings...)
		}
	}

	return result
}

// auditGoMod parses a go.mod file for dependency information.
func auditGoMod(file FileEntry) []Finding {
	var findings []Finding

	f, err := os.Open(file.AbsPath)
	if err != nil {
		return findings
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inRequire := false
	depCount := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "require (" {
			inRequire = true
			continue
		}
		if line == ")" && inRequire {
			inRequire = false
			continue
		}
		if inRequire && line != "" && !strings.HasPrefix(line, "//") {
			depCount++
		}
		// Single-line require
		if strings.HasPrefix(line, "require ") && !strings.Contains(line, "(") {
			depCount++
		}
	}

	if depCount > 0 {
		severity := SeverityInfo
		if depCount > 10 {
			severity = SeverityWarn
		}
		if depCount > 30 {
			severity = SeverityFail
		}

		findings = append(findings, Finding{
			Rule:     "dependency-count",
			Severity: severity,
			Message:  fmt.Sprintf("%d direct dependencies in go.mod", depCount),
			File:     file.Path,
			Detail:   "Consider whether all dependencies are necessary",
		})
	}

	return findings
}

// auditPackageJSON parses a package.json for dependency stats.
func auditPackageJSON(file FileEntry) []Finding {
	var findings []Finding

	data, err := os.ReadFile(file.AbsPath)
	if err != nil {
		return findings
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		findings = append(findings, Finding{
			Rule:     "malformed-manifest",
			Severity: SeverityWarn,
			Message:  "Could not parse package.json",
			File:     file.Path,
			Detail:   err.Error(),
		})
		return findings
	}

	depCount := len(pkg.Dependencies)
	if depCount > 0 {
		severity := SeverityInfo
		if depCount > 15 {
			severity = SeverityWarn
		}
		if depCount > 50 {
			severity = SeverityFail
		}
		findings = append(findings, Finding{
			Rule:     "dependency-count",
			Severity: severity,
			Message:  fmt.Sprintf("%d runtime dependencies in package.json", depCount),
			File:     file.Path,
		})
	}

	// Check for unpinned versions
	wildcardCount := 0
	for name, version := range pkg.Dependencies {
		if version == "*" || version == "latest" || strings.HasPrefix(version, ">=") {
			wildcardCount++
			if wildcardCount <= 3 { // cap findings to avoid noise
				findings = append(findings, Finding{
					Rule:     "unpinned-version",
					Severity: SeverityWarn,
					Message:  fmt.Sprintf("Unpinned dependency: %s@%s", name, version),
					File:     file.Path,
					Detail:   "Pin to a specific version for reproducible builds",
				})
			}
		}
	}
	if wildcardCount > 3 {
		findings = append(findings, Finding{
			Rule:     "unpinned-version",
			Severity: SeverityWarn,
			Message:  fmt.Sprintf("... and %d more unpinned dependencies", wildcardCount-3),
			File:     file.Path,
		})
	}

	// Zero-dep opportunities
	zeroDepReplacements := map[string]string{
		"chalk":       "ANSI escape codes (stdlib I/O)",
		"colors":      "ANSI escape codes (stdlib I/O)",
		"left-pad":    "String.prototype.padStart()",
		"is-even":     "n % 2 === 0",
		"is-odd":      "n % 2 !== 0",
		"is-number":   "typeof n === 'number'",
		"rimraf":      "fs.rmSync(path, {recursive: true})",
		"mkdirp":      "fs.mkdirSync(path, {recursive: true})",
		"uuid":        "crypto.randomUUID()",
		"dotenv":      "Manual .env parser (~20 lines)",
		"lodash":      "Native Array/Object methods",
		"underscore":  "Native Array/Object methods",
	}
	for name := range pkg.Dependencies {
		if replacement, ok := zeroDepReplacements[name]; ok {
			findings = append(findings, Finding{
				Rule:     "zero-dep-opportunity",
				Severity: SeverityInfo,
				Message:  fmt.Sprintf("%q can be replaced with: %s", name, replacement),
				File:     file.Path,
			})
		}
	}

	return findings
}

// auditRequirementsTxt parses a Python requirements.txt.
func auditRequirementsTxt(file FileEntry) []Finding {
	var findings []Finding

	f, err := os.Open(file.AbsPath)
	if err != nil {
		return findings
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	depCount := 0
	unpinnedCount := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") {
			continue
		}
		depCount++

		// Check if version is pinned (contains ==)
		if !strings.Contains(line, "==") {
			unpinnedCount++
		}
	}

	if depCount > 0 {
		severity := SeverityInfo
		if depCount > 10 {
			severity = SeverityWarn
		}
		findings = append(findings, Finding{
			Rule:     "dependency-count",
			Severity: severity,
			Message:  fmt.Sprintf("%d dependencies in requirements.txt", depCount),
			File:     file.Path,
		})
	}

	if unpinnedCount > 0 {
		findings = append(findings, Finding{
			Rule:     "unpinned-version",
			Severity: SeverityWarn,
			Message:  fmt.Sprintf("%d dependencies without pinned versions (missing ==)", unpinnedCount),
			File:     file.Path,
			Detail:   "Pin with == for reproducible installs",
		})
	}

	return findings
}
