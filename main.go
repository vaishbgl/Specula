// Specula — a zero-dependency project linter.
//
// Entry point: thin dispatch only. All logic lives in other files.
// Errors go to stderr. Only main() calls os.Exit().
package main

import (
	"fmt"
	"os"
)

var version = "0.1.0"

func main() {
	cfg, err := parseArgs(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nRun 'specula --help' for usage.\n")
		os.Exit(1)
	}

	switch cfg.Command {
	case "help":
		printUsage()
		os.Exit(0)
	case "version":
		fmt.Printf("specula %s\n", version)
		os.Exit(0)
	case "lint":
		if err := runLint(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "diff":
		if err := runDiff(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "demo":
		if err := runDemo(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}
}

// runLint orchestrates the full lint pipeline.
func runLint(cfg Config) error {
	return RunLint(cfg)
}

// runDiff compares two lint reports.
func runDiff(cfg Config) error {
	return RunDiff(cfg)
}

// runDemo runs lint against built-in sample data.
func runDemo(cfg Config) error {
	return RunDemo(cfg)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "specula %s — a zero-dependency project linter\n\n", version)
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  specula lint [flags] <path>    Lint a project directory\n")
	fmt.Fprintf(os.Stderr, "  specula diff <a.json> <b.json> Compare two lint reports\n")
	fmt.Fprintf(os.Stderr, "  specula demo                   Run against built-in sample data\n\n")
	fmt.Fprintf(os.Stderr, "Run 'specula lint --help' for lint-specific flags.\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	fmt.Fprintf(os.Stderr, "  --version, -v     Print version\n")
	fmt.Fprintf(os.Stderr, "  --help, -h        Show this help\n")
}
