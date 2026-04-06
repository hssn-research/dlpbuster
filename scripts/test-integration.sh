#!/usr/bin/env bash
# Integration test — requires network access and a running listener
set -euo pipefail

BINARY="${1:-./bin/dlpbuster}"

if [ ! -f "$BINARY" ]; then
  echo "Build binary first: make build"
  exit 1
fi

echo "=== Integration Test: dlpbuster ==="
echo ""

# Test list command
echo "[1] list channels"
$BINARY list
echo ""

# Test run with webhook (if configured)
echo "[2] run --channel webhook (skips if not configured)"
$BINARY run --channel webhook --timeout 5 || true

echo ""
echo "Integration test complete."
