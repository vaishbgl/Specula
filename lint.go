package main

import (
	"fmt"
	"os"

	"github.com/vaishbgl/specula/modules"
)

// RunLint orchestrates the full project linting pipeline.
func RunLint(cfg Config) error {
	// 1. Walk directory
	scanResult, err := Walk(cfg.Target, DefaultWalkOptions())
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	if cfg.Verbose {
		fmt.Fprintf(os.Stderr, "debug: found %d files in %s (duration: %v)\n",
			len(scanResult.Files), cfg.Target, scanResult.Stats.Duration)
		for _, w := range scanResult.Warnings {
			fmt.Fprintf(os.Stderr, "debug: warning: %s\n", w)
		}
	}

	// Convert ScannedFiles to FileEntries for modules
	fileEntries := make([]modules.FileEntry, len(scanResult.Files))
	for i, f := range scanResult.Files {
		fileEntries[i] = modules.FileEntry{
			Path:    f.Path,
			AbsPath: f.AbsPath,
			Size:    f.Size,
		}
	}

	moduleInput := modules.ModuleInput{
		RootDir: scanResult.Root,
		Files:   fileEntries,
	}

	// 2. Execute lint modules
	var results []modules.ModuleResult

	// Module map for filtered or full execution
	allModules := []struct {
		key string
		fn  func(modules.ModuleInput) modules.ModuleResult
	}{
		{"security", modules.RunSecretScanner},
		{"deps", modules.RunDepsAudit},
		{"tests", modules.RunTestReadiness},
		{"code", modules.RunCodeQuality},
	}

	for _, m := range allModules {
		if cfg.Module != "" && cfg.Module != m.key {
			continue
		}
		res := m.fn(moduleInput)
		results = append(results, res)
	}

	// 3. Compute deterministic grade
	gradeReport := ComputeGrade(results)

	// 4. Output results
	if cfg.Output != "" {
		if err := WriteReportToFile(cfg.Output, gradeReport, results, scanResult.Stats, cfg.Target, cfg.JSON); err != nil {
			return fmt.Errorf("failed writing output: %w", err)
		}
		if !cfg.JSON {
			fmt.Fprintf(os.Stderr, "Report saved to %s\n", cfg.Output)
		}
	}

	if cfg.JSON {
		if err := RenderJSONReport(os.Stdout, gradeReport, results, scanResult.Stats, cfg.Target); err != nil {
			return fmt.Errorf("failed rendering json: %w", err)
		}
	} else {
		RenderTerminalReport(os.Stdout, gradeReport, results, scanResult.Stats, cfg.Target, cfg.NoColor)
	}

	// 5. Evaluate fail-under threshold
	if cfg.FailUnder > 0 && gradeReport.Overall < cfg.FailUnder {
		return fmt.Errorf("overall grade %d is below required threshold of %d", gradeReport.Overall, cfg.FailUnder)
	}

	// If there are critical failures and no explicit fail-under override was set to allow it
	if cfg.FailUnder == 0 && gradeReport.TotalFails > 0 {
		return fmt.Errorf("lint failed with %d critical issue(s)", gradeReport.TotalFails)
	}

	return nil
}
