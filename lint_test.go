package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/vaishbgl/specula/modules"
)

func TestRenderTerminalReport(t *testing.T) {
	var buf bytes.Buffer
	report := GradeReport{
		Overall:    88,
		Grade:      "B",
		TotalFails: 0,
		TotalWarns: 2,
		TotalInfos: 1,
		ModuleScores: []ModuleScore{
			{Name: "Secret Scanner", Score: 100, Status: "PASS", TopMsg: "No issues"},
			{Name: "Dependency Audit", Score: 75, Status: "PASS", TopMsg: "3 dependencies"},
		},
	}

	results := []modules.ModuleResult{
		{
			Name: "Dependency Audit",
			Findings: []modules.Finding{
				{Rule: "unpinned-version", Severity: modules.SeverityWarn, Message: "Unpinned dep: lodash"},
			},
		},
	}

	stats := ScanStats{
		FilesScanned: 15,
		Duration:     120 * time.Millisecond,
		BytesTotal:   45000,
	}

	RenderTerminalReport(&buf, report, results, stats, "/path/to/project", true)
	output := buf.String()

	if !strings.Contains(output, "SPECULA") {
		t.Error("expected banner containing SPECULA")
	}
	if !strings.Contains(output, "Grade: 🟢 B (88/100)") {
		t.Errorf("expected grade line in output, got:\n%s", output)
	}
	if !strings.Contains(output, "Secret Scanner") {
		t.Error("expected module name in table")
	}
	if !strings.Contains(output, "Top Actionable Fixes") {
		t.Error("expected fix suggestions section")
	}
}

func TestRenderJSONReport(t *testing.T) {
	var buf bytes.Buffer
	report := GradeReport{
		Overall: 92,
		Grade:   "A",
	}
	results := []modules.ModuleResult{
		{Name: "Secret Scanner", Findings: nil},
	}
	stats := ScanStats{
		FilesScanned: 10,
		Duration:     50 * time.Millisecond,
	}

	err := RenderJSONReport(&buf, report, results, stats, ".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed LintReportJSON
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed parsing rendered JSON: %v", err)
	}

	if parsed.Grade.Overall != 92 {
		t.Errorf("got overall score %d, want 92", parsed.Grade.Overall)
	}
	if parsed.FilesScanned != 10 {
		t.Errorf("got files scanned %d, want 10", parsed.FilesScanned)
	}
}

func TestRunLint_EndToEndClean(t *testing.T) {
	dir := t.TempDir()

	// Write clean minimal project
	writeFile(t, dir, "main.go", "package main\n\nfunc main() {}\n")
	writeFile(t, dir, "main_test.go", "package main\n\nimport \"testing\"\n\nfunc TestMain(t *testing.T) {}\n")

	cfg := Config{
		Command: "lint",
		Target:  dir,
		NoColor: true,
	}

	err := RunLint(cfg)
	if err != nil {
		t.Fatalf("unexpected error running lint on clean project: %v", err)
	}
}

func TestRunLint_FailUnderThreshold(t *testing.T) {
	dir := t.TempDir()

	// Write project with a secret (which incurs FAIL deduction -25 -> score 75)
	writeFile(t, dir, "main.go", "package main\nconst key = \"AKIAIOSFODNN7EXAMPLE\"\n")

	cfg := Config{
		Command:   "lint",
		Target:    dir,
		NoColor:   true,
		FailUnder: 90, // score will be < 90
	}

	err := RunLint(cfg)
	if err == nil {
		t.Fatal("expected error due to score being below fail-under threshold, got nil")
	}
	if !strings.Contains(err.Error(), "below required threshold") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRunLint_SaveOutputFile(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "report.json")

	writeFile(t, dir, "main.go", "package main\nfunc main() {}\n")
	writeFile(t, dir, "main_test.go", "package main\nimport \"testing\"\nfunc TestX(t *testing.T){}\n")

	cfg := Config{
		Command: "lint",
		Target:  dir,
		JSON:    true,
		Output:  outPath,
		NoColor: true,
	}

	err := RunLint(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify report file was created
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed reading output file: %v", err)
	}

	var parsed LintReportJSON
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed parsing report file: %v", err)
	}
	if parsed.Grade.Grade != "A" {
		t.Errorf("got grade %s (score %d), want 'A'", parsed.Grade.Grade, parsed.Grade.Overall)
	}
}
