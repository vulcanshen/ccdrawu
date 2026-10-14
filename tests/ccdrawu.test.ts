import { expect, test } from 'claude-code/testing'

import {
  bar,
  commandNames,
  countdown,
  duration,
  isNewer,
  modeLabel,
  parseShortstat,
  shortPath,
  truncateStart,
  usageColor,
} from '../hooks/format'
import { registerCommands } from '../hooks/commands'
import { ITEMS, view, type Snapshot } from '../hooks/items'
import { columnRows, enabledItems, inlineValue, parseConfig } from '../hooks/layout'
import { locations } from '../hooks/paths'

const PANE = {
  title: 'Status',
  isFocused: false,
  bodyColumns: 32,
  placement: 'dock' as const,
  scroll: { offset: 0, bodyRows: 30 },
  view: {},
}

const MOUNT = {
  plugin: 'ccdrawu',
  surface: 'terminal',
  component: 'Pane',
  requestId: 'ccdrawu',
  props: PANE,
} as const

const FULL: Snapshot = {
  usage: {
    startedAt: 0,
    contextPercent: 42,
    limits: [
      { kind: 'five_hour', percentUsed: 18, resetsAt: new Date(3_600_000).toISOString() },
      { kind: 'seven_day', percentUsed: 40 },
      { kind: 'spend_limit', percentUsed: 5 },
    ],
    costUsd: 1.2,
  },
  model: { name: 'Opus 5.5', effort: 'high' },
  git: { branch: 'main', added: 12, removed: 3 },
  path: '~/sideproj/ccdrawu',
  turns: 14,
  cache: { read: 92, total: 100 },
  compacts: 1,
  agents: 2,
  version: { current: '2.1.299', latest: '2.1.300' },
  modes: { permission: 'auto', editor: 'vim', vim: 'INSERT' },
  now: 60_000,
}

test('reset countdowns read as days, hours or minutes', async () => {
  expect(countdown(3 * 86_400_000 + 4 * 3_600_000 + 5 * 60_000)).toBe('3d4h')
  expect(countdown(4 * 3_600_000 + 31 * 60_000)).toBe('4h31m')
  expect(countdown(12 * 60_000)).toBe('12m')
  expect(countdown(30_000)).toBe('now')
  expect(duration(30_000)).toBe('0m')
})

test('the bar fills in proportion and keeps its width', async () => {
  expect(bar(50, 10)).toEqual({ filled: '█████', empty: '░░░░░' })
  expect(bar(130, 4)).toEqual({ filled: '████', empty: '' })
  expect(bar(-5, 4)).toEqual({ filled: '', empty: '░░░░' })
})

test('usage turns green, yellow, peach, then red', async () => {
  expect(usageColor(49)).toBe('#a6e3a1')
  expect(usageColor(50)).toBe('#f9e2af')
  expect(usageColor(69)).toBe('#f9e2af')
  expect(usageColor(70)).toBe('#fab387')
  expect(usageColor(89)).toBe('#fab387')
  expect(usageColor(90)).toBe('#f38ba8')
})

test('git shortstat, paths, modes and versions are read as shown', async () => {
  expect(parseShortstat(' 3 files changed, 12 insertions(+), 3 deletions(-)\n')).toEqual({ added: 12, removed: 3 })
  expect(parseShortstat(' 1 file changed, 1 insertion(+)\n')).toEqual({ added: 1, removed: 0 })
  expect(parseShortstat('')).toEqual({ added: 0, removed: 0 })
  expect(shortPath('/Users/me/proj', '/Users/me')).toBe('~/proj')
  expect(shortPath('/Users/meow', '/Users/me')).toBe('/Users/meow')
  expect(truncateStart('~/sideproj/ccdrawu', 10)).toBe('…j/ccdrawu')
  expect(truncateStart('~/proj', 10)).toBe('~/proj')
  expect(modeLabel('auto')).toBe('auto')
  expect(modeLabel('acceptEdits')).toBe('accept edits')
  expect(modeLabel('somethingNew')).toBe('somethingNew')
  expect(isNewer('2.1.300', '2.1.299')).toBe(true)
  expect(isNewer('2.1.299', '2.1.299')).toBe(false)
  expect(isNewer('2.1.29', '2.1.299')).toBe(false)
})

test('the alias adds one command name and ignores what cannot be one', async () => {
  expect(commandNames(undefined)).toEqual(['ccdrawu'])
  expect(commandNames('')).toEqual(['ccdrawu'])
  expect(commandNames('ccd')).toEqual(['ccdrawu', 'ccd'])
  expect(commandNames(' /ccd ')).toEqual(['ccdrawu', 'ccd'])
  expect(commandNames('ccdrawu')).toEqual(['ccdrawu'])
  expect(commandNames('two words')).toEqual(['ccdrawu'])
})

