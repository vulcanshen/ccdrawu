import { atom, read, update } from 'claude-code'
import type { EngineInterface as Engine, Register } from 'claude-code'

import type {
  Background,
  CacheInfo,
  Config,
  GitInfo,
  ModelInfo,
  Modes,
  ServiceStatus,
  SessionInfo,
  SkillsInfo,
  TokenInfo,
  Usage,
  VersionInfo,
} from '../types'
import {
  BALL,
  BALL_INK,
  BALL_LEFT,
  drawnCells,
  LIFT_ROOM,
  PET_HEIGHT,
  PET_START,
  PET_TICK,
  petCells,
  petCoat,
  petEvent,
  petLook,
  petSpecies,
  petTick,
  TRACK,
  type PetEvent,
  type Run,
} from './pets'
import { registerCommands } from './commands'
import {
  bar,
  countBackground,
  parseShortstat,
  parseStatusV2,
  percentText,
  repoFromRemote,
  shortPath,
  usageColor,
} from './format'
import { C, ITEMS, view, type Seg, type Snapshot, type View } from './items'
import { chunk, COLUMN_GAP, columnRows, drawerShows, enabledItems, inlineValue, parseConfig, type Placement, segWidth, showsAnywhere } from './layout'
import { locations } from './paths'
import { statuslineFile } from './statusline'

const PANE = 'ccdrawu'
const PANE_OPEN = { id: PANE, title: 'Status', columns: 34 } as const
const LABEL = 9 // key column width beside the transcript: keys of 8 characters, and a space; wider while a longer one shows

const usage = atom({ plugin: 'ccdrawu', key: 'usage' } as const, null as Usage | null)
const model = atom({ plugin: 'ccdrawu', key: 'model' } as const, null as ModelInfo | null)
const git = atom({ plugin: 'ccdrawu', key: 'git' } as const, null as GitInfo | null)
const path = atom({ plugin: 'ccdrawu', key: 'path' } as const, '')
const turns = atom({ plugin: 'ccdrawu', key: 'turns' } as const, 0)
const cache = atom({ plugin: 'ccdrawu', key: 'cache' } as const, null as CacheInfo | null)
const tokens = atom({ plugin: 'ccdrawu', key: 'tokens' } as const, { input: 0, output: 0, cacheRead: 0, cacheWrite: 0 } as TokenInfo)
const sessionInfo = atom({ plugin: 'ccdrawu', key: 'session' } as const, {} as SessionInfo)
const service = atom({ plugin: 'ccdrawu', key: 'status' } as const, null as ServiceStatus | null)
const skills = atom({ plugin: 'ccdrawu', key: 'skills' } as const, { count: 0 } as SkillsInfo)
const compacts = atom({ plugin: 'ccdrawu', key: 'compacts' } as const, 0)
const agents = atom({ plugin: 'ccdrawu', key: 'agents' } as const, 0)
const background = atom({ plugin: 'ccdrawu', key: 'background' } as const, {} as Background)
const version = atom({ plugin: 'ccdrawu', key: 'version' } as const, null as VersionInfo | null)
const drawer = atom({ plugin: 'ccdrawu', key: 'drawer' } as const, null as string | null)
const modes = atom({ plugin: 'ccdrawu', key: 'modes' } as const, {} as Modes)
const now = atom({ plugin: 'ccdrawu', key: 'now' } as const, 0)
const pet = atom({ plugin: 'ccdrawu', key: 'pet' } as const, PET_START)
const config = atom({ plugin: 'ccdrawu', key: 'config' } as const, {} as Config)

async function refreshUsage($: Engine) {
  const u = await $.session.usage()
  await update($, usage, () => ({
    startedAt: u.startedAt,
    contextPercent: u.context.percent,
    contextTokens: u.context.tokens,
    contextWindow: u.context.window,
    limits: u.rateLimits.map(l => ({ kind: l.kind, percentUsed: l.percentUsed, resetsAt: l.resetsAt })),
    costUsd: u.cost?.usd,
  }))
}

