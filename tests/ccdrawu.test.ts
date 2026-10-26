import { expect, test, type Mounted } from 'claude-code/testing'

import {
  bar,
  compact,
  commandNames,
  countBackground,
  countdown,
  duration,
  effortColor,
  filesText,
  directoryName,
  isNewer,
  LINE,
  modeLabel,
  parseShortstat,
  parseStatusV2,
  repoFromRemote,
  shortPath,
  SQUARE,
  SQUARE_FILL,
  syncText,
  tasksText,
  truncateStart,
  usageColor,
} from '../hooks/format'
import { registerCommands } from '../hooks/commands'
import { ITEMS, view, type Snapshot } from '../hooks/items'
import { cellWidth, columnRows, drawerShows, drawerWanted, enabledItems, inlineValue, parseConfig, segWidth, showsAnywhere } from '../hooks/layout'
import { locations } from '../hooks/paths'
import { statuslineFile, statuslineItems } from '../hooks/statusline'
import { parseYaml } from '../hooks/yaml'
import {
  BALL_LINES,
  ballAt,
  KIT_HEIGHT,
  KIT_WIDTH,
  kitCells,
  KIT_MAX,
  KITS,
  kitsFor,
  kitsTick,
  LIFT_ROOM,
  HEART,
  HEART_COLOR,
  PAT_LINES,
  PET_HEIGHT,
  PET_START,
  PET_TICK,
  petCells,
  petCoat,
  petEvent,
  petLook,
  PETS,
  petSpecies,
  petTick,
  POSES,
  snore,
  SNORE_COLOR,
  SPECIES,
  TRACK,
  type Pose,
  type Run,
  type Species,
} from '../hooks/pets'
import type { Kit, PetState } from '../types'

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
  session: { id: '46daa4a7-76e0-44f1-be38-3216d29934ea', outputStyle: 'default' },
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
  expect(bar(50, 10)).toEqual({ filled: SQUARE_FILL.repeat(5), empty: SQUARE.repeat(5) })
  expect(bar(130, 4)).toEqual({ filled: SQUARE_FILL.repeat(4), empty: '' })
  expect(bar(-5, 4)).toEqual({ filled: '', empty: SQUARE.repeat(4) })
  // A terminal with no Nerd Font picks the line; a style it does not know is the squares
  expect(bar(50, 4, 'line')).toEqual({ filled: LINE.repeat(2), empty: LINE.repeat(2) })
  expect(bar(50, 4, 'dots')).toEqual({ filled: SQUARE_FILL.repeat(2), empty: SQUARE.repeat(2) })
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
  expect(directoryName('~/sideproj/ccdrawu')).toBe('ccdrawu')
  expect(directoryName('~\\proj\\')).toBe('proj')
  expect(directoryName('~')).toBe('~')
  expect(directoryName('/')).toBe('/')
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
    expect(item.label.length).toBeLessThanOrEqual(item.id === 'directory' ? 9 : 8)
    expect(item.short.length).toBeGreaterThan(0)
    expect(item.short.length).toBeLessThanOrEqual(4)
    // The pet is no row: the pane draws it itself, first of all and on by default
    if (item.id === 'pet') expect(view(item.id, FULL)).toBeUndefined()
    else expect(view(item.id, FULL)).toBeDefined()
  }
  expect(ITEMS[0]).toMatchObject({ id: 'pet', dock: true, inline: true, statusline: false })
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
  expect(tasksText(2, { shell: 1 }, ', ')).toBe('2 agents, 1 shell')
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

test('the new items show what they hold, and - while they hold nothing, but Sync and Repo say what is missing', async () => {
  const shown = (id: string, s: Snapshot = FULL) => view(id, s)?.value.map(g => g.text).join('')
  expect(shown('style')).toBe('default')
  expect(shown('id')).toBe('46daa4a7')
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
  expect(shown('rate', { ...FULL, now: 30_000 })).toBe('-')
  const empty: Snapshot = {
    ...FULL,
    tokens: { input: 0, output: 0, cacheRead: 0, cacheWrite: 0 },
    session: {},
    status: null,
    skills: { count: 0 },
    git: { branch: 'main', added: 0, removed: 0 },
  }
  // Files has nothing to count only when git status itself failed
  for (const id of ['style', 'id', 'tokens', 'speed', 'files', 'status']) {
    expect(shown(id, empty)).toBe('-')
  }
  expect(shown('skills', empty)).toBe('none')
  // So are the first rows: a new session hears its mode only with the first prompt
  const fresh: Snapshot = { ...empty, usage: null, model: null, modes: {}, cache: null, version: null, drawer: null }
  for (const id of ['model', 'effort', 'mode', 'session', 'weekly', 'spend', 'cost', 'cache', 'version', 'ccdrawu']) {
    expect(shown(id, fresh)).toBe('-')
  }
  expect(shown('uptime', fresh)).toBe('- · 14 turns')
  // and a limit's resets row with them
  expect(view('session', fresh)?.resets).toBe('-')
  // A git row turned on is always there: it says what is missing, or that there is no repo
  expect(shown('sync', empty)).toBe('no upstream')
  expect(shown('commit', empty)).toBe('no commits')
  expect(shown('repo', empty)).toBe('no origin')
  for (const id of ['git', 'sync', 'files', 'commit', 'repo']) {
    expect(shown(id, { ...FULL, git: null })).toBe('not a repo')
  }
  // New items start off: the drawer does not grow by itself
  const added = ['style', 'id', 'window', 'rate', 'tokens', 'speed', 'skills', 'sync', 'files', 'commit', 'repo', 'status']
  for (const item of ITEMS.filter(i => added.includes(i.id))) {
    expect(item.dock || item.inline).toBe(false)
  }
})

