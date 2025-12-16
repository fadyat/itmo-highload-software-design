---
name: Bug report
about: Create a report to help us fix an issue
title: "[BUG] "
labels: bug
assignees: ""
---

**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:

1. Go to '...'
2. Run command `...`
3. Provide input `...`
4. See error

If possible, provide a minimal reproduction case (small command line, input file, or code snippet).

**Expected behavior**
A clear and concise description of what you expected to happen.

**Actual behavior**
What actually happened. Include error messages, stack traces, and unexpected outputs.

**Logs & output**
Paste relevant log lines, error messages, and command output. Use code blocks:

```
# example
$ ./cli run "echo hi"
panic: ...
```

**Environment (please complete the following information):**

- OS: (e.g. macOS 14.1, Ubuntu 22.04)
- Go version: (e.g. go1.25.0)
- Commit/tag (if known): e.g. `main@abcdef`
- How you installed / built (e.g. `go build ./...`, binary, package manager)

**Reproduction files**
If the bug depends on a file, configuration or input, please attach minimal reproduction files or point to a gist.

**Temporary workaround**
If you found a workaround, please describe it.

**Severity**

- [ ] Trivial (typo / minor cosmetic)
- [ ] Low (minor impact)
- [ ] Medium (feature degraded)
- [ ] High (major functionality broken)
- [ ] Critical (data loss / crash / security)

**Additional context**
Add any other context about the problem here (related issues, links, screenshots).

---

<!--
Note: A separate template for feature requests is recommended (feature_request.md).
Below is a feature request template — please create `.github/ISSUE_TEMPLATE/feature_request.md`
with the following content if you want to add a feature request template file.
-->

---

name: Feature request
about: Suggest an idea for this project
title: "[FEATURE] "
labels: enhancement
assignees: ''

---

**Is your feature request related to a problem? Please describe.**
A clear and concise description of the problem this feature would solve.

**Describe the solution you'd like**
Describe the desired behavior and any user-visible changes. Provide examples of commands, configuration, APIs, or output.

**Alternatives considered**
Describe alternatives you've considered and why they are not sufficient.

**Design notes (optional)**
If you have design ideas or sketches (architecture notes, interfaces, packages), include them here. Small diagrams or pseudocode are helpful.

**Prior art / references**
Links to other projects, RFCs or discussions that influenced or inspired the request.

**Potential impact & acceptance criteria**
When would this feature be considered implemented? List tests, compatibility considerations, and how it should behave.

**Additional context**
Any other context or screenshots.
