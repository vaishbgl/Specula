# Standard Library Replacement Log

Specula uses **zero third-party packages**. Every package that would normally be
`go get`'d was replaced with a standard library equivalent. This document records
those substitutions with honest trade-off analysis.

## Substitution Table

| # | Package Normally Used | Weekly Downloads | Replaced By (stdlib) | Implementation | Trade-off |
|:---:|---|---|---|---|---|
| 1 | `github.com/spf13/cobra` | ~30k Go imports/week | `flag.NewFlagSet` + manual `switch` dispatch | Custom subcommand router in [`cli.go`](cli.go) (~140 lines). Supports `lint`, `diff`, `demo` subcommands, each with its own `flag.FlagSet`, shared `Config` struct, and `--help` per subcommand. | No shell completions, no automatic help generation, no middleware. Manual `switch` statement for dispatch. |
| 2 | `github.com/fatih/color` | ~20k Go imports/week | `fmt.Fprintf` with VT100 ANSI escape sequences (`\033[31m`, etc.) | `ColorHelper` struct in [`output.go`](output.go) (~15 lines). Methods: `Bold()`, `Red()`, `Green()`, `Yellow()`, `Cyan()`, `Gray()`, `Dim()`. Respects `--no-color` flag and `NO_COLOR` env. | 8 basic colors only; no 256-color, no truecolor, no Windows terminal detection. |
| 3 | `github.com/olekukonko/tablewriter` | ~15k Go imports/week | `fmt.Fprintf` with Unicode box-drawing characters (`┌─┬─┐`, `│`, `└─┴─┘`) | `renderModuleTable()` in [`output.go`](output.go) (~30 lines). Fixed-column layout with header, separator, and data rows. | Fixed column widths; no auto-sizing, no text wrapping, no alignment options. |
| 4 | `github.com/bmatcuk/doublestar` | ~5k Go imports/week | `filepath.WalkDir` + `strings.HasPrefix` + hardcoded skip list | `Walk()` in [`walker.go`](walker.go) (~80 lines). Built-in skip list covers 15+ common noise dirs (`.git`, `node_modules`, `vendor`, `__pycache__`, `target`, `.venv`, etc.). `filepath.Match` used for extension checks. | No user-configurable glob syntax (`**/*.go`); exclusions are directory-name-based only. |
| 5 | `github.com/zricethezav/gitleaks` | ~3k Go imports/week | `regexp.Compile` with 9 hand-crafted patterns + `crypto/sha256` fingerprinting | `RunSecretScanner()` in [`modules/secret.go`](modules/secret.go) (~140 lines). Detects AWS keys, GitHub/OpenAI/Slack tokens, private keys, Bearer tokens, generic `API_KEY=` assignments. SHA-256 fingerprints enable regression diffing without storing raw secrets. Shannon entropy helper for randomness heuristics. | Pattern-based heuristics only; no git history scanning, no entropy-only detection, no custom rule loading. |
| 6 | `github.com/sirupsen/logrus` | ~50k Go imports/week | `fmt.Fprintf(os.Stderr, ...)` for errors; `--verbose` flag for debug output | Throughout [`main.go`](main.go), [`lint.go`](lint.go). Errors to `stderr`, normal output to `stdout`. `cfg.Verbose` gates debug lines. | No structured log levels (debug/info/warn/error), no JSON log output, no log file rotation. |
| 7 | `github.com/stretchr/testify` | ~100k Go imports/week | `testing.T` with table-driven subtests (`t.Run`) | All `*_test.go` files (~63 tests). Uses `t.Errorf`, `t.Fatalf`, `t.Helper`, `t.TempDir`. | No `require.Equal`/`assert.Contains` helpers; manual assertion messages throughout. |
| 8 | `depcheck` / `npm-check` (npm) | 1M+/week (npm) | `encoding/json.Decoder` + `bufio.Scanner` + `strings.SplitN` | `RunDepsAudit()` in [`modules/deps.go`](modules/deps.go) (~230 lines). Hand-parses `go.mod` require blocks, `package.json` dependency objects, and `requirements.txt` pin formats. Includes 12 zero-dep replacement suggestions (e.g. `chalk` → ANSI, `dotenv` → `os.Getenv`). | No transitive dependency resolution; direct deps only. `package.json` parser is regex-based, not a full JSON tokenizer. |
| 9 | `github.com/google/go-cmp` | ~20k Go imports/week | `encoding/json.Unmarshal` + `fmt.Sprintf` fingerprint comparison | `computeDiff()` in [`diff.go`](diff.go) (~50 lines). Creates fingerprints via `fmt.Sprintf("%s|%s|%d", rule, file, line)` and compares `map[string]DiffFinding` sets. | Only supports our own `LintReportJSON` schema; not a generic deep differ. |
| 10 | `github.com/joho/godotenv` | ~10k Go imports/week | `bufio.Scanner` + `strings.SplitN` + `strings.TrimSpace` | `parseEnvKeys()` in [`modules/docs.go`](modules/docs.go) (~25 lines). Reads `.env` and `.env.example` files line-by-line, skips comments (`#`) and blanks, extracts key names only (values intentionally discarded for security). | No multiline value support, no variable expansion (`${VAR}`), no quote stripping. |

## Test Dependencies

Go has a built-in test framework (`testing` package). **No third-party test
dependencies were used.** All 63 tests use only:
- `testing.T` — assertions and test lifecycle
- `os` — temp file creation for test fixtures
- `path/filepath` — constructing test paths
- `bytes`, `strings`, `fmt`, `encoding/json`, `time` — test data formatting

## Vendored Code

**None.** All code in this repository was written during the hackathon window.
No external source code was copied, vendored, or embedded.
