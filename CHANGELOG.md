# Changelog

Each release names the version of terminu design (tdp) it follows.

## Unreleased

- The watcher Claude Code keeps on the comments of an Artifact the session published is no longer taken for a monitor: it drew an owl, and Tasks read `1 monitor`, while no monitor ran. Only a monitor the Monitor tool started counts, as in Claude Code's own count of work running.
- The install script, `install.sh`, finds the latest release where the releases page's latest points, not through GitHub's API, as `ccdrawu update` and `install.ps1` do, and says so when it can't.

## v1.4.2 — tdp v1.3.5

- `ccdrawu update` finds the latest release where the releases page's latest points, not through GitHub's API: the API answers 60 lookups an hour to everything sharing a network address, then `403 Forbidden`, and `ccdrawu update` stopped there. It is terminu-go's update now, the family's one (tdp E1), with ccdrawu's own part after it: the plugin in Claude Code taken to the same release.
- The Windows install script, `install.ps1`, finds the latest release the same way, and says so when it can't.
- `ccdrawu update` on a version newer than the latest release, such as a beta, says so instead of calling it the latest (tdp E1).
- Toasts stack, the newest at the bottom, at most three: what `ccdrawu` says as it starts, such as a file it moved and values that did not fit, is a toast each, no longer rows of one (tdp E3, toast.2).
- Built on terminu-go v1.3.7 (tdp v1.3.5).

## v1.4.1 — tdp v1.2.20

- An owl looks left and right now and then while it rests, and where it goes while it flies.
- Built on terminu-go v1.2.21 (tdp v1.2.20).

## v1.4.0 — tdp v1.2.19

- The sea snail is smaller, 6×3, smaller than a small one of the pet's kind.
- **Shell** and **Monitor** in `ccdrawu`'s preferences pick how a background shell and monitor show beside the transcript, or `none`: for a shell, the domed sea snail (6×3, 7×4, 6×4 or 5×3), a spire, horn or cone sea snail, a hermit crab, or the snail it once was; for a monitor, the owl. They are kept in `config.yaml` as `pet.shell` and `pet.monitor`; open sessions follow within a second. What is not shown the pet leaves alone, and it does not keep it awake.

## v1.3.3 — tdp v1.2.19

- A shell left in the background is a sea snail now, not a snail: a domed shell resting on the ground at the back, its body out at the front, the front of its foot what moves as it crawls. It comes, wanders, draws into its shell and goes as the snail did.
- A new sea snail or owl takes a hue unlike those of the others of its kind on the strip, so two of a kind seldom look alike; before, each one's hue was picked at random on its own.
- Built on terminu-go v1.2.20 (tdp v1.2.19).

## v1.3.2 — tdp v1.2.18

- A command you are to type is in backticks wherever ccdrawu writes it in a sentence, in the color around it: ``run `ccdrawu doctor` `` in the drawer's and the status line's red text (blue in 1.3.1), and in the TUI's confirms, its Alias prompt, the pet's description, the drawer's toast and what the commands print (tdp M5).
- The commands print as every terminu app's do (tdp E1): a wrong use, such as an unknown command or `ccdrawu install` without `claude`, says what is wrong, then a blank line and the usage, with exit status 2; each line of an error starts with `ccdrawu: ` and a capital.
- `ccdrawu uninstall claude`, with values in `config.yaml` it can't use, says on stderr each value it changed and each key it does not read, as `ccdrawu doctor` does and in the TUI's words. Both write paths in the home folder with `~`, as the TUI does.
- While `ccdrawu` is open, the plugin changing its list of items is an INFO in the App Log, naming the items it added and took out; the same list written again is none. One it can't read is thrown away for the plugin to write anew, as before, now with an Error toast the first time, and an INFO once it reads again (tdp note.7).
- Built on terminu-go v1.2.19 (tdp v1.2.18).

## v1.3.1 — tdp v1.2.17

- An owl now keeps time: it rests 10 seconds, flies 10 seconds along the top of the strip, its wings beating, and comes down where it is to rest again. It used to fly off now and then, at random, straight to another perch.
- In the red hint of the drawer and the status line, `ccdrawu doctor` is blue, set apart from the words around it.

## v1.3.0 — tdp v1.2.17

- `ccdrawu doctor` sets right what the drawer and the status line say is wrong with `config.yaml`, as `ccdrawu` does when it opens: the values it can't use go back to their defaults, yours kept in `config.yaml.bak`, and it says on stderr each value it changed. A file that can't be read it leaves as it is, and says why.
- The red text at the top of the drawer and at the end of the status line now ends with `run ccdrawu doctor`.