test('every pose of every pet is 12 x 8 pixels, every mark inked in every coat, and draws 12 x 4 cells', async () => {
  for (const species of SPECIES) {
    const pet = PETS[species]
    expect(pet.coats.length).toBeGreaterThan(1)
    for (const pose of POSES) {
      const frame = pet.frames[pose]
      expect(frame.length).toBe(8)
      for (const row of frame) expect(row.length).toBe(12)
      const marks = new Set(frame.join('').replace(/\./g, ''))
      for (let i = 0; i < pet.coats.length; i++) {
        const coat = petCoat(species, i)
        for (const mark of marks) expect(coat[mark]).toBeDefined()
      }
      for (const faceLeft of [false, true]) {
        const cells = petCells(species, pose, petCoat(species, 0), faceLeft)
        expect(cells.length).toBe(4)
        for (const row of cells) expect(cellsWide(row)).toBe(12)
      }
    }
    // On the ground, its feet on the bottom row
    expect(pet.frames.walk1[7]).not.toBe('............')
    // The limbs are their own ink in every pose, a color every coat names apart from the fur, so a
    // paw a swipe raises does not run into the head
    for (const pose of POSES) expect(pet.frames[pose].join('')).toContain('q')
    for (let i = 0; i < pet.coats.length; i++) expect(petCoat(species, i).q).not.toBe(petCoat(species, i)['#'])
    // Asleep it leaves the top row of cells for its snore, which keeps the row 12 wide
    expect(pet.frames.sleep.slice(0, 2).join('')).toBe('.'.repeat(24))
    for (const n of [1, 2, 3]) {
      expect(cellsWide(topRow(species, 'sleep', false, n))).toBe(12)
    }
    // Sitting it leaves the corner over its head free for a heart
    for (const faceLeft of [false, true]) {
      const top = topRow(species, 'sit', faceLeft)
      expect(top[faceLeft ? 0 : top.length - 1]).toHaveProperty('gap')
    }
  }
  // Facing left is the same drawing the other way about
  const right = petCells('cat', 'walk1', petCoat('cat', 0))
  const left = petCells('cat', 'walk1', petCoat('cat', 0), true)
  expect(left[3]).toEqual([...(right[3] ?? [])].reverse())
  // The snore is text over the head, growing z, zz, zzZ: at the right facing right, at the left facing left
  const text = (n: number, faceLeft: boolean) => topRow('cat', 'sleep', faceLeft, n).flatMap(r => ('gap' in r ? [] : [r.ch]))
  expect([1, 2, 3, 4].map(n => text(n, false)[0])).toEqual(['z', 'zz', 'zzZ', 'z'])
  expect(text(3, true)).toEqual(['Zzz'])
  expect(topRow('cat', 'sleep', false, 3)).toEqual([{ gap: 9 }, { ch: 'zzZ', fg: SNORE_COLOR, bg: undefined }])
  // Beside the transcript the box is taller than the pet: standing, the room is over it; up the wall,
  // under it
  const tall = (lift: number) => petCells('cat', 'swipe1', petCoat('cat', 0), false, undefined, PET_HEIGHT + LIFT_ROOM, lift)
  expect(tall(0).length).toBe(5)
  expect(tall(0)[0]).toEqual([{ gap: 12 }])
  expect(tall(2)[4]).toEqual([{ gap: 12 }])
  expect(tall(2)[0]).not.toEqual([{ gap: 12 }])
  // A session's coat number counts round any animal's coats
  expect(petCoat('penguin', PETS.penguin.coats.length + 1)).toEqual(PETS.penguin.coats[1])
})

// How many cells a row of runs takes, and the top row of a pose, its snore n along when asleep
const cellsWide = (row: readonly Run[]) => row.reduce((w, r) => w + ('gap' in r ? r.gap : r.ch.length), 0)
const topRow = (species: Species, pose: Pose, faceLeft: boolean, n = 0) =>
  petCells(species, pose, petCoat(species, 0), faceLeft, n > 0 ? snore(n) : undefined)[0] ?? []

