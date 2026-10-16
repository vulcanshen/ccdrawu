// Which items show where, from the config the TUI writes, and how they fall into columns above the prompt.
// None touches `$`, so tests can call them.

import type { Config } from '../types'
import { truncateStart } from './format'
import { C, ITEMS, NO_VALUE, type Item, type Seg, type View } from './items'

export type Placement = 'dock' | 'inline'

// ~/.config/ccdrawu/config.json: { "dock": [ids], "inline": [ids], "alias": name, "pet": { "species" },
// "bar": style, "drawer": { "sidebarOff", "inlineOff" }, "statusline": { "on", ... } }. Undefined when it is not JSON, as when caught half written:
// read it again on the next tick
export function parseConfig(text: string): Config | undefined {
  let raw: unknown
  try {
    raw = JSON.parse(text)
  } catch {
    return undefined
  }
  const config: Config = {}
  if (typeof raw !== 'object' || raw === null) return config
  for (const placement of ['dock', 'inline'] as const) {
    const ids = (raw as Record<string, unknown>)[placement]
    if (Array.isArray(ids)) config[placement] = ids.filter((id): id is string => typeof id === 'string')
  }
  const alias = (raw as Record<string, unknown>).alias
  if (typeof alias === 'string') config.alias = alias
  const style = (raw as Record<string, unknown>).bar
  if (typeof style === 'string') config.bar = style
  const pet = (raw as Record<string, unknown>).pet
  if (typeof pet === 'object' && pet !== null) {
    const { species } = pet as Record<string, unknown>
    config.pet = typeof species === 'string' ? { species } : {}
  }
  const drawer = (raw as Record<string, unknown>).drawer
  if (typeof drawer === 'object' && drawer !== null) {
    const d = drawer as Record<string, unknown>
    // 0.1.1's off turned the whole drawer off
    const off = d.off === true
    config.drawer = { sidebarOff: off || d.sidebarOff === true, inlineOff: off || d.inlineOff === true }
  }
  const statusline = (raw as Record<string, unknown>).statusline
  if (typeof statusline === 'object' && statusline !== null) {
    config.statusline = { on: (statusline as Record<string, unknown>).on === true }
  }
  return config
}

// In the order the config lists them (set in ccdrawu), the ones the catalog knows, each once; a placement
// the config leaves out takes the catalog's defaults, in its order
export function enabledItems(config: Config, placement: Placement): Item[] {
  const ids = config[placement]
  if (ids === undefined) return ITEMS.filter(item => item[placement])
  return [...new Set(ids)].flatMap(id => ITEMS.find(item => item.id === id) ?? [])
}

// Whether the config lets the drawer show beside the transcript (dock) or above the prompt (inline)
export function drawerShows(config: Config, placement: Placement): boolean {
  return placement === 'dock' ? config.drawer?.sidebarOff !== true : config.drawer?.inlineOff !== true
}

// /ccdrawu turns the drawer on or off for the session, wherever it sits, over the config; null until it
// does, and again once the config's Drawer changes (the last of the two wins)
export type Manual = 'on' | 'off' | null

// Whether the drawer belongs in a place: as /ccdrawu last set it for the session, else as the config has it
export function drawerWanted(config: Config, placement: Placement, manual: Manual): boolean {
  return manual === null ? drawerShows(config, placement) : manual === 'on'
}

// Whether an item shows beside the transcript or above the prompt
export function showsAnywhere(config: Config, id: string): boolean {
  return (['dock', 'inline'] as const).some(p => enabledItems(config, p).some(item => item.id === id))
}

const INLINE_PATH = 32 // above the prompt a path keeps its last 32 characters
export const COLUMN_GAP = ' │ '

// Above the prompt: no bars, what the sidebar puts below after the value, the reset countdown after
// the percentage, but not one not known: on one row it holds no row in place
export function inlineValue(v: View): Seg[] {
  const cut = v.cutStart ? v.value.map(s => ({ ...s, text: truncateStart(s.text, INLINE_PATH) })) : v.value
  const value = v.below === undefined ? cut : [...cut, { text: '  ', color: C.text }, ...v.below]
  if (v.resets === undefined || v.resets === NO_VALUE) return value
  return [...value, { text: ' · ', color: C.surface }, { text: v.resets, color: C.text }]
}

// Cells a text takes on the terminal: East Asian wide and fullwidth characters (CJK, the ｜ between
// task counts) take two, the rest one
export function cellWidth(text: string): number {
  let n = 0
  for (const ch of text) {
    const c = ch.codePointAt(0) ?? 0
    const wide =
      (c >= 0x1100 && c <= 0x115f) ||
      (c >= 0x2e80 && c <= 0xa4cf) ||
      (c >= 0xac00 && c <= 0xd7a3) ||
      (c >= 0xf900 && c <= 0xfaff) ||
      (c >= 0xfe30 && c <= 0xfe4f) ||
      (c >= 0xff00 && c <= 0xff60) ||
      (c >= 0xffe0 && c <= 0xffe6) ||
      (c >= 0x20000 && c <= 0x3fffd)
    n += wide ? 2 : 1
  }
  return n
}

export function segWidth(segs: readonly Seg[]): number {
  return segs.reduce((n, s) => n + cellWidth(s.text), 0)
}

// Above the prompt the items fill columns top to bottom, side by side: the fewest rows whose columns
// fit in `width`, each column a key (its longest plus a space) and a value; one column when none fit
export function columnRows(cells: readonly { label: number; value: number }[], width: number): number {
  for (let rows = 1; rows < cells.length; rows++) {
    let total = -COLUMN_GAP.length
    for (let i = 0; i < cells.length; i += rows) {
      const column = cells.slice(i, i + rows)
      total += COLUMN_GAP.length + Math.max(...column.map(c => c.label)) + 1 + Math.max(...column.map(c => c.value))
    }
    if (total <= width) return rows
  }
  return Math.max(1, cells.length)
}

export function chunk<T>(list: readonly T[], size: number): T[][] {
  const out: T[][] = []
  for (let i = 0; i < list.length; i += size) out.push(list.slice(i, i + size))
  return out
}
