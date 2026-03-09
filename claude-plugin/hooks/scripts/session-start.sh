#!/usr/bin/env bash
set -euo pipefail

# Check canary CLI availability
if ! command -v canary &>/dev/null; then
  echo "canary CLI not found on PATH. Install from: https://github.com/devnw/canary"
  echo "Some /canary.* commands will not work without the CLI."
  exit 0
fi

echo "canary is available: $(command -v canary)"

# Print version
echo ""
echo "Version:"
canary --version 2>/dev/null || true

# Check for project configuration
if [ -f ".canary/project.yaml" ]; then
  echo ""
  echo "CANARY project detected."
fi