test('the pet wanders while Claude works, sits when it is done, sleeps when nothing goes on, and answers a pat', async () => {
  const rand = () => 0
  // Random draws in turn, the last one again once they run out
  const draws = (...values: number[]) => {
    let i = 0
    return () => values[Math.min(i++, values.length - 1)] ?? 0
  }
  const quiet = () => 0.99 // nothing left to chance happens
  const walk = (s: PetState, from: number, ticks: number) => {
    let t = from
    for (let i = 0; i < ticks; i++) s = petTick(s, (t += PET_TICK), quiet)
    return s
  }
  // Away from either wall
  let s: PetState = { ...PET_START, coat: 0, active: 0, x: 5, target: 5 }
  // Idle and just started: sitting, and a tick changes nothing (no redraw)
  s = petTick(s, 1_000, quiet)
  expect(petLook(s).pose).toBe('sit')
  expect(petTick(s, 1_200, quiet)).toBe(s)

  // Claude at work: it picks a place and goes there, its legs taking turns, facing its way
  s = petEvent(s, 'busy', 2_000, 'cat', rand)
  s = petTick(s, 2_200, draws(0.9, 0.5)) // straight on, to the middle
  expect([s.target, s.run]).toEqual([10, false])
  s = walk(s, 2_200, 3)
  expect([s.mode, s.x, s.faceLeft]).toEqual(['walk', 8, false])
  expect(petLook(s).pose).not.toBe(petLook(walk(s, 2_800, 1)).pose)
  s = walk(s, 2_800, 2)
  expect(s.x).toBe(10)
  // There it rests one to three seconds, sitting, then goes on to the next place, facing it
  let t = 4_400
  // or, now and then, swipes at the air a moment, one way and the other
  const swipes = petTick({ ...s }, t, draws(0.1, 0.5))
  expect([swipes.mode, petLook(swipes).pose === petLook(petTick(swipes, t + 200, quiet)).pose]).toEqual(['play', false])
  expect(petTick(swipes, t + 1_400, quiet).mode).not.toBe('play')
  s = petTick(s, t, draws(0.3, 0.5, 0.1)) // a rest of 2 s, then to 2
  expect([s.mode, s.restUntil, s.target]).toEqual(['sit', t + 2_000, 2])
  expect(petTick(s, t + 1_000, quiet).mode).toBe('sit')
  expect([petTick(s, t + 2_200, quiet).mode, petTick(s, t + 2_200, quiet).faceLeft]).toEqual(['walk', true])
  // or hops where it is
  expect(petTick({ ...s, restUntil: 0, target: s.x }, t, draws(0.5, 0.5)).mode).toBe('jump')
  // A far place it may run to, two steps a tick
  const far = petTick({ ...s, x: 0, target: 0, restUntil: 0 }, t, draws(0.9, 0.99, 0.1))
  expect([far.target, far.run]).toEqual([TRACK, true])
  expect(walk(far, t, 1).x).toBe(2)
  // At the ends of its walk it does as anywhere else: no climbing
  expect(petTick({ ...s, x: TRACK, target: TRACK, restUntil: 0 }, 50_000, draws(0.3, 0.5, 0.1)).mode).toBe('sit')

  // A tool fails: a jump, and a line meanwhile, then on its way again. The jump is the pet as it
  // strides, not squashed, lifted 1, 2, 1 pixels and down
  s = petEvent(s, 'error', 10_000, 'cat', rand)
  expect(petTick(s, 10_200, quiet).mode).toBe('jump')
  expect(petTick(s, 10_200, quiet).line).toBe('Uh oh...')
  expect([10_000, 10_200, 10_400, 10_600].map(at => petLook(petTick(s, at, quiet)).lift)).toEqual([1, 2, 1, 0])
  expect(petLook(petTick(s, 10_200, quiet)).pose).toBe('walk2')
  expect(petTick(s, 10_800, quiet).mode).not.toBe('jump')

  // A turn the user stopped: it stops where it is and sits, saying nothing; one an error ended: it
  // jumps, worried, and sits
  const stopped = petEvent({ ...s, target: s.x + 5 }, 'stop', 15_000, 'cat', rand)
  expect(stopped.target).toBe(stopped.x)
  expect(petTick(stopped, 15_200, quiet).mode).toBe('sit')
  expect([stopped.line, stopped.lineUntil]).toEqual([s.line, s.lineUntil])
  const failed = petEvent(s, 'failed', 15_000, 'cat', rand)
  expect([petTick(failed, 15_200, quiet).mode, petTick(failed, 15_200, quiet).line]).toEqual(['jump', 'Uh oh...'])
  expect(petTick(failed, 15_800, quiet).mode).toBe('sit')

  // Done: it stops and sits, and says so for eight seconds
  s = petEvent(s, 'done', 20_000, 'cat', rand)
  s = petTick(s, 20_200, quiet)
  expect([s.mode, s.line, s.target]).toEqual(['sit', 'All done!', s.x])
  expect(petTick(s, 28_200, quiet).line).toBeUndefined()

  // Idle, now and then it strolls a little, looks the other way, or swipes at the air
  const strolls = petTick(s, 30_000, draws(0, 0.2, s.x < 10 ? 0.9 : 0.1))
  expect(Math.abs(strolls.target - s.x)).toBeGreaterThanOrEqual(4)
  expect(petTick(strolls, 30_200, quiet).mode).toBe('walk')
  expect(petTick(s, 30_000, draws(0, 0.5)).faceLeft).toBe(!s.faceLeft)
  expect(petTick(s, 30_000, draws(0, 0.9)).mode).toBe('play')

  // Idle, it winds down: still sitting a minute on, lying down after two, asleep after ten
  expect(petTick(s, 20_000 + 60_000, quiet).mode).toBe('sit')
  const lying = petTick(s, 20_000 + 2 * 60_000, quiet)
  expect([lying.mode, petLook(lying).pose]).toEqual(['lie', 'lie'])
  // and anything going on gets it up
  expect(petTick(petEvent(lying, 'busy', 150_000, 'cat', rand), 150_200, draws(0.9, 0.5)).mode).not.toBe('lie')

  // Ten minutes of nothing: asleep, snoring z, zz, zzZ
  s = petTick(s, 20_000 + 10 * 60_000, quiet)
  expect(s.mode).toBe('sleep')
  expect(petLook(s).over?.text).toMatch(/^z/)
  const snores = new Set([0, 800, 1600].map(d => petLook(petTick(s, 20_000 + 10 * 60_000 + d, quiet)).over?.text))
  expect(snores.size).toBe(3)

  // A pat wakes it: it sits, a heart over its head, and says what its kind says
  const patted = petTick(petEvent(s, 'pat', 700_000, 'duck', rand), 700_200, quiet)
  expect(patted.mode).toBe('sit')
  expect(petLook(patted).over).toEqual({ text: HEART, color: HEART_COLOR })
  expect(patted.line).toBe(PAT_LINES.duck[0])
  // and while Claude works it stops for the pat, then goes on its way
  const busyPat = petEvent(petEvent({ ...s, target: s.x + 5 }, 'busy', 700_000, 'cat', rand), 'pat', 700_000, 'cat', rand)
  expect(petTick(busyPat, 700_200, quiet).mode).toBe('sit')
  const after = petTick(busyPat, 703_200, draws(0.9, 0.5))
  expect(after.target).not.toBe(after.x)
  expect(petTick(after, 703_400, quiet).mode).toBe('walk')

  // The config names the animal; one it does not know is a cat
  expect([petSpecies('penguin'), petSpecies('dragon'), petSpecies(undefined)]).toEqual(['penguin', 'cat', 'cat'])
  expect(parseConfig('{"pet": {"species": "dog", "name": "Mochi"}}')?.pet).toEqual({ species: 'dog' })
})

