import { atom, read, update } from 'claude-code'
import type { EngineInterface as Engine, Register } from 'claude-code'

import type { CacheInfo, GitInfo, ModelInfo, Modes, Usage, VersionInfo } from '../types'
import {
  bar,
  commandNames,
  countdown,
  duration,
  isNewer,
  limitLabel,
  modeLabel,
  parseShortstat,
  percentText,
  shortPath,
  usageColor,
} from './format'

const PANE = 'statusdrawer'
const PANE_OPEN = { id: PANE, title: 'Status', columns: 34 } as const
const LABEL = 9 // key column width; the longest key is "Thinking "

// Catppuccin Mocha: keys in blue, values in text.
// The pane's background is not drawn here: it is the theme's composerSidebarBackground.
const C = {
  text: '#cdd6f4',
  blue: '#89b4fa',
  lavender: '#b4befe',
  sky: '#89dceb',
  flamingo: '#f2cdcd',
  surface: '#45475a',
  green: '#a6e3a1',
  red: '#f38ba8',
  peach: '#fab387',
}

const usage = atom({ plugin: 'statusdrawer', key: 'usage' } as const, null as Usage | null)
const model = atom({ plugin: 'statusdrawer', key: 'model' } as const, null as ModelInfo | null)
const git = atom({ plugin: 'statusdrawer', key: 'git' } as const, null as GitInfo | null)
const path = atom({ plugin: 'statusdrawer', key: 'path' } as const, '')
const turns = atom({ plugin: 'statusdrawer', key: 'turns' } as const, 0)
const cache = atom({ plugin: 'statusdrawer', key: 'cache' } as const, null as CacheInfo | null)
const compacts = atom({ plugin: 'statusdrawer', key: 'compacts' } as const, 0)
const agents = atom({ plugin: 'statusdrawer', key: 'agents' } as const, 0)
const version = atom({ plugin: 'statusdrawer', key: 'version' } as const, null as VersionInfo | null)
const modes = atom({ plugin: 'statusdrawer', key: 'modes' } as const, {} as Modes)
const now = atom({ plugin: 'statusdrawer', key: 'now' } as const, 0)

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
// writes vim.mode to ~/.claude/statusdrawer/vim/<session id> each time the status line runs; this reads it
let home: string | undefined
let vimFile: string | undefined

async function locateFiles($: Engine) {
  home = await $.env.get('HOME')
  const id = (await $.session.id()).replace(/\.jsonl$/, '')
  vimFile = `${home}/.claude/statusdrawer/vim/${id}`
}

// Keep the bridge at a fixed path, ~/.claude/statusdrawer/statusline.sh, so a statusLine setting
// pointing at it survives plugin updates; rewritten only when the shipped copy changed
async function installBridge($: Engine) {
  if (home === undefined) return
  const shipped = await $.fs.read(`${$.plugin.root}/statusline/statusline.sh`)
  const target = `${home}/.claude/statusdrawer/statusline.sh`
  const current = await $.fs.read(target).catch(() => undefined)
  if (current !== shipped) await $.fs.write(target, shipped)
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
        name === 'sdrawer' ? 'Toggle the status drawer' : 'Toggle the status drawer (alias of /sdrawer)'
      await $.command.register({ name, description, immediate: true })
    }

    await tick($)
    await locateFiles($).catch(() => undefined)
    quietly(installBridge($))
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
    // Leave out the paddingX on either side, or a bar runs past the width and wraps to a second row
    const inner = Math.max(22, e.props.bodyColumns - 2)
    const barWidth = Math.max(4, inner - LABEL - 5)

    const u = await read($, usage)
    const m = await read($, model)
    const g = await read($, git)
    const p = await read($, path)
    const t = await read($, turns)
    const c = await read($, cache)
    const k = await read($, compacts)
    const a = await read($, agents)
    const v = await read($, version)
    const md = await read($, modes)
    const at = await read($, now)

    const label = (text: string) => <Text color={C.blue}>{text.padEnd(LABEL)}</Text>
    const meter = (name: string, percent: number) => {
      const b = bar(percent, barWidth)
      return (
        <Box flexDirection="row">
          {label(name)}
          <Text color={usageColor(percent)}>{b.filled}</Text>
          <Text color={C.surface}>{b.empty}</Text>
          <Text color={C.text}>{percentText(percent).padStart(5)}</Text>
        </Box>
      )
    }
    const row = (name: string, value: string, color: string = C.text) => (
      <Box flexDirection="row">
        {label(name)}
        <Text color={color}>{value}</Text>
      </Box>
    )
    const gap = () => <Text> </Text>

    const cachePercent = c && c.total > 0 ? (c.read / c.total) * 100 : undefined
    const editor = md.editor === 'vim' && md.vim ? `vim · ${md.vim}` : md.editor

    // No Fragments here: this surface lays a Fragment out as a column, so siblings go in a row Box
    return (
      <Box flexDirection="column" paddingX={1}>
        <Box flexDirection="row">
          {label('Model')}
          <Text color={C.lavender} bold>
            {m?.name ?? '…'}
          </Text>
        </Box>
        {m?.effort !== undefined && row('Thinking', m.effort, C.sky)}
        {md.permission !== undefined && row('Mode', modeLabel(md.permission))}
        {editor !== undefined && row('Editor', editor)}
        {gap()}
        {meter('Context', u?.contextPercent ?? 0)}
        {(u?.limits ?? []).map(l => (
          <Box flexDirection="column">
            {meter(limitLabel(l.kind), l.percentUsed)}
            {l.resetsAt !== undefined && (
              <Box flexDirection="row">
                {label('')}
                <Text color={C.flamingo}>{'resets '}</Text>
                <Text color={C.text}>{countdown(Date.parse(l.resetsAt) - at)}</Text>
              </Box>
            )}
          </Box>
        ))}
        {u?.costUsd !== undefined && row('Cost', '$' + u.costUsd.toFixed(2))}
        {gap()}
        {row('Uptime', (u ? duration(at - u.startedAt) : '…') + ` · ${t} turns`)}
        {cachePercent !== undefined && row('Cache', percentText(cachePercent))}
        {row('Compact', String(k))}
        {row('Agents', a > 0 ? `${a} running` : 'none')}
        {gap()}
        {g ? (
          <Box flexDirection="row">
            {label('Git')}
            <Text color={C.text}>{g.branch}</Text>
            <Text color={C.green}>{'  +' + g.added}</Text>
            <Text color={C.red}>{' -' + g.removed}</Text>
          </Box>
        ) : (
          row('Git', 'not a repo')
        )}
        <Box flexDirection="row">
          {label('Path')}
          <Text color={C.text} wrap="truncate-start">
            {p}
          </Text>
        </Box>
        {gap()}
        {v && (
          <Box flexDirection="row">
            {label('Version')}
            <Text color={C.text}>{v.current}</Text>
            {v.latest !== undefined && isNewer(v.latest, v.current) && (
              <Text color={C.peach}>{'  ↑ ' + v.latest}</Text>
            )}
          </Box>
        )}
      </Box>
    )
  })
}