## v1.2.1 — tdp v1.2.17

- The release is tagged as an annotated tag, its message this heading, as the family releases now (harness project.7). The tags of v1.2.0 and before stay as they were.

## v1.2.0 — tdp v1.2.17

- Snails and owls: beside the transcript, a snail crawls onto the pet's strip for each shell left running in the background, and an owl flies down onto it for each monitor, each in colors picked at random; an owl now and then flies off, its wings beating, to perch somewhere else, on the git boxes too. While Claude is idle the pet now and then goes over to one on the ground and paws at it, and a snail draws into its shell, an owl spreads its wings; with them about it does not fall asleep. Once its task is done, a snail crawls off and an owl flies off.
- Left alone, the pet now and then strolls, looks around or swipes, as it was meant to; it hardly ever did. The same went for a kit's hops and rests, which came less often than meant.
- **Tasks** counts a monitor the Monitor tool started as a monitor; it counted it as a shell.
- In the Claude desktop app, a drawer minimized to a widget shows as it does above the prompt: the pet and columns of text.
- Follows Claude Code 2.1.328's plugin API, which named a type of its own as ccdrawu had named one, and seats a pane in a third place, the desktop app's widget.
- The plugin has an icon, the one `ccdrawu` shows at its start.
- A `config.yaml` that reads but has values ccdrawu can't use is no longer one it can't read (tdp E3): at the start, `ccdrawu` copies it to `config.yaml.bak`, writes it again with the default for each such value and says so in a Warning toast, each value a WARN in the App Log with its key, what you wrote and what is used now. Before, some of them were marked red, others quietly read as their defaults, and a value of the wrong kind (`refresh: ten`) made the whole file one it could not read. `ccdrawu uninstall claude` does the same and says so on stderr. When the copy can't be made, the file is left as it is and an Error toast says why. A theme that does not fit is off, which gives Claude Code its own theme back.
- What one start has to say, a move and values that did not fit, comes in one toast, a row each, and the App Log has an entry for each.
- An item id in `config.yaml` that the plugin's list of items does not have is taken out at the start, as a value that does not fit.
- A key in `config.yaml` that ccdrawu does not read, one mistyped or one it read once, stays in the file when `ccdrawu` writes it, where it was taken out, and each start has an INFO of it in the App Log (tdp E3).
- The drawer and the status line change no file when `config.yaml` can't be read or has values they can't use (tdp E3, outside the TUI): they use the defaults and say which file has what wrong in red, the drawer at its top, the status line at its end; `ccdrawu` sets it right when it next opens. The drawer said nothing, and the status line printed nothing at all for a file it could not read.
- Built on terminu-go v1.2.18 (tdp v1.2.17).

## v1.1.1 — tdp v1.2.0

- The git boxes look more like boxes, and are a darker brown, so orange and brown pets read apart from them: a column wider, lighter inside, with a frame round the sides and the bottom, darker. A box for a file is open in the middle of the top; a taped one is shut under a lid, with a pixel of it down the middle where the flaps meet, and no longer told apart by being light inside.
- The git boxes pile three layers high, nine at most, in the room over them in the strip that stood empty.
- A kit walks in from beyond an end of the pet's strip as its subagent starts, and off past one when it is done; it was there all at once, and gone all at once.
- Where the git boxes get higher or lower, and now and then at a step among them, a kit picks a layer of them at random, the ground in front of them too, and walks along it, in front of the boxes on it; it went up onto the top of them and walked over the outside of the pile alone. The layers are counted from the boxes there, up to the highest with room over it for a kit: the second, and a kit walks in front of the third.
- The pet, a kit and the ball, in front of the boxes or a kit, no longer leave black around their feet: where they take only half a cell, the other half shows what is behind them.

## v1.1.0 — tdp v1.2.0

- Git boxes: beside the transcript, a box stands at the right end of the pet's strip for each file not committed, and a box light inside, taped shut, for each commit not pushed, six at most in two layers. A commit turns boxes of files into a taped one; a push takes the taped ones away. With no upstream there are no taped boxes, as what is not pushed is not known.
- Kits hop now and then, and jump up onto the boxes in their way, walk over them and drop off them.
- Beside the transcript, the pet's strip is two rows taller, room for the boxes and a kit on them.
- Built on terminu-go v1.2.0 (tdp v1.2.0). In `ccdrawu`, the App Log and the key reference (`?`) have `/` search, `n|N` going to the next and the previous hit.
- `ccdrawu statusline`, when `config.yaml` can't be read, writes which file and why to stderr; as before, it prints no line and exits non-zero.
- `ccdrawu uninstall claude`, when `config.yaml` or `originals.yaml` can't be read, stops before it changes anything and says which file and why. It used to take the plugin out of Claude Code anyway and leave ccdrawu's theme and status line set there.
- The repo's templates come from terminu-harness, where tdp v1.2.0 moved them; only their first comments change.
- The developer notes give way to terminu-harness's record of the app, in `harness-assets/`.

