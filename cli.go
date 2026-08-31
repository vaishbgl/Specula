package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
)

// Config holds all parsed CLI configuration.
type Config struct {
	Command   string // "lint", "diff", "demo"
	Target    string // path to lint (for lint command)
	Module    string // single module filter
	JSON      bool   // JSON output
	Output    string // output file path
	FailUnder int    // exit 1 if score below this
	NoColor   bool   // disable ANSI colors
	Verbose   bool   // debug output
}

// parseArgs parses os.Args into a Config struct.
// Returns the config and any error encountered.
func parseArgs(args []string) (Config, error) {
	if len(args) < 2 {
		return Config{}, fmt.Errorf("no command specified")
	}

	cmd := args[1]

	switch cmd {
	case "--help", "-h":
		return Config{Command: "help"}, nil
	case "--version", "-v":
		return Config{Command: "version"}, nil
	case "lint":
		return parseLintArgs(args[2:])
	case "diff":
		return parseDiffArgs(args[2:])
	case "demo":
		return Config{Command: "demo"}, nil
	default:
		return Config{}, fmt.Errorf("unknown command %q", cmd)
	}
}

// parseLintArgs parses flags for the lint subcommand.
func parseLintArgs(args []string) (Config, error) {
	fs := flag.NewFlagSet("lint", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	cfg := Config{Command: "lint"}

	fs.StringVar(&cfg.Module, "module", "", "Run a single module (security, deps, tests, code, docs, setup)")
	fs.BoolVar(&cfg.JSON, "json", false, "Output results as JSON")
	fs.StringVar(&cfg.Output, "output", "", "Save report to file")
	fs.BoolVar(&cfg.NoColor, "no-color", false, "Disable colored output")
	fs.BoolVar(&cfg.Verbose, "verbose", false, "Enable debug output")

	// Custom flag for --fail-under since it needs int parsing
	failUnderStr := fs.String("fail-under", "0", "Exit code 1 if grade score < n")

	// main prints the usage text for --help after parsing succeeds.
	fs.Usage = func() {}

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return Config{Command: "lint-help"}, nil
		}
		return Config{}, err
	}

	// Parse fail-under
	if *failUnderStr != "0" {
		n, err := strconv.Atoi(*failUnderStr)
		if err != nil {
			return Config{}, fmt.Errorf("--fail-under requires an integer, got %q", *failUnderStr)
		}
		if n < 0 || n > 100 {
			return Config{}, fmt.Errorf("--fail-under must be 0-100, got %d", n)
		}
		cfg.FailUnder = n
	}

	// Respect NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		cfg.NoColor = true
	}

	// Positional argument: target path
	remaining := fs.Args()
	if len(remaining) == 0 {
		return Config{}, fmt.Errorf("lint requires a target path, e.g.: specula lint .")
	}
	if len(remaining) > 1 {
		return Config{}, fmt.Errorf("lint accepts exactly one target path")
	}
	if cfg.Module != "" && !isValidModule(cfg.Module) {
		return Config{}, fmt.Errorf("unknown module %q (valid: security, deps, tests, code, docs, setup)", cfg.Module)
	}
	cfg.Target = remaining[0]

	return cfg, nil
}

// parseDiffArgs parses arguments for the diff subcommand.
func parseDiffArgs(args []string) (Config, error) {
	if len(args) != 2 {
		return Config{}, fmt.Errorf("diff requires two JSON report files, e.g.: specula diff before.json after.json")
	}
	// We'll store both paths in Target (before) and Output (after) for now.
	// A proper struct field will be added when diff is implemented.
	return Config{
		Command: "diff",
		Target:  args[0],
		Output:  args[1],
	}, nil
}

func isValidModule(name string) bool {
	for _, module := range []string{"security", "deps", "tests", "code", "docs", "setup"} {
		if name == module {
			return true
		}
	}
	return false
}

func printLintUsage() {
	fmt.Fprintf(os.Stderr, "Usage: specula lint [flags] <path>\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	fmt.Fprintf(os.Stderr, "  --module <name>   Run a single module (security, deps, tests, code, docs, setup)\n")
	fmt.Fprintf(os.Stderr, "  --json            Output results as JSON\n")
	fmt.Fprintf(os.Stderr, "  --output <file>   Save report to file\n")
	fmt.Fprintf(os.Stderr, "  --fail-under <n>  Exit code 1 if grade score < n (0-100)\n")
	fmt.Fprintf(os.Stderr, "  --no-color        Disable colored output (also respects NO_COLOR env)\n")
	fmt.Fprintf(os.Stderr, "  --verbose         Enable debug output\n\n")
	fmt.Fprintf(os.Stderr, "Example:\n")
	fmt.Fprintf(os.Stderr, "  specula lint --json --fail-under 80 .\n")
}
