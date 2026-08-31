package modules

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// RunDocsCheck checks README, LICENSE, and environment file drift.
func RunDocsCheck(input ModuleInput) ModuleResult {
	result := ModuleResult{Name: "Docs & Env Check"}

	hasREADME := false
	hasLICENSE := false
	hasEnvExample := false
	hasEnv := false
	var readmePath, envExamplePath, envPath string

	for _, file := range input.Files {
		base := strings.ToUpper(fileBaseName(file.Path))
		switch {
		case base == "README.MD" || base == "README" || base == "README.TXT" || base == "README.RST":
			hasREADME = true
			readmePath = file.Path
		case base == "LICENSE" || base == "LICENSE.MD" || base == "LICENSE.TXT":
			hasLICENSE = true
		case fileBaseName(file.Path) == ".env.example" || fileBaseName(file.Path) == ".env.sample":
			hasEnvExample = true
			envExamplePath = file.AbsPath
		case fileBaseName(file.Path) == ".env":
			hasEnv = true
			envPath = file.AbsPath
		}
	}

	// README checks
	if !hasREADME {
		result.Findings = append(result.Findings, Finding{
			Rule:     "missing-readme",
			Severity: SeverityFail,
			Message:  "No README file found",
			Detail:   "Add README.md with project description, installation, and usage",
		})
	} else {
		findings := auditReadme(readmePath, input.RootDir)
		result.Findings = append(result.Findings, findings...)
	}

	// LICENSE check
	if !hasLICENSE {
		result.Findings = append(result.Findings, Finding{
			Rule:     "missing-license",
			Severity: SeverityWarn,
			Message:  "No LICENSE file found",
			Detail:   "Add an OSI-approved license (e.g. MIT, Apache-2.0)",
		})
	}

	// Environment drift check
	if hasEnvExample && hasEnv {
		driftFindings := checkEnvDrift(envExamplePath, envPath)
		result.Findings = append(result.Findings, driftFindings...)
	} else if hasEnv && !hasEnvExample {
		result.Findings = append(result.Findings, Finding{
			Rule:     "env-no-example",
			Severity: SeverityInfo,
			Message:  ".env found but no .env.example — collaborators won't know required variables",
			Detail:   "Add .env.example with placeholder values",
		})
	}

	return result
}

// auditReadme checks README for minimum length and key sections.
func auditReadme(relPath, rootDir string) []Finding {
	var findings []Finding

	absPath := rootDir + "/" + relPath
	f, err := os.Open(absPath)
	if err != nil {
		return findings
	}
	defer f.Close()

	content, err := os.ReadFile(absPath)
	if err != nil {
		return findings
	}

	text := string(content)
	lower := strings.ToLower(text)
	wordCount := len(strings.Fields(text))

	if wordCount < 100 {
		findings = append(findings, Finding{
			Rule:     "readme-too-short",
			Severity: SeverityWarn,
			Message:  fmt.Sprintf("README has only %d words (minimum: 100)", wordCount),
			File:     relPath,
			Detail:   "Add installation steps, usage examples, and project description",
		})
	}

	// Check for key sections
	sections := map[string]string{
		"install":   "installation section (install / getting started)",
		"usage":     "usage section (usage / how to use)",
		"licen":     "license section",
	}
	for keyword, desc := range sections {
		if !strings.Contains(lower, keyword) {
			findings = append(findings, Finding{
				Rule:     "readme-missing-section",
				Severity: SeverityInfo,
				Message:  fmt.Sprintf("README appears to be missing a %s", desc),
				File:     relPath,
			})
		}
	}

	return findings
}

// checkEnvDrift compares keys in .env.example vs .env for missing/extra entries.
func checkEnvDrift(examplePath, envPath string) []Finding {
	var findings []Finding

	exampleKeys := parseEnvKeys(examplePath)
	envKeys := parseEnvKeys(envPath)

	// Keys in example but not in .env
	for key := range exampleKeys {
		if !envKeys[key] {
			findings = append(findings, Finding{
				Rule:     "env-drift",
				Severity: SeverityWarn,
				Message:  fmt.Sprintf("Key %q is in .env.example but missing from .env", key),
				Detail:   "Your local .env may be out of date",
			})
		}
	}

	return findings
}

// parseEnvKeys reads an env file and returns a set of variable names.
// Values are intentionally discarded for safety.
func parseEnvKeys(path string) map[string]bool {
	keys := make(map[string]bool)
	f, err := os.Open(path)
	if err != nil {
		return keys
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) >= 1 {
			key := strings.TrimSpace(parts[0])
			if key != "" {
				keys[key] = true
			}
		}
	}
	return keys
}
