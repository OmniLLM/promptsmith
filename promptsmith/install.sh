#!/usr/bin/env bash
# promptsmith installer — puts the CLI on your PATH and scaffolds a config file.
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="${PROMPTSMITH_BIN_DIR:-$HOME/.local/bin}"
CFG_DIR="$HOME/.config/promptsmith"
CFG_FILE="$CFG_DIR/config.json"

mkdir -p "$BIN_DIR" "$CFG_DIR"

# Symlink the CLI onto PATH
ln -sf "$HERE/promptsmith.py" "$BIN_DIR/promptsmith"
chmod +x "$HERE/promptsmith.py"
echo "linked $BIN_DIR/promptsmith -> $HERE/promptsmith.py"

# Scaffold config if missing
if [[ ! -f "$CFG_FILE" ]]; then
  cat > "$CFG_FILE" <<'JSON'
{
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

echo "done. Try: promptsmith --raw \"write a tweet about cats\""
