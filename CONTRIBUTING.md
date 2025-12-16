# Contributing to CLI Shell

First off — thank you for your interest in contributing! Your help makes this project better. This document explains how to contribute in a way that makes review and acceptance smooth.

Contents
- Code of Conduct
- How to file a good issue
- How to propose changes (pull requests)
- Development workflow (fork, branch, commit conventions)
- Tests and quality checks
- Reviewing and CI expectations
- Reporting security issues
- Licensing & copyright

---

## Code of conduct

Please read and follow the project's `CODE_OF_CONDUCT.md`. We expect contributors to be respectful and collaborative. If a `CODE_OF_CONDUCT.md` file is not present, reach out to the maintainers before contributing and treat all interactions as professional and courteous.

---

## Filing a good issue

A useful issue helps maintainers reproduce the problem and decide on the correct fix quickly.

When opening an issue, include:
- A concise, descriptive title.
- A short description of the problem or feature request.
- Steps to reproduce (commands, input, environment).
- Expected behavior vs actual behavior (include error output if any).
- Go version and OS (e.g., `go version`, `uname -a` or Windows details).
- If the issue is a regression, the last known-good version (if known).

For feature requests:
- Explain the motivation and typical usage.
- Provide example commands or APIs.
- If applicable, sketch a minimal implementation approach.

Label issues appropriately if you can (bug, enhancement, question).

---

## Proposing changes (Pull Requests)

We accept changes via pull requests (PRs). Follow these guidelines to increase the chance of a fast review.

1. Fork the repository and create a branch:
   - Branch name convention: `feat/<short-desc>`, `fix/<short-desc>`, `chore/<desc>`, `test/<desc>`, etc.
   - Keep each PR focused on a single change/idea.

2. Keep your fork up-to-date with `main` (or `master`) and rebase or merge frequently.

3. Commit message guidelines:
   - Keep commits small and logically grouped.
   - Use clear messages. Prefer the Conventional Commits style:
     - `feat: add example command`
     - `fix(exec): close pipe readers on exit`
     - `test(edge): add concurrent pipeline stress test`
   - Include a short description in the first line (<=72 chars) and an optional body with motivation and details.

4. Ensure your code compiles and tests pass locally.

5. Run formatters and linters locally before opening a PR:
   - `gofmt -w .`
   - `golangci-lint run ./...`
   - `go vet ./...`

6. Add or update tests for changes:
   - Unit tests should live next to code in `_test.go` files.
   - Keep tests deterministic where possible.
   - If you add heavy/stress tests, mark them clearly (comment or build tag) and explain why they are heavy in the PR description.

7. Update docs:
   - If your change affects user-facing behavior, update `README.md` or `docs/` accordingly.
   - Add examples for public API/CLI behavior if relevant.

8. Open the PR:
   - Use a clear title.
   - In the PR description include:
     - What problem this fixes (link to issue if any).
     - What changed and why.
     - How to test (commands to run).
     - Any known limitations.

9. Respond to review comments and iterate. Keep an eye on CI results and fix failures promptly.

---

## Development workflow & commands

Clone and work locally:
```bash
git clone <repo-url>
cd term3

# run all tests (30s overall timeout)
go test ./... -v -timeout 30s

# run single package tests, e.g. the shell package
go test ./shell -v -run TestName -timeout 30s

# format and vet
gofmt -w .
go vet ./...

# run linters (requires golangci-lint)
golangci-lint run ./...
```

If heavy tests are causing CI fluctuations, either:
- mark them with build tags and exclude them from default CI, or
- ensure CI provides enough resources/time, and document how maintainers can run them locally.

---

## Tests & quality checks

- Every bug fix or new feature should include tests.
- Prefer table-driven tests for variations.
- Keep tests deterministic: avoid relying on timing or external network access.
- Use `t.TempDir()` for temporary files.
- Use `require`/`assert` libraries consistently (this project uses `github.com/stretchr/testify`).
- If you add long-running or resource-heavy tests, note that in PR and consider separating them or gating them behind a selective flag.

CI runs the following checks:
- `golangci-lint run ./...`
- `go vet ./...`
- `go test ./... -v -timeout 30s`

Ensure your PR is green before requesting final review.

---

## Reviewing process

- Maintainers will review PRs and may request changes.
- Please keep PRs small and focused to speed up review.
- Address review comments with follow-up commits; avoid force-pushing multiple times to keep history readable unless requested.
- Squash/rebase may be performed by maintainers before merging to keep history tidy.

---

## Security issues

- Do NOT disclose security vulnerabilities in public issues. Contact maintainers privately (email) or use the platform's private vulnerability reporting if available.
- If you find a secret accidentally committed (e.g., private key), open an issue and coordinate with maintainers to rotate secrets and remove them from history.
- Avoid committing secrets in the first place — use environment variables and secret stores.

---

## License & Contributor License

- This project is licensed under the MIT License (see `LICENSE`).
- By contributing, you affirm that you have the right to grant the project license to your contributions.
- If your organization requires a Contributor License Agreement (CLA), maintainers will notify you.

---

## Other notes & best practices

- Keep external dependencies minimal and vetted for license compatibility.
- Do not commit build artifacts or local IDE configuration (use `.gitignore`).
- Add tests for regressions as bugs are fixed.
- Use meaningful IDs in tests (avoid fragile string matching where unnecessary).
- When in doubt, open an issue to discuss big changes before investing time in a PR.

---

## Contact & additional resources

- If you have questions before contributing, open an issue labelled `question`.
- For urgent or security communication, reach out to repository maintainers (listed in `docs/design.md`) or the project admin.

Thank you for contributing — your improvements and fixes are appreciated!
