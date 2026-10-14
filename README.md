# ccdrawu

[![License](https://img.shields.io/github/license/vulcanshen/ccdrawu)](LICENSE)

[繁體中文](README-zh_TW.md)

**A sidebar for Claude Code.** Usage limits, context, model, cost, git and session stats sit in a pane beside the transcript, so they no longer crowd the space under your prompt.

```
Model    claude-opus-5-5
Effort   xhigh
Mode     auto
Context  ▰▰▰▰▰▰▰▰▱▱▱▱▱▱  65%
Session  ▰▱▱▱▱▱▱▱▱▱▱▱▱▱   6%
         resets 2h12m
Weekly   ▰▰▰▱▱▱▱▱▱▱▱▱▱▱  23%
         resets 3d17h
Cost     $35.46
Uptime   11h39m · 43 turns
Cache    100%
Compact  0
Tasks    none
Git      main  +12 -3
Path     ~/projects/app
Claude   2.1.299  ↑ 2.1.300
ccdrawu  0.1.0
```

## What it shows

| Row | What it is |
|---|---|
| Model, Effort | The session's model, and the last request's effort level: green for `low` through red for `max` |
| Mode | The permission mode: `auto`, `plan`, `accept edits`, … |
| Context, Session, Weekly | How full the context window is, and your 5-hour and weekly plan limits with their reset countdowns. Bars turn green, yellow, peach, then red at 50%, 70% and 90% |
| Cost | What the session has cost so far |
| Uptime | How long the session has run, and how many prompts you sent |
| Cache | How much of the last request the prompt cache served |
| Compact | How many times the conversation was compacted |
| Tasks | Subagents running now, and the shells, monitors and workflows still running in the background when Claude last replied, a count per type: `2 agents｜1 shell` |
| Git, Path | The branch with its added and removed lines, and the working directory |
| Claude, ccdrawu | Your Claude Code version, with the newest release when one is out, and this drawer's version |

## Requirements

- Claude Code 2.1.287 or later
- Fullscreen rendering: run `/tui fullscreen` once. Without it the pane sits above the prompt instead of beside the transcript
- A terminal wide enough for a sidebar: a pane opened on its own waits until the terminal is about 110–144 columns wide
- The `ccdrawu` command, to pick what the drawer shows; it needs a [Nerd Font](https://www.nerdfonts.com) in your terminal
- macOS, Linux or Windows

## Install

At the Claude Code prompt:

```
/plugin install ccdrawu --marketplace vulcanshen/ccdrawu
```

Answer `y` to add the marketplace, then choose the user scope to have the drawer in every session.

Then install the `ccdrawu` command:

```sh
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/vulcanshen/ccdrawu/main/install.sh | sh

# Homebrew (macOS / Linux)
brew install vulcanshen/tap/ccdrawu
```

```powershell
# Windows PowerShell
irm https://raw.githubusercontent.com/vulcanshen/ccdrawu/main/install.ps1 | iex

# Scoop (Windows)
scoop bucket add vulcanshen https://github.com/vulcanshen/scoop-bucket
scoop install ccdrawu
```

## Usage

- The drawer opens by itself when a session starts.
- `/ccdrawu` opens or closes it, even while Claude is working.
- Click the `×` at its top right to close it, or press `Ctrl+X Tab` to focus it and then `Ctrl+X X`.
- `Ctrl+X ←` and `Ctrl+X →` make it wider or narrower while it has focus.

### A shorter command

Run `ccdrawu`, move to **alias** under Settings, press `Enter` twice, and type a name such as `ccd`. Within a second `/ccd` works as well as `/ccdrawu`, in every open session. A name you change or remove keeps working in a session until you restart it.

The alias is kept in `~/.config/ccdrawu/config.json` as `"alias": "ccd"`, so you can also set it there.

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

- `~/.config/ccdrawu/config.json`: what you picked in `ccdrawu`, the items and the alias. `CCDRAWU__CONFIG` names another folder for it, and so does `XDG_CONFIG_HOME`.
- `~/.ccdrawu/`: what the plugin writes, the list of items `ccdrawu` offers (`catalog.json`). `CCDRAWU__DATA` names another folder for it.

On Windows, `~` is your user folder (`%USERPROFILE%`).

Every 6 hours the drawer asks the npm registry for Claude Code's newest version.

## Limitations

- After you switch the permission mode with `shift+tab`, **Mode** changes once Claude finishes the tools of the step it is on; while Claude is idle, it changes when you send the next prompt.
- The colors are Catppuccin Mocha and cannot be changed yet.
- Where an organization manages Claude Code, **Mode** and **Compact** may stay empty, and **Context** catches up with a compaction only after the next reply: those events skip plugins the user installed.
- Claude Code's mod interface is in early access and may change between releases.

## Uninstall

```
claude plugin uninstall ccdrawu@ccdrawu
```

Then remove the `ccdrawu` command:

```sh
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/vulcanshen/ccdrawu/main/uninstall.sh | sh
```

```powershell
# Windows PowerShell
irm https://raw.githubusercontent.com/vulcanshen/ccdrawu/main/uninstall.ps1 | iex
```

Both ask before removing `~/.config/ccdrawu/` and `~/.ccdrawu/`. With Homebrew or Scoop, `brew uninstall ccdrawu` or `scoop uninstall ccdrawu`.

## More

- [Developer notes](docs/dev-remarks.md)
- [License: GPL-3.0](LICENSE)
