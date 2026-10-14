import { expect, test } from 'claude-code/testing'

import {
  bar,
  commandNames,
  countdown,
  duration,
  isNewer,
  limitLabel,
  modeLabel,
  parseShortstat,
  shortPath,
  usageColor,
} from '../hooks/format'

const PANE = {
  title: 'Status',
  isFocused: false,
  bodyColumns: 32,
  placement: 'dock' as const,
  scroll: { offset: 0, bodyRows: 30 },
  view: {},
}

const MOUNT = {
  plugin: 'statusdrawer',
  surface: 'terminal',
  component: 'Pane',
  requestId: 'statusdrawer',
  props: PANE,
} as const

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

test('git shortstat, paths, modes, limits and versions are read as shown', async () => {
  expect(parseShortstat(' 3 files changed, 12 insertions(+), 3 deletions(-)\n')).toEqual({ added: 12, removed: 3 })
  expect(parseShortstat(' 1 file changed, 1 insertion(+)\n')).toEqual({ added: 1, removed: 0 })
  expect(parseShortstat('')).toEqual({ added: 0, removed: 0 })
  expect(shortPath('/Users/me/proj', '/Users/me')).toBe('~/proj')
  expect(shortPath('/Users/meow', '/Users/me')).toBe('/Users/meow')
  expect(limitLabel('five_hour')).toBe('Session')
  expect(limitLabel('seven_day')).toBe('Weekly')
  expect(modeLabel('auto')).toBe('auto')
  expect(modeLabel('acceptEdits')).toBe('accept edits')
  expect(modeLabel('somethingNew')).toBe('somethingNew')
  expect(isNewer('2.1.300', '2.1.299')).toBe(true)
  expect(isNewer('2.1.299', '2.1.299')).toBe(false)
  expect(isNewer('2.1.29', '2.1.299')).toBe(false)
})

test('the alias adds one command name and ignores what cannot be one', async () => {
  expect(commandNames(undefined)).toEqual(['sdrawer'])
  expect(commandNames('')).toEqual(['sdrawer'])
  expect(commandNames('sdr')).toEqual(['sdrawer', 'sdr'])
  expect(commandNames(' /sdr ')).toEqual(['sdrawer', 'sdr'])
  expect(commandNames('sdrawer')).toEqual(['sdrawer'])
  expect(commandNames('two words')).toEqual(['sdrawer'])
})

test('the pane draws on the terminal before any figures arrive, a key on every row', async $ => {
  const ui = await $.ui.mount(MOUNT)
  expect(await ui.find({ type: 'Text', text: /not a repo/ })).toBeDefined()
  for (const key of ['Model', 'Context', 'Uptime', 'Git', 'Path']) {
    expect(await ui.find({ type: 'Text', text: new RegExp('^' + key + ' ') })).toBeDefined()
  }
  expect((await ui.find({ type: 'Box' }))?.props.backgroundColor).toBeUndefined()
  await ui.unmount()
})