async function tick($: Engine) {
  const t = await $.clock.now()
  await update($, now, () => t)
}

async function refreshTurns($: Engine) {
  const n = await $.session.turns()
  await update($, turns, () => n)
}

// Before a resumed session sends its first request, take the effort /model saved in settings;
// turn.step's real value replaces it once a request goes out
async function refreshModelName($: Engine) {
  const name = await $.session.model()
  const settings = (await $.settings.read()) as unknown as {
    effortLevel?: string
    modelSettings?: Record<string, { effortLevel?: string } | undefined>
  }
  const configured = settings.modelSettings?.[name]?.effortLevel ?? settings.effortLevel
  await update($, model, m => ({ name, effort: m?.effort ?? configured }))
}

let home: string | undefined
let dataDir: string | undefined
let configFile: string | undefined

async function locateFiles($: Engine) {
  // Windows has USERPROFILE, and often no HOME; the TUI's os.UserHomeDir reads them in this order too
  home = (await $.env.get('USERPROFILE')) || (await $.env.get('HOME'))
  if (!home) return
  const found = locations({
    home,
    data: await $.env.get('CCDRAWU__DATA'),
    config: await $.env.get('CCDRAWU__CONFIG'),
    xdgConfig: await $.env.get('XDG_CONFIG_HOME'),
  })
  dataDir = found.data
  configFile = found.configFile
}

// Write `text` to <data>/<name> unless it already holds it
async function keepFile($: Engine, name: string, text: string) {
  if (dataDir === undefined) return
  const target = `${dataDir}/${name}`
  const current = await $.fs.read(target).catch(() => undefined)
  if (current !== text) await $.fs.write(target, text)
}

// The TUI lists what this version of the mod can show from catalog.json, so it needs no update of its own
async function installCatalog($: Engine) {
  await keepFile($, 'catalog.json', JSON.stringify({ version: 1, items: ITEMS }, null, 2) + '\n')
}

// The TUI writes ~/.config/ccdrawu/config.json; every open session looks at it each second and takes
// a change at once. A file not there means the catalog's defaults
let configStamp: number | undefined
let statusWasShown = false // whether the config read last had the Status row on
// Where the config read last let the drawer show, before the first read
let drawerWas: { dock: boolean; inline: boolean } | undefined
// The command names this session has registered: /ccdrawu, and each alias the config has named
const commands = new Set<string>()

async function refreshConfig($: Engine) {
  const file = configFile
  if (file === undefined) return
  const stamp = await $.fs.stat(file).then(
    s => s.mtimeMs,
    () => 0,
  )
  if (stamp === configStamp) return
  const parsed = stamp === 0 ? {} : parseConfig(await $.fs.read(file))
  // Half written: keep what shows and read it again next second
  if (parsed === undefined) return
  configStamp = stamp
  await update($, config, () => parsed)
  await registerCommands(spec => $.command.register(spec), commands, parsed.alias)
  // The Status row just turned on, or on from the start: its reading now, not at the next 5 minutes
  const shown = showsAnywhere(parsed, 'status')
  if (shown && !statusWasShown) quietly(refreshStatus($))
  statusWasShown = shown
  // The drawer just turned on or off in a place: open sessions open it, close it, or take it out of a
  // place it is off in. The first read leaves it as session.start has it, so a reload keeps a drawer
  // /ccdrawu opened
  const shows = { dock: drawerShows(parsed, 'dock'), inline: drawerShows(parsed, 'inline') }
  if (drawerWas !== undefined && (shows.dock !== drawerWas.dock || shows.inline !== drawerWas.inline)) {
    asked = undefined
    if (!shows.dock && !shows.inline) await closePane($)
    else if (!(await $.ui.panes()).some(p => p.id === PANE)) void $.ui.open(PANE_OPEN)
    else await keepPlace($)
  }
  drawerWas = shows
}

