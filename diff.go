package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/vaishbgl/specula/modules"
)

// DiffFinding represents a change between two reports.
type DiffFinding struct {
	Status      string `json:"status"` // "resolved" or "regressed"
	Rule        string `json:"rule"`
	File        string `json:"file"`
	Message     string `json:"message"`
	Fingerprint string `json:"fingerprint"`
}

// DiffReport is the output of comparing two reports.
type DiffReport struct {
	ScoreBefore int           `json:"score_before"`
	ScoreAfter  int           `json:"score_after"`
	ScoreDelta  int           `json:"score_delta"`
	Resolved    []DiffFinding `json:"resolved"`
	Regressed   []DiffFinding `json:"regressed"`
}

// RunDiff compares two JSON report files and shows regressions/improvements.
func RunDiff(cfg Config) error {
	before, err := loadReport(cfg.Target)
	if err != nil {
		return fmt.Errorf("load before-report %q: %w", cfg.Target, err)
	}

	after, err := loadReport(cfg.Output)
	if err != nil {
		return fmt.Errorf("load after-report %q: %w", cfg.Output, err)
	}

	diff := computeDiff(before, after)

	c := NewColorHelper(cfg.NoColor)

	fmt.Fprintln(os.Stdout)
	fmt.Fprintf(os.Stdout, "  %s\n", c.Cyan("╔══════════════════════════════════════════╗"))
	fmt.Fprintf(os.Stdout, "  %s  %s    %s\n", c.Cyan("║"), c.Bold("⬡  SPECULA  ·  Regression Report"), c.Cyan("║"))
	fmt.Fprintf(os.Stdout, "  %s\n\n", c.Cyan("╚══════════════════════════════════════════╝"))

	deltaStr := fmt.Sprintf("%+d", diff.ScoreDelta)
	deltaColor := c.Green
	if diff.ScoreDelta < 0 {
		deltaColor = c.Red
	}

	fmt.Fprintf(os.Stdout, "  Score: %s → %s (%s)\n\n",
		c.Bold(fmt.Sprintf("%d", diff.ScoreBefore)),
		c.Bold(fmt.Sprintf("%d", diff.ScoreAfter)),
		deltaColor(deltaStr),
	)

	if len(diff.Resolved) > 0 {
		fmt.Fprintf(os.Stdout, "  %s (%d)\n", c.Green("✅ Resolved"), len(diff.Resolved))
		for _, d := range diff.Resolved {
			loc := d.File
			if loc == "" {
				loc = "project-wide"
			}
			fmt.Fprintf(os.Stdout, "    [%s] %s in %s\n", d.Rule, d.Message, loc)
		}
		fmt.Fprintln(os.Stdout)
	}

	if len(diff.Regressed) > 0 {
		fmt.Fprintf(os.Stdout, "  %s (%d)\n", c.Red("❌ Regressed"), len(diff.Regressed))
		for _, d := range diff.Regressed {
			loc := d.File
			if loc == "" {
				loc = "project-wide"
			}
			fmt.Fprintf(os.Stdout, "    [%s] %s in %s\n", d.Rule, d.Message, loc)
		}
		fmt.Fprintln(os.Stdout)
	}

	if len(diff.Resolved) == 0 && len(diff.Regressed) == 0 {
		fmt.Fprintf(os.Stdout, "  %s No changes in findings between reports.\n\n", c.Green("✅"))
	}

	// Exit 1 if regressions found (useful in CI)
	if len(diff.Regressed) > 0 {
		return fmt.Errorf("%d regression(s) detected", len(diff.Regressed))
	}

	return nil
}

// computeDiff builds the diff report from two JSON reports.
func computeDiff(before, after LintReportJSON) DiffReport {
	diff := DiffReport{
		ScoreBefore: before.Grade.Overall,
		ScoreAfter:  after.Grade.Overall,
		ScoreDelta:  after.Grade.Overall - before.Grade.Overall,
	}

	// Build fingerprint sets
	beforeFPs := buildFingerprintSet(before.Results)
	afterFPs := buildFingerprintSet(after.Results)

	// Resolved: in before but not in after
	for fp, d := range beforeFPs {
		if _, stillPresent := afterFPs[fp]; !stillPresent {
			d.Status = "resolved"
			diff.Resolved = append(diff.Resolved, d)
		}
	}

	// Regressed: in after but not in before
	for fp, d := range afterFPs {
		if _, wasBefore := beforeFPs[fp]; !wasBefore {
			d.Status = "regressed"
			diff.Regressed = append(diff.Regressed, d)
		}
	}

	return diff
}

// buildFingerprintSet creates a map of fingerprint -> DiffFinding from module results.
func buildFingerprintSet(results []modules.ModuleResult) map[string]DiffFinding {
	set := make(map[string]DiffFinding)
	for _, r := range results {
		for _, f := range r.Findings {
			fp := fingerprintFinding(f)
			set[fp] = DiffFinding{
				Rule:        f.Rule,
				File:        f.File,
				Message:     f.Message,
				Fingerprint: fp,
			}
		}
	}
	return set
}

// fingerprintFinding creates a stable string key for a finding.
func fingerprintFinding(f modules.Finding) string {
	return fmt.Sprintf("%s|%s|%d", f.Rule, f.File, f.Line)
}

// loadReport reads and parses a JSON lint report file.
func loadReport(path string) (LintReportJSON, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return LintReportJSON{}, err
	}
	var report LintReportJSON
	if err := json.Unmarshal(data, &report); err != nil {
		return LintReportJSON{}, fmt.Errorf("parse JSON: %w", err)
	}
	return report, nil
}
