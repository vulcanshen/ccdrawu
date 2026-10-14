import { expect, test } from 'claude-code/testing'

import {
  bar,
  compact,
  commandNames,
  countBackground,
  countdown,
  duration,
  effortColor,
  filesText,
  isNewer,
  modeLabel,
  parseShortstat,
  parseStatusV2,
  repoFromRemote,
  shortPath,
  syncText,
  tasksText,
  truncateStart,
  usageColor,
} from '../hooks/format'
import { registerCommands } from '../hooks/commands'
import { ITEMS, view, type Snapshot } from '../hooks/items'
import { cellWidth, columnRows, enabledItems, inlineValue, parseConfig, segWidth } from '../hooks/layout'
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
    contextTokens: 420_000,
    contextWindow: 1_000_000,
    limits: [
      { kind: 'five_hour', percentUsed: 18, resetsAt: new Date(3_600_000).toISOString() },
      { kind: 'seven_day', percentUsed: 40 },
      { kind: 'spend_limit', percentUsed: 5 },
    ],
    costUsd: 1.2,
  },
  model: { name: 'Opus 5.5', effort: 'high' },
  git: {
    branch: 'main',
    added: 12,
    removed: 3,
    sha: 'e73ac92',
    ahead: 2,
    behind: 0,
    staged: 1,
    unstaged: 2,
    untracked: 0,
    repo: 'vulcanshen/ccdrawu',
  },
  path: '~/sideproj/ccdrawu',
  turns: 14,
  cache: { read: 92, total: 100 },
  tokens: { input: 12_300, output: 3_400, cacheRead: 1_200_000, cacheWrite: 45_000, speed: 61.4 },
  session: { id: '46daa4a7-76e0-44f1-be38-3216d29934ea', account: 'me@example.com', outputStyle: 'default' },
  status: { indicator: 'none', description: 'All Systems Operational' },
  skills: { count: 3, last: 'commit' },
  compacts: 1,
  agents: 2,
  background: { shell: 1 },
  version: { current: '2.1.299', latest: '2.1.300' },
  drawer: '0.1.0',
  modes: { permission: 'auto' },
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
  expect(bar(50, 10)).toEqual({ filled: '▰▰▰▰▰', empty: '▱▱▱▱▱' })
  expect(bar(130, 4)).toEqual({ filled: '▰▰▰▰', empty: '' })
  expect(bar(-5, 4)).toEqual({ filled: '', empty: '▱▱▱▱' })
})

test('usage turns green, yellow, peach, then red', async () => {
  expect(usageColor(49)).toBe('#a6e3a1')
  expect(usageColor(50)).toBe('#f9e2af')
  expect(usageColor(69)).toBe('#f9e2af')
  expect(usageColor(70)).toBe('#fab387')
  expect(usageColor(89)).toBe('#fab387')
  expect(usageColor(90)).toBe('#f38ba8')
})

test('effort turns from green at low to red at max', async () => {
  expect(['low', 'medium', 'high', 'xhigh', 'max'].map(effortColor)).toEqual([
    '#a6e3a1',
    '#f9e2af',
    '#fab387',
    '#eba0ac',
    '#f38ba8',
  ])
  expect(effortColor('16000')).toBe('#89dceb')
})

