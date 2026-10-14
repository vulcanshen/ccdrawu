// The drawer's items: the catalog (written to ~/.ccdrawu/catalog.json for the TUI to list)
// and what each one shows. None touches `$`, so tests can call them.

import type { Background, CacheInfo, GitInfo, Limit, ModelInfo, Modes, Usage, VersionInfo } from '../types'
import { countdown, duration, effortColor, isNewer, modeLabel, percentText, tasksText, usageColor } from './format'

// Catppuccin Mocha: keys in blue, values in text.
// The pane's background is not drawn here: it is the theme's composerSidebarBackground.
export const C = {
  text: '#cdd6f4',
  blue: '#89b4fa',
  lavender: '#b4befe',
  flamingo: '#f2cdcd',
  surface: '#45475a',
  green: '#a6e3a1',
  red: '#f38ba8',
  peach: '#fab387',
}

export type Item = {
  id: string
  label: string // the key column in the pane: 8 characters at most
  name: string // what the TUI lists
  group: string // the TUI lists the items under these, in catalog order
  description: string
  dock: boolean // on by default beside the transcript
  inline: boolean // on by default above the prompt
}

// The order here is the order on screen
export const ITEMS: readonly Item[] = [
  { id: 'model', label: 'Model', name: 'Model', group: 'Model', description: 'The model answering in this session', dock: true, inline: true },
  { id: 'effort', label: 'Effort', name: 'Effort level', group: 'Model', description: "The last request's effort level, green for low to red for max", dock: true, inline: true },
  { id: 'mode', label: 'Mode', name: 'Permission mode', group: 'Model', description: 'auto, plan, accept edits, ...; after shift+tab it shows once Claude finishes the tools of a step, or with the next prompt', dock: true, inline: true },
  { id: 'context', label: 'Context', name: 'Context', group: 'Usage', description: 'How full the context window is', dock: true, inline: true },
  { id: 'session', label: 'Session', name: 'Session limit', group: 'Usage', description: 'The 5-hour usage limit, and when it resets', dock: true, inline: true },
  { id: 'weekly', label: 'Weekly', name: 'Weekly limit', group: 'Usage', description: 'The 7-day usage limit, and when it resets', dock: true, inline: true },
  { id: 'spend', label: 'Spend', name: 'Spend limit', group: 'Usage', description: "A gateway's spend limit, where there is one", dock: true, inline: true },
  { id: 'cost', label: 'Cost', name: 'Session cost', group: 'Usage', description: 'What this session has cost so far, as /cost totals it', dock: true, inline: true },
  { id: 'uptime', label: 'Uptime', name: 'Uptime and turns', group: 'Activity', description: 'How long the session has run, and how many turns', dock: true, inline: true },
  { id: 'cache', label: 'Cache', name: 'Cache hit rate', group: 'Activity', description: 'How much of the last request the prompt cache served', dock: true, inline: true },
  { id: 'compact', label: 'Compact', name: 'Compactions', group: 'Activity', description: 'How many times the context was compacted', dock: true, inline: true },
  { id: 'tasks', label: 'Tasks', name: 'Background tasks', group: 'Activity', description: "Subagents running now, and shells, monitors and workflows in the background as of Claude's last reply", dock: true, inline: true },
  { id: 'git', label: 'Git', name: 'Git branch and changes', group: 'Workspace', description: 'The branch, and lines added and removed since HEAD', dock: true, inline: true },
  { id: 'path', label: 'Path', name: 'Working directory', group: 'Workspace', description: "The session's working directory", dock: true, inline: true },
  { id: 'version', label: 'Claude', name: 'Claude Code version', group: 'Version', description: 'This version, and the latest when a newer one is out', dock: true, inline: true },
  { id: 'ccdrawu', label: 'ccdrawu', name: 'ccdrawu version', group: 'Version', description: "This drawer's own version", dock: true, inline: true },
]

// Everything an item may show, as the drawer holds it
export type Snapshot = {
  usage: Usage | null
  model: ModelInfo | null
  git: GitInfo | null
  path: string
  turns: number
  cache: CacheInfo | null
  compacts: number
  agents: number
  background: Background // by type, as of Claude's last reply
  version: VersionInfo | null
  drawer: string | null // ccdrawu's own version, from its plugin.json
  modes: Modes
  now: number
}

export type Seg = { text: string; color: string; bold?: boolean }

// What an item shows. Beside the transcript a meter draws a bar, `resets` takes a row of its own
// and `cutStart` keeps the end of a value too long for the row; above the prompt all is text on one row
export type View = { value: Seg[]; meter?: number; resets?: string; cutStart?: boolean }

function limitView(limit: Limit | undefined, now: number): View | undefined {
  if (limit === undefined) return undefined
  return {
    value: [{ text: percentText(limit.percentUsed), color: usageColor(limit.percentUsed) }],
    meter: limit.percentUsed,
    resets: limit.resetsAt === undefined ? undefined : countdown(Date.parse(limit.resetsAt) - now),
  }
}

// An item's view, or undefined while it has nothing to show (or for an id the catalog does not have)
export function view(id: string, s: Snapshot): View | undefined {
  const text = (value: string, color: string = C.text): View => ({ value: [{ text: value, color }] })
  const limit = (kind: string) => limitView(s.usage?.limits.find(l => l.kind === kind), s.now)

  switch (id) {
    case 'model':
      return { value: [{ text: s.model?.name ?? '…', color: C.lavender, bold: true }] }
    case 'effort':
      return s.model?.effort === undefined ? undefined : text(s.model.effort, effortColor(s.model.effort))
    case 'mode':
      return s.modes.permission === undefined ? undefined : text(modeLabel(s.modes.permission))
    case 'context': {
      const percent = s.usage?.contextPercent ?? 0
      return { value: [{ text: percentText(percent), color: usageColor(percent) }], meter: percent }
    }
    case 'session':
      return limit('five_hour')
    case 'weekly':
      return limit('seven_day')
    case 'spend':
      return limit('spend_limit')
    case 'cost':
      return s.usage?.costUsd === undefined ? undefined : text('$' + s.usage.costUsd.toFixed(2))
    case 'uptime':
      return text((s.usage ? duration(s.now - s.usage.startedAt) : '…') + ` · ${s.turns} turns`)
    case 'cache':
      return s.cache && s.cache.total > 0 ? text(percentText((s.cache.read / s.cache.total) * 100)) : undefined
    case 'compact':
      return text(String(s.compacts))
    case 'tasks':
      return text(tasksText(s.agents, s.background))
    case 'git':
      if (!s.git) return text('not a repo')
      return {
        value: [
          { text: s.git.branch, color: C.text },
          { text: '  +' + s.git.added, color: C.green },
          { text: ' -' + s.git.removed, color: C.red },
        ],
      }
    case 'path':
      return { ...text(s.path), cutStart: true }
    case 'version': {
      const v = s.version
      if (!v) return undefined
      const newer = v.latest !== undefined && isNewer(v.latest, v.current)
      return { value: [{ text: v.current, color: C.text }, ...(newer ? [{ text: '  ↑ ' + v.latest, color: C.peach }] : [])] }
    }
    case 'ccdrawu':
      return s.drawer ? text(s.drawer) : undefined
  }
  return undefined
}