test('a ball thrown in rolls to a stop from the far side; the pet runs over, paws at it, and has it', async () => {
  const quiet = () => 0.99
  let s: PetState = { ...PET_START, coat: 0, active: 0, x: 2, target: 2 }
  s = petEvent(s, 'ball', 1_000, 'dog', () => 0.5)
  const ball = s.ball
  // From the far side to a stop in the middle stretch; the pet will stand on its nearer side
  expect(ball).toMatchObject({ from: 1_000, start: TRACK, rest: 10, goal: 5, line: BALL_LINES.dog[1] })
  if (ball === undefined) return
  // It bounces in, 2, 1, 0, 1, 0 pixels off the ground, slowing to a stop
  expect([0, 200, 400, 600, 800, 1_000].map(d => ballAt(ball, 1_000 + d).lift)).toEqual([2, 1, 0, 1, 0, 0])
  expect([ballAt(ball, 1_000).pos, ballAt(ball, 2_200).pos, ballAt(ball, 9_000).pos]).toEqual([TRACK, 10, 10])

  const seen: PetState[] = []
  for (let t = 1_200; t <= 4_000; t += PET_TICK) seen.push((s = petTick(s, t, quiet)))
  // It runs, two steps a tick, and waits there for the ball to stop
  expect(seen.slice(0, 2).map(p => [p.mode, p.x])).toEqual([
    ['walk', 4],
    ['walk', 5],
  ])
  expect(seen[2]?.mode).toBe('sit')
  // Its place along the walk ticks as the ball rolls in, so the drawing follows it
  expect(seen[0]?.ballPos).toBeGreaterThan(seen[3]?.ballPos ?? 0)
  // Stopped: it paws at it, facing it, one way and the other
  const playing = seen.filter(p => p.mode === 'play')
  expect(playing.length).toBeGreaterThan(4)
  expect(playing.every(p => !p.faceLeft)).toBe(true)
  expect(new Set(playing.map(p => petLook(p).pose))).toEqual(new Set(['swipe1', 'swipe2']))
  // Then it has it: the ball goes, and it says so
  expect([s.mode, s.ball, s.line]).toEqual(['sit', undefined, BALL_LINES.dog[1]])

  // A ball not had in 20 seconds goes
  const left = petEvent({ ...PET_START, active: 0, x: 2, target: 2, heartUntil: 99_000 }, 'ball', 0, 'cat', quiet)
  expect(petTick({ ...left, heartUntil: 99_000 }, 20_200, quiet).ball).toBeUndefined()
})