test('git shortstat, paths, modes and versions are read as shown', async () => {
  expect(parseShortstat(' 3 files changed, 12 insertions(+), 3 deletions(-)\n')).toEqual({ added: 12, removed: 3 })
  expect(parseShortstat(' 1 file changed, 1 insertion(+)\n')).toEqual({ added: 1, removed: 0 })
  expect(parseShortstat('')).toEqual({ added: 0, removed: 0 })
  expect(shortPath('/Users/me/proj', '/Users/me')).toBe('~/proj')
  expect(shortPath('/Users/meow', '/Users/me')).toBe('/Users/meow')
  expect(shortPath('C:\\Users\\me\\proj', 'C:\\Users\\me')).toBe('~\\proj')
  expect(shortPath('C:\\Users\\meow', 'C:\\Users\\me')).toBe('C:\\Users\\meow')
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

test('tasks count subagents live and the rest of the background by type, parted by ｜', async () => {
  const reported = [
    { type: 'shell' },
    { type: 'subagent' },
    { type: 'shell' },
    { type: 'workflow' },
    { type: 'MCP task' },
  ]
  expect(countBackground(reported)).toEqual({ shell: 2, workflow: 1, 'MCP task': 1 })
  expect(tasksText(2, { shell: 1 })).toBe('2 agents｜1 shell')
  expect(tasksText(1, { workflow: 1, 'MCP task': 2, shell: 2, monitor: 1 })).toBe(
    '1 agent｜2 shells｜1 monitor｜1 workflow｜2 MCP tasks',
  )
  expect(tasksText(0, { shell: 0 })).toBe('none')
  expect(tasksText(0, {})).toBe('none')
  // ｜ and CJK take two cells, so columns above the prompt still line up
  expect(cellWidth('2 agents｜1 shell')).toBe(17)
  expect(cellWidth('~/專案')).toBe(6)
  expect(segWidth([{ text: '1 agent｜', color: '' }, { text: 'ab', color: '' }])).toBe(11)
})

test('counts read short, and git status, remotes and file counts are read as shown', async () => {
  expect([950, 1000, 1500, 12_340, 152_000, 999_999, 1_200_000].map(compact)).toEqual([
    '950',
    '1k',
    '1.5k',
    '12k',
    '152k',
    '1M',
    '1.2M',
  ])
  const status = [
    '# branch.oid e73ac924a1b2c3d4e5f6',
    '# branch.head main',
    '# branch.upstream origin/main',
    '# branch.ab +2 -3',
    '1 M. N... 100644 100644 100644 aaa bbb hooks/items.ts',
    '1 .M N... 100644 100644 100644 aaa bbb README.md',
    '1 MM N... 100644 100644 100644 aaa bbb tests/a.ts',
    '? notes.txt',
  ].join('\n')
  expect(parseStatusV2(status)).toEqual({ sha: 'e73ac92', ahead: 2, behind: 3, staged: 2, unstaged: 2, untracked: 1 })
  // No upstream, no commit yet: no ahead or behind, no sha
  expect(parseStatusV2('# branch.oid (initial)\n# branch.head main\n')).toEqual({ staged: 0, unstaged: 0, untracked: 0 })
  for (const url of ['git@github.com:vulcanshen/ccdrawu.git', 'https://github.com/vulcanshen/ccdrawu', 'ssh://git@host:22/vulcanshen/ccdrawu.git\n']) {
    expect(repoFromRemote(url)).toBe('vulcanshen/ccdrawu')
  }
  expect(filesText(1, 2, 0)).toBe('1 staged · 2 changed')
  expect(filesText(0, 0, 0)).toBe('clean')
  expect(syncText(2, 3)).toBe('↑2 ↓3')
  expect(syncText(0, 0)).toBe('up to date')
})

test('the new items show what they hold, and nothing while they hold nothing, but Sync and Repo say what is missing', async () => {
  const shown = (id: string, s: Snapshot = FULL) => view(id, s)?.value.map(g => g.text).join('')
  expect(shown('style')).toBe('default')
  expect(shown('id')).toBe('46daa4a7')
  expect(shown('account')).toBe('me@example.com')
  expect(shown('window')).toBe('420k / 1M')
  expect(shown('rate')).toBe('$72.00/h')
  expect(shown('tokens')).toBe('in 12k · out 3.4k · cache 1.2M')
  expect(shown('speed')).toBe('61 tok/s')
  expect(shown('skills')).toBe('3 · last commit')
  expect(shown('sync')).toBe('↑2')
  expect(shown('files')).toBe('1 staged · 2 changed')
  expect(shown('commit')).toBe('e73ac92')
  expect(shown('repo')).toBe('vulcanshen/ccdrawu')
  expect(shown('status')).toBe('All Systems Operational')
  expect(view('status', { ...FULL, status: { indicator: 'major', description: 'Major outage' } })?.value[0]?.color).toBe('#f38ba8')
  // Before a minute has passed, the hourly rate would jump about
  expect(shown('rate', { ...FULL, now: 30_000 })).toBeUndefined()
  const empty: Snapshot = {
    ...FULL,
    tokens: { input: 0, output: 0, cacheRead: 0, cacheWrite: 0 },
    session: {},
    status: null,
    skills: { count: 0 },
    git: { branch: 'main', added: 0, removed: 0 },
  }
  // Files has nothing to count only when git status itself failed
  for (const id of ['style', 'id', 'account', 'tokens', 'speed', 'files', 'status']) {
    expect(shown(id, empty)).toBeUndefined()
  }
  expect(shown('skills', empty)).toBe('none')
  // A git row turned on is always there: it says what is missing, or that there is no repo
  expect(shown('sync', empty)).toBe('no upstream')
  expect(shown('commit', empty)).toBe('no commits')
  expect(shown('repo', empty)).toBe('no origin')
  for (const id of ['git', 'sync', 'files', 'commit', 'repo']) {
    expect(shown(id, { ...FULL, git: null })).toBe('not a repo')
  }
  // New items start off: the drawer does not grow by itself
  const added = ['style', 'id', 'account', 'window', 'rate', 'tokens', 'speed', 'skills', 'sync', 'files', 'commit', 'repo', 'status']
  for (const item of ITEMS.filter(i => added.includes(i.id))) {
    expect(item.dock || item.inline).toBe(false)
  }
})

test('Claude Code and ccdrawu have a version row each', async () => {
  const shown = (id: string, s: Snapshot) => view(id, s)?.value.map(g => g.text).join('')
  expect(shown('version', FULL)).toBe('2.1.299  ↑ 2.1.300')
  expect(shown('ccdrawu', FULL)).toBe('0.1.0')
  expect(shown('ccdrawu', { ...FULL, drawer: null })).toBeUndefined()
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
  expect(await ui.find({ type: 'Text', text: /▰|▱/ })).toBeDefined()
  // One row after another: no blank row between groups
  expect(await ui.find({ type: 'Text', text: /^ $/ })).toBeUndefined()
  expect((await ui.find({ type: 'Box' }))?.props.backgroundColor).toBeUndefined()
  await ui.unmount()
})

test('above the prompt the pane draws columns of text with no bars', async $ => {
  const ui = await $.ui.mount({ ...MOUNT, props: { ...PANE, placement: 'inline', bodyColumns: 100 } })
  expect(await ui.find({ type: 'Text', text: /│/ })).toBeDefined()
  expect(await ui.find({ type: 'Text', text: /▰|▱/ })).toBeUndefined()
  expect(await ui.find({ type: 'Text', text: /^Context / })).toBeDefined()
  await ui.unmount()
})
