#!/usr/bin/env bash
# promptsmith installer — builds the Go binary, installs it on your PATH,
# and scaffolds a config file.
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="${PROMPTSMITH_BIN_DIR:-$HOME/.local/bin}"
CFG_DIR="$HOME/.config/promptsmith"
CFG_FILE="$CFG_DIR/config.json"

command -v go >/dev/null 2>&1 || { echo "error: Go toolchain not found. Install Go >=1.21." >&2; exit 1; }

mkdir -p "$BIN_DIR" "$CFG_DIR"

echo "building promptsmith..."
( cd "$HERE" && go build -o "$BIN_DIR/pps" . )
echo "installed $BIN_DIR/pps"

# Scaffold config if missing
if [[ ! -f "$CFG_FILE" ]]; then
  cat > "$CFG_FILE" <<'JSON'
{
  "provider": "custom",
  "api_shape": "openai-compatible",
  "base_url": "http://localhost:5000/v1",
  "model": "gpt-5.5"
}
JSON
  echo "wrote default config $CFG_FILE"
else
  echo "config already exists: $CFG_FILE (left untouched)"
fi

case ":$PATH:" in
  *":$BIN_DIR:"*) ;;
  *) echo "NOTE: $BIN_DIR is not on your PATH. Add: export PATH=\"$BIN_DIR:\$PATH\"" ;;
esac

echo "done. Try: pps --raw \"write a tweet about cats\""