test("a kit of the pet's kind walks for each subagent running, three at most, each in a coat of its own, and leaves once it is done", async () => {
  // Every kit is 6 pixels wide and stands in a box 6 high, inked in every coat, and draws 6 x 3 cells in
  // its own coat's colors; every kind has coats enough for the pet and three kits apart
  for (const species of SPECIES) {
    expect(PETS[species].coats.length).toBeGreaterThan(KIT_MAX)
    for (const pose of ['walk1', 'walk2'] as const) {
      const frame = KITS[species][pose]
      expect(frame.length % 2).toBe(0)
      expect(frame.length).toBeLessThanOrEqual(KIT_HEIGHT)
      for (const row of frame) expect(row.length).toBe(KIT_WIDTH)
      const marks = new Set(frame.join('').replace(/\./g, ''))
      for (let i = 0; i < PETS[species].coats.length; i++) {
        const coat = petCoat(species, i)
        for (const mark of marks) expect(coat[mark]).toBeDefined()
        expect(coat.q).not.toBe(coat['#'])
      }
    }
    const colors = new Set(Object.values(petCoat(species, 1)))
    const cells = kitCells(species, { id: 'a', coat: 1, x: 0, target: 5, stride: 1, faceLeft: true, restUntil: 0, leaving: false })
    expect(cells.length).toBe(KIT_HEIGHT / 2)
    for (const row of cells) {
      expect(cellsWide(row)).toBe(KIT_WIDTH)
      for (const r of row) if ('ch' in r) for (const c of [r.fg, r.bg]) if (c) expect(colors.has(c)).toBe(true)
    }
  }

  const rand = () => 0.2
  const ticks = (list: Kit[], from: number, n: number) => {
    let t = from
    for (let i = 0; i < n; i++) list = kitsTick(list, (t += PET_TICK), rand)
    return list
  }
  // Three subagents at most get a kit, each in at an end of the walk, in a coat neither the pet (coat 0)
  // nor another kit has
  let ks = kitsFor([], ['a', 'b', 'c', 'd'], 'cat', 0, rand)
  expect(ks.map(k => [k.id, k.coat, k.x, k.target])).toEqual([
    ['a', 1, 0, 4],
    ['b', 2, 0, 4],
    ['c', 3, 0, 4],
  ])
  expect(kitsFor(ks, ['a', 'b', 'c', 'd'], 'cat', 0, rand)).toBe(ks)
  // A step a tick to where each goes, its legs taking turns; there a rest, then the next place
  ks = ticks(ks, 0, 4)
  expect(ks.map(k => [k.x, k.stride])).toEqual([
    [4, 4],
    [4, 4],
    [4, 4],
  ])
  ks = kitsTick(ks, 1_000, rand)
  expect(ks[1]).toMatchObject({ x: 4, target: TRACK, restUntil: 2_400 })
  expect(kitsTick(ks, 1_200, rand)).toBe(ks)
  // a and c done: they head for the nearer end, resting no more, and go once there; b rests on
  ks = kitsFor(ks, ['b', 'd'], 'cat', 0, rand)
  expect(ks.map(k => [k.id, k.leaving, k.target])).toEqual([
    ['a', true, 0],
    ['b', false, TRACK],
    ['c', true, 0],
  ])
  ks = ticks(ks, 1_200, 5)
  expect(ks.map(k => k.id)).toEqual(['b'])
  // Room again: the subagent still running gets its kit, in a coat b and the pet do not have
  expect(kitsFor(ks, ['b', 'd'], 'cat', 0, rand).map(k => [k.id, k.coat])).toEqual([
    ['b', 2],
    ['d', 1],
  ])
})

test('Claude Code and ccdrawu have a version row each', async () => {
  const shown = (id: string, s: Snapshot) => view(id, s)?.value.map(g => g.text).join('')
  expect(shown('version', FULL)).toBe('2.1.299  ↑ 2.1.300')
  expect(shown('ccdrawu', FULL)).toBe('0.1.0')
  expect(shown('ccdrawu', { ...FULL, drawer: null })).toBe('-')
})

