#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
staticcheck -f json ./... | sh ../../../ci/scripts/staticcheck-to-gitlab.sh > staticcheck-report.json
