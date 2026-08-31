// Specula — a zero-dependency project linter.
// Entry point: thin dispatch only. All logic lives in other files.
package main

import (
	"fmt"
	"os"
)

var version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "lint":
		fmt.Fprintf(os.Stderr, "specula %s: lint command not yet implemented\n", version)
		os.Exit(1)
	case "diff":
		fmt.Fprintf(os.Stderr, "specula %s: diff command not yet implemented\n", version)
		os.Exit(1)
	case "demo":
		fmt.Fprintf(os.Stderr, "specula %s: demo command not yet implemented\n", version)
		os.Exit(1)
	case "--version", "-v":
		fmt.Printf("specula %s\n", version)
	case "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "specula: unknown command %q\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "specula %s — a zero-dependency project linter\n\n", version)
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  specula lint [flags] <path>    Lint a project directory\n")
	fmt.Fprintf(os.Stderr, "  specula diff <a.json> <b.json> Compare two lint reports\n")
	fmt.Fprintf(os.Stderr, "  specula demo                   Run against built-in sample data\n\n")
	fmt.Fprintf(os.Stderr, "Flags (for lint):\n")
	fmt.Fprintf(os.Stderr, "  --module <name>   Run a single module\n")
	fmt.Fprintf(os.Stderr, "  --json            Output results as JSON\n")
	fmt.Fprintf(os.Stderr, "  --output <file>   Save report to file\n")
	fmt.Fprintf(os.Stderr, "  --fail-under <n>  Exit code 1 if score < n\n")
	fmt.Fprintf(os.Stderr, "  --no-color        Disable colored output\n")
	fmt.Fprintf(os.Stderr, "  --verbose         Enable debug output\n")
	fmt.Fprintf(os.Stderr, "  --version, -v     Print version\n")
	fmt.Fprintf(os.Stderr, "  --help, -h        Show this help\n")
}