test('files sit where tdp E3 puts them, unless CCDRAWU__ variables say otherwise', async () => {
  const config = (dir: string) => ({ configFile: `${dir}/config.yaml`, oldConfigFile: `${dir}/config.json` })
  expect(locations({ home: '/h', os: 'darwin' })).toEqual({ ...config('/h/.config/ccdrawu'), cache: '/h/Library/Caches/ccdrawu' })
  expect(locations({ home: '/h', os: 'linux', xdgConfig: '/x', xdgCache: '/xc', cache: '' })).toEqual({
    ...config('/x/ccdrawu'),
    cache: '/xc/ccdrawu',
  })
  // A relative XDG_CACHE_HOME is not taken, as Go's os.UserCacheDir does not
  expect(locations({ home: '/h', os: 'linux', xdgCache: 'rel' }).cache).toBe('/h/.cache/ccdrawu')
  expect(locations({ home: 'C:\\U', os: 'windows', localAppData: 'C:\\U\\AppData\\Local' }).cache).toBe('C:\\U\\AppData\\Local/ccdrawu/cache')
  expect(locations({ home: '/h', os: 'linux', xdgConfig: '/x', config: '/c', cache: '/k' })).toEqual({ ...config('/c'), cache: '/k' })
})

test('config.yaml reads as the ccdrawu command writes it, and as a person writes YAML by hand', async () => {
  // As the command writes it (gopkg.in/yaml.v3): block style, quotes only where needed, a kept status line
  // of many rows as a literal block
  const written = [
    'dock:',
    '  - pet',
    '  - model',
    'inline: null',
    'alias: "true"',
    'pet:',
    '  species: dog',
    'bar: line',
    'drawer:',
    '  inlineOff: true',
    'statusline:',
    '  on: true',
    '  items: []',
    "  separator: ' · '",
    '  esc: "a\\eb"',
    '  before: |-',
    '    {',
    '        "type": "command"',
    '      }',
    '  refresh: 30',
    '',
  ].join('\n')
  expect(parseYaml(written)).toEqual({
    dock: ['pet', 'model'],
    inline: null,
    alias: 'true',
    pet: { species: 'dog' },
    bar: 'line',
    drawer: { inlineOff: true },
    statusline: { on: true, items: [], separator: ' · ', esc: 'a\x1bb', before: '{\n    "type": "command"\n  }', refresh: 30 },
  })
  expect(parseConfig(written)).toEqual({
    dock: ['pet', 'model'],
    alias: 'true',
    pet: { species: 'dog' },
    bar: 'line',
    drawer: { sidebarOff: false, inlineOff: true },
    statusline: { on: true },
  })
  // By hand: comments, a sequence as deep as its key, flow, quotes of both kinds, a mapping in a sequence
  const hand = "# mine\n---\nalias: 'it''s'   # a name\ndock:\n- model\n- 'git'\ninline: [cost, \"context\"]\nx:\n  - a: 1\n    b: [2]\n"
  expect(parseYaml(hand)).toEqual({ alias: "it's", dock: ['model', 'git'], inline: ['cost', 'context'], x: [{ a: 1, b: [2] }] })
  // Not YAML, or YAML this does not read, is no config yet
  for (const bad of ['dock: [', 'a: *ref', 'a: 1\n---\nb: 2', '\tdock: []', 'a: "\\q"']) {
    expect(parseConfig(bad)).toBeUndefined()
  }
  // The config.json of 0.1.4 and before, read as JSON until the TUI moves it
  expect(parseConfig('{"dock": ["git"], "alias": "ccd"}', true)).toEqual({ dock: ['git'], alias: 'ccd' })
  expect(parseConfig('dock: [git]', true)).toBeUndefined()
})

