#!/usr/bin/env bash
set -euo pipefail

# golangci-lint wrapper script
# Ensures consistent linting behavior across development and CI environments

GOLANGCI_LINT_VERSION="v2.9.0"

# Check if golangci-lint is installed and matches the expected version
NEEDS_INSTALL=false
if ! command -v golangci-lint &>/dev/null; then
    NEEDS_INSTALL=true
elif ! golangci-lint version 2>&1 | grep -q "${GOLANGCI_LINT_VERSION#v}"; then
    echo "golangci-lint version mismatch, upgrading to ${GOLANGCI_LINT_VERSION}..."
    NEEDS_INSTALL=true
fi

if [ "$NEEDS_INSTALL" = true ]; then
    echo "Installing golangci-lint ${GOLANGCI_LINT_VERSION}..."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b "$(go env GOPATH)/bin" "${GOLANGCI_LINT_VERSION}"
fi

# Run golangci-lint with the provided arguments
exec golangci-lint "$@"