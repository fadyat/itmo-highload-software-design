#!/usr/bin/env bash
# Helper to run tests, linters and license checks for the term3 module.
#
# Place this file at: term3/scripts/run_tests.sh
#
# Usage:
#   ./run_tests.sh            -> run lint + tests (default)
#   ./run_tests.sh lint       -> run linters only
#   ./run_tests.sh test       -> run tests only
#   ./run_tests.sh licenses   -> run license checks only (if tool available)
#   ./run_tests.sh all        -> run lint + licenses + tests
#   ./run_tests.sh help       -> show this help
#
# Notes:
# - Tests are run with `go test ./... -v -timeout 30s -count=1`
# - Linter: `golangci-lint` if installed. Script will attempt to install a known-good version
#   if the tool is missing or appears incompatible with the repository config.
# - License check: tries `go-licenses` (github.com/google/go-licenses) if installed;
#   otherwise prints instructions how to install it.
#
# This script is intentionally conservative about environment changes and prints
# actionable messages when helpers are missing.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$(dirname "$0")")" && pwd)"
TIMEOUT="${TEST_TIMEOUT:-30s}"
GOLANGCI="${GOLANGCI:-golangci-lint}"
GOLANGCI_VERSION="${GOLANGCI_VERSION:-v1.59.0}"
GOLICENSES="${GOLICENSES:-go-licenses}"
GO_CMD="${GO_CMD:-go}"
DEFAULT_PKGS="./..."

# Colors (may be disabled in CI)
if [ -t 1 ]; then
  GREEN='\033[0;32m'
  YELLOW='\033[0;33m'
  RED='\033[0;31m'
  BLUE='\033[0;34m'
  NC='\033[0m'
else
  GREEN=''
  YELLOW=''
  RED=''
  BLUE=''
  NC=''
fi

usage() {
  cat <<EOF
Usage: $(basename "$0") [command]

Commands:
  lint        Run golangci-lint (installing it if missing or incompatible)
  test        Run 'go test' for the module (timeout: ${TIMEOUT})
  licenses    Check third-party licenses using 'go-licenses' (if installed)
  all         Run lint -> licenses -> tests
  help        Show this help

Examples:
  # Run linters and tests (default)
  ./run_tests.sh

  # Run only tests
  ./run_tests.sh test

  # Run license check
  ./run_tests.sh licenses
EOF
}

info()  { printf "${BLUE}==>${NC} %s\n" "$*"; }
ok()    { printf "${GREEN}✔%s${NC}\n" "$*"; }
warn()  { printf "${YELLOW}⚠%s${NC}\n" "$*"; }
err()   { printf "${RED}✖%s${NC}\n" "$*"; }

install_golangci() {
  info "Installing golangci-lint ${GOLANGCI_VERSION} to GOPATH/bin using 'go install'"
  # Determine GOPATH; fallback to $HOME/go
  GOPATH_VAL="$(${GO_CMD} env GOPATH 2>/dev/null || echo '')"
  if [ -z "${GOPATH_VAL}" ]; then
    GOPATH_VAL="${HOME}/go"
  fi
  BIN_DIR="${GOPATH_VAL}/bin"
  export PATH="${BIN_DIR}:$PATH"

  # Use `go install` so the binary is built with the active toolchain and placed in $GOPATH/bin
  set +e
  ${GO_CMD} install github.com/golangci/golangci-lint/cmd/golangci-lint@"${GOLANGCI_VERSION}"
  INST_RC=$?
  set -e

  if [ ${INST_RC} -ne 0 ]; then
    err "golangci-lint installation via 'go install' failed (exit ${INST_RC})."
    return ${INST_RC}
  fi

  # Ensure it's available in PATH
  if command -v "${GOLANGCI}" >/dev/null 2>&1; then
    ok "golangci-lint installed: $(${GOLANGCI} --version || true)"
    return 0
  else
    err "golangci-lint installed but binary not found in PATH (${BIN_DIR})."
    return 2
  fi
}

