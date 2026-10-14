#!/bin/sh
# ccdrawu uninstaller for macOS / Linux (Windows: uninstall.ps1).
# Usage: curl -fsSL https://raw.githubusercontent.com/vulcanshen/ccdrawu/main/uninstall.sh | sh

set -e

CANDIDATES="$HOME/.local/bin/ccdrawu /usr/local/bin/ccdrawu"

FOUND=""
for path in $CANDIDATES; do
  if [ -f "$path" ]; then
    FOUND="$path"
    break
  fi
done

if [ -z "$FOUND" ]; then
  echo "ccdrawu not found in expected locations."
  echo "Checked: $CANDIDATES"
  exit 1
fi

# Take the drawer out of Claude Code while the command is still here; with no
# plugin there (or no Claude Code) it says so and the uninstall goes on.
"$FOUND" uninstall claude || true

rm "$FOUND"
echo "removed $FOUND"

# ask offers to remove one directory: the answer is read from the terminal,
# not stdin — under `curl ... | sh` stdin is the script itself, so a plain
# `read` would consume script text (or hit EOF) instead of the keypress. No
# controlling terminal (cron, nohup) -> keep it, the safe default.
ask() {
  dir="$1"; what="$2"
  [ -d "$dir" ] || return 0
  printf "Remove %s in %s? [y/N]: " "$what" "$dir"
  if read -r answer < /dev/tty 2>/dev/null; then
    case "$answer" in
      y|Y|yes|YES) rm -rf "$dir"; echo "removed $dir" ;;
      *) echo "kept $dir" ;;
    esac
  else
    echo ""
    echo "kept $dir (no terminal to confirm on)"
  fi
}

# What the user picked: the drawer's items and the alias.
if [ -n "$CCDRAWU__CONFIG" ]; then
  CONFIG_DIR="$CCDRAWU__CONFIG"
elif [ -n "$XDG_CONFIG_HOME" ]; then
  CONFIG_DIR="$XDG_CONFIG_HOME/ccdrawu"
else
  CONFIG_DIR="$HOME/.config/ccdrawu"
fi
ask "$CONFIG_DIR" "ccdrawu settings (the drawer's items and the alias)"

# What the plugin writes: the catalog.
ask "${CCDRAWU__DATA:-$HOME/.ccdrawu}" "ccdrawu data (the catalog)"

echo ""
echo "ccdrawu uninstalled."