test('the config picks items per placement, in the order it lists them', async () => {
  expect(parseConfig('{"dock": ["git", "model", 3], "inline": "x"}')).toEqual({ dock: ['git', 'model'] })
  expect(parseConfig('{"dock": [')).toBeUndefined()
  expect(parseConfig('null')).toEqual({})
  expect(parseConfig('{"dock": null, "alias": "ccd"}')).toEqual({ alias: 'ccd' })
  expect(parseConfig('{"alias": 3}')).toEqual({})
  expect(parseConfig('{"bar": "line"}')).toEqual({ bar: 'line' })
  expect(parseConfig('{"bar": 1}')).toEqual({})
  // The statusline's block is the TUI's and the command's; the mod reads only whether it is on
  expect(parseConfig('{"statusline": {"on": true, "items": ["model"]}}')).toEqual({ statusline: { on: true } })
  expect(parseConfig('{"statusline": {"on": "yes"}}')).toEqual({ statusline: { on: false } })
  // The drawer shows in each place unless the config turns it off there; 0.1.1's off is off in both
  expect(parseConfig('{"drawer": {"inlineOff": true}}')).toEqual({ drawer: { sidebarOff: false, inlineOff: true } })
  expect(parseConfig('{"drawer": {"sidebarOff": 1}}')).toEqual({ drawer: { sidebarOff: false, inlineOff: false } })
  expect(parseConfig('{"drawer": {"off": true}}')).toEqual({ drawer: { sidebarOff: true, inlineOff: true } })
  expect([drawerShows({}, 'dock'), drawerShows({}, 'inline')]).toEqual([true, true])
  const narrowOff = parseConfig('{"drawer": {"inlineOff": true}}')!
  expect([drawerShows(narrowOff, 'dock'), drawerShows(narrowOff, 'inline')]).toEqual([true, false])
  // /ccdrawu sets it for the session in both places over the config; until it does, the config goes
  expect([drawerWanted(narrowOff, 'dock', null), drawerWanted(narrowOff, 'inline', null)]).toEqual([true, false])
  expect([drawerWanted(narrowOff, 'dock', 'on'), drawerWanted(narrowOff, 'inline', 'on')]).toEqual([true, true])
  expect([drawerWanted({}, 'dock', 'off'), drawerWanted({}, 'inline', 'off')]).toEqual([false, false])
  expect(parseConfig('{"statusline": 3}')).toEqual({})
  expect(enabledItems({ dock: ['git', 'model', 'gone', 'git'] }, 'dock').map(i => i.id)).toEqual(['git', 'model'])
  expect(enabledItems({ dock: [] }, 'dock')).toEqual([])
  // Status, off by default, is asked of status.claude.com only once a placement shows it; the pet is on
  expect(showsAnywhere({}, 'status')).toBe(false)
  expect(showsAnywhere({ inline: ['status'] }, 'status')).toBe(true)
  expect(showsAnywhere({ dock: [], inline: [] }, 'pet')).toBe(false)
  expect(showsAnywhere({}, 'pet')).toBe(true)
  expect(enabledItems({}, 'inline').map(i => i.id)).toEqual(ITEMS.filter(i => i.inline).map(i => i.id))
})

test('above the prompt a limit reads as text with its countdown, and columns take the fewest rows that fit', async () => {
  const session = view('session', FULL)!
  expect(inlineValue(session).map(s => s.text)).toEqual(['18%', ' · ', '59m'])
  // A countdown not known holds the sidebar's resets row, but is left out here
  expect(view('weekly', FULL)?.resets).toBe('-')
  expect(inlineValue(view('weekly', FULL)!).map(s => s.text)).toEqual(['40%'])
  expect(inlineValue(view('path', { ...FULL, path: '~/' + 'x'.repeat(40) })!)[0]!.text).toHaveLength(32)
  // The branch whole, its diff on a row of its own in the sidebar and after it here
  const branch = 'customization-yuanta-' + 'x'.repeat(30)
  const git = view('git', { ...FULL, git: { ...FULL.git!, branch } })!
  expect(git.value.map(s => s.text)).toEqual([branch])
  expect(git.below?.map(s => s.text)).toEqual(['+12', ' -3'])
  expect(inlineValue(git).map(s => s.text).join('')).toBe(branch + '  +12 -3')

  // Four cells of key 5 and value 5: one row needs 4 * 11 + 3 * 3 = 53 cells
  const cells = Array.from({ length: 4 }, () => ({ label: 5, value: 5 }))
  expect(columnRows(cells, 53)).toBe(1)
  expect(columnRows(cells, 52)).toBe(2)
  expect(columnRows(cells, 10)).toBe(4)
  expect(columnRows([], 10)).toBe(1)
})

test('the statusline file holds each item as it reads above the prompt, the pet aside, with no time in it', async () => {
  const items = statuslineItems(FULL)
  expect(Object.keys(items)).toEqual(ITEMS.filter(i => i.id !== 'pet').map(i => i.id))
  expect(items.session!.map(s => s.text)).toEqual(['18%', ' · ', '59m'])
  expect(items.model).toEqual([{ text: 'Opus 5.5', color: '#b4befe', bold: true }])
  // Tasks part their counts with a comma, as the statusline parts its items with |
  expect(items.tasks!.map(s => s.text).join('')).toBe('2 agents, 1 shell')
  // An item with nothing to show reads -, as in the drawer
  const bare = statuslineItems({ ...FULL, model: null, session: {}, drawer: null })
  for (const id of ['model', 'effort', 'id', 'ccdrawu']) expect(bare[id]).toEqual([{ text: '-', color: '#cdd6f4' }])

  const text = statuslineFile(FULL, true)
  expect(JSON.parse(text)).toEqual({ version: 1, docked: true, items })
  expect(statuslineFile({ ...FULL }, true)).toBe(text)
  // The drawer moving from beside the transcript is a change the file takes
  expect(JSON.parse(statuslineFile(FULL, false)).docked).toBe(false)
})

