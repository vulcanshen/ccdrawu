// What `ccdrawu statusline` lays out: each item's value on one row, as it shows above the prompt, kept in
// <cache>/sessions/<session id>.json while the config has the statusline on. The command picks the items,
// their order, the labels and the separator; the values and their colors are worked out here alone.
// None touches `$`, so tests can call them.

import { tasksText } from './format'
import { C, ITEMS, view, type Seg, type Snapshot } from './items'
import { inlineValue } from './layout'

// Every item, the pet aside: it draws nothing on one row of text. Tasks part
// their counts with ", " here, as the statusline parts its items with |
export function statuslineItems(s: Snapshot): Record<string, Seg[]> {
  const items: Record<string, Seg[]> = {}
  for (const item of ITEMS) {
    if (item.id === 'pet') continue
    const v = item.id === 'tasks' ? { value: [{ text: tasksText(s.agents, s.background, ', '), color: C.text }] } : view(item.id, s)
    if (v) items[item.id] = inlineValue(v)
  }
  return items
}

// The file's text, with no time in it, so a session whose figures stand still leaves the file alone.
// docked: whether the drawer shows beside the transcript now, for a statusline on only when it does not
export function statuslineFile(s: Snapshot, docked: boolean): string {
  return JSON.stringify({ version: 1, docked, items: statuslineItems(s) }) + '\n'
}
