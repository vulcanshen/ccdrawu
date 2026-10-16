// The drawer's items: the catalog (written to ~/.ccdrawu/catalog.json for the TUI to list)
// and what each one shows. None touches `$`, so tests can call them.

import type { Background, CacheInfo, GitInfo, Limit, ModelInfo, Modes, ServiceStatus, SessionInfo, SkillsInfo, TokenInfo, Usage, VersionInfo } from '../types'
import { compact, countdown, directoryName, duration, effortColor, filesText, isNewer, modeLabel, percentText, syncText, tasksText, usageColor } from './format'

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
  label: string // the key column in the pane: 9 characters at most, and 8 for all but Directory
  short: string // the label in the statusline when its format is short: 4 characters at most
  name: string // what the TUI lists
  group: string // the TUI lists the items under these, in catalog order
  description: string
  dock: boolean // on by default beside the transcript
  inline: boolean // on by default above the prompt
  statusline: boolean // on by default in `ccdrawu statusline`
}

// The order here is the order on screen
export const ITEMS: readonly Item[] = [
  // Drawn by the pane itself (pets.ts), over the rows beside the transcript and left of the columns above the prompt
  { id: 'pet', label: 'Pet', short: 'pet', name: 'Pet', group: 'Pet', description: 'A pixel pet: it walks while Claude works, sits when Claude is done, sleeps when nothing goes on; /ccdrawu pat pats it', dock: true, inline: true, statusline: false },
  { id: 'model', label: 'Model', short: 'mdl', name: 'Model', group: 'Model', description: 'The model answering in this session', dock: true, inline: true, statusline: true },
  { id: 'effort', label: 'Effort', short: 'eff', name: 'Effort level', group: 'Model', description: "The last request's effort level, green for low to red for max", dock: true, inline: true, statusline: false },
  { id: 'mode', label: 'Mode', short: 'mode', name: 'Permission mode', group: 'Model', description: 'auto, plan, accept edits, ...; after shift+tab it shows once Claude finishes the tools of a step, or with the next prompt', dock: true, inline: true, statusline: false },
  { id: 'style', label: 'Style', short: 'sty', name: 'Output style', group: 'Model', description: 'The output style set in /config', dock: false, inline: false, statusline: false },
  { id: 'id', label: 'ID', short: 'id', name: 'Session ID', group: 'Session', description: "The start of this session's id", dock: false, inline: false, statusline: false },
  { id: 'account', label: 'Account', short: 'acct', name: 'Account', group: 'Session', description: 'The Claude account signed in', dock: false, inline: false, statusline: false },
  { id: 'context', label: 'Context', short: 'ctx', name: 'Context', group: 'Usage', description: 'How full the context window is', dock: true, inline: true, statusline: true },
  { id: 'window', label: 'Window', short: 'win', name: 'Context size', group: 'Usage', description: 'Tokens in the context window, out of its size', dock: false, inline: false, statusline: false },
  { id: 'session', label: 'Session', short: '5h', name: 'Session limit', group: 'Usage', description: 'The 5-hour usage limit, and when it resets', dock: true, inline: true, statusline: true },
  { id: 'weekly', label: 'Weekly', short: '7d', name: 'Weekly limit', group: 'Usage', description: 'The 7-day usage limit, and when it resets', dock: true, inline: true, statusline: true },
  { id: 'spend', label: 'Spend', short: 'spnd', name: 'Spend limit', group: 'Usage', description: "A gateway's spend limit, where there is one", dock: true, inline: true, statusline: false },
  { id: 'cost', label: 'Cost', short: 'cost', name: 'Session cost', group: 'Usage', description: 'What this session has cost so far, as /cost totals it', dock: true, inline: true, statusline: true },
  { id: 'rate', label: 'Rate', short: 'rate', name: 'Cost rate', group: 'Usage', description: 'What the session costs an hour, so far', dock: false, inline: false, statusline: false },
  { id: 'tokens', label: 'Tokens', short: 'tok', name: 'Tokens', group: 'Usage', description: 'Input, output and cache tokens of the main conversation since the drawer loaded', dock: false, inline: false, statusline: false },
  { id: 'speed', label: 'Speed', short: 'spd', name: 'Token speed', group: 'Usage', description: 'Output tokens a second in the last reply', dock: false, inline: false, statusline: false },
  { id: 'uptime', label: 'Uptime', short: 'up', name: 'Uptime and turns', group: 'Activity', description: 'How long the session has run, and how many turns', dock: true, inline: true, statusline: false },
  { id: 'cache', label: 'Cache', short: 'hit', name: 'Cache hit rate', group: 'Activity', description: 'How much of the last request the prompt cache served', dock: true, inline: true, statusline: false },
  { id: 'compact', label: 'Compact', short: 'cmp', name: 'Compactions', group: 'Activity', description: 'How many times the context was compacted', dock: true, inline: true, statusline: false },
  { id: 'tasks', label: 'Tasks', short: 'bg', name: 'Background tasks', group: 'Activity', description: "Subagents running now, and shells, monitors and workflows in the background as of Claude's last reply", dock: true, inline: true, statusline: false },
  { id: 'skills', label: 'Skills', short: 'skl', name: 'Skills', group: 'Activity', description: 'Skills run in this session, and the last one', dock: false, inline: false, statusline: false },
  { id: 'git', label: 'Git', short: 'git', name: 'Git branch and changes', group: 'Workspace', description: 'The branch, and lines added and removed since HEAD', dock: true, inline: true, statusline: true },
  { id: 'sync', label: 'Sync', short: 'sync', name: 'Ahead and behind', group: 'Workspace', description: 'Commits ahead of and behind the upstream', dock: false, inline: false, statusline: false },
  { id: 'files', label: 'Files', short: 'chg', name: 'Changed files', group: 'Workspace', description: 'Files staged, changed and new', dock: false, inline: false, statusline: false },
  { id: 'commit', label: 'Commit', short: 'sha', name: 'Commit', group: 'Workspace', description: 'The short hash of HEAD', dock: false, inline: false, statusline: false },
  { id: 'repo', label: 'Repo', short: 'repo', name: 'Repository', group: 'Workspace', description: "origin's owner/repo, and the worktree when the session is in one", dock: false, inline: false, statusline: false },
  { id: 'path', label: 'Path', short: 'path', name: 'Working directory', group: 'Workspace', description: "The session's working directory", dock: true, inline: true, statusline: false },
  { id: 'directory', label: 'Directory', short: 'dir', name: 'Directory name', group: 'Workspace', description: "The working directory's own name, without the path to it", dock: false, inline: false, statusline: false },
  { id: 'version', label: 'Claude', short: 'cc', name: 'Claude Code version', group: 'About', description: 'This version, and the latest when a newer one is out', dock: true, inline: true, statusline: false },
  { id: 'ccdrawu', label: 'ccdrawu', short: 'drw', name: 'ccdrawu version', group: 'About', description: "This drawer's own version", dock: true, inline: true, statusline: false },
  { id: 'status', label: 'Status', short: 'stat', name: 'Claude status', group: 'About', description: 'status.claude.com: all clear in green, an incident in peach or red', dock: false, inline: false, statusline: false },
]

