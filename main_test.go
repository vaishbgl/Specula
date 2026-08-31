package main

import (
	"testing"
)

func TestParseArgs_Help(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"--help flag", []string{"specula", "--help"}, "help"},
		{"-h flag", []string{"specula", "-h"}, "help"},
		{"--version flag", []string{"specula", "--version"}, "version"},
		{"-v flag", []string{"specula", "-v"}, "version"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := parseArgs(tc.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Command != tc.want {
				t.Errorf("got command %q, want %q", cfg.Command, tc.want)
			}
		})
	}
}

func TestParseArgs_NoCommand(t *testing.T) {
	_, err := parseArgs([]string{"specula"})
	if err == nil {
		t.Fatal("expected error for no command, got nil")
	}
}

func TestParseArgs_UnknownCommand(t *testing.T) {
	_, err := parseArgs([]string{"specula", "bogus"})
	if err == nil {
		t.Fatal("expected error for unknown command, got nil")
	}
}

func TestParseLintArgs_Basic(t *testing.T) {
	cfg, err := parseArgs([]string{"specula", "lint", "."})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Command != "lint" {
		t.Errorf("got command %q, want 'lint'", cfg.Command)
	}
	if cfg.Target != "." {
		t.Errorf("got target %q, want '.'", cfg.Target)
	}
}

func TestParseLintArgs_AllFlags(t *testing.T) {
	cfg, err := parseArgs([]string{
		"specula", "lint",
		"--json",
		"--no-color",
		"--verbose",
		"--module", "security",
		"--output", "report.json",
		"--fail-under", "80",
		"/tmp/project",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.JSON {
		t.Error("expected JSON=true")
	}
	if !cfg.NoColor {
		t.Error("expected NoColor=true")
	}
	if !cfg.Verbose {
		t.Error("expected Verbose=true")
	}
	if cfg.Module != "security" {
		t.Errorf("got module %q, want 'security'", cfg.Module)
	}
	if cfg.Output != "report.json" {
		t.Errorf("got output %q, want 'report.json'", cfg.Output)
	}
	if cfg.FailUnder != 80 {
		t.Errorf("got fail-under %d, want 80", cfg.FailUnder)
	}
	if cfg.Target != "/tmp/project" {
		t.Errorf("got target %q, want '/tmp/project'", cfg.Target)
	}
}

func TestParseLintArgs_MissingPath(t *testing.T) {
	_, err := parseArgs([]string{"specula", "lint"})
	if err == nil {
		t.Fatal("expected error for missing target path, got nil")
	}
}

func TestParseLintArgs_InvalidFailUnder(t *testing.T) {
	_, err := parseArgs([]string{"specula", "lint", "--fail-under", "abc", "."})
	if err == nil {
		t.Fatal("expected error for non-integer fail-under, got nil")
	}
}

func TestParseLintArgs_FailUnderOutOfRange(t *testing.T) {
	_, err := parseArgs([]string{"specula", "lint", "--fail-under", "150", "."})
	if err == nil {
		t.Fatal("expected error for fail-under > 100, got nil")
	}
}

func TestParseDiffArgs_Basic(t *testing.T) {
	cfg, err := parseArgs([]string{"specula", "diff", "before.json", "after.json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Command != "diff" {
		t.Errorf("got command %q, want 'diff'", cfg.Command)
	}
	if cfg.Target != "before.json" {
		t.Errorf("got target %q, want 'before.json'", cfg.Target)
	}
}

func TestParseDiffArgs_MissingFiles(t *testing.T) {
	_, err := parseArgs([]string{"specula", "diff", "only_one.json"})
	if err == nil {
		t.Fatal("expected error for diff with only one file, got nil")
	}
}

func TestParseArgs_Demo(t *testing.T) {
	cfg, err := parseArgs([]string{"specula", "demo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Command != "demo" {
		t.Errorf("got command %q, want 'demo'", cfg.Command)
	}
}
