# ccdrawu

[![License](https://img.shields.io/github/license/vulcanshen/ccdrawu)](LICENSE)

[繁體中文](README-zh_TW.md)

**A sidebar for Claude Code.** Usage limits, context, model, cost, git and session stats sit in a pane beside the transcript, so they no longer crowd the space under your prompt.

```
Model    claude-opus-5-5
Thinking xhigh
Mode     auto
Editor   vim · INSERT

Context  ████████░░░░░░  65%
Session  █░░░░░░░░░░░░░   6%
         resets 2h12m
Weekly   ███░░░░░░░░░░░  23%
         resets 3d17h
Cost     $35.46

Uptime   11h39m · 43 turns
Cache    100%
Compact  0
Agents   none

Git      main  +12 -3
Path     ~/projects/app

Version  2.1.299  ↑ 2.1.300
```

## What it shows

| Row | What it is |
|---|---|
| Model, Thinking | The session's model and its effort level |
| Mode | The permission mode: `auto`, `plan`, `accept edits`, … |
| Editor | `vim` or `normal`; with vim, its live state when the [status line bridge](#live-vim-mode-optional) is on |
| Context, Session, Weekly | How full the context window is, and your 5-hour and weekly plan limits with their reset countdowns. Bars turn green, yellow, peach, then red at 50%, 70% and 90% |
| Cost | What the session has cost so far |
| Uptime | How long the session has run, and how many prompts you sent |
| Cache | How much of the last request the prompt cache served |
| Compact | How many times the conversation was compacted |
| Agents | Subagents running now |
| Git, Path | The branch with its added and removed lines, and the working directory |
| Version | Your Claude Code version, with the newest release when one is out |

## Requirements

- Claude Code 2.1.287 or later
- Fullscreen rendering: run `/tui fullscreen` once. Without it the pane sits above the prompt instead of beside the transcript
- A terminal wide enough for a sidebar: a pane opened on its own waits until the terminal is about 110–144 columns wide
- `jq`, only for the live vim mode

## Install

At the Claude Code prompt:

```
/plugin install ccdrawu --marketplace vulcanshen/ccdrawu
```

Answer `y` to add the marketplace, then choose the user scope to have the drawer in every session.

## Usage

- The drawer opens by itself when a session starts.
- `/ccdrawu` opens or closes it, even while Claude is working.
- Click the `×` at its top right to close it, or press `Ctrl+X Tab` to focus it and then `Ctrl+X X`.
- `Ctrl+X ←` and `Ctrl+X →` make it wider or narrower while it has focus.

### A shorter command

`/config` has a **Command alias** row for ccdrawu. Type a name such as `ccd`, and `/ccd` works as well as `/ccdrawu`.

### Live vim mode (optional)

A plugin cannot read whether vim is in INSERT or NORMAL mode, but a status line command can. ccdrawu ships a small one that draws nothing and passes the mode on. Point your status line at it in `~/.claude/settings.json`:

```json
"statusLine": {
  "type": "command",
  "command": "sh ~/.claude/ccdrawu/statusline.sh"
}
```

This replaces any status line you have now. It prints nothing, but Claude Code keeps an empty line where the status line was.

### Match the background

The drawer's background comes from your Claude Code theme, not from the plugin. To make it the same color as your terminal, create a custom theme in `~/.claude/themes/`, for example `~/.claude/themes/my-theme.json`:

```json
{
  "name": "My theme",
  "base": "dark-ansi",
  "overrides": { "composerSidebarBackground": "#11111b" }
}
```

Then pick it with `/theme`. Use your current theme as `base` and your terminal's background color as the value.

Inside tmux, Claude Code limits itself to 256 colors, so a color like `#11111b` comes out as a different one. If your tmux passes truecolor through, add `"CLAUDE_CODE_TMUX_TRUECOLOR": "1"` to the `env` block of `~/.claude/settings.json` and restart Claude Code.

## Where it keeps things

`~/.claude/ccdrawu/` holds the status line bridge (`statusline.sh`, copied there by the plugin) and one small file per session with its vim mode (`vim/`). Files untouched for a week are removed.

Every 6 hours the drawer asks the npm registry for Claude Code's newest version.

## Limitations

- After you switch the permission mode with `shift+tab`, **Mode** changes when you send the next prompt or Claude's reply ends.
- The colors are Catppuccin Mocha and cannot be changed yet.
- Where an organization manages Claude Code, **Mode** and **Compact** may stay empty: those events skip plugins the user installed.
- Claude Code's mod interface is in early access and may change between releases.

## Uninstall

```
claude plugin uninstall ccdrawu@ccdrawu
```

Then remove the `statusLine` entry if you added it, and the `~/.claude/ccdrawu/` folder.

## More

- [Developer notes](docs/dev-remarks.md)
- [License: GPL-3.0](LICENSE)
