#!/usr/bin/env bash
# Install CANARY Cursor plugin into ~/.cursor/plugins for Cursor IDE to load it.
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGIN_NAME="canary-requirements"
TARGET="${HOME}/.cursor/plugins/${PLUGIN_NAME}"
mkdir -p "${HOME}/.cursor/plugins"
if [[ -d "$TARGET" ]]; then
  rm -rf "$TARGET"
fi
cp -r "$SCRIPT_DIR" "$TARGET"
echo "Installed CANARY Cursor plugin to $TARGET"
echo "Restart Cursor (or reload window) and enable the plugin in Settings → Plugins."
