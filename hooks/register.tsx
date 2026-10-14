import { atom, read, update } from 'claude-code'
import type { EngineInterface as Engine, Register } from 'claude-code'

import type { CacheInfo, Config, GitInfo, ModelInfo, Modes, Usage, VersionInfo } from '../types'
import { bar, commandNames, parseShortstat, percentText, shortPath, usageColor } from './format'
import { C, ITEMS, view, type Seg, type Snapshot, type View } from './items'
import { chunk, COLUMN_GAP, columnRows, enabledItems, inlineValue, parseConfig, segWidth } from './layout'
import { locations } from './paths'

const PANE = 'ccdrawu'
const PANE_OPEN = { id: PANE, title: 'Status', columns: 34 } as const
const LABEL = 9 // key column width beside the transcript; the longest key is "Thinking "

const usage = atom({ plugin: 'ccdrawu', key: 'usage' } as const, null as Usage | null)
const model = atom({ plugin: 'ccdrawu', key: 'model' } as const, null as ModelInfo | null)
const git = atom({ plugin: 'ccdrawu', key: 'git' } as const, null as GitInfo | null)
const path = atom({ plugin: 'ccdrawu', key: 'path' } as const, '')
const turns = atom({ plugin: 'ccdrawu', key: 'turns' } as const, 0)
const cache = atom({ plugin: 'ccdrawu', key: 'cache' } as const, null as CacheInfo | null)
const compacts = atom({ plugin: 'ccdrawu', key: 'compacts' } as const, 0)
const agents = atom({ plugin: 'ccdrawu', key: 'agents' } as const, 0)
const version = atom({ plugin: 'ccdrawu', key: 'version' } as const, null as VersionInfo | null)
const modes = atom({ plugin: 'ccdrawu', key: 'modes' } as const, {} as Modes)
const now = atom({ plugin: 'ccdrawu', key: 'now' } as const, 0)
const config = atom({ plugin: 'ccdrawu', key: 'config' } as const, {} as Config)