test('the pane draws on the terminal before any figures arrive, a key on every row', async $ => {
  const ui = await $.ui.mount(MOUNT)
  expect(await ui.find({ type: 'Text', text: /not a repo/ })).toBeDefined()
  // Each row turned on is there from the start, the figures not come yet reading -
  for (const key of ITEMS.filter(i => i.dock && i.id !== 'pet').map(i => i.label)) {
    expect(await ui.find({ type: 'Text', text: new RegExp('^' + key + ' ') })).toBeDefined()
  }
  expect(await ui.find({ type: 'Text', text: '-' })).toBeDefined()
  expect(await ui.find({ type: 'Text', text: /^resets $/ })).toBeDefined()
  expect(await ui.find({ type: 'Text', text: /\u{F445}|\u{F51D}/u })).toBeDefined()
  // The pet, on by default, in pixels in a box of 12 x 4 that keeps its height
  expect(await ui.find({ type: 'Text', text: /▀|▄/ })).toBeDefined()
  expect((await ui.find({ type: 'Box', key: 'pet' }))?.props).toMatchObject({ width: 12, height: 5 })
  // One row after another: no blank row between groups
  expect(await ui.find({ type: 'Text', text: /^ $/ })).toBeUndefined()
  expect((await ui.find({ type: 'Box' }))?.props.backgroundColor).toBeUndefined()
  await ui.unmount()
})

test('above the prompt the pane draws columns of text with no bars', async $ => {
  const ui = await $.ui.mount({ ...MOUNT, props: { ...PANE, placement: 'inline', bodyColumns: 100 } })
  expect(await ui.find({ type: 'Text', text: /│/ })).toBeDefined()
  expect(await ui.find({ type: 'Text', text: /\u{F445}|\u{F51D}/u })).toBeUndefined()
  expect(await ui.find({ type: 'Text', text: /^Context / })).toBeDefined()
  // and the pet left of the columns
  expect((await ui.find({ type: 'Box', key: 'pet' }))?.props).toMatchObject({ width: 12, height: 4 })
  await ui.unmount()
})

// Waits for the drawing to hold a Text that reads text, once the work a hook left running wrote it
// (each find gives that work its turn)
async function drawn(ui: Mounted<'terminal', 'Pane'>, text: string) {
  for (let i = 0; i < 200; i++) {
    if (await ui.find({ type: 'Text', text })) return
  }
  const texts = (await ui.findAll({ type: 'Text' })).map(t => t.text)
  throw new Error(`no ${JSON.stringify(text)} among ${JSON.stringify(texts)}`)
}

test('a compaction empties the context, though the engine answers the fill before it a moment longer', async ($, on) => {
  let fill: { tokens?: number; percent?: number } = { tokens: 640_000, percent: 64 }
  const context = () => ({ ...fill, window: 1_000_000 })
  on('session.usage', () => ({ value: { startedAt: 0, context: context(), rateLimits: [] } }))
  on('session.measure', (_, e) => ({ changed: e.changed }))
  on('classic.PostCompact', () => ({}))
  const measure = () => $.session.measure({ context: context(), rateLimits: [], changed: ['context'] })
  const ui = await $.ui.mount({ ...MOUNT, props: { ...PANE, placement: 'inline', bodyColumns: 100 } })

  await measure()
  await drawn(ui, '64%')
  // At PostCompact $.session.usage() still answers the fill from before (Claude Code 2.1.300)
  await $.classic.PostCompact({ trigger: 'manual', compact_summary: '' })
  await drawn(ui, '0%')
  // The next response measures it again
  fill = { tokens: 30_000, percent: 3 }
  await measure()
  await drawn(ui, '3%')
  await ui.unmount()
})

test('after /clear the drawer reads the new session again, as it did when the first one started', async ($, on) => {
  let claude = '2.1.300'
  on('command.register', (_, e) => ({ value: { command: e.name } }))
  on('clock.now', () => ({ value: 0 }))
  // No timers: what reads the version again is the clear alone
  on('clock.every', () => ({ deny: 'no timers here' }))
  on('ui.open', () => ({ value: { isPlaced: true } }))
  on('session.version', () => ({ value: { version: claude, base: claude } }))
  on('session.start', (_, e) => ({ cwd: e.cwd }))
  on('classic.SessionStart', () => ({}))
  const ui = await $.ui.mount(MOUNT)
  await $.session.start({ cwd: '/x', surface: 'terminal', isInteractive: true })
  await drawn(ui, '2.1.300')
  // The process goes on under a new session, which holds none of the old one's values, and no
  // session.start comes for it (Claude Code 2.1.302)
  claude = '2.1.302'
  await $.classic.SessionStart({ source: 'clear' })
  await drawn(ui, '2.1.302')
  await ui.unmount()
})
