# Changelog

All notable changes to this project will be documented in this file.

The format is based on "Keep a Changelog" and follows Semantic Versioning.

## [Unreleased]

### Added
- Comprehensive CI workflow (GitHub Actions) with linting and tests.
- `.golangci.yml` configuration for static analysis and linters.
- `README.md` with project overview, build/test/run instructions and badges.
- `LICENSE` (MIT) in repository root.
- `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `CODEOWNERS` and issue/PR templates under `.github/`.
- Helper scripts under `scripts/` (test/lint/license helpers).
- `.gitattributes` and `.gitattributes` entries to normalize line endings and binary handling.
- `CHANGELOG.md` (this file).
- New edge and stress tests for pipeline/assignment/exit behavior: `term3/shell/edge_test.go`.

### Changed
- Fixed deadlock in pipeline execution when a builtin `exit` appears in the middle of a pipeline:
  - Improved handling of pipe readers/writers closure to ensure upstream writers do not block.
  - Adjusted `ExecutePipeline` result collection and pipe cleanup to avoid goroutine hangs.
- Adjusted `TestAssignmentThenEcho` expectation to match implemented semantics (command-local assignments do not mutate the global environment).
- Reformatted and ran `gofmt` across the codebase.

### Removed
- N/A

### Notes
- Tests are run with an overall timeout recommended at 30s for CI: `go test ./... -v -timeout 30s`.
- Heavy/stress tests have been added (concurrency/large-input). If needed, they can be gated with build tags or run selectively via `-run` to avoid CI resource issues.

---

## [0.1.0] - Initial implementation

### Added
- Basic shell interpreter core:
  - Lexer: tokenization with single/double quotes, recognition of `|` and `=`.
  - Parser: pipeline and command structure with assignments and args.
  - Executor: builtin commands (`echo`, `cat`, `wc`, `pwd`, `exit`) and support for piping and launching external commands.
  - Environment management: `Env` with cloning and substitution behavior.
- Unit tests covering lexing, parsing, builtins, pipeline flow and variable substitution.
- Project layout:
  - `cmd/cli/` - small CLI entrypoint
  - `shell/` - core implementation and tests
  - `docs/design.md` - architecture and design notes
- `go.mod` and dependency on `github.com/stretchr/testify` for tests.

---

## How to use this changelog
- New releases should be added as separate sections using `## [x.y.z] - YYYY-MM-DD`.
- Move entries from `[Unreleased]` into a new release section when publishing a release.
- Keep entries concise and grouped by categories (`Added`, `Changed`, `Fixed`, `Removed`, `Security`).

---

## Authors & Contributors
- Primary authors listed in `docs/design.md` (repository maintainers).
- Contributions are welcome — see `CONTRIBUTING.md` for details on the workflow and expectations.
