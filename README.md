# ccdrawu

[![License](https://img.shields.io/github/license/vulcanshen/ccdrawu)](LICENSE)

[繁體中文](README-zh_TW.md)

**A sidebar for Claude Code.** Usage limits, context, model, cost, git and session stats sit in a pane beside the transcript, so they no longer crowd the space under your prompt.

```
Model    claude-opus-5-5
Effort   xhigh
Mode     auto
Context  ■■■■■■■■■□□□□□  65%
Session  ■□□□□□□□□□□□□□   6%
         resets 2h12m
Weekly   ■■■□□□□□□□□□□□  23%
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
| Pet | A pixel-art pet that keeps you company: see [The pet](#the-pet) |
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

These start off; turn them on in `ccdrawu`:

| Row | What it is |
|---|---|
| Style | The output style set in `/config` |
| ID, Account | The start of this session's id, and the Claude account signed in |
| Window | Tokens in the context window, out of its size: `420k / 1M` |
| Rate | What the session costs an hour so far |
| Tokens, Speed | Input, output and cache tokens of the conversation since the drawer loaded, and the last reply's output tokens a second |
| Skills | Skills run in this session, and the last one |
| Sync, Files, Commit, Repo | Commits ahead of and behind the upstream, files staged, changed and new, the short hash of HEAD, and origin's owner/repo with the worktree when you are in one; `no upstream`, `no commits` or `no origin` when there is none, and `not a repo` outside one |
| Status | [status.claude.com](https://status.claude.com): all clear in green, an incident in peach or red |

## Requirements

- Claude Code 2.1.287 or later
- Fullscreen rendering: run `/tui fullscreen` once. Without it the pane sits above the prompt instead of beside the transcript
- A terminal wide enough for a sidebar: a pane opened on its own waits until the terminal is about 110–144 columns wide
- A [Nerd Font](https://www.nerdfonts.com) in your terminal: the drawer draws its bars with it, and so does the `ccdrawu` command. Without one, set **Bar** to `line` in **preferences** under Settings: each choice there shows its own bar, so you can see which one your terminal draws
- The `ccdrawu` command, to pick what the drawer shows
- macOS, Linux or Windows

## Install

First the `ccdrawu` command:

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

Then add the drawer to Claude Code:

```sh
ccdrawu install claude
```

It installs the ccdrawu plugin for your user, so every session has the drawer. Sessions already open take it with `/reload-plugins`. Then run `ccdrawu` to pick what the drawer shows.

## Usage

- The drawer opens by itself when a session starts.
- `/ccdrawu` opens or closes it, even while Claude is working.
- Click the `×` at its top right to close it, or press `Ctrl+X Tab` to focus it and then `Ctrl+X X`.
- `Ctrl+X ←` and `Ctrl+X →` make it wider or narrower while it has focus.

### Pick what it shows

Run `ccdrawu`, pick **sidebar** (beside the transcript) or **inline** (above the prompt) under Drawer and press `Enter`. The right side shows **Items**, the ones on, and **Order**, the same in the order the drawer shows them. Press `Enter` on **Items** for the list of every item: `Space` checks or unchecks one, `Enter` applies, `Esc` leaves things as they were, and `/` filters as you type. An item you turn on goes after the one before it in that list. Press `Enter` on **Order** to change the order: `D` picks an item up, `j` and `k` move it, `Enter` puts it down, and `Enter` again saves the order (`Esc` puts things back). Open sessions follow within a second.

### The pet

A pixel-art pet lives at the top of the drawer, a cat unless you pick another. Above the prompt it takes the leftmost column, without its lines or the ball.

- While Claude works, it wanders about: it stops, hops, runs and swipes at the air.
- When Claude is done, it sits down and says so. If you stop Claude it just sits; if the reply ends in an error, or a tool fails, it jumps. It has something to say after a compaction too.
- Left alone, it lies down after 2 minutes and falls asleep after 10, and now and then it strolls, looks around or swipes.
- `/ccdrawu pat` pats it: it stops, and a heart shows over it.
- `/ccdrawu ball` throws it a ball, which it runs after and plays with.

What it says comes from a fixed set of lines, so it costs no tokens. To pick the animal (cat, dog, rabbit, duck or penguin), run `ccdrawu` and set **Pet** in **preferences** under Settings; open sessions follow within a second. Its coat is picked at random each session. To hide it, uncheck **Pet** in the list of items.

### A shorter command

Run `ccdrawu`, move to **preferences** under Settings and press `Enter`, then `Enter` on **Alias** and type a name such as `ccd`. Within a second `/ccd` works as well as `/ccdrawu`, in every open session. A name you change or remove keeps working in a session until you restart it.

The alias is kept in `~/.config/ccdrawu/config.json` as `"alias": "ccd"`, so you can also set it there.

### Match the background

The drawer's background comes from your Claude Code theme, not from the plugin. To make it the same color as your terminal, run `ccdrawu` and, in **preferences** under Settings:

- **Claude theme**: pick the built-in Claude Code theme to stand on (`dark-ansi`, `dark`, …).
- **Drawer background**: pick your terminal's background color; `x` on it goes back to the theme's own.

ccdrawu writes its own theme, that built-in one with your background (`~/.claude/themes/ccdrawu.json`), and selects it in Claude Code. Setting Claude theme back to `off`, or `ccdrawu uninstall claude`, selects the theme you had before. Open Claude Code sessions change at once.

Inside tmux, Claude Code limits itself to 256 colors, so a color like `#11111b` comes out as a different one. If your tmux passes truecolor through, add `"CLAUDE_CODE_TMUX_TRUECOLOR": "1"` to the `env` block of `~/.claude/settings.json` and restart Claude Code.

## Where it keeps things

- `~/.config/ccdrawu/config.json`: what you picked in `ccdrawu`: the items, the alias, the bar and the pet. `CCDRAWU__CONFIG` names another folder for it, and so does `XDG_CONFIG_HOME`.
- `~/.ccdrawu/`: what the plugin writes, the list of items `ccdrawu` offers (`catalog.json`). `CCDRAWU__DATA` names another folder for it.

On Windows, `~` is your user folder (`%USERPROFILE%`).

Every 6 hours the drawer asks the npm registry for Claude Code's newest version.

## Limitations

- After you switch the permission mode with `shift+tab`, **Mode** changes once Claude finishes the tools of the step it is on; while Claude is idle, it changes when you send the next prompt.
- The colors are Catppuccin Mocha and cannot be changed yet.
- Where an organization manages Claude Code, **Mode**, **Compact** and **Skills** may stay empty, **Context** catches up with a compaction only after the next reply, and the pet does not jump when a tool fails or speak after a compaction: those events skip plugins the user installed.
- Claude Code's mod interface is in early access and may change between releases.

## Uninstall

First take the drawer out of Claude Code:

```sh
ccdrawu uninstall claude
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