test('an alias in the config registers one more command, and a name once registered stays', async () => {
  const registered: string[] = []
  const register = async (spec: { name: string }) => registered.push(spec.name)
  const commands = new Set<string>()
  await registerCommands(register, commands, undefined)
  await registerCommands(register, commands, 'ccd')
  await registerCommands(register, commands, 'ccd')
  await registerCommands(register, commands, '')
  await registerCommands(register, commands, 'cd2')
  expect(registered).toEqual(['ccdrawu', 'ccd', 'cd2'])
  expect([...commands]).toEqual(['ccdrawu', 'ccd', 'cd2'])
})

test('every catalog item has a unique id, a short key and something to show', async () => {
  expect(new Set(ITEMS.map(i => i.id)).size).toBe(ITEMS.length)
  for (const item of ITEMS) {
    expect(item.label.length).toBeLessThanOrEqual(8)
    expect(view(item.id, FULL)).toBeDefined()
  }
  expect(view('no-such-item', FULL)).toBeUndefined()
})

test('files sit where terminu design E3 puts them, unless CCDRAWU__ variables say otherwise', async () => {
  expect(locations({ home: '/h' })).toEqual({ data: '/h/.ccdrawu', configFile: '/h/.config/ccdrawu/config.json' })
  expect(locations({ home: '/h', xdgConfig: '/x', data: '' })).toEqual({ data: '/h/.ccdrawu', configFile: '/x/ccdrawu/config.json' })
  expect(locations({ home: '/h', xdgConfig: '/x', config: '/c', data: '/d' })).toEqual({ data: '/d', configFile: '/c/config.json' })
})

test('the config picks items per placement, in the catalog order', async () => {
  expect(parseConfig('{"dock": ["git", "model", 3], "inline": "x"}')).toEqual({ dock: ['git', 'model'] })
  expect(parseConfig('{"dock": [')).toBeUndefined()
  expect(parseConfig('null')).toEqual({})
  expect(parseConfig('{"dock": null, "alias": "ccd"}')).toEqual({ alias: 'ccd' })
  expect(parseConfig('{"alias": 3}')).toEqual({})
  expect(enabledItems({ dock: ['git', 'model', 'gone'] }, 'dock').map(i => i.id)).toEqual(['model', 'git'])
  expect(enabledItems({ dock: [] }, 'dock')).toEqual([])
  expect(enabledItems({}, 'inline').map(i => i.id)).toEqual(ITEMS.filter(i => i.inline).map(i => i.id))
})

test('above the prompt a limit reads as text with its countdown, and columns take the fewest rows that fit', async () => {
  const session = view('session', FULL)!
  expect(inlineValue(session).map(s => s.text)).toEqual(['18%', ' · ', '59m'])
  expect(inlineValue(view('path', { ...FULL, path: '~/' + 'x'.repeat(40) })!)[0]!.text).toHaveLength(32)

  // Four cells of key 5 and value 5: one row needs 4 * 11 + 3 * 3 = 53 cells
  const cells = Array.from({ length: 4 }, () => ({ label: 5, value: 5 }))
  expect(columnRows(cells, 53)).toBe(1)
  expect(columnRows(cells, 52)).toBe(2)
  expect(columnRows(cells, 10)).toBe(4)
  expect(columnRows([], 10)).toBe(1)
})

test('the pane draws on the terminal before any figures arrive, a key on every row', async $ => {
  const ui = await $.ui.mount(MOUNT)
  expect(await ui.find({ type: 'Text', text: /not a repo/ })).toBeDefined()
  for (const key of ['Model', 'Context', 'Uptime', 'Git', 'Path']) {
    expect(await ui.find({ type: 'Text', text: new RegExp('^' + key + ' ') })).toBeDefined()
  }
  expect(await ui.find({ type: 'Text', text: /█|░/ })).toBeDefined()
  expect((await ui.find({ type: 'Box' }))?.props.backgroundColor).toBeUndefined()
  await ui.unmount()
})

test('above the prompt the pane draws columns of text with no bars', async $ => {
  const ui = await $.ui.mount({ ...MOUNT, props: { ...PANE, placement: 'inline', bodyColumns: 100 } })
  expect(await ui.find({ type: 'Text', text: /│/ })).toBeDefined()
  expect(await ui.find({ type: 'Text', text: /█|░/ })).toBeUndefined()
  expect(await ui.find({ type: 'Text', text: /^Context / })).toBeDefined()
  await ui.unmount()
})