## v1.0.1 — tdp v1.1.0

- Built on terminu-go v1.1.0 (tdp v1.1.0). Nothing on the screen or in the keys changes.
- `ccdrawu update` orders versions as semver does: a beta, `1.0.2-beta.1`, is newer than the release before it, `1.0.1`, and older than its own, `1.0.2`, so on a beta it is never offered an older release. It took a version's parts as numbers alone, and read `1.0.1-beta.1` as 1.0.0.1. The latest release it looks up is never a beta, so on a stable release you are never offered one.
- A beta is a GitHub pre-release: Homebrew, Scoop and the install scripts stay on the latest stable release. A release's binaries no longer read `+dirty` in their build information (`go version -m`), and the build takes the path it writes the version into from `go.mod`.
- The repo no longer tracks Claude Code's files: `.gitignore` is the tdp template's again, keeping `CLAUDE.md` and `.claude/` out of git, and the developer notes no longer point to notes kept off GitHub or to a path on one machine.
- The README's terminu family section links tdp, terminu-go and terminu-harness, where it listed the family's other apps.

## v1.0.1-beta.2 — tdp v1.1.0-beta.3

- Built on terminu-go v1.1.0-beta.3 (tdp v1.1.0-beta.3). The release build takes the path it writes the version into from `go.mod`, as every app of the family now does; for ccdrawu, on v1, it is the same path, and `ccdrawu version` reads as it did.

## v1.0.1-beta.1 — tdp v1.1.0-beta.2

- Built on terminu-go v1.1.0-beta.2 (tdp v1.1.0-beta.2). The repo no longer tracks Claude Code's files: `.gitignore` is the tdp template's again, keeping `CLAUDE.md` and `.claude/` out of git, and the developer notes no longer point to notes kept off GitHub or to a path on one machine.
- `ccdrawu update` orders versions as semver does. On a beta it is never offered an older release: on 1.0.2-beta.1 it would have taken 1.0.1 for newer. The release a beta leads to counts as newer than the beta. The latest release it looks up is never a beta, so on a stable release you are never offered one.
- A beta is a GitHub pre-release: Homebrew, Scoop and the install scripts stay on the latest stable release. A release's binaries no longer read `+dirty` in their build information (`go version -m`).

## v1.0.0 — tdp v1.0.0

