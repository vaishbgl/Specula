package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/vaishbgl/specula/modules"
)

// ANSI color codes
const (
	ansiReset   = "\033[0m"
	ansiBold    = "\033[1m"
	ansiDim     = "\033[2m"
	ansiRed     = "\033[31m"
	ansiGreen   = "\033[32m"
	ansiYellow  = "\033[33m"
	ansiBlue    = "\033[34m"
	ansiMagenta = "\033[35m"
	ansiCyan    = "\033[36m"
	ansiGray    = "\033[90m"
)

// ColorHelper handles colored terminal output respecting no-color settings.
type ColorHelper struct {
	Disabled bool
}

// NewColorHelper returns a ColorHelper based on config.
func NewColorHelper(noColor bool) ColorHelper {
	return ColorHelper{Disabled: noColor}
}

func (c ColorHelper) Color(code, text string) string {
	if c.Disabled {
		return text
	}
	return code + text + ansiReset
}

func (c ColorHelper) Bold(text string) string   { return c.Color(ansiBold, text) }
func (c ColorHelper) Dim(text string) string    { return c.Color(ansiDim, text) }
func (c ColorHelper) Green(text string) string  { return c.Color(ansiGreen, text) }
func (c ColorHelper) Yellow(text string) string { return c.Color(ansiYellow, text) }
func (c ColorHelper) Red(text string) string    { return c.Color(ansiRed, text) }
func (c ColorHelper) Cyan(text string) string   { return c.Color(ansiCyan, text) }
func (c ColorHelper) Gray(text string) string   { return c.Color(ansiGray, text) }

// LintReportJSON is the top-level schema for JSON export.
type LintReportJSON struct {
	Version      string                 `json:"specula_version"`
	ScannedPath  string                 `json:"scanned_path"`
	Timestamp    string                 `json:"timestamp"`
	DurationMs   int64                  `json:"duration_ms"`
	FilesScanned int                    `json:"files_scanned"`
	BytesScanned int64                  `json:"bytes_scanned"`
	Grade        GradeReport            `json:"grade"`
	Results      []modules.ModuleResult `json:"modules"`
}

// RenderTerminalReport prints the human-friendly report to w.
func RenderTerminalReport(w io.Writer, report GradeReport, results []modules.ModuleResult, stats ScanStats, path string, noColor bool) {
	c := NewColorHelper(noColor)

	fmt.Fprintln(w)
	// Banner
	fmt.Fprintf(w, "  %s\n", c.Cyan("╔══════════════════════════════════════════╗"))
	fmt.Fprintf(w, "  %s   %s  %s    %s\n", c.Cyan("║"), c.Bold("⬡  SPECULA"), c.Dim("·  Project Lint Report"), c.Cyan("║"))
	fmt.Fprintf(w, "  %s\n\n", c.Cyan("╚══════════════════════════════════════════╝"))

	// Grade and Stats
	gradeEmoji := "🟢"
	gradeColor := c.Green
	if report.Overall < 75 {
		gradeEmoji = "🟡"
		gradeColor = c.Yellow
	}
	if report.Overall < 60 || report.TotalFails > 0 {
		gradeEmoji = "🔴"
		gradeColor = c.Red
	}

	fmt.Fprintf(w, "  Grade: %s %s (%d/100)\n",
		gradeEmoji,
		gradeColor(c.Bold(report.Grade)),
		report.Overall,
	)

	durationStr := fmt.Sprintf("%.2fs", stats.Duration.Seconds())
	if stats.Duration < time.Second {
		durationStr = fmt.Sprintf("%dms", stats.Duration.Milliseconds())
	}
	fmt.Fprintf(w, "  %s\n\n", c.Dim(fmt.Sprintf("Scanned: %d files · %s · 0 network calls", stats.FilesScanned, durationStr)))

	// Module Breakdown Table
	renderModuleTable(w, report.ModuleScores, c)

	// Findings Summary
	fmt.Fprintf(w, "\n  Findings: %s %s · %s %s · %s %s\n\n",
		c.Red(fmt.Sprintf("%d", report.TotalFails)), c.Bold("FAIL"),
		c.Yellow(fmt.Sprintf("%d", report.TotalWarns)), c.Bold("WARN"),
		c.Cyan(fmt.Sprintf("%d", report.TotalInfos)), c.Bold("INFO"),
	)

	// Fix Suggestions Footer
	renderFixSuggestions(w, results, c)
}

