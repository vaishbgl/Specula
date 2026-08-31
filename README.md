# Specula

> **The engineering-readiness gate for your release.** One command, six
> checkpoints: leaked secrets, dependency bloat, weak test coverage, code
> rot, incomplete docs, and missing CI. Specula explains every point deducted,
> so a fix is never a guessing game.
>
> **Zero Dependency Hackathon · Track A (Developer Tools & CLI)** — built
> 100% with the Go standard library. No `go get`, no network calls, no config.

---

## TL;DR for Judges

| Question | Answer |
|---|---|
| What is it? | A CLI linter that grades a project's release-readiness (A–F) |
| Why does it matter? | Catches the problems code review misses — a leaked key, a stub test, a missing README — before they ship |
| Dependencies | **Zero.** `go list -m all` prints one line: this module |
| Build | `make build` on a fresh clone, Go 1.22+, nothing else |
| Quality | 63 tests, race-detector clean, `go vet` clean, deterministic output |
| Demo | `specula demo` — built-in broken sample project, no setup needed |
| Dogfooding | Run `specula lint .` on this repo: **A (98/100), 0 failures** |

## The Problem

Before a release, every team asks the same questions: *did we leak a key? is
the test suite real? does every dependency earn its place? does the README
actually tell a newcomer how to run this?* Today those checks are scattered
across different tools — or skipped entirely until a leaked key or a broken
release forces the conversation.

Specula bundles six checks into one deterministic pass, prints a per-rule
report you can screenshot in a PR, and exits non-zero in CI when a project
drops below the bar. There is no config file to maintain, no plugin to
install, and no network access — ever.

## Demo

> *Demo video link will be added before submission.*
>
> Until then, `specula demo` needs zero setup: it writes a broken sample
> project to a temp dir (hardcoded API key, 5 dependencies, no tests, 8-word
> README) and shows every problem it finds:

```
$ specula demo

  ⚠️  Running in DEMO mode — scanning built-in sample project

  ╔══════════════════════════════════════════╗
  ║   ⬡  SPECULA  ·  Project Lint Report    ║
  ╚══════════════════════════════════════════╝

  Grade: 🔴 B (83/100)
  Scanned: 3 files · 0ms · 0 network calls

  ┌──────────────────────┬───────┬─────────┬──────────────────────────────────────┐
  │ Module               │ Score │ Status  │ Top Finding                          │
  ├──────────────────────┼───────┼─────────┼──────────────────────────────────────┤
  │ Secret Scanner       │   75  │ ❌ FAIL  │ Possible OpenAI API Key detected     │
  │ Dependency Audit     │   97  │ ✅ PASS  │ 5 direct dependencies in go.mod      │
  │ Test Readiness       │   75  │ ❌ FAIL  │ No test files found (1 source fil... │
  │ Code & Context Lint  │   97  │ ✅ PASS  │ 2 TODO/FIXME/HACK/XXX markers acr... │
  │ Docs & Env Check     │   71  │ ⚠️  WARN │ README has only 8 words (minimum:... │
  │ Project Setup        │   84  │ ✅ PASS  │ No CI configuration found            │
  └──────────────────────┴───────┴─────────┴──────────────────────────────────────┘

  Findings: 2 FAIL · 3 WARN · 7 INFO

  ┌─ Top Actionable Fixes ──────────────────────────────────────────┐
  │  1. 🔴 HIGH  Possible OpenAI API Key detected (main.go)           │
  │  2. 🔴 HIGH  No test files found (1 source files detected)        │
  │  3. 🟡 MED   README has only 8 words (minimum: 100) (README.md)   │
  └──────────────────────────────────────────────────────────────────┘
```

## Quick Start

**Prerequisites:** Go 1.22+ — no other dependencies.

```bash
git clone https://github.com/vaishbgl/specula.git
cd specula
make build            # produces ./specula
./specula lint .      # grades this repo: A (98/100)
```

## Usage