- Your picks are in `~/.config/ccdrawu/config.yaml`, where they were in `config.json`; what the plugin writes for `ccdrawu`, the list of items and the status line's values, is in your system's cache folder (`~/Library/Caches/ccdrawu` on macOS, `~/.cache/ccdrawu` on Linux, `%LOCALAPPDATA%\ccdrawu\cache` on Windows), where it was in `~/.ccdrawu`. The first time you run `ccdrawu`, it moves them and says from where to where; where it can't, it names both places for you to move them, and goes on reading the old ones. A `config.json` left beside a `config.yaml` is named with it each time `ccdrawu` starts, until you merge what you want of it into `config.yaml` and remove it; `config.yaml` is what is read. Until then the drawer and the status line read the old places too. `CCDRAWU__CACHE` names another folder for the cache.
- The Claude Code theme and status line you had before ccdrawu's, kept to put back, are in `~/.ccdrawu/originals.yaml` (`CCDRAWU__DATA` names another folder), where they were in the config as `theme.before` and `statusline.before`: they can't be made again, so they are kept apart from your picks. The first time you run `ccdrawu`, or `ccdrawu uninstall claude`, moves them, and `ccdrawu` says so; until then they are read where they were. The file is there only while ccdrawu's theme or status line is on.
- `!` opens the App Log: what `ccdrawu` has done since it started, each save, what it wrote into Claude Code, every message it showed, and what went wrong. Until you look, the footer reads `!:1 warning` or `!:2 errors` in place of `!:log`.
- A `config.yaml` that can't be read is no longer worked on: `ccdrawu` says why and asks first, before it moves anything. `Enter` renames it `config.yaml.bak` (`.bak.1` and on where one is there already) and starts as the first time you run it: with the defaults, and what is in the old places moved; `Esc` quits, with the reason on the terminal. `originals.yaml` is asked about the same way, and its `.bak` stays beside it. A list of items that can't be read is thrown away, said in the App Log, and the plugin writes it again; the pages no longer read a red failure.
- Built on terminu-go v1.0.0 (tdp v1.0.0). With a popup open, the first `q` or `Ctrl-C` asks `Quit ccdrawu?` before it quits. The **Items** popup starts with `Select all`. A message of several rows, what could not move or what moved, reads a row each; where `claude` says several lines, the error popup and **Plugin** keep the first, and the App Log all of them. The message after an install or an uninstall from **Plugin** reads `Claude Code:` and Claude Code's own words. `ccdrawu` needs a terminal of at least 40×12, and says so in the middle of a smaller one, with how to quit: `Ctrl-C` twice. A setting that can't be saved keeps its popup open with what you typed, the error over it. A change that is done says so in green (`SUCCESS`). Before ccdrawu has run in Claude Code, the empty pages say so in grey, not red. **Refresh** has its unit, `seconds`, under the value as you type. In **Order**, `h` and `l` move as `k` and `j` do. A control character in a value set by hand reads in red, `\x1b`. Opened with no **Drawer background** set, the color picker's old reads `none`. `?` lists the movement keys only where they move something. In the color picker for the drawer background `h|j|k|l` move among R, G and B only, and `Tab` goes to `Choose`; text in a confirm, a toast and an error popup wraps at spaces, and between wide characters.
- The repo, from the tdp v0.8.0 project templates: `family.mk` (with `make dev` since tdp v0.9.4), the release workflow, goreleaser, `install.sh` and `uninstall.sh`, `internal/version`. `uninstall.sh` also offers to remove a cache folder, where there is one.
- **Drawer background** reads as a swatch of the color and its hex, where it read the hex alone.
- **Claude theme** reads the theme picked; `Enter` opens the list to pick from, where all seven were drawn on the page (tdp: more than five is a select).
- **Refresh** takes digits only, and says the range, 1-60, before you type; on the page it reads `10 seconds`, where it read `every 10 seconds`.
- A value set by hand in `config.yaml` that ccdrawu does not take, an alias, a background, a refresh or a separator, turns its label red; a control character in it no longer goes to the terminal as it was. **Plugin** turns red too where ccdrawu can't tell whether Claude Code has the plugin.
- A change that can't be saved, or a theme that can't be set, is not shown as if it were: the page goes back to what is saved.
- An install or uninstall that fails says which in its title, `INSTALL FAILED` or `UNINSTALL FAILED`, where both read `CLAUDE CODE SAID NO`; the alias's errors name the alias.
- On the right side, what `Enter` does on the value under the cursor is in the `Space` menu and in `?` too: `[Enter] Edit`, or `Choose` on an option.
- Before ccdrawu has run in Claude Code, the empty pages name the key that installs the plugin: `[Enter]` on **Plugin** in preferences.
- The README runs in the family's order: requirements, the first run, updating and removing under Install, then a quick start; troubleshooting comes right after usage. It says the terminal needs truecolor, and names the variables for an icon two cells wide, `CCDRAWU__ICON_WIDTH`, `CCDRAWU__POWERLINE_WIDTH` and the family's `TERMINU__` ones; Claude Code's keys read as Claude Code writes them, `ctrl+x tab`.
- After `/clear` the drawer and the status line keep their values, and the drawer the rows you picked in `ccdrawu`. Model, Session, Weekly and Cost read `-` until the next reply, the versions until Claude Code started again, the drawer went back to its default rows, and the pet fell asleep.

## [0.1.4] - 2026-10-16
### Added
- Kits: while subagents run, a small one of the pet's kind walks behind the pet for each, up to three, beside the transcript, each in a coat of its own, never the pet's. It comes in at an end of the walk as its subagent starts, and walks off the nearer end once it is done; the pet goes on as before, drawn over them.
### Changed
- A row turned on is there from the start, in the drawer and the status line, and reads `-` until its value comes, or where there is none, as Spend without a spend limit. Rows were missing until then, Mode until the first prompt of a new session, and the rows under them moved down when they came. Model and Uptime read `-` where they read `…`. In the sidebar Session, Weekly and Spend keep their resets row the same way.
- Spend starts off, as few go through a gateway with a spend limit and the rest would see `-` there always. Where you picked your rows in `ccdrawu`, they stay as you picked them.
### Removed
- Account: Claude Code tells the account only to plugins an organization installs, so for one you installed yourself it never had a value. Where your config still lists it, it is left out.

## [0.1.3] - 2026-10-16
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
- `docs/icon.svg`: the terminu family's mark, ccdrawu's own.