# Run golangci-lint if available, or install it if missing/incompatible
run_lint() {
  info "Running linters"

  # Ensure go is available before attempting install
  if ! command -v "${GO_CMD}" >/dev/null 2>&1; then
    err "Go executable not found. Ensure 'go' is installed and in PATH."
    return 2
  fi

  # If not installed, attempt install
  if ! command -v "${GOLANGCI}" >/dev/null 2>&1; then
    warn "golangci-lint not found. Attempting to install ${GOLANGCI_VERSION}..."
    install_golangci
  fi

  # Try running golangci-lint, capture output and exit status
  info "Executing: ${GOLANGCI} run ./..."
  LOUT=""
  set +e
  LOUT="$(cd "${ROOT_DIR}" && "${GOLANGCI}" run ./... 2>&1)"
  LRC=$?
  set -e

  if [ ${LRC} -eq 0 ]; then
    ok "golangci-lint finished successfully"
    return 0
  fi

  # If we detect an incompatible config error, reinstall a known-good version and retry
  echo "${LOUT}" | grep -qi "unsupported version of the configuration\|unsupported version" && INCOMPAT=1 || INCOMPAT=0

  if [ ${INCOMPAT} -eq 1 ]; then
    warn "Detected an incompatible golangci-lint configuration error. Reinstalling ${GOLANGCI_VERSION} and retrying..."
    if install_golangci; then
      info "Retrying golangci-lint run..."
      set +e
      LOUT="$(cd "${ROOT_DIR}" && "${GOLANGCI}" run ./... 2>&1)"
      LRC=$?
      set -e
      if [ ${LRC} -eq 0 ]; then
        ok "golangci-lint finished successfully after reinstall"
        return 0
      else
        err "golangci-lint still failing after reinstall. Output:"
        echo "${LOUT}"
        return ${LRC}
      fi
    else
      err "Failed to install golangci-lint. See output above."
      return 2
    fi
  fi

  # For other failures, print output and return non-zero
  err "golangci-lint reported issues (exit code ${LRC}). Output:"
  echo "${LOUT}"
  return ${LRC}
}

# Run go test with a global timeout for the test process
run_tests() {
  info "Running go tests for module (packages: ${DEFAULT_PKGS})"
  (cd "${ROOT_DIR}" && "${GO_CMD}" test ${DEFAULT_PKGS} -v -timeout "${TIMEOUT}" -count=1)
  ok "go tests passed"
}

# Try to run go-licenses to enumerate dependency licenses
run_license_check() {
  info "Running license check"

  if command -v "${GOLICENSES}" >/dev/null 2>&1; then
    # Print a CSV of dependencies and licenses. This requires GOPROXY/modules available.
    (cd "${ROOT_DIR}" && "${GOLICENSES}" csv ./... | tee licenses.csv)
    ok "License list written to licenses.csv"
    warn "Please review licenses.csv for compatibility with your project license."
  else
    warn "go-licenses not found in PATH."
    cat <<EOF
Suggested install:
  go install github.com/google/go-licenses@latest

After installation you can run:
  cd ${ROOT_DIR}
  go-licenses csv ./... > licenses.csv
  # or inspect unique licenses:
  go-licenses csv ./... | awk -F, '{print \$2}' | sort | uniq -c | sort -nr
EOF
  fi
}

# Run sanity checks: go is available and module appears present
sanity_check() {
  if ! command -v "${GO_CMD}" >/dev/null 2>&1; then
    err "Go executable not found. Ensure 'go' is installed and in PATH."
    exit 2
  fi

  if [ ! -f "${ROOT_DIR}/go.mod" ]; then
    warn "go.mod not found in ${ROOT_DIR}. Are you in the right module?"
  fi
}

main() {
  cmd="${1:-all}"

  sanity_check

  case "${cmd}" in
    ""|all)
      run_lint
      run_license_check
      run_tests
      ;;
    lint)
      run_lint
      ;;
    test)
      run_tests
      ;;
    licenses)
      run_license_check
      ;;
    help|-h|--help)
      usage
      ;;
    *)
      echo "Unknown command: ${cmd}"
      usage
      exit 2
      ;;
  esac
}

# If script runs directly, execute main with arguments
if [ "${BASH_SOURCE[0]}" = "$0" ]; then
  main "$@"
fi
