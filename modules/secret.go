package modules

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"
)

// secretPatterns are regex patterns matching common secret/key formats.
var secretPatterns = []struct {
	name    string
	pattern *regexp.Regexp
}{
	{"AWS Access Key", regexp.MustCompile(`AKIA[0-9A-Z]{16}`)},
	{"GitHub Token", regexp.MustCompile(`ghp_[A-Za-z0-9]{36}`)},
	{"GitHub OAuth", regexp.MustCompile(`gho_[A-Za-z0-9]{36}`)},
	{"GitHub App Token", regexp.MustCompile(`(ghu|ghs|ghr)_[A-Za-z0-9]{36}`)},
	{"OpenAI API Key", regexp.MustCompile(`sk-[A-Za-z0-9]{20,}`)},
	{"Slack Token", regexp.MustCompile(`xox[bpors]-[A-Za-z0-9-]{10,}`)},
	{"Private Key Header", regexp.MustCompile(`-----BEGIN (RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`)},
	{"Bearer Token", regexp.MustCompile(`[Bb]earer\s+[A-Za-z0-9\-_.~+/]{20,}`)},
	{"Generic API Key Assignment", regexp.MustCompile(`(?i)(api[_-]?key|api[_-]?secret|auth[_-]?token|access[_-]?token)\s*[:=]\s*["']?[A-Za-z0-9\-_.]{16,}["']?`)},
}

// sensitiveFilePatterns are filenames that should not be in the repo tree.
var sensitiveFilePatterns = []string{
	".env",
	".env.local",
	".env.production",
	".env.staging",
	"credentials.json",
	"service-account.json",
	"id_rsa",
	"id_ed25519",
}

// RunSecretScanner scans files for potential secrets and sensitive files.
// It never retains raw secret values — only redacted previews and fingerprints.
func RunSecretScanner(input ModuleInput) ModuleResult {
	result := ModuleResult{Name: "Secret Scanner"}

	for _, file := range input.Files {
		// Check for sensitive filenames
		baseName := fileBaseName(file.Path)
		for _, sensitive := range sensitiveFilePatterns {
			if baseName == sensitive {
				result.Findings = append(result.Findings, Finding{
					Rule:     "sensitive-file",
					Severity: SeverityWarn,
					Message:  fmt.Sprintf("Sensitive file %q found in project tree", baseName),
					File:     file.Path,
					Detail:   "Consider adding this to .gitignore",
				})
			}
		}

		// Scan file content for secret patterns
		findings := scanFileForSecrets(file)
		result.Findings = append(result.Findings, findings...)
	}

	return result
}

// scanFileForSecrets scans a single file for secret patterns.
func scanFileForSecrets(file FileEntry) []Finding {
	var findings []Finding

	f, err := os.Open(file.AbsPath)
	if err != nil {
		return findings
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		for _, sp := range secretPatterns {
			if sp.pattern.MatchString(line) {
				// Generate a stable fingerprint without storing the secret
				fingerprint := secretFingerprint(sp.name, file.Path, lineNum)

				findings = append(findings, Finding{
					Rule:     "secret-pattern",
					Severity: SeverityFail,
					Message:  fmt.Sprintf("Possible %s detected", sp.name),
					File:     file.Path,
					Line:     lineNum,
					Detail:   fmt.Sprintf("fingerprint:%s", fingerprint),
				})
				break // one finding per line, avoid duplicates
			}
		}
	}

	return findings
}

// secretFingerprint creates a one-way hash for diffing without exposing values.
func secretFingerprint(ruleName, filePath string, line int) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s:%s:%d", ruleName, filePath, line)
	sum := h.Sum(nil)
	return fmt.Sprintf("%x", sum[:8]) // 16-char hex prefix
}

// shannonEntropy calculates the Shannon entropy of a string.
// Higher values (> 4.5) suggest randomness / possible secrets.
func shannonEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}
	freq := make(map[rune]float64)
	for _, r := range s {
		freq[r]++
	}
	length := float64(len([]rune(s)))
	entropy := 0.0
	for _, count := range freq {
		p := count / length
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}
	return entropy
}

// fileBaseName returns the last component of a path.
func fileBaseName(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return path
	}
	return parts[len(parts)-1]
}
