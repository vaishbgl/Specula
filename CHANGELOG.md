# Changelog

All notable changes to Specula are documented in this file.

## [1.0.0] - 2026-08-31

### Added

- `specula lint` — walks a project directory and scores six readiness modules:
  secret scanning, dependency audit, test readiness, code quality, docs & env
  check, and project setup.
- Weighted letter grade (A–F) with transparent, per-rule deduction caps.
- Terminal table renderer with color support and `NO_COLOR` handling.
- JSON report export with `--json` and `--output`.
- CI-friendly `--fail-under` threshold that exits non-zero when the score
  drops below a required bar.
- `specula diff` — regression gate comparing two JSON reports via stable
  fingerprints.
- `specula demo` — built-in sample project for a zero-setup demonstration.
- `make build-reproducible` — byte-identical reproducible build layout.
- 63 tests covering CLI parsing, the directory walker, every lint module, the
  grading engine, both output renderers, the diff gate, and end-to-end runs.
- `.github/workflows/ci.yml` — GitHub Actions job with gofmt, vet, race tests,
  and build verification.

### Changed

- Every feature is implemented exclusively with the Go standard library;
  `go list -m all` reports a single module with zero dependencies.
