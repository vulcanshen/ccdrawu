#!/bin/sh
# statusdrawer's status line bridge. Claude Code runs the status line again on each new message,
# vim mode switch and permission mode change, handing it JSON on stdin. This draws nothing: it only
# writes vim.mode where the statusdrawer mod reads it, since a mod cannot read vim's live state.
dir="$HOME/.claude/statusdrawer/vim"
mkdir -p "$dir"

jq -r '[.session_id // "", .vim.mode // ""] | @tsv' | {
  IFS="$(printf '\t')" read -r id mode
  [ -n "$id" ] || exit 0
  # Write a temporary file and rename it, so the mod never reads a half-written one
  printf '%s' "$mode" > "$dir/$id.tmp" && mv "$dir/$id.tmp" "$dir/$id"
}

# Clear out sessions untouched for a week
find "$dir" -type f -mtime +7 -delete 2>/dev/null
exit 0
