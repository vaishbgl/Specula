package modules

import (
	"os"
	"path/filepath"
	"testing"
)

// --- Secret Scanner Tests ---

func TestSecretScanner_NoSecrets(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "main.go", `package main

func main() {
    fmt.Println("hello world")
}
`)

	input := ModuleInput{
		RootDir: dir,
		Files: []FileEntry{
			{Path: "main.go", AbsPath: filepath.Join(dir, "main.go"), Size: 100},
		},
	}

	result := RunSecretScanner(input)
	if len(result.Findings) != 0 {
		t.Errorf("expected 0 findings, got %d: %+v", len(result.Findings), result.Findings)
	}
}

func TestSecretScanner_DetectsAWSKey(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "config.go", `package config
const awsKey = "AKIAIOSFODNN7EXAMPLE"
`)

	input := ModuleInput{
		RootDir: dir,
		Files: []FileEntry{
			{Path: "config.go", AbsPath: filepath.Join(dir, "config.go"), Size: 100},
		},
	}

	result := RunSecretScanner(input)
	if len(result.Findings) == 0 {
		t.Fatal("expected at least 1 finding for AWS key")
	}
	if result.Findings[0].Severity != SeverityFail {
		t.Errorf("expected FAIL severity, got %s", result.Findings[0].Severity)
	}
	if result.Findings[0].Rule != "secret-pattern" {
		t.Errorf("expected rule 'secret-pattern', got %q", result.Findings[0].Rule)
	}
}

func TestSecretScanner_DetectsGitHubToken(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "ci.sh", `#!/bin/bash
export GITHUB_TOKEN=ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghij
`)

	input := ModuleInput{
		RootDir: dir,
		Files: []FileEntry{
			{Path: "ci.sh", AbsPath: filepath.Join(dir, "ci.sh"), Size: 100},
		},
	}

	result := RunSecretScanner(input)
	if len(result.Findings) == 0 {
		t.Fatal("expected at least 1 finding for GitHub token")
	}
}

func TestSecretScanner_DetectsPrivateKey(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "key.pem", `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA...
-----END RSA PRIVATE KEY-----
`)

	input := ModuleInput{
		RootDir: dir,
		Files: []FileEntry{
			{Path: "key.pem", AbsPath: filepath.Join(dir, "key.pem"), Size: 100},
		},
	}

	result := RunSecretScanner(input)
	if len(result.Findings) == 0 {
		t.Fatal("expected at least 1 finding for private key")
	}
}

func TestSecretScanner_DetectsSensitiveFile(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, ".env", `DATABASE_URL=postgres://localhost/db`)

	input := ModuleInput{
		RootDir: dir,
		Files: []FileEntry{
			{Path: ".env", AbsPath: filepath.Join(dir, ".env"), Size: 30},
		},
	}

	result := RunSecretScanner(input)
	found := false
	for _, f := range result.Findings {
		if f.Rule == "sensitive-file" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'sensitive-file' finding for .env")
	}
}

func TestSecretScanner_FingerprintIsStable(t *testing.T) {
	fp1 := secretFingerprint("AWS Access Key", "config.go", 5)
	fp2 := secretFingerprint("AWS Access Key", "config.go", 5)
	if fp1 != fp2 {
		t.Errorf("fingerprints should be stable, got %q and %q", fp1, fp2)
	}

	fp3 := secretFingerprint("AWS Access Key", "other.go", 5)
	if fp1 == fp3 {
		t.Error("fingerprints should differ for different files")
	}
}

func TestShannonEntropy(t *testing.T) {
	// Low entropy (repetitive)
	low := shannonEntropy("aaaaaaaaaa")
	if low > 1.0 {
		t.Errorf("expected low entropy for repeated chars, got %f", low)
	}

	// High entropy (random-looking)
	high := shannonEntropy("aB3x9Zq7Lm2Yk5P")
	if high < 3.5 {
		t.Errorf("expected high entropy for random-looking string, got %f", high)
	}

	// Empty
	empty := shannonEntropy("")
	if empty != 0 {
		t.Errorf("expected 0 entropy for empty string, got %f", empty)
	}
}

// --- Dependency Audit Tests ---

func TestDepsAudit_EmptyGoMod(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "go.mod", `module example.com/myapp

go 1.22
`)

	input := ModuleInput{
		RootDir: dir,
		Files: []FileEntry{
			{Path: "go.mod", AbsPath: filepath.Join(dir, "go.mod"), Size: 100},
		},
	}

	result := RunDepsAudit(input)
	for _, f := range result.Findings {
		if f.Rule == "dependency-count" {
			t.Errorf("expected no dependency-count finding for empty go.mod, got: %s", f.Message)
		}
	}
}

func TestDepsAudit_GoModWithDeps(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "go.mod", `module example.com/myapp

go 1.22

require (
	github.com/spf13/cobra v1.8.0
	github.com/stretchr/testify v1.9.0
	golang.org/x/tools v0.20.0
)
`)

	input := ModuleInput{
		RootDir: dir,
		Files: []FileEntry{
			{Path: "go.mod", AbsPath: filepath.Join(dir, "go.mod"), Size: 200},
		},
	}

	result := RunDepsAudit(input)
	found := false
	for _, f := range result.Findings {
		if f.Rule == "dependency-count" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected dependency-count finding for go.mod with 3 deps")
	}
}

func TestDepsAudit_PackageJSON(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "package.json", `{
  "name": "test-app",
  "dependencies": {
    "express": "^4.18.0",
    "chalk": "*",
    "lodash": "latest"
  }
}`)

	input := ModuleInput{
		RootDir: dir,
		Files: []FileEntry{
			{Path: "package.json", AbsPath: filepath.Join(dir, "package.json"), Size: 200},
		},
	}

	result := RunDepsAudit(input)

	// Should find: dependency-count, unpinned-version (chalk, lodash), zero-dep-opportunity (chalk, lodash)
	rules := make(map[string]int)
	for _, f := range result.Findings {
		rules[f.Rule]++
	}

	if rules["dependency-count"] == 0 {
		t.Error("expected dependency-count finding")
	}
	if rules["unpinned-version"] == 0 {
		t.Error("expected unpinned-version finding")
	}
	if rules["zero-dep-opportunity"] == 0 {
		t.Error("expected zero-dep-opportunity finding for chalk/lodash")
	}
}

func TestDepsAudit_MalformedPackageJSON(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "package.json", `{this is not valid json}`)

	input := ModuleInput{
		RootDir: dir,
		Files: []FileEntry{
			{Path: "package.json", AbsPath: filepath.Join(dir, "package.json"), Size: 30},
		},
	}

	result := RunDepsAudit(input)
	found := false
	for _, f := range result.Findings {
		if f.Rule == "malformed-manifest" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected malformed-manifest finding")
	}
}

func TestDepsAudit_RequirementsTxt(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "requirements.txt", `flask==2.3.0
requests>=2.28.0
numpy
# this is a comment

pandas==1.5.3
`)

	input := ModuleInput{
		RootDir: dir,
		Files: []FileEntry{
			{Path: "requirements.txt", AbsPath: filepath.Join(dir, "requirements.txt"), Size: 100},
		},
	}

	result := RunDepsAudit(input)
	rules := make(map[string]int)
	for _, f := range result.Findings {
		rules[f.Rule]++
	}
	if rules["dependency-count"] == 0 {
		t.Error("expected dependency-count finding")
	}
	if rules["unpinned-version"] == 0 {
		t.Error("expected unpinned-version finding for requests and numpy")
	}
}

// --- helper ---

func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