// Where /ccdrawu opened the drawer: it stays there, the place off or not, until it moves. pending until
// the drawer is drawn
let asked: 'pending' | Placement | undefined
// Whether the terminal docks panes beside the transcript, as the pane's last drawing says
let fullscreen = false

async function closePane($: Engine) {
  placedAs = undefined
  await $.ui.close({ id: PANE })
}

// The drawer out of a place the config has it off in. Above the prompt, in a terminal that docks panes,
// it opens again unasked: the engine holds it undrawn until the terminal is wide enough to dock it
async function keepPlace($: Engine) {
  const place = placedAs
  if (place !== 'dock' && place !== 'inline') return
  if (asked === place) return
  asked = undefined
  const pane = (await $.ui.panes()).find(p => p.id === PANE)
  const conf = await read($, config)
  if (pane?.isPlaced !== true || drawerShows(conf, place)) return
  await closePane($)
  if (place === 'inline' && fullscreen && drawerShows(conf, 'dock')) void $.ui.open(PANE_OPEN)
}

// The pet on either side of the prompt: it has no need of a tick, or of events, when it shows nowhere
async function petShown($: Engine) {
  return showsAnywhere(await read($, config), 'pet')
}

// A tick of the pet: written only when what it shows changed, worked out again on the value it has
// then, so an event that came in between is kept
async function stepPet($: Engine) {
  if (!(await petShown($))) return
  const t = await $.clock.now()
  const s = await read($, pet)
  const next = petTick(s, t)
  if (next === s) return
  await update($, pet, current => (current === s ? next : petTick(current, t)))
}

async function petHears($: Engine, event: PetEvent) {
  if (!(await petShown($))) return
  const t = await $.clock.now()
  const species = petSpecies((await read($, config)).pet?.species)
  await update($, pet, s => petEvent(s, event, t, species, Math.random))
}

// No API reports the permission mode; Claude Code's own hook events carry permission_mode,
// so remember the main conversation's from them
async function rememberMode($: Engine, e: { permission_mode?: string; agent_id?: string }) {
  const mode = e.permission_mode
  if (e.agent_id !== undefined || mode === undefined) return
  // Read before each tool call too: leave the drawing alone when nothing changed
  if ((await read($, modes)).permission === mode) return
  await update($, modes, m => ({ ...m, permission: mode }))
}

async function refreshGit($: Engine) {
  const cwd = await $.session.cwd()
  await update($, path, () => shortPath(cwd, home))

  const head = await $.process.run(['git', 'rev-parse', '--abbrev-ref', 'HEAD'], { timeoutMs: 3_000 })
  if (head.exitCode !== 0) {
    await update($, git, () => null)
    return
  }
  // Before the first commit there is no HEAD: compare with the index instead
  let diff = await $.process.run(['git', 'diff', 'HEAD', '--shortstat'], { timeoutMs: 3_000 })
  if (diff.exitCode !== 0) diff = await $.process.run(['git', 'diff', '--shortstat'], { timeoutMs: 3_000 })
  const branch = head.stdout.trim()
  const status = await $.process.run(['git', 'status', '--porcelain=v2', '--branch'], { timeoutMs: 3_000 })
  const place = await repoPlace($, cwd)
  await update($, git, () => ({
    branch,
    ...parseShortstat(diff.stdout),
    ...(status.exitCode === 0 ? parseStatusV2(status.stdout) : {}),
    ...place,
  }))
}

// origin's owner/repo, and the worktree's folder when the session is in a linked one; they change with the
// working directory alone, so they are looked up again only when it moves
let placeOf: { cwd: string; repo?: string; worktree?: string } | undefined