```bash
# Grade the current directory
specula lint .

# Grade with JSON output, fail CI if below 80
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
| `--json` | `false` | Output structured JSON instead of the terminal table |
| `--output <path>` | (none) | Save report to a file (auto-detects JSON by `.json` extension) |
| `--fail-under <n>` | `0` | Exit 1 if the overall score is below `n` (0–100) |
| `--module <name>` | (all) | Run one module: `security`, `deps`, `tests`, `code`, `docs`, `setup` |
| `--no-color` | `false` | Disable ANSI color (also respects `NO_COLOR`) |
| `--verbose` | `false` | Debug diagnostics on stderr |
| `-v` / `--version` | | Print version and exit |
| `-h` / `--help` | | Print usage and exit |

## Dogfooding

Specula grades its own repository. Actual output of `specula lint --no-color .`
run on this repo at submission time:

```
  ╔══════════════════════════════════════════╗
  ║   ⬡  SPECULA  ·  Project Lint Report    ║
  ╚══════════════════════════════════════════╝

  Grade: 🟢 A (98/100)
  Scanned: 32 files · 1ms · 0 network calls

  ┌──────────────────────┬───────┬─────────┬──────────────────────────────────────┐
  │ Module               │ Score │ Status  │ Top Finding                          │
  ├──────────────────────┼───────┼─────────┼──────────────────────────────────────┤
  │ Secret Scanner       │  100  │ ✅ PASS  │ No issues                            │
  │ Dependency Audit     │  100  │ ✅ PASS  │ No issues                            │
  │ Test Readiness       │   97  │ ✅ PASS  │ Test ratio: 5 tests / 15 sources ... │
  │ Code & Context Lint  │   94  │ ✅ PASS  │ File has 587 lines (threshold: 500)  │
  │ Docs & Env Check     │  100  │ ✅ PASS  │ No issues                            │
  │ Project Setup        │   97  │ ✅ PASS  │ CI configuration detected            │
  └──────────────────────┴───────┴─────────┴──────────────────────────────────────┘

  Findings: 0 FAIL · 0 WARN · 4 INFO
```

Deliberately honest: the two INFO deductions are its own documented rules
(test-file ratio reporting and a 587-line test suite file) — that's the
system working as designed.

## What Each Module Checks

| Module | Catches |
|---|---|
| **Secret Scanner** (25%) | AWS keys, GitHub/OpenAI/Slack tokens, private key headers, Bearer tokens, plaintext `API_KEY=` assignments, sensitive files (`.env`, `id_rsa`, …). Findings report a SHA-256 fingerprint, never the raw secret. |
| **Dependency Audit** (20%) | Parses `go.mod`, `package.json`, `requirements.txt`. Flags bloat (with severity tiers), unpinned versions, and 12 ready-made zero-dependency replacements (`chalk`→ANSI, `dotenv`→`os.Getenv`, …). |
| **Test Readiness** (20%) | Test files per ecosystem (Go, JS/TS, Python, Java/Kotlin, Rust), test-to-source ratio, empty stub tests. |
| **Code & Context Lint** (15%) | TODO/FIXME/HACK aggregation, files over 500 lines, byte-identical duplicate files, and AI-context rot (broken file references inside `.cursorrules`/`CLAUDE.md`). |
| **Project Setup** (10%) | CI presence (GitHub Actions, GitLab, CircleCI, Bitbucket), `SECURITY.md`, `CHANGELOG`. |
| **Docs & Env Check** (10%) | Missing/undersized README, missing LICENSE, `.env` ↔ `.env.example` key drift. |

Each module starts at 100. FAIL findings deduct 25 (max −25 per rule), WARN
10 (max −20), INFO 3 (max −9). Modules floor at 25, and the overall score is a
weighted average: **A ≥ 90, B ≥ 75, C ≥ 60, D ≥ 45, F < 45**. Every number in
a report is reproducible from the listed rules — no black box.

## Zero-Dependency Proof

This is the hackathon's core requirement, so here is the entire dependency
manifest:

```bash
go list -m all
# github.com/vaishbgl/specula
```

There is no second line and there has never been one. Every import in the
codebase resolves to a Go standard library package. No vendored code, no
generated wrappers, no `//go:embed` trickery.