async function refreshUsage($: Engine) {
  const u = await $.session.usage()
  await update($, usage, () => ({
    startedAt: u.startedAt,
    contextPercent: u.context.percent,
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

// Whether vim is on is settings' editorMode
async function refreshEditor($: Engine) {
  const { editorMode } = (await $.settings.read()) as unknown as { editorMode?: string }
  await update($, modes, m => ({ ...m, editor: editorMode ?? 'normal' }))
}

// A mod cannot read vim's live state (INSERT / NORMAL). The status line bridge, statusline/statusline.sh,
// writes vim.mode to <data>/vim/<session id> each time the status line runs; this reads it
let home: string | undefined
let dataDir: string | undefined
let configFile: string | undefined
let vimFile: string | undefined

async function locateFiles($: Engine) {
  home = await $.env.get('HOME')
  if (home === undefined) return
  const found = locations({
    home,
    data: await $.env.get('CCDRAWU__DATA'),
    config: await $.env.get('CCDRAWU__CONFIG'),
    xdgConfig: await $.env.get('XDG_CONFIG_HOME'),
  })
  dataDir = found.data
  configFile = found.configFile
  const id = (await $.session.id()).replace(/\.jsonl$/, '')
  vimFile = `${dataDir}/vim/${id}`
}

// Write `text` to <data>/<name> unless it already holds it
async function keepFile($: Engine, name: string, text: string) {
  if (dataDir === undefined) return
  const target = `${dataDir}/${name}`
  const current = await $.fs.read(target).catch(() => undefined)
  if (current !== text) await $.fs.write(target, text)
}

// Keep the bridge at a fixed path, ~/.ccdrawu/statusline.sh, so a statusLine setting
// pointing at it survives plugin updates
async function installBridge($: Engine) {
  await keepFile($, 'statusline.sh', await $.fs.read(`${$.plugin.root}/statusline/statusline.sh`))
}

// The TUI lists what this version of the mod can show from catalog.json, so it needs no update of its own
async function installCatalog($: Engine) {
  await keepFile($, 'catalog.json', JSON.stringify({ version: 1, items: ITEMS }, null, 2) + '\n')
}

// The TUI writes ~/.config/ccdrawu/config.json; every open session looks at it each second and takes
// a change at once. A file not there means the catalog's defaults
let configStamp: number | undefined

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
}

async function refreshVim($: Engine) {
  if (vimFile === undefined) return
  const state = (await $.fs.read(vimFile)).trim() || undefined
  if ((await read($, modes)).vim === state) return
  await update($, modes, m => ({ ...m, vim: state }))
}

// No API reports the permission mode; Claude Code's own hook events carry permission_mode,
// so remember the main conversation's from them
async function rememberMode($: Engine, e: { permission_mode?: string; agent_id?: string }) {
  const mode = e.permission_mode
  if (e.agent_id !== undefined || mode === undefined) return
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
  await update($, git, () => ({ branch, ...parseShortstat(diff.stdout) }))
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

// A failed background refresh (git timing out, offline for the version check) affects its row alone
function quietly(work: Promise<unknown>) {
  void work.catch(() => undefined)
}

export const register: Register = (on, options) => {
  const names = commandNames(options.alias)

  on('session.start', async ($, e, next) => {
    for (const name of names) {
      const description =
        name === 'ccdrawu' ? 'Toggle the status drawer' : 'Toggle the status drawer (alias of /ccdrawu)'
      await $.command.register({ name, description, immediate: true })
    }

    await tick($)
    await locateFiles($).catch(() => undefined)
    quietly(installBridge($))
    quietly(installCatalog($))
    quietly(refreshConfig($))
    quietly(refreshUsage($))
    quietly(refreshModelName($))
    quietly(refreshEditor($))
    quietly(refreshGit($))
    quietly(refreshAgents($))
    quietly(refreshVersion($))
    quietly(refreshTurns($))

    // Countdowns and uptime every 30 seconds; git, path and subagents every 5; the latest version every 6 hours
    $.clock.every(30_000, () => quietly(tick($)))
    $.clock.every(5_000, () => {
      quietly(refreshGit($))
      quietly(refreshAgents($))
    })
    $.clock.every(6 * 60 * 60_000, () => quietly(refreshVersion($)))
    // vim: the status line already waits 300ms to batch changes, so reading every half second is enough
    $.clock.every(500, () => quietly(refreshVim($)))
    $.clock.every(1_000, () => quietly(refreshConfig($)))

    // In a terminal too narrow for a sidebar the engine holds the pane until it widens
    void $.ui.open(PANE_OPEN)

    return next(e)
  })

  // Every slash command passes through here, so a failure must leave other commands alone
  on('command.run', async ($, e, next) => {
    if (!names.includes(e.command)) return next(e)
    const isOpen = (await $.ui.panes()).some(p => p.id === PANE)
    if (isOpen) await $.ui.close({ id: PANE })
    else await $.ui.open(PANE_OPEN)
    return {}
  }).catch(($, e, next) => (names.includes(e.command) ? {} : next(e)))

  // Usage, context and cost changed (after each turn, or a rate limit moved)
  on('session.measure', async ($, e, next) => {
    const result = await next(e)
    quietly(refreshUsage($))
    return result
  })

  // Each request of the main conversation: its effort, and how much the prompt cache served
  on('turn.step', async function* ($, e, next) {
    const result = yield* next(e)
    if (e.agentId === undefined) {
      const effort = e.effort === undefined ? undefined : String(e.effort)
      await update($, model, m => ({ name: m?.name ?? e.model, effort }))
      const u = result.usage
      if (u) {
        const total = u.input_tokens + u.cache_read_input_tokens + u.cache_creation_input_tokens
        await update($, cache, () => ({ read: u.cache_read_input_tokens, total }))
      }
    }
    return result
  })

  on('turn.complete', async ($, e, next) => {
    const result = await next(e)
    if (e.agentId === undefined) {
      quietly(refreshTurns($))
      quietly(refreshModelName($))
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

  on('session.compact', async ($, e, next) => {
    const result = await next(e)
    if (e.agentId === undefined && e.trigger !== 'precompute' && !('skip' in result && result.skip)) {
      quietly(update($, compacts, n => n + 1))
    }
    return result
  }).catch(($, e, next) => next(e))

  // Permission mode: recorded at session start, on each prompt and when a reply ends
  // (after shift+tab it shows once the next of these comes)
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
    return result
  }).catch(($, e, next) => next(e))

  on('ui.render', { component: 'Pane', requestId: PANE }, async ($, e) => {
    const { Box, Text } = $.ui.resolve(e)
    // Leave out the paddingX on either side, or a row runs past the width and wraps
    const inner = Math.max(22, e.props.bodyColumns - 2)
    const placement = e.props.placement

    const s: Snapshot = {
      usage: await read($, usage),
      model: await read($, model),
      git: await read($, git),
      path: await read($, path),
      turns: await read($, turns),
      cache: await read($, cache),
      compacts: await read($, compacts),
      agents: await read($, agents),
      version: await read($, version),
      modes: await read($, modes),
      now: await read($, now),
    }
    const shown = enabledItems(await read($, config), placement).flatMap(item => {
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

    // Above the prompt: text alone, in columns side by side
    if (placement === 'inline') {
      const cells = shown.map(({ item, view: v }) => {
        const value = inlineValue(v)
        return { label: item.label, value, width: segWidth(value) }
      })
      const rows = columnRows(
        cells.map(c => ({ label: c.label.length, value: c.width })),
        inner,
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
        <Box flexDirection="column" paddingX={1}>
          {lines}
        </Box>
      )
    }

    // Beside the transcript: a key column, bars for meters, a blank row between groups
    const barWidth = Math.max(4, inner - LABEL - 5)
    const label = (text: string) => <Text color={C.blue}>{text.padEnd(LABEL)}</Text>
    const dockRow = (key: string, v: View) => {
      if (v.meter !== undefined) {
        const b = bar(v.meter, barWidth)
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
    let group: string | undefined
    for (const { item, view: v } of shown) {
      if (group !== undefined && item.group !== group) rows.push(<Text> </Text>)
      group = item.group
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
