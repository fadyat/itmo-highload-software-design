# CLI Shell

[![CI](https://img.shields.io/badge/ci-%20not%20configured-lightgrey)](#)
[![Go Report Card](https://goreportcard.com/badge/github.com/fadyat/cli)](#)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Small command-line interpreter implemented in Go. Supports a handful of built-in commands, variable assignments/substitution, and pipelines. The project is intended as an educational exercise and as a compact reference implementation for shell-like behavior.

Supported builtins

- `echo` — print its arguments followed by newline
- `cat` — print file content (or `-` to read stdin)
- `wc` — counts lines, words and bytes
- `pwd` — print working directory
- `exit` — request shell exit with optional code
- `grep` — search files or stdin using regular expressions; supports flags `-w` (whole-word), `-i` (case-insensitive), and `-A` (print N lines after a match)

Features

- Tokenization with single and double quotes
    - single quotes: no substitution
    - double quotes: variable substitution allowed
- Variable assignments: `NAME=value` either as standalone assignment or command-local
- Pipeline support: `cmd1 | cmd2 | cmd3`
- Local pipeline environment snapshotting — assignments inside a pipeline do not mutate the global env (except a single assignment-only pipeline)

Repository layout

- `cmd/cli/` — small CLI entrypoint
- `shell/` — lexer, parser, executor, builtins, and tests
- `docs/` — design notes
- `LICENSE` — MIT license
- (this README) — project overview and instructions

Quick start (development)
Prerequisites

- Go 1.25+ (module-aware)
- git

Build

```sh
# from repository root (the 'term3' directory)
go build ./...
```

Run (CLI)

```sh
# build the CLI binary and run it
go run ./cmd/cli
# or
./your-built-binary
```

Tests
Run the full test suite locally:

```sh
# run all tests with verbose output, single-run, 30s timeout per overall test process
go test ./... -v -timeout 30s -count=1
```

Notes:

- Some edge tests are somewhat heavy (concurrency/stress). Use `-run` to execute specific tests when debugging:
    - e.g. `go test ./shell -run TestExitInMiddleRepeated -v`

Linting and static analysis
We recommend `golangci-lint` or `staticcheck` for quality checks.

Install `golangci-lint`:

```sh
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

Run linters:

```sh
golangci-lint run ./...
# or
go vet ./...
staticcheck ./...
```

CI
A basic GitHub Actions workflow is recommended:

- run `go test ./... -v -timeout 30s`
- run `golangci-lint`
- optionally, check module licenses
  Add a file `.github/workflows/ci.yml` similar to the example in the project templates.

Contribution guide
Thank you for your interest in contributing. Suggested workflow:

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/your-change`
3. Write tests for any new behavior and ensure they pass locally
4. Run linters and fix issues
5. Open a pull request describing the change and motivation

Please keep commits atomic and messages descriptive. If your change is large, prefer multiple small commits or a clean branch history.

Coding style

- Code is formatted with `gofmt` / `go fmt`. Please run `gofmt` before committing.
- Prefer clear, small functions. Add unit tests for complex behavior.
- Keep exported identifiers documented with comments.

License
This project is licensed under the MIT License — see the `LICENSE` file in the repository root.

Third-party dependencies and license compatibility

- Dependencies are listed in `go.mod`.
- Before adding non-trivial third-party libraries, verify license compatibility with MIT (or your chosen project license). Tools like `go-licenses` or `licensee` can help enumerate dependency licenses.

Security and secrets

- Do not commit secrets (private keys, tokens, passwords). Use environment variables or CI secret stores for credentials.
- Consider running secret scanners (`truffleHog`, `detect-secrets`) on repository history if migrating an existing project.

Documentation

- `docs/design.md` contains an architectural overview and design notes.
- Please keep `docs/` updated for significant design or API changes.

Note about argument-parsing library choice

- The `grep` builtin uses a dedicated command-line argument parsing library to correctly handle flags and their values (as required by the assignment). The chosen library is `github.com/jessevdk/go-flags`. It was selected because:
    - it is lightweight and focused on argument parsing (no heavy CLI framework),
    - it supports short and long options and parsing slices of arguments without requiring a full application context,
    - it is well-tested and commonly used in Go projects,
    - its license is permissive and compatible with this project's MIT license.
- The README and `docs/` mention this choice and rationale so reviewers can see why a library was used instead of ad-hoc parsing.

Releases & tags

- Tag stable releases with semantic versioning (e.g., `v0.1.0`) and add GitHub Releases describing changes.
- Include CHANGELOG entries for user-visible changes.

Repository maintenance suggestions

- Add a basic CI workflow (`.github/workflows/ci.yml`) to ensure tests and linters run on PRs.
- Add a `CODE_OF_CONDUCT.md` and `CONTRIBUTING.md` if the project is expected to attract external contributors.
- Populate repository topics and About on GitHub to improve discoverability.

Common troubleshooting

- If tests hang: try running specific test with `-run` to isolate. Use `-v` for verbose logging. Example: `go test ./shell -run TestLargeInputCloseOnExit -v`
- If you hit resource limits on CI, consider marking heavy tests with build tags or splitting them into a separate job.

Contact / Authors

- Authors listed in `docs/design.md`.
- For issues, open GitHub Issues in the repository.

Acknowledgements

- This project uses `github.com/stretchr/testify` for test assertions.
- See `go.mod` for additional indirect dependencies.

Appendix: useful commands

```sh
# format code
gofmt -w .

# run unit tests
go test ./... -v -timeout 30s

# run lint
golangci-lint run ./...

# list modules
go list -m all
```

If you want, I can:

- add a `.github/workflows/ci.yml` workflow file with the suggested CI steps,
- add a `.golangci.yml` configuration and enable a sensible set of linters,
- create `CONTRIBUTING.md` and `CODE_OF_CONDUCT.md` templates,
- add badges for CI and linters once CI is configured.

Tell me which of these you'd like me to add next and I will prepare the files.
