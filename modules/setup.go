package modules

import (
	"fmt"
	"strings"
)

// RunProjectSetup checks for CI configuration and security policy.
func RunProjectSetup(input ModuleInput) ModuleResult {
	result := ModuleResult{Name: "Project Setup"}

	hasCI := false
	hasSecurity := false
	hasChangelog := false

	for _, file := range input.Files {
		path := file.Path
		base := strings.ToUpper(fileBaseName(path))

		// CI detection
		if strings.Contains(path, ".github/workflows/") && (strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml")) {
			hasCI = true
		}
		// GitLab CI
		if fileBaseName(path) == ".gitlab-ci.yml" {
			hasCI = true
		}
		// CircleCI
		if strings.Contains(path, ".circleci/") {
			hasCI = true
		}
		// Bitbucket Pipelines
		if fileBaseName(path) == "bitbucket-pipelines.yml" {
			hasCI = true
		}

		// Security policy
		if base == "SECURITY.MD" || base == "SECURITY" {
			hasSecurity = true
		}

		// Changelog
		if base == "CHANGELOG.MD" || base == "CHANGELOG" || base == "CHANGELOG.TXT" {
			hasChangelog = true
		}
	}

	if !hasCI {
		result.Findings = append(result.Findings, Finding{
			Rule:     "no-ci",
			Severity: SeverityWarn,
			Message:  "No CI configuration found",
			Detail:   "Add .github/workflows/, .gitlab-ci.yml, or .circleci/config.yml",
		})
	} else {
		result.Findings = append(result.Findings, Finding{
			Rule:     "ci-present",
			Severity: SeverityInfo,
			Message:  "CI configuration detected",
		})
	}

	if !hasSecurity {
		result.Findings = append(result.Findings, Finding{
			Rule:     "no-security-policy",
			Severity: SeverityInfo,
			Message:  "No SECURITY.md found",
			Detail:   fmt.Sprintf("Add a security policy to help users report vulnerabilities"),
		})
	}

	if !hasChangelog {
		result.Findings = append(result.Findings, Finding{
			Rule:     "no-changelog",
			Severity: SeverityInfo,
			Message:  "No CHANGELOG found",
			Detail:   "Maintain a changelog for user-facing changes",
		})
	}

	return result
}