// renderModuleTable draws a box-aligned table of module scores.
func renderModuleTable(w io.Writer, scores []ModuleScore, c ColorHelper) {
	fmt.Fprintf(w, "  ┌──────────────────────┬───────┬─────────┬──────────────────────────────────────┐\n")
	fmt.Fprintf(w, "  │ %-20s │ %-5s │ %-7s │ %-36s │\n", "Module", "Score", "Status", "Top Finding")
	fmt.Fprintf(w, "  ├──────────────────────┼───────┼─────────┼──────────────────────────────────────┤\n")

	for _, ms := range scores {
		statusStr := "✅ PASS"
		switch ms.Status {
		case "WARN":
			statusStr = "⚠️  WARN"
		case "FAIL":
			statusStr = "❌ FAIL"
		case "SKIPPED":
			statusStr = "⏭️  SKIP"
		}

		topMsg := ms.TopMsg
		if len(topMsg) > 36 {
			topMsg = topMsg[:33] + "..."
		}

		fmt.Fprintf(w, "  │ %-20s │  %3d  │ %-7s │ %-36s │\n",
			ms.Name,
			ms.Score,
			statusStr,
			topMsg,
		)
	}

	fmt.Fprintf(w, "  └──────────────────────┴───────┴─────────┴──────────────────────────────────────┘\n")
}

// renderFixSuggestions displays the highest priority fixes.
func renderFixSuggestions(w io.Writer, results []modules.ModuleResult, c ColorHelper) {
	type fixItem struct {
		priority string
		msg      string
		file     string
	}
	var fixes []fixItem

	for _, r := range results {
		for _, f := range r.Findings {
			if f.Severity == modules.SeverityFail {
				fixes = append(fixes, fixItem{
					priority: c.Red("🔴 HIGH"),
					msg:      f.Message,
					file:     f.File,
				})
			} else if f.Severity == modules.SeverityWarn && len(fixes) < 4 {
				fixes = append(fixes, fixItem{
					priority: c.Yellow("🟡 MED "),
					msg:      f.Message,
					file:     f.File,
				})
			}
		}
	}

	if len(fixes) == 0 {
		return
	}

	fmt.Fprintf(w, "  ┌─ %s ──────────────────────────────────────────┐\n", c.Bold("Top Actionable Fixes"))
	count := 0
	for i, fix := range fixes {
		if count >= 3 {
			break
		}
		detail := fix.msg
		if fix.file != "" {
			detail = fmt.Sprintf("%s (%s)", fix.msg, fix.file)
		}
		if len(detail) > 55 {
			detail = detail[:52] + "..."
		}
		fmt.Fprintf(w, "  │  %d. %s  %-52s │\n", i+1, fix.priority, detail)
		count++
	}
	fmt.Fprintf(w, "  └──────────────────────────────────────────────────────────────────┘\n")
}

// RenderJSONReport formats the findings into JSON and writes to w.
func RenderJSONReport(w io.Writer, report GradeReport, results []modules.ModuleResult, stats ScanStats, path string) error {
	out := LintReportJSON{
		Version:      version,
		ScannedPath:  path,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		DurationMs:   stats.Duration.Milliseconds(),
		FilesScanned: stats.FilesScanned,
		BytesScanned: stats.BytesTotal,
		Grade:        report,
		Results:      results,
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

// WriteReportToFile handles saving output to a file path.
func WriteReportToFile(filePath string, report GradeReport, results []modules.ModuleResult, stats ScanStats, path string, asJSON bool) error {
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("create output file %q: %w", filePath, err)
	}
	defer f.Close()

	if asJSON || strings.HasSuffix(filePath, ".json") {
		return RenderJSONReport(f, report, results, stats, path)
	}

	RenderTerminalReport(f, report, results, stats, path, true)
	return nil
}
