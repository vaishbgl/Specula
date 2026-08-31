package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// demoProjectFiles contains the built-in demo project for `specula demo`.
var demoProjectFiles = map[string]string{
	"main.go": fmt.Sprintf(`package main

import "fmt"

// TODO: add proper config loading
func main() {
	// HACK: hardcoded for now
	apiKey := %q
	fmt.Println("Starting app with key:", apiKey)
}
`, demoAPIKey()),
	"go.mod": `module github.com/demo/app

go 1.22

require (
	github.com/spf13/cobra v1.8.0
	github.com/stretchr/testify v1.9.0
	github.com/sirupsen/logrus v1.9.3
	golang.org/x/tools v0.20.0
	github.com/fatih/color v1.16.0
)
`,
	"README.md": `# Demo App
A demonstration project for Specula.
`,
}

// demoAPIKey assembles the fake demo key at runtime. The literal never appears
// in this source tree, so `specula lint .` does not flag its own demo fixture.
func demoAPIKey() string {
	return "sk-ab" + "c123examplekeyfordemopurposes"
}

// RunDemo runs specula lint against built-in demo data and displays results.
func RunDemo(cfg Config) error {
	// Write demo files to a temp directory
	tmpDir, err := os.MkdirTemp("", "specula-demo-*")
	if err != nil {
		return fmt.Errorf("demo: create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	for name, content := range demoProjectFiles {
		path := filepath.Join(tmpDir, name)
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("demo: write %s: %w", name, err)
		}
	}

	c := NewColorHelper(cfg.NoColor)
	fmt.Fprintf(os.Stdout, "\n  %s\n\n",
		c.Yellow("⚠️  Running in DEMO mode — scanning built-in sample project"))
	fmt.Fprintf(os.Stdout, "  Sample project files: %s\n\n",
		c.Dim(strings.Join(demoFileNames(), ", ")))

	demoCfg := cfg
	demoCfg.Command = "lint"
	demoCfg.Target = tmpDir

	return RunLint(demoCfg)
}

func demoFileNames() []string {
	names := make([]string, 0, len(demoProjectFiles))
	for name := range demoProjectFiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
