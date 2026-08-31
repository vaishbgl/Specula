# STDLIB.md — Zero Dependency Substitution Log

Specula uses **zero third-party packages**. Every package that would normally be `go get`'d
was replaced with a standard library equivalent. This document records those substitutions
with honest trade-off analysis.

| # | Package Replaced | Weekly Downloads (npm/pkg.go.dev) | Stdlib Replacement | File | Trade-off |
|:---:|---|---|---|---|---|
| 1 | `github.com/spf13/cobra` | ~30k Go imports/week | `flag.FlagSet` + manual subcommand dispatch | `cli.go` | No automatic nested subcommands or shell completions |
| 2 | `github.com/fatih/color` | ~20k Go imports/week | ANSI VT100 escape codes via `fmt.Fprintf` | `output.go` | Limited to 8 basic colors; no 256-color or truecolor |
| 3 | `github.com/olekukonko/tablewriter` | ~15k Go imports/week | Hand-rolled Unicode box-drawing character table | `output.go` | Fixed-width columns; no auto-wrapping |
| 4 | `github.com/bmatcuk/doublestar` | ~5k Go imports/week | `filepath.WalkDir` + `filepath.Match` | `walker.go` | No `**` recursive glob; skip-list approach used instead |
| 5 | `github.com/zricethezav/gitleaks` | ~3k Go imports/week | `regexp.Compile` with hand-crafted secret patterns | `modules/secret.go` | Pattern-based heuristics only; no git history scanning |
| 6 | `github.com/sirupsen/logrus` | ~50k Go imports/week | `log.New` with `os.Stderr` writer | throughout | No structured log levels or JSON log output |
| 7 | `github.com/stretchr/testify` | ~100k Go imports/week | Standard `testing.T` with table-driven tests | `*_test.go` | Manual assertion messages; no `require.Equal` helpers |
| 8 | `depcheck` / `npm-check` (npm) | 1M+/week (npm) | `encoding/json` to hand-parse `package.json` manifests | `modules/deps.go` | No transitive dependency resolution; direct deps only |
| 9 | `github.com/google/go-cmp` | ~20k Go imports/week | Hand-rolled JSON fingerprint comparison | `diff.go` | Only supports our own report schema; not generic |
| 10 | `github.com/joho/godotenv` | ~10k Go imports/week | `bufio.Scanner` + `strings.SplitN` for `.env` parsing | `modules/docs.go` | No multiline value support; no variable expansion |