// Everything an item may show, as the drawer holds it
export type Snapshot = {
  usage: Usage | null
  model: ModelInfo | null
  git: GitInfo | null
  path: string
  turns: number
  cache: CacheInfo | null
  tokens: TokenInfo
  session: SessionInfo
  status: ServiceStatus | null
  skills: SkillsInfo
  compacts: number
  agents: number
  background: Background // by type, as of Claude's last reply
  version: VersionInfo | null
  drawer: string | null // ccdrawu's own version, from its plugin.json
  modes: Modes
  now: number
}

export type Seg = { text: string; color: string; bold?: boolean }

// What an item shows. Beside the transcript a meter draws a bar, `resets` and `below` take a row of
// their own (so a long value wraps whole) and `cutStart` keeps the end of a value too long for the row;
// above the prompt all is text on one row
export type View = { value: Seg[]; meter?: number; resets?: string; below?: Seg[]; cutStart?: boolean }

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
        value: [{ text: s.git.branch, color: C.text }],
        below: [
          { text: '+' + s.git.added, color: C.green },
          { text: ' -' + s.git.removed, color: C.red },
        ],
      }
    case 'path':
      return { ...text(s.path), cutStart: true }
    case 'directory':
      return text(directoryName(s.path))
    case 'version': {
      const v = s.version
      if (!v) return undefined
      const newer = v.latest !== undefined && isNewer(v.latest, v.current)
      return { value: [{ text: v.current, color: C.text }, ...(newer ? [{ text: '  ↑ ' + v.latest, color: C.peach }] : [])] }
    }
    case 'ccdrawu':
      return s.drawer ? text(s.drawer) : undefined
    case 'style':
      return s.session.outputStyle ? text(s.session.outputStyle) : undefined
    case 'id':
      return s.session.id ? text(s.session.id.slice(0, 8)) : undefined
    case 'account':
      return s.session.account ? text(s.session.account) : undefined
    case 'window': {
      const u = s.usage
      if (u?.contextWindow === undefined) return undefined
      return text(`${compact(u.contextTokens ?? 0)} / ${compact(u.contextWindow)}`)
    }
    case 'rate': {
      // From a minute on: before that one request makes the hourly figure jump about
      const hours = s.usage ? (s.now - s.usage.startedAt) / 3_600_000 : 0
      if (s.usage?.costUsd === undefined || hours < 1 / 60) return undefined
      return text(`$${(s.usage.costUsd / hours).toFixed(2)}/h`)
    }
    case 'tokens': {
      const t = s.tokens
      if (t.input + t.output + t.cacheRead + t.cacheWrite === 0) return undefined
      return text(`in ${compact(t.input)} · out ${compact(t.output)} · cache ${compact(t.cacheRead + t.cacheWrite)}`)
    }
    case 'speed':
      return s.tokens.speed === undefined ? undefined : text(`${Math.round(s.tokens.speed)} tok/s`)
    case 'skills':
      return text(s.skills.count === 0 ? 'none' : `${s.skills.count}` + (s.skills.last ? ` · last ${s.skills.last}` : ''))
    case 'sync':
      if (!s.git) return text('not a repo')
      return text(s.git.ahead === undefined ? 'no upstream' : syncText(s.git.ahead, s.git.behind ?? 0))
    case 'files':
      if (!s.git) return text('not a repo')
      return s.git.staged === undefined ? undefined : text(filesText(s.git.staged, s.git.unstaged ?? 0, s.git.untracked ?? 0))
    case 'commit':
      if (!s.git) return text('not a repo')
      return text(s.git.sha ?? 'no commits')
    case 'repo': {
      const g = s.git
      if (!g) return text('not a repo')
      if (!g.repo && !g.worktree) return text('no origin')
      return text([g.repo, g.worktree].filter(Boolean).join(' · '))
    }
    case 'status': {
      if (!s.status) return undefined
      const color = s.status.indicator === 'none' ? C.green : s.status.indicator === 'minor' ? C.peach : C.red
      return text(s.status.description, color)
    }
  }
  return undefined
}
