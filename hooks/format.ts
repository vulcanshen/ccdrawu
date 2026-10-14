// Pure helpers for the drawer: times, bars, paths, git stats, versions. None touches `$`, so tests can call them.

const MINUTE = 60_000
const HOUR = 60 * MINUTE
const DAY = 24 * HOUR

// Time left: 3d4h, 4h31m, 12m; under a minute (or past) reads "now"
export function countdown(ms: number): string {
  if (ms < MINUTE) return 'now'
  const d = Math.floor(ms / DAY)
  const h = Math.floor((ms % DAY) / HOUR)
  const m = Math.floor((ms % HOUR) / MINUTE)
  if (d > 0) return `${d}d${h}h`
  if (h > 0) return `${h}h${String(m).padStart(2, '0')}m`
  return `${m}m`
}

// Time elapsed, shaped like countdown; under a minute reads "0m"
export function duration(ms: number): string {
  return ms < MINUTE ? '0m' : countdown(ms)
}

// A bar's filled and empty parts, `width` cells in all
export function bar(percent: number, width: number): { filled: string; empty: string } {
  const n = Math.round((Math.min(100, Math.max(0, percent)) / 100) * width)
  return { filled: '█'.repeat(n), empty: '░'.repeat(width - n) }
}

export function percentText(percent: number): string {
  return `${Math.round(percent)}%`
}

// Usage color (Catppuccin Mocha): green under 50%, yellow under 70%, peach under 90%, red from 90%
export function usageColor(percent: number): string {
  if (percent < 50) return '#a6e3a1'
  if (percent < 70) return '#f9e2af'
  if (percent < 90) return '#fab387'
  return '#f38ba8'
}

const MODE_LABELS: Record<string, string> = {
  default: 'default',
  acceptEdits: 'accept edits',
  bypassPermissions: 'bypass',
  plan: 'plan',
  dontAsk: "don't ask",
  auto: 'auto',
}

// A permission mode's display name; one it does not know is shown as it is
export function modeLabel(mode: string): string {
  return MODE_LABELS[mode] ?? mode
}

export function limitLabel(kind: string): string {
  if (kind === 'five_hour') return 'Session'
  if (kind === 'seven_day') return 'Weekly'
  return kind
}

export function shortPath(path: string, home: string | undefined): string {
  if (home && (path === home || path.startsWith(home + '/'))) return '~' + path.slice(home.length)
  return path
}

// `git diff --shortstat` prints " 3 files changed, 12 insertions(+), 3 deletions(-)"
export function parseShortstat(text: string): { added: number; removed: number } {
  const added = /(\d+) insertions?\(\+\)/.exec(text)
  const removed = /(\d+) deletions?\(-\)/.exec(text)
  return { added: added ? Number(added[1]) : 0, removed: removed ? Number(removed[1]) : 0 }
}

// Whether `latest` is newer than `current`, comparing numeric parts: 2.1.299 < 2.1.300
export function isNewer(latest: string, current: string): boolean {
  const a = latest.split('.').map(n => parseInt(n, 10) || 0)
  const b = current.split('.').map(n => parseInt(n, 10) || 0)
  for (let i = 0; i < Math.max(a.length, b.length); i++) {
    const x = a[i] ?? 0
    const y = b[i] ?? 0
    if (x !== y) return x > y
  }
  return false
}

// The command is always /sdrawer; the alias setting adds one more name (a leading slash is dropped)
export function commandNames(alias: unknown): string[] {
  const extra = typeof alias === 'string' ? alias.trim().replace(/^\/+/, '') : ''
  if (extra === '' || extra === 'sdrawer' || !/^[a-z0-9][a-z0-9-]*$/i.test(extra)) return ['sdrawer']
  return ['sdrawer', extra]
}
