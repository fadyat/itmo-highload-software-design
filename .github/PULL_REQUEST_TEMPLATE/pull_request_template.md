<!--
Pull Request Template

Please fill out the sections below. Keep the PR focused and small where possible.
-->

## Summary

Short description of what this PR does and why. Link related issue(s) if any.

- Related issue: # (link)
- Type: (bugfix / feature / docs / test / chore / refactor)

## Changes

Describe the main changes introduced by this PR. Bullet points are fine.

- Change 1
- Change 2

## Motivation & Context

Why is this change needed? What problem does it solve? Any background or design notes.

## How Has This Been Tested?

Describe the tests that you ran to verify your changes. Provide instructions so reviewers can reproduce.

- Unit tests: `go test ./... -v -timeout 30s`
- Specific test(s) run: `go test ./shell -run TestName -v`
- Manual testing steps:
  1. Step one
  2. Step two

If you added new tests, ensure they are deterministic and do not rely on external network resources.

## Checklist (replace [ ] with [x] when done)

- [ ] My code follows the project's coding style (gofmt / goimports done)
- [ ] I ran `golangci-lint run ./...` and addressed the reported issues
- [ ] I ran `go vet ./...` and fixed warnings
- [ ] I added unit tests for new/changed functionality
- [ ] All existing and new tests pass: `go test ./... -v -timeout 30s`
- [ ] I updated documentation in `README.md` or `docs/` if applicable
- [ ] I added an entry to CHANGELOG.md (if applicable)

## Types of changes

Choose one or more (remove irrelevant):

- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing behavior to change)
- [ ] Documentation update
- [ ] Tests / CI
- [ ] Chore / Refactor

If this is a breaking change, describe what is breaking and how to migrate.

## Impact on Users / Migration Notes

If this PR changes public behavior, config, or APIs, document migration steps and user-visible effects.

## Screenshots (if applicable)

If this PR changes UI/UX or CLI output, include before/after screenshots or sample output.

## Release notes (optional)

Suggested short release note entry for maintainers:

> Short summary of change suitable for release notes.

## Reviewer Guidance

- Areas of the codebase changed: list packages/files reviewers should focus on.
- Potential risks or edge-cases to pay attention to.
- Any follow-up tasks required after merging.

---

Thank you for your contribution! Please ensure CI is green before requesting a final review.
