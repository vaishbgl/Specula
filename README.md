# Specula

> **A zero-dependency project readiness gate and linter — no `go get` required.**

Specula scans your project and produces a deterministic, secret-safe, CI-safe health report.
It tells you exactly why your project scored 72 instead of 82, and what to fix first.

## Installation

```bash
git clone https://github.com/vaishbgl/specula
cd specula
make build
./specula --help
```

Requires **Go 1.22+**. No internet access, no package manager, no config file.

## Usage

```bash
# Lint the current directory
specula lint .

# Lint with JSON output and exit code enforcement
specula lint --json --fail-under 80 .

# Run a single module only
specula lint --module security .

# Compare two reports (regression gate in CI)
specula diff before.json after.json

# Run against built-in demo data (no setup required)
specula demo

# Save report to file
specula lint --output report.json .
```

## What Specula Checks

| Module | What It Finds |
|---|---|
| 🔒 Secret Scanner | AWS keys, GitHub tokens, OpenAI keys, private keys, `.env` files |
| 📦 Dependency Audit | Wildcard versions, dep counts, zero-dep replacement suggestions |
| 🧪 Test Readiness | Missing tests, low test ratio, empty test stubs (Go, JS, Python, Java, Rust) |
| 🔍 Code & Context Lint | TODOs, 500+ LOC files, duplicate files, AI context rot |
| 📝 Docs & Env Check | README completeness, LICENSE, `.env.example` vs `.env` drift |
| ⚙️ Project Setup | CI config, SECURITY.md, CHANGELOG |

## Grading

Each module starts at **100**. Deductions are applied per rule:

| Severity | Deduction | Max per Rule |
|---|---|---|
| FAIL | -25 | -25 |
| WARN | -10 | -20 |
| INFO | -3 | -9 |

Module floor: 25. Overall score: weighted average of applicable modules. Letter grades: A (≥90), B (≥75), C (≥60), D (≥45), F (<45).

## Sample Output

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
```

## Reproducible Build

Build is byte-identical across clean environments:

```bash
make build-reproducible
sha256sum specula
# Verify again — hashes must match
sha256sum specula
```

**Verified hash (both builds):** `8ec63ce513942546c359b18c8c950ee908d58a587375fa4baca50cfb40da591e`

Flags used: `CGO_ENABLED=0 -trimpath -buildvcs=false -ldflags="-w -s -X main.version=0.1.0"`

## CI Integration

```yaml
# .github/workflows/specula.yml
- name: Run Specula
  run: |
    make build
    ./specula lint --fail-under 75 .
```

## Zero Dependencies

```
go list -m all
# github.com/vaishbgl/specula
```

No third-party packages. See [STDLIB.md](STDLIB.md) for the full substitution log.

## Testing

```bash
make test          # Run all tests with race detector
make test-capture  # Run and save to test-results.txt
```

## License

MIT — see [LICENSE](LICENSE)
