# ccdrawu

<p align="center"><img src="docs/icon.svg" width="128" alt="ccdrawu icon" /></p>

[![GitHub Release](https://img.shields.io/github/v/release/vulcanshen/ccdrawu)](https://github.com/vulcanshen/ccdrawu/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/vulcanshen/ccdrawu)](https://go.dev/)
[![License](https://img.shields.io/badge/license-GPL--3.0-blue)](LICENSE)

**Language**: English · [繁體中文](README-zh_TW.md)

**A sidebar and a status line for Claude Code.** Usage limits, context, model, cost, git and session stats sit in a pane beside the transcript, so they no longer crowd the space under your prompt. Or, if you would rather have them under the prompt, in Claude Code's status line, laid out the way you pick.

![demo](docs/demo.gif)

## What it shows

| Row | What it is |
|---|---|
| Pet | A pixel-art pet that keeps you company: see [The pet](#the-pet) |
| Model, Effort | The session's model, and the last request's effort level: green for `low` through red for `max` |
| Mode | The permission mode: `auto`, `plan`, `accept edits`, … |
| Context, Session, Weekly | How full the context window is, and your 5-hour and weekly plan limits with their reset countdowns. Bars turn green, yellow, peach, then red at 50%, 70% and 90% |
| Spend | Where you go through a gateway with a spend limit: how much of it is used, and when it resets |
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
| Directory | The working directory's own name, without the path to it: `ccdrawu` |
| Status | [status.claude.com](https://status.claude.com): all clear in green, an incident in peach or red |

The status line can show any of them but the pet: see [The status line](#the-status-line).

## Requirements

- Claude Code 2.1.287 or later
- Fullscreen rendering, for the drawer: run `/tui fullscreen` once. Without it the pane sits above the prompt instead of beside the transcript (see [Troubleshooting](#troubleshooting)). The status line needs neither this nor a wide terminal
- A terminal wide enough for a sidebar: a pane opened on its own waits until the terminal is about 110–144 columns wide
- A [Nerd Font](https://www.nerdfonts.com) as your terminal's font: the drawer's bars and the marks in the `ccdrawu` command (which choice is picked, which items are ticked) are its icons. Without one, see [Troubleshooting](#troubleshooting)
- The `ccdrawu` command, to pick what the drawer and the status line show. Claude Code runs it for the status line, so it has to be on your `PATH`
- macOS, Linux or Windows (on Windows, tried in Windows Terminal: see [Limitations](#limitations))

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

It installs the ccdrawu plugin for your user, so every session has the drawer. Sessions already open take it with `/reload-plugins`. Then run `ccdrawu` to pick what the drawer shows, or to turn on the status line.

You can also add only the plugin, from inside Claude Code: `/plugin install ccdrawu --marketplace vulcanshen/ccdrawu`. This is not recommended: the drawer then shows its default rows, and picking the rows, their order, the pet, the bar or the theme takes the `ccdrawu` command, or editing `~/.config/ccdrawu/config.json` by hand. The status line needs the `ccdrawu` command too.

## Usage

- The drawer opens by itself when a session starts.
- **Drawer**, atop the sidebar page and the inline page in `ccdrawu`, turns it off in that place alone. Off on the inline page, a terminal narrowed below the sidebar's width hides the drawer instead of moving it above the prompt, and widening it again brings the drawer back beside the transcript. Turned back on while the terminal is narrow, the drawer does not come back above the prompt by itself: see [Limitations](#limitations). Off on both, sessions start without it. Open sessions follow within a second.
- `/ccdrawu` opens or closes it, even while Claude is working, and even where **Drawer** is off: it stays where it opens until it moves.
- Click the `×` at its top right to close it, or press `Ctrl+X Tab` to focus it and then `Ctrl+X X`.
- `Ctrl+X ←` and `Ctrl+X →` make it wider or narrower while it has focus.

### Pick what it shows

Run `ccdrawu`, pick **sidebar** (beside the transcript) or **inline** (above the prompt) under Drawer and press `Enter`. The right side shows **Items**, the ones on, and **Order**, the same in the order the drawer shows them. Press `Enter` on **Items** for the list of every item: `Space` checks or unchecks one, `Enter` applies, `Esc` leaves things as they were, and `/` filters as you type. An item you turn on goes after the one before it in that list. Press `Enter` on **Order** to change the order: `D` picks an item up, `j` and `k` move it, `Enter` puts it down, and `Enter` again saves the order (`Esc` puts things back). Open sessions follow within a second.

![settings](docs/settings.gif)

### The status line

ccdrawu can also be Claude Code's status line, the row under the prompt: with the drawer, or on its own with **Drawer** set to `off`.

Run `ccdrawu`, pick **statusline** under Statusline and press `Enter`, then set **Statusline** to `on`. ccdrawu sets `ccdrawu statusline` as Claude Code's status line in `~/.claude/settings.json`. If you already have one, such as ccstatusline, it asks before replacing it and keeps it: setting **Statusline** back to `off`, or `ccdrawu uninstall claude`, puts yours back. Open sessions change at once.

Set it to `on when narrow` instead to have it only while the terminal is too narrow for the sidebar, under about 144 columns, and the drawer is not beside the transcript. Otherwise the row under the prompt goes away. With **Drawer** off on the inline page too, a narrow terminal shows the status line in place of the sidebar.

```
Model claude-opus-5-5 | Context 42% | Session 18% · 2h31m | Weekly 7% · 3d4h | Cost $1.20 | Git main  +12 -3
```

- **Items** and **Order**: what it shows, and in which order, picked as for the drawer. Model, Context, Session, Weekly, Cost and Git to start with; the pet stays in the drawer.
- **Format**: how each item reads: `label` (`Context 42%`), `short` (`ctx 42%`) or `value` (`42%`).
- **Color**: `highlighted`, each value in the drawer's colors, or `dimmed`, all of it in Claude Code's dim color.
- **Separator**: what stands between the items, `|` unless you set another.
- **Refresh**: how many seconds Claude Code waits before running it again while nothing happens, 10 unless you set another; it also runs it as the conversation moves on.

A line too long for the terminal is cut at the end with `…`.

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

- `~/.config/ccdrawu/config.json`: what you picked in `ccdrawu`: the items, the alias, the bar, the pet, whether the drawer opens by itself, and the status line's settings. `CCDRAWU__CONFIG` names another folder for it, and so does `XDG_CONFIG_HOME`.
- `~/.ccdrawu/`: what the plugin writes: the list of items `ccdrawu` offers (`catalog.json`) and, while the status line is on, each session's values for it (`sessions/`; a session's file is removed after a week without use). `CCDRAWU__DATA` names another folder for it.

On Windows, `~` is your user folder (`%USERPROFILE%`).

Every 6 hours the drawer asks the npm registry for Claude Code's newest version. With **Status** on, it also asks status.claude.com every 5 minutes; it makes no other requests.

## Troubleshooting

### The drawer sits above the prompt

Claude Code decides where the drawer goes, not ccdrawu: beside the transcript only in fullscreen rendering and in a terminal about 110–144 columns wide or more, above the prompt otherwise.

- Run `/tui fullscreen`. Claude Code restarts, picks your conversation up where it was, and keeps the choice for later sessions. `/tui default` goes back.
- If `/tui fullscreen` says Claude Code already uses it but the drawer still sits above the prompt, quit Claude Code and start it again: `claude -c` continues the conversation.
- Widen the terminal.

### Bars or marks show as boxes or `�`

Or the bars run into each other. The terminal's font is not a Nerd Font, so the terminal borrows those icons from another font, or can't draw them at all. Install one from [nerdfonts.com](https://www.nerdfonts.com) and pick it as your terminal's font; in Windows Terminal that is Settings → your profile → Appearance → Font face.

A Nerd Font can also lack a few of these icons. Maple Mono NF CN has no mark for a choice that is not picked, so in the `ccdrawu` command those show as `�` while the picked one shows. JetBrains Mono Nerd Font has been tried and draws them all.

Without a Nerd Font, set **Bar** to `line` in **preferences** under Settings: each choice there shows its own bar, so you can see which one your terminal draws. The marks in the `ccdrawu` command have no such setting; there, only the color tells which choice is picked.

### Rows wrap, with pieces at the left edge of the screen

The sidebar's rows with bars wrap and push the rest out of place, and the `ccdrawu` command leaves parts of earlier screens behind. Seen in Termius's local terminal on Windows: use Windows Terminal there.

### The sidebar is light gray

And its text is hard to read. Claude Code paints the sidebar from its theme, a light theme paints it light gray, and ccdrawu's colors are made for a dark background. Pick a dark theme with `/theme`, or set **Claude theme** in **preferences** under Settings, which can also give the sidebar your terminal's background: see [Match the background](#match-the-background).

### The sidebar's background is not the terminal's

See [Match the background](#match-the-background).

### The status line is an empty row

Claude Code runs `ccdrawu statusline`, and the `ccdrawu` it finds on your `PATH` is an older one, from before the status line. `ccdrawu --version` says which one runs; update it the way you installed it, or remove the older one.

### The status line shows only Model, Context and Cost

The ccdrawu plugin is not running in that session, so the status line shows what Claude Code hands it. Install the plugin with `ccdrawu install claude`; a session started before it takes it with `/reload-plugins`.

## Limitations

- With fullscreen rendering, Claude Code puts the drawer above the prompt only when you open it with `/ccdrawu`, or when the terminal narrows while it sits beside the transcript. So if you turn **Drawer** on the inline page back on while the terminal is too narrow for the sidebar, the drawer stays out of sight: open it with `/ccdrawu`, or widen the terminal and narrow it again. Open sessions say so in a toast when it happens.
- After you switch the permission mode with `shift+tab`, **Mode** changes once Claude finishes the tools of the step it is on; while Claude is idle, it changes when you send the next prompt.
- The status line can be a few seconds behind the drawer: Claude Code runs it again as the conversation moves on, or after **Refresh** seconds.
- The colors are Catppuccin Mocha, made for a dark background, and cannot be changed yet: on a light terminal the text is hard to read. Under a light Claude Code theme, see [Troubleshooting](#troubleshooting).
- Where an organization manages Claude Code, **Mode**, **Compact** and **Skills** may stay empty, **Context** catches up with a compaction only after the next reply, and the pet does not jump when a tool fails or speak after a compaction: those events skip plugins the user installed.
- Claude Code's mod interface is in early access and may change between releases.
- On Windows, ccdrawu has been tried in Windows Terminal. Termius's local terminal draws it out of place (see [Troubleshooting](#troubleshooting)), and other Windows terminals have not been tried yet; reports are welcome.

## Update

The `ccdrawu` command and the plugin in Claude Code update apart, and go together: a newer command with an older plugin can leave a setting without effect.

1. Update the command the way you installed it: `brew upgrade ccdrawu`, `scoop update ccdrawu`, or the install script again.
2. Update the plugin:

   ```sh
   ccdrawu install claude
   ```

   With the plugin there already, this updates it. Claude Code does not update it by itself.
3. In sessions already open, run `/reload-plugins`. New sessions have the new plugin from the start.

## Uninstall

First take the drawer out of Claude Code:

```sh
ccdrawu uninstall claude
```

This also puts back the theme and the status line you had before ccdrawu's.

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

- [Changelog](CHANGELOG.md)
- [Developer notes](docs/dev-remarks.md)
- [License: GPL-3.0](LICENSE)