- [`deps-proof.txt`](deps-proof.txt) — captured verification output
- [`STDLIB.md`](STDLIB.md) — 10-entry substitution log: the package normally
  used, its adoption stats, the stdlib primitive that replaced it, and an
  honest trade-off for each (e.g. `cobra`→`flag.FlagSet`, `gitleaks`→9
  hand-written regexes, `tablewriter`→hand-rolled box drawing)

## Reproducible Build

`make build-reproducible` produces a byte-identical binary across clean
environments (same OS, architecture, and Go version):

```bash
make build-reproducible
shasum -a 256 specula
962fac2b04c6d11f4a297884600b15b7f4c6d6f4cb78273c17f18f3ae8cc08f8  specula
# rebuild — the hash must be identical
make build-reproducible
shasum -a 256 specula
962fac2b04c6d11f4a297884600b15b7f4c6d6f4cb78273c17f18f3ae8cc08f8  specula
```

Hash above: `darwin/arm64` with Go 1.22.0. Flags:
`CGO_ENABLED=0 -trimpath -buildvcs=false -ldflags="-w -s -X main.version=0.1.0"`.

## Architecture

Data flow: `parse flags → walk files → run modules → compute grade → render output`

```
main.go          Thin dispatch — only file that ever calls os.Exit()
cli.go           Arg parsing via flag.FlagSet, Config struct, per-command help
walker.go        filepath.WalkDir with ignore rules, binary detection, size limits
lint.go          Orchestrator — walk → modules → grade → output
grade.go         Deterministic scoring engine (weighted average, per-rule caps)
output.go        ANSI terminal renderer + JSON serializer (io.Writer injection)
diff.go          Regression gate — fingerprint comparison of two JSON reports
demo.go          Built-in sample project for zero-setup demonstration
modules/
  types.go       Shared types: Finding, Severity, ModuleInput, ModuleResult
  secret.go      Secret pattern detection with SHA-256 fingerprinting
  deps.go        Manifest parsing (go.mod, package.json, requirements.txt)
  tests.go       Cross-language test file detection and ratio analysis
  codequality.go TODO aggregation, duplicate detection, AI-context rot
  docs.go        README/LICENSE checks, .env drift detection
  setup.go       CI workflow and security policy detection
```

Design choices judges may want to know:

- **Determinism** — files are walked in sorted order, deductions are capped
  per rule, and secrets are reported as stable fingerprints. The same tree
  always produces the same report, which is what makes `specula diff` a
  trustworthy regression gate.
- **Exit codes** — `0` pass, `1` on findings/threshold/parse errors. Errors go
  to stderr; reports go to stdout; nothing is ever written to the scanned
  directory.
- **No network, no config, no plugins** — a single static binary is the whole
  product, so it runs the same in a laptop terminal and a CI sandbox.

## CI Integration

```yaml
# .github/workflows/specula.yml
- name: Run Specula
  run: |
    make build
    ./specula lint --fail-under 75 .
```

For an upgrade-blocking gate, store a baseline report and diff against it:

```bash
./specula lint --output baseline.json .
./specula diff baseline.json report.json   # exit 1 on any regression
```

## Testing

```bash
make test           # 63 tests with the race detector
```

Coverage: happy path, empty directories, hidden-directory skipping, binary and
oversized file detection, malformed input, secret matching and fingerprint
stability, dependency parsing, test-ratio math, duplicate detection, context
rot, README/LICENSE/env checks, CI detection, the full grading engine, both
renderers, diff fingerprinting, and end-to-end lint runs including
`--fail-under` enforcement.

## Limitations

- **Secret detection is heuristic.** It scans working-tree files only; it does
  not inspect Git history and cannot guarantee every secret is found. False
  positives are possible on fixture strings.
- **Dependency checks are shallow.** Direct dependencies in common manifest
  files only; no transitive resolution.
- **File scanning is conservative.** Binary files, files > 512 KB (> 10 MB for
  lockfiles), and hidden/noise directories are skipped.
- **No custom rule configuration.** All rules are built-in — there is no
  `.specula.toml` or plugin system.
- **Performance.** Single-threaded; repositories beyond ~10k source files
  expect a few seconds.

## License

MIT — see [LICENSE](LICENSE)