async function repoPlace($: Engine, cwd: string): Promise<{ repo?: string; worktree?: string }> {
  if (placeOf?.cwd !== cwd) {
    const remote = await $.process.run(['git', 'remote', 'get-url', 'origin'], { timeoutMs: 3_000 })
    const dirs = await $.process.run(
      ['git', 'rev-parse', '--path-format=absolute', '--git-dir', '--git-common-dir', '--show-toplevel'],
      { timeoutMs: 3_000 },
    )
    const [gitDir, commonDir, top] = dirs.stdout.trim().split('\n')
    placeOf = {
      cwd,
      repo: remote.exitCode === 0 ? repoFromRemote(remote.stdout) : undefined,
      worktree: dirs.exitCode === 0 && gitDir !== commonDir ? top?.split(/[\\/]/).pop() : undefined,
    }
  }
  return { repo: placeOf.repo, worktree: placeOf.worktree }
}

// The session's id, the account signed in and the output style
async function refreshSession($: Engine) {
  const id = (await $.session.id()).replace(/\.jsonl$/, '')
  const who = await $.session.identity().catch(() => undefined)
  const style = (await $.config.list()).find(r => r.key === 'outputStyle')?.value
  await update($, sessionInfo, () => ({
    id,
    account: who?.email,
    outputStyle: typeof style === 'string' ? style : undefined,
  }))
}

// status.claude.com's overall status, from its public status page API
// Claude's status, asked of status.claude.com only while the Status row shows somewhere: a request out
// that the many who leave the row off never make
async function refreshStatus($: Engine) {
  if (!showsAnywhere(await read($, config), 'status')) return
  const res = await $.http.fetch('https://status.claude.com/api/v2/status.json')
  if (!res.ok) return
  const s = (JSON.parse(res.text) as { status?: { indicator?: string; description?: string } }).status
  if (s?.indicator !== undefined && s.description !== undefined) {
    const status = { indicator: s.indicator, description: s.description }
    await update($, service, () => status)
  }
}

async function refreshAgents($: Engine) {
  const list = await $.agent.list()
  const running = list.filter(a => a.status === 'pending' || a.status === 'running' || a.status === 'waiting')
  await update($, agents, () => running.length)
}

async function refreshVersion($: Engine) {
  const { base, version: full } = await $.session.version()
  const current = base ?? full
  await update($, version, v => ({ current, latest: v?.latest }))
  const res = await $.http.fetch('https://registry.npmjs.org/@anthropic-ai/claude-code/latest')
  if (!res.ok) return
  const latest = (JSON.parse(res.text) as { version?: string }).version
  await update($, version, () => ({ current, latest }))
}

// $.plugin has no version: read this plugin's own plugin.json
async function refreshDrawer($: Engine) {
  const manifest = JSON.parse(await $.fs.read(`${$.plugin.root}/.claude-plugin/plugin.json`)) as { version?: string }
  await update($, drawer, () => manifest.version ?? null)
}

// Everything the items may show, as the drawer holds it now
async function readSnapshot($: Engine): Promise<Snapshot> {
  return {
    usage: await read($, usage),
    model: await read($, model),
    git: await read($, git),
    path: await read($, path),
    turns: await read($, turns),
    cache: await read($, cache),
    tokens: await read($, tokens),
    session: await read($, sessionInfo),
    status: await read($, service),
    skills: await read($, skills),
    compacts: await read($, compacts),
    agents: await read($, agents),
    background: await read($, background),
    version: await read($, version),
    drawer: await read($, drawer),
    modes: await read($, modes),
    now: await read($, now),
  }
}

// While the config has the statusline on, keep what it shows in <data>/sessions/<session id>.json for
// `ccdrawu statusline` to lay out, written only when it changed. The id is asked each time: after /clear
// the session goes on under a new one, and Claude Code hands the command that one
let statuslineWritten: { file: string; text: string } | undefined
// Where the pane was drawn last: dock, beside the transcript, or inline, above the prompt; undefined
// once ccdrawu closes it
let placedAs: Placement | undefined

