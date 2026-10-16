# Changelog

## [Unreleased]
### Added
- `ccdrawu update`: updates ccdrawu to its latest release the way it was installed, through Homebrew, or, from the install script, in place, checked against the release's checksums; then the plugin in Claude Code to the same release, where Claude Code has it. Installed with Scoop, it says what to run, since Scoop does not update ccdrawu while it runs. A folder it cannot write in, it says so, and never asks for more rights.
### Changed
- `/ccdrawu` turns the drawer on or off for the whole session, wherever it sits, over **Drawer**. Opened where **Drawer** is off, it stays as the terminal widens and narrows, where it used to go once it moved; closed, it stays closed through `/reload-plugins` too. Changing **Drawer** in `ccdrawu` takes over again.
### Fixed
- Context reads 0% as soon as a compaction ends, in the drawer and the status line, until the next reply measures it; it kept the fill from before the compaction until then.
- Git in the sidebar shows the added and removed lines on a row of their own under the branch, as Session and Weekly show their resets, so a branch too long for its row wraps whole; they were cut apart across the wrap. Above the prompt and in the status line Git stays on one row.

## [0.1.2] - 2026-10-16
### Added
- **Statusline** `on when narrow`: the status line only while the terminal is too narrow for the sidebar, under about 144 columns, and the drawer is not beside the transcript. Otherwise the row goes away.
### Changed
- `ccdrawu install claude` installs the plugin released with it, from that release's tag, where it took whatever the main branch held; a plugin installed before, from the main branch or an older release, moves to it. Run again, it updates the plugin, which Claude Code does not do by itself; it did nothing there before. The README says how to update, and to install the plugin with `ccdrawu install claude` only.
- **Drawer** atop the sidebar page and atop the inline page are a setting each, where they were one. Off on the inline page alone, a terminal narrowed below the sidebar's width hides the drawer, instead of moving it above the prompt, until it widens again. Turned back on in a terminal that narrow, the drawer stays out of sight until `/ccdrawu`, as Claude Code places it, and open sessions say so in a toast. 0.1.1's off reads as off on both.
### Fixed
- `/ccdrawu` shows a drawer that waits for a wider terminal at once; it took two.
- `ccdrawu uninstall claude` puts back your theme and status line even when the ccdrawu plugin is already gone from Claude Code; it used to stop at "The ccdrawu plugin is not in Claude Code.", leaving Claude Code's status line on a ccdrawu about to be removed.
- The status line you had before ccdrawu's comes back exactly as it was written when you turn ccdrawu's off after `ccdrawu` was closed and opened again, or with `ccdrawu uninstall claude`; its spacing used to change.

## [0.1.1] - 2026-10-15
### Added
- The status line: ccdrawu can be Claude Code's status line, the row under the prompt, with the drawer or on its own. Turned on under Statusline in `ccdrawu`, it sets `ccdrawu statusline` as Claude Code's status line; one you had is replaced only after asking, and comes back when the status line is turned off or with `ccdrawu uninstall claude`. Its page has Items and Order, as the drawer's (Model, Context, Session, Weekly, Cost and Git to start with); Format, `label` (`Context 42%`), `short` (`ctx 42%`) or `value` (`42%`); Color, `highlighted` or `dimmed`; Separator, `|` unless you set another; and Refresh, 10 seconds unless you set another. While it is on, the plugin keeps each session's values in `~/.ccdrawu/sessions/`, each file removed after a week without use; where the plugin is not running, it shows the model, context and cost Claude Code hands it.
- **Drawer**, atop the sidebar and inline pages: `off` starts sessions without the drawer and closes it in open ones; `/ccdrawu` still opens and closes it.
- Directory, the working directory's own name without the path to it, off until picked.
- Titles over the columns of every page in `ccdrawu`: SETTING and VALUE, or KEY and VALUE on preferences.

## [0.1.0] - 2026-10-15
### Added
- The drawer: a pane beside the Claude Code transcript (Claude Code 2.1.287 or later, in fullscreen rendering) that opens by itself when a session starts; `/ccdrawu` opens or closes it, even while Claude is working. Where it can't sit beside the transcript it sits above the prompt, its rows in columns.
- Its rows: Model; Effort, from green at `low` to red at `max`; Mode, the permission mode; Context, Session and Weekly, bars of the context window and of the 5-hour and weekly plan limits with their reset countdowns, green, yellow, peach, then red at 50%, 70% and 90%; Spend, the same for a gateway's spend limit, where there is one; Cost; Uptime and the prompts sent; Cache, how much of the last request the prompt cache served; Compact, the compactions; Tasks, the subagents running and the shells, monitors and workflows left in the background; Git, the branch with its added and removed lines; Path; Claude, the version with the newest release when one is out; and ccdrawu's own version.
- More rows, off until picked: Style, the output style; ID and Account; Window, the tokens in the context window out of its size; Rate, the cost an hour; Tokens and Speed; Skills; Sync, Files, Commit and Repo, which say `no upstream`, `no commits`, `no origin` or `not a repo` where there is none; and Status, from status.claude.com.
- A pixel pet at the top of the drawer, or in the leftmost column above the prompt: a cat, dog, rabbit, duck or penguin, in a coat picked at random each session. It wanders while Claude works, sits and says so when Claude is done, jumps when a reply ends in an error or a tool fails, lies down after 2 minutes alone and falls asleep after 10. `/ccdrawu pat` pats it; `/ccdrawu ball` throws it a ball. Its lines are fixed, so it costs no tokens.
- Bars drawn in Nerd Font squares, or, set in `ccdrawu`, in a line any monospace font has.
- The `ccdrawu` command, the settings screen. Under Drawer, sidebar and inline each have Items, every row the drawer can show, checked in one popup, and Order, the ones checked in the order they show, moved about in a reorder popup. Under Settings, preferences holds the rest: an Alias, one more name for `/ccdrawu` such as `/ccd`; a Claude theme, one of Claude Code's built-in themes, with a Drawer background of your own over it, written as `~/.claude/themes/ccdrawu.json` and selected in Claude Code, the theme you had coming back with `off`; Bar; Pet, the animal; and Plugin, whether Claude Code has the drawer, installed or uninstalled from there. Every change is saved at once, and open Claude Code sessions follow within a second.
- `ccdrawu install claude` and `ccdrawu uninstall claude`, putting the drawer into Claude Code for your user and taking it out again; uninstalling puts your own theme back. `CCDRAWU__MARKETPLACE` points the install at another marketplace.
- Settings in `~/.config/ccdrawu/config.json` (`CCDRAWU__CONFIG`, or `XDG_CONFIG_HOME`), and what the plugin writes in `~/.ccdrawu/` (`CCDRAWU__DATA`); on Windows, `~` is your user folder.
- macOS, Linux and Windows (not yet tried on a Windows machine), on amd64 and arm64: an install script for each (`install.sh`, `install.ps1`) and its uninstall script, Homebrew (`vulcanshen/tap/ccdrawu`) and Scoop (`vulcanshen/scoop-bucket`).
- `docs/icon.svg`: the terminu family's mark, ccdrawu's own, which the settings screen's splash draws.
