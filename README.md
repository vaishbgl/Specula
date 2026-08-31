# Specula

> A zero-dependency project readiness gate and linter — scans your codebase for
> leaked secrets, dependency bloat, missing tests, and documentation gaps, then
> produces an explainable health score with actionable fix suggestions.
> Built for the Zero Dependency Hackathon — standard library only.

## What It Does

Specula answers the question every team lead asks before a release: *"Is this
project actually ready to ship?"*  It scans a project directory and checks six
dimensions of engineering readiness — secrets, dependencies, test coverage,
code quality, documentation, and CI setup — then produces a weighted letter
grade (A–F) with per-rule deductions that are fully transparent and
deterministic.  No network calls, no config file, no package manager.

**Who it's for:** Engineers who want a single `specula lint .` command in CI to
catch the problems that code review often misses — a hardcoded API key, a
missing README section, a test file with zero assertions, a dependency that
could be replaced with 10 lines of stdlib.

## Demo

> *Demo video link will be added before submission.*

## Installation & Build

**Prerequisites:** Go 1.22+ — no other dependencies.

```bash
git clone https://github.com/vaishbgl/specula.git
cd specula
make build        # produces ./specula binary
```

## Usage

```bash
# Lint the current directory
specula lint .

# Lint with JSON output and exit code enforcement
specula lint --json --fail-under 80 .

# Run a single module only
specula lint --module security .

# Compare two reports (regression gate for CI)
specula diff before.json after.json

# Run against built-in demo data (no setup required)
specula demo

# Save report to file
specula lint --output report.json .
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `--json` | `false` | Output structured JSON instead of terminal table |
| `--output <path>` | (none) | Save report to a file (auto-detects JSON by `.json` extension) |
| `--fail-under <n>` | `0` | Exit with code 1 if overall score is below `n` (0–100) |
| `--module <name>` | (all) | Run only one module: `security`, `deps`, `tests`, `code`, `docs`, `setup` |
| `--no-color` | `false` | Disable ANSI color output (also respects `NO_COLOR` env) |
| `--verbose` | `false` | Print debug diagnostics to stderr |
| `-v` / `--version` | | Print version and exit |
| `-h` / `--help` | | Print usage and exit |

## Output

```
  ╔══════════════════════════════════════════╗
  ║   ⬡  SPECULA  ·  Project Lint Report    ║
  ╚══════════════════════════════════════════╝

  Grade: 🟢 A (92/100)
  Scanned: 24 files · 6ms · 0 network calls

  ┌──────────────────────┬───────┬─────────┬──────────────────────────────────────┐
  │ Module               │ Score │ Status  │ Top Finding                          │
  ├──────────────────────┼───────┼─────────┼──────────────────────────────────────┤
  │ Secret Scanner       │  100  │ ✅ PASS │ No issues                            │
  │ Dependency Audit     │  100  │ ✅ PASS │ No issues                            │
  │ Test Readiness       │   97  │ ✅ PASS │ Test ratio: 5 tests / 11 sources ... │
  │ Code & Context Lint  │   97  │ ✅ PASS │ 14 TODO/FIXME markers across 6 fil.. │
  │ Docs & Env Check     │  100  │ ✅ PASS │ No issues                            │
  │ Project Setup        │   75  │ ✅ PASS │ No SECURITY.md found                 │
  └──────────────────────┴───────┴─────────┴──────────────────────────────────────┘

  Findings: 0 FAIL · 1 WARN · 3 INFO

  ┌─ Top Actionable Fixes ──────────────────────────────────────────┐
  │  1. 🟡 MED   No CI configuration found                           │
  │  2. 🟡 MED   No SECURITY.md found                                │
  └──────────────────────────────────────────────────────────────────┘
```

## Architecture

Data flow: `parse flags → walk files → run modules → compute grade → render output`

```
main.go          Thin dispatch — only file that calls os.Exit()
cli.go           Argument parsing via flag.FlagSet, Config struct
walker.go        Recursive filepath.WalkDir with ignore list, binary detection, size limits
lint.go          Orchestrator — walk → modules → grade → output
grade.go         Deterministic scoring engine (weighted average, per-rule deduction caps)
output.go        ANSI terminal renderer + JSON serializer (io.Writer injection)
diff.go          Regression diff gate (fingerprint comparison between two JSON reports)
demo.go          Built-in sample project for zero-setup demonstration
modules/
  types.go       Shared types: Finding, Severity, ModuleInput, ModuleResult
  secret.go      Secret pattern detection with SHA-256 fingerprinting
  deps.go        Dependency manifest parsing (go.mod, package.json, requirements.txt)
  tests.go       Cross-language test file detection and ratio analysis
  codequality.go TODO aggregation, duplicate detection, AI context rot scanning
  docs.go        README/LICENSE checks, .env drift detection
  setup.go       CI workflow and security policy detection
```

## Grading

Each module starts at **100**. Deductions are applied per rule:

| Severity | Deduction | Max per Rule |
|---|---|---|
| FAIL | -25 | -25 |
| WARN | -10 | -20 |
| INFO | -3 | -9 |

Module floor: 25. Overall score is a weighted average of all modules:

| Module | Weight |
|---|---|
| Secret Scanner | 25% |
| Dependency Audit | 20% |
| Test Readiness | 20% |
| Code & Context Lint | 15% |
| Project Setup | 10% |
| Docs & Env Check | 10% |

Letter grades: A (≥90), B (≥75), C (≥60), D (≥45), F (<45).

## Zero-Dependency Proof

```bash
go list -m all
# github.com/vaishbgl/specula
```

See [`deps-proof.txt`](deps-proof.txt) for the full verification output.
See [`STDLIB.md`](STDLIB.md) for the complete substitution log (10 entries).

## Limitations

- **Secret detection is heuristic.** It scans working-tree files only; it does
  not inspect Git history and cannot guarantee every secret is found. False
  positives are possible on test fixture strings.
- **Dependency checks are shallow.** They examine direct dependencies in common
  manifest files only; transitive dependency resolution is not performed.
- **File scanning is conservative.** Binary files (detected by null-byte check),
  oversized files (>512 KB source / >10 MB manifest), and hidden/noise
  directories are automatically skipped.
- **No custom rule configuration.** All rules are built-in. There is no
  `.specula.toml` or plugin system for adding custom checks.
- **Performance.** Scanning is single-threaded. On repositories with >10k
  source files, expect ~2–5s scan time (vs. sub-second for compiled tools with
  parallelism like `golangci-lint`).

## Reproducible Build

Build is byte-identical across clean environments (same OS, architecture, Go version):

```bash
make build-reproducible
shasum -a 256 specula
# Verify again — hashes must match
make build-reproducible
shasum -a 256 specula
```

Flags used: `CGO_ENABLED=0 -trimpath -buildvcs=false -ldflags="-w -s -X main.version=0.1.0"`

## CI Integration

```yaml
# .github/workflows/specula.yml
- name: Run Specula
  run: |
    make build
    ./specula lint --fail-under 75 .
```

## Testing

```bash
make test          # Run all 63 tests with race detector
make test-capture  # Run and save to test-results.txt
```

Tests cover: happy path, empty directories, missing files, hidden directory
skipping, malformed input, binary file detection, secret pattern matching,
dependency parsing, test ratio calculation, duplicate detection, AI context rot,
README/LICENSE checks, CI detection, grading engine, JSON output, end-to-end
integration, and fail-under threshold enforcement.

## License

MIT — see [LICENSE](LICENSE)