async function writeStatusline($: Engine) {
  if (dataDir === undefined || !(await read($, config)).statusline?.on) return
  const id = (await $.session.id()).replace(/\.jsonl$/, '')
  const file = `${dataDir}/sessions/${id}.json`
  // Beside the transcript only while the pane is open and drawn there: not closed, nor waiting for a
  // wider terminal, nor above the prompt
  const pane = (await $.ui.panes()).find(p => p.id === PANE)
  const docked = pane?.isPlaced === true && placedAs === 'dock'
  const text = statuslineFile(await readSnapshot($), docked)
  if (statuslineWritten?.file === file && statuslineWritten.text === text) return
  await $.fs.write(file, text)
  statuslineWritten = { file, text }
}

// A failed background refresh (git timing out, offline for the version check) affects its row alone
function quietly(work: Promise<unknown>) {
  void work.catch(() => undefined)
}

export const register: Register = on => {
  on('session.start', async ($, e, next) => {
    // The alias follows once the config is read
    await registerCommands(spec => $.command.register(spec), commands, undefined)

    await tick($)
    // The pet's coat, once a session (a reload keeps it), and it starts awake. A reload keeps the state
    // an older module wrote too: fields it did not have yet take their defaults
    const started = await $.clock.now()
    await update($, pet, p => ({
      ...PET_START,
      ...p,
      coat: p.coat < 0 ? Math.floor(Math.random() * 1000) : p.coat,
      active: started,
    }))
    await locateFiles($).catch(() => undefined)
    quietly(installCatalog($))
    // Read before the drawer opens, which it does only while the config has it on
    await refreshConfig($).catch(() => undefined)
    quietly(refreshUsage($))
    quietly(refreshModelName($))
    quietly(refreshGit($))
    quietly(refreshAgents($))
    quietly(refreshVersion($))
    quietly(refreshDrawer($))
    quietly(refreshTurns($))
    quietly(refreshSession($))

    // Countdowns and uptime every 30 seconds; git, path and subagents every 5; Claude's status every 5
    // minutes while its row shows (first when the config turns it on); the latest version every 6 hours;
    // the config, and the statusline's file while it is on, every second
    $.clock.every(30_000, () => quietly(tick($)))
    $.clock.every(5_000, () => {
      quietly(refreshGit($))
      quietly(refreshAgents($))
    })
    $.clock.every(5 * 60_000, () => quietly(refreshStatus($)))
    $.clock.every(6 * 60 * 60_000, () => quietly(refreshVersion($)))
    $.clock.every(1_000, () => quietly(refreshConfig($)))
    $.clock.every(1_000, () => quietly(writeStatusline($)))
    $.clock.every(PET_TICK, () => quietly(stepPet($)))

    // In a terminal too narrow for a sidebar the engine holds the pane until it widens. With the drawer
    // off in both places the session starts without it, and /ccdrawu opens it; off in one, the pane
    // drawn there is taken out of it (keepPlace)
    const conf = await read($, config)
    if (drawerShows(conf, 'dock') || drawerShows(conf, 'inline')) void $.ui.open(PANE_OPEN)

    return next(e)
  })

  // Every slash command passes through here, so a failure must leave other commands alone
  on('command.run', async ($, e, next) => {
    if (!commands.has(e.command)) return next(e)
    // /ccdrawu pat pats the pet, /ccdrawu ball throws it a ball; anything else opens or closes the drawer
    const word = e.args.trim().toLowerCase()
    if (word === 'pat' || word === 'ball') {
      await petHears($, word)
      return {}
    }
    // A pane waiting undrawn for a wider terminal reads as closed: opened again, asked, it is drawn
    const isShown = (await $.ui.panes()).some(p => p.id === PANE && p.isPlaced)
    if (isShown) {
      asked = undefined
      await closePane($)
    } else {
      asked = 'pending'
      await $.ui.open(PANE_OPEN)
    }
    return {}
  }).catch(($, e, next) => (commands.has(e.command) ? {} : next(e)))

  // Usage, context and cost changed (after each turn, or a rate limit moved)
  on('session.measure', async ($, e, next) => {
    const result = await next(e)
    quietly(refreshUsage($))
    return result
  })

  // Each request of the main conversation: its effort, how much the prompt cache served, the tokens it
  // took, and its output tokens a second (the time from asking to the last token, waiting included)
  on('turn.step', async function* ($, e, next) {
    if (e.agentId === undefined) quietly(petHears($, 'busy'))
    const start = await $.clock.now()
    const result = yield* next(e)
    if (e.agentId === undefined) {
      const effort = e.effort === undefined ? undefined : String(e.effort)
      await update($, model, m => ({ name: m?.name ?? e.model, effort }))
      const u = result.usage
      if (u) {
        const total = u.input_tokens + u.cache_read_input_tokens + u.cache_creation_input_tokens
        await update($, cache, () => ({ read: u.cache_read_input_tokens, total }))
        const seconds = ((await $.clock.now()) - start) / 1000
        await update($, tokens, t => ({
          input: t.input + u.input_tokens,
          output: t.output + u.output_tokens,
          cacheRead: t.cacheRead + u.cache_read_input_tokens,
          cacheWrite: t.cacheWrite + u.cache_creation_input_tokens,
          speed: seconds > 0 && u.output_tokens > 0 ? u.output_tokens / seconds : t.speed,
        }))
      }
    }
    return result
  })

  on('turn.complete', async ($, e, next) => {
    const result = await next(e)
    if (e.agentId === undefined) {
      quietly(petHears($, e.reason === 'answer' ? 'done' : e.reason === 'aborted' ? 'stop' : 'failed'))
      quietly(refreshTurns($))
      quietly(refreshModelName($))
      // The output style may have changed in /config meanwhile
      quietly(refreshSession($))
    }
    quietly(refreshAgents($))
    return result
  })

  // The events below hold back what the engine does; if the drawer fails, .catch lets it go on
  // (when the hook already called next, that call's result is reused)
  on('agent.spawn', async ($, e, next) => {
    const result = await next(e)
    quietly(refreshAgents($))
    return result
  }).catch(($, e, next) => next(e))

  // A skill run in this session: count it, and keep its name
  on('skill.prompt', async ($, e, next) => {
    const result = await next(e)
    quietly(update($, skills, s => ({ count: s.count + 1, last: e.skill })))
    return result
  }).catch(($, e, next) => next(e))

  on('session.compact', async ($, e, next) => {
    const result = await next(e)
    if (e.agentId === undefined && e.trigger !== 'precompute' && !('skip' in result && result.skip)) {
      quietly(update($, compacts, n => n + 1))
      quietly(petHears($, 'compact'))
    }
    return result
  }).catch(($, e, next) => next(e))

  // A tool of Claude's own failed, not one the user stopped: the pet jumps
  on('classic.PostToolUseFailure', async ($, e, next) => {
    const result = await next(e)
    if (e.agent_id === undefined && e.is_interrupt !== true) quietly(petHears($, 'error'))
    return result
  }).catch(($, e, next) => next(e))

  // A compaction starts the context over, and no turn ends to measure it: read it again
  on('classic.PostCompact', async ($, e, next) => {
    const result = await next(e)
    if (e.agent_id === undefined) quietly(refreshUsage($))
    return result
  }).catch(($, e, next) => next(e))

  // Permission mode: recorded at session start, on each prompt, after each step's tool calls and
  // when a reply ends (after shift+tab it shows once the next of these comes). classic.PreToolUse
  // carries no permission_mode; PostToolBatch comes once a step, where PostToolUse comes per tool
  on('classic.PostToolBatch', async ($, e, next) => {
    const result = await next(e)
    quietly(rememberMode($, e))
    return result
  }).catch(($, e, next) => next(e))

  on('classic.SessionStart', async ($, e, next) => {
    const result = await next(e)
    quietly(rememberMode($, e))
    return result
  }).catch(($, e, next) => next(e))

  on('classic.UserPromptSubmit', async ($, e, next) => {
    const result = await next(e)
    quietly(rememberMode($, e))
    return result
  }).catch(($, e, next) => next(e))

  on('classic.Stop', async ($, e, next) => {
    const result = await next(e)
    quietly(rememberMode($, e))
    if (e.agent_id === undefined) quietly(update($, background, () => countBackground(e.background_tasks ?? [])))
    return result
  }).catch(($, e, next) => next(e))

  on('ui.render', { component: 'Pane', requestId: PANE }, async ($, e) => {
    const { Box, Text } = $.ui.resolve(e)
    // Leave out the paddingX on either side, or a row runs past the width and wraps
    const inner = Math.max(22, e.props.bodyColumns - 2)
    const placement = e.props.placement
    fullscreen = e.viewport?.isFullscreen === true
    if (asked === 'pending') asked = placement
    // Drawn somewhere new: out of it if the config has the drawer off there
    if (placement !== placedAs) {
      placedAs = placement
      quietly(keepPlace($))
    }

    const s = await readSnapshot($)
    const conf = await read($, config)
    const enabled = enabledItems(conf, placement)
    // The pet is not a row: it is drawn on its own, over the rows or left of the columns
    const shown = enabled.flatMap(item => {
      const v = view(item.id, s)
      return v ? [{ item, view: v }] : []
    })

    // No Fragments here: this surface lays a Fragment out as a column, so siblings go in a row Box
    const segs = (list: readonly Seg[]) =>
      list.map(g => (
        <Text color={g.color} bold={g.bold === true}>
          {g.text}
        </Text>
      ))

    // The pet, in this session's coat of the animal the config names
    const hasPet = enabled.some(item => item.id === 'pet')
    const p = await read($, pet)
    const species = petSpecies(conf.pet?.species)
    const look = petLook(p)
    // Columns from the left of the strip for a place along the walk, spread over the width there is
    const along = (pos: number) => Math.round((pos / TRACK) * Math.max(0, inner - 12))
    const ball = p.ball === undefined ? undefined : { pos: p.ballPos, lift: p.ballLift }
    // A pet's box keeps its size, each row one high: a row of nothing but gaps (a sleeping pet's top
    // row with no snore yet) would otherwise fold to nothing and the rows under it jump up. Beside the
    // transcript it is taller than the pet, room for it to jump into; above the prompt there is none
    const cellRows = (rows: Run[][]) =>
      rows.map(row => (
        <Box flexDirection="row" height={1} flexShrink={0}>
          {row.map(r =>
            'gap' in r ? (
              <Box width={r.gap} flexShrink={0} />
            ) : (
              <Text color={r.fg} backgroundColor={r.bg}>
                {r.ch}
              </Text>
            ),
          )}
        </Box>
      ))
    const drawPet = (pixels: number) => (
      <Box key="pet" flexDirection="column" width={12} height={pixels / 2} flexShrink={0}>
        {cellRows(
          petCells(species, look.pose, petCoat(species, p.coat), p.faceLeft, look.over, pixels, pixels > PET_HEIGHT ? look.lift : 0),
        )}
      </Box>
    )

    // Above the prompt: text alone, in columns side by side
    if (placement === 'inline') {
      // The pet takes the leftmost columns, 12 and 2 to spare, and walks on the spot
      const petWidth = hasPet ? 14 : 0
      const cells = shown.map(({ item, view: v }) => {
        const value = inlineValue(v)
        return { label: item.label, value, width: segWidth(value) }
      })
      const rows = columnRows(
        cells.map(c => ({ label: c.label.length, value: c.width })),
        inner - petWidth,
      )
      const columns = chunk(cells, rows).map(list => ({
        list,
        label: Math.max(...list.map(c => c.label.length)) + 1,
        value: Math.max(...list.map(c => c.width)),
      }))
      const lines = Array.from({ length: Math.min(rows, cells.length) }, (_, r) => (
        <Box flexDirection="row">
          {columns.flatMap((col, i) => {
            const cell = col.list[r]
            if (cell === undefined) return []
            const isLast = i === columns.length - 1
            return [
              <Text color={C.blue}>{cell.label.padEnd(col.label)}</Text>,
              ...segs(cell.value),
              ...(isLast ? [] : [<Text>{' '.repeat(col.value - cell.width)}</Text>, <Text color={C.surface}>{COLUMN_GAP}</Text>]),
            ]
          })}
        </Box>
      ))
      return (
        <Box flexDirection="row" paddingX={1}>
          {hasPet ? (
            <Box flexDirection="column" flexShrink={0} marginRight={2}>
              {drawPet(PET_HEIGHT)}
            </Box>
          ) : null}
          <Box flexDirection="column">{lines}</Box>
        </Box>
      )
    }

    // Beside the transcript: a key column and bars for meters, one row after another
    const keyWidth = Math.max(LABEL, ...shown.map(({ item }) => item.label.length + 1))
    const barWidth = Math.max(4, inner - keyWidth - 5)
    // The key keeps its width: a value too long for the row (a path) would squeeze it, and its padding
    // would wrap to an empty second row
    const label = (text: string) => (
      <Box flexShrink={0}>
        <Text color={C.blue}>{text.padEnd(keyWidth)}</Text>
      </Box>
    )
    const dockRow = (key: string, v: View) => {
      if (v.meter !== undefined) {
        const b = bar(v.meter, barWidth, conf.bar)
        return (
          <Box flexDirection="row">
            {label(key)}
            <Text color={usageColor(v.meter)}>{b.filled}</Text>
            <Text color={C.surface}>{b.empty}</Text>
            <Text color={C.text}>{percentText(v.meter).padStart(5)}</Text>
          </Box>
        )
      }
      return (
        <Box flexDirection="row">
          {label(key)}
          {v.cutStart ? (
            <Text color={C.text} wrap="truncate-start">
              {v.value.map(g => g.text).join('')}
            </Text>
          ) : (
            segs(v.value)
          )}
        </Box>
      )
    }

    const rows = []
    // The pet over the rows: what it says, then the strip it walks along, its place along its walk spread
    // over the width there is
    if (hasPet) {
      rows.push(
        <Box flexDirection="row" height={1} flexShrink={0}>
          {p.line === undefined ? null : <Text color={C.text}>{p.line.slice(0, inner)}</Text>}
        </Box>,
        <Box flexDirection="row" height={(PET_HEIGHT + LIFT_ROOM) / 2} flexShrink={0}>
          <Box width={along(p.x)} flexShrink={0} />
          {drawPet(PET_HEIGHT + LIFT_ROOM)}
          {/* A ball over the strip, out of the flow: where a pet at its place would have its middle */}
          {ball === undefined ? null : (
            <Box
              key="ball"
              position="absolute"
              top={0}
              left={along(ball.pos) + BALL_LEFT}
              width={4}
              height={(PET_HEIGHT + LIFT_ROOM) / 2}
              flexDirection="column"
            >
              {cellRows(drawnCells(BALL, BALL_INK, PET_HEIGHT + LIFT_ROOM, ball.lift))}
            </Box>
          )}
        </Box>,
      )
    }
    for (const { item, view: v } of shown) {
      rows.push(dockRow(item.label, v))
      if (v.resets !== undefined) {
        rows.push(
          <Box flexDirection="row">
            {label('')}
            <Text color={C.flamingo}>{'resets '}</Text>
            <Text color={C.text}>{v.resets}</Text>
          </Box>,
        )
      }
    }
    return (
      <Box flexDirection="column" paddingX={1}>
        {rows}
      </Box>
    )
  })
}
