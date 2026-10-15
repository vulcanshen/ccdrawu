// The pets: five animals, 12 x 8 pixels each, drawn two pixels to a cell (▀ ▄), so 12 columns by 4 rows.
// Each faces right, its feet on the bottom row; a pet facing left is drawn the other way about. A
// jump is no drawing of its own: the pet as it strides, lifted off the ground (user, 2026-10-15: drawn
// apart it looked squashed). The marks are inks a coat colors: '#' the fur, 's' the second color
// (stripes, an ear), 'w' the light one (chest, belly), 'x' the more white a black-and-white cat
// has, 'e' the eye, 'n' the nose, 'p' pink, 'b' a bill and feet, and 'q' the limbs: legs,
// paws and feet together, the duck's wing and the penguin's flipper, in every pose, a color each coat
// names apart from the fur, as such an animal's would be (white socks, cream legs), so they stand out
// from the body and a paw a swipe raises does not run into the head (user, 2026-10-15: the limbs are
// colored from the first, all of them). A session picks one coat of its pet at random.

import type { PetBall, PetMode, PetState } from '../types'

export const SPECIES = ['cat', 'dog', 'rabbit', 'duck', 'penguin'] as const
export type Species = (typeof SPECIES)[number]
export const POSES = ['walk1', 'walk2', 'sit', 'sleep', 'lie', 'swipe1', 'swipe2'] as const
export type Pose = (typeof POSES)[number]
export type Coat = Record<string, string>

type Pet = { frames: Record<Pose, string[]>; coats: Coat[] }

export const PETS: Record<Species, Pet> = {
  // locku's cat; its sitting drawn lower, to fit in 8
  cat: {
    frames: {
      walk1: [
        'x.......#..#',
        '#.......####',
        '.s......##e#',
        '.##s#s#s##xx',
        '.##s#s#s#www',
        '..#xxxxxxww.',
        '..q.q...q.q.',
        '..q.q...q.q.',
      ],
      walk2: [
        'x.......#..#',
        '#.......####',
        '.s......##e#',
        '.##s#s#s##xx',
        '.##s#s#s#www',
        '..#xxxxxxww.',
        '.q..q..q..q.',
        'q...q..q...q',
      ],
      sit: [
        '.....#..#...',
        '.....####...',
        '.....##e#...',
        '.....##xx...',
        '...s#sww....',
        '..#s#sxx....',
        'x.#s#s#qq...',
        '.qq####qq...',
      ],
      sleep: [
        '............',
        '............',
        '............',
        '.......#..#.',
        '......####..',
        '..#s#s##xx..',
        '.##s#s#s##ww',
        'x#xxxxxxx#qq',
      ],
      lie: [
        '............',
        '............',
        '........#..#',
        '........####',
        '........##e#',
        '..#s#s#s##xx',
        '.##s#s#s#www',
        'x#xxxxxxx.qq',
      ],
      swipe1: [
        '.....#..#...',
        '.....####...',
        '.....##e#...',
        '.....##xx...',
        '...s#swwqq..',
        '..#s#sxx....',
        'x.#s#s#qq...',
        '.qq####q....',
      ],
      swipe2: [
        '.....#..#...',
        '.....####.q.',
        '.....##e#q..',
        '.....##xxq..',
        '...s#sww....',
        '..#s#sxx....',
        'x.#s#s#qq...',
        '.qq####q....',
      ],
    },
    // locku's five: an orange tabby, an amber, a white, a grey-blue, a black-and-white
    coats: [
      { q: '#f4f1ea', '#': '#f6b06a', s: '#a8460c', w: '#f6b06a', x: '#f6b06a', e: '#7ed957' },
      { q: '#9a958e', '#': '#7a4a2a', s: '#e8b830', w: '#9a958e', x: '#7a4a2a', e: '#7ed957' },
      { q: '#c4c0b8', '#': '#f4f1ea', s: '#8c8a86', w: '#f4f1ea', x: '#f4f1ea', e: '#5fa8ff' },
      { q: '#f4f1ea', '#': '#9fb2c8', s: '#4e5d70', w: '#9fb2c8', x: '#9fb2c8', e: '#7ed957' },
      { q: '#f4f1ea', '#': '#55504c', s: '#55504c', w: '#f4f1ea', x: '#f4f1ea', e: '#7ed957' },
    ],
  },
  dog: {
    frames: {
      walk1: [
        '........###.',
        '#......s##e#',
        '.#.....s###n',
        '.#########w.',
        '.########ww.',
        '..########..',
        '..q.q...q.q.',
        '..q.q...q.q.',
      ],
      walk2: [
        '........###.',
        '#......s##e#',
        '.#.....s###n',
        '.#########w.',
        '.########ww.',
        '..########..',
        '.q..q..q..q.',
        'q...q..q...q',
      ],
      sit: [
        '......###...',
        '.....s##e#..',
        '.....s###n..',
        '......##w...',
        '#....###ww..',
        '.#..####ww..',
        '..######w...',
        '..q###.qq...',
      ],
      sleep: [
        '............',
        '............',
        '............',
        '........###.',
        '......s####n',
        '..######s##.',
        '.##########.',
        '#qq######qq.',
      ],
      lie: [
        '............',
        '............',
        '........###.',
        '.......s##e#',
        '#......s###n',
        '.#########w.',
        '.##########.',
        '.q########qq',
      ],
      swipe1: [
        '......###...',
        '.....s##e#..',
        '.....s###n..',
        '......##w...',
        '#....###wqqq',
        '.#..####w...',
        '..######w...',
        '..q###.q....',
      ],
      swipe2: [
        '......###...',
        '.....s##e#q.',
        '.....s###nq.',
        '......##wq..',
        '#....###w...',
        '.#..####w...',
        '..######w...',
        '..q###.q....',
      ],
    },
    // a shiba, a black-and-tan, a white
    coats: [
      { q: '#f4ead8', '#': '#d98b4a', s: '#a8602c', w: '#f4ead8', e: '#3a2a20', n: '#2b2b2b' },
      { q: '#d9a066', '#': '#6b625a', s: '#3e3934', w: '#d9a066', e: '#f0c060', n: '#2b2b2b' },
      { q: '#d8c8b0', '#': '#eee8dc', s: '#d8c8b0', w: '#ffffff', e: '#3a2a20', n: '#2b2b2b' },
    ],
  },
  rabbit: {
    frames: {
      walk1: [
        '........#.#.',
        '........p.p.',
        '........#.#.',
        '.......##e#.',
        '..#########p',
        'w#########..',
        '.########...',
        '..qq...qq...',
      ],
      walk2: [
        '........#.#.',
        '........p.p.',
        '........#.#.',
        '.......##e#.',
        '..#########p',
        'w#########..',
        '.#######....',
        '.qq.....qq..',
      ],
      sit: [
        '......#.#...',
        '......p.p...',
        '......#.#...',
        '.....##e#...',
        '.....####p..',
        '...#####....',
        'w.######....',
        '..qq.qq.....',
      ],
      sleep: [
        '............',
        '............',
        '............',
        '............',
        '.....####...',
        '...##pp####.',
        'w##########p',
        '.qq######qq.',
      ],
      lie: [
        '............',
        '........#.#.',
        '........p.p.',
        '........#.#.',
        '.......##e#.',
        '..#########p',
        'w##########.',
        '.qq######qq.',
      ],
      swipe1: [
        '......#.#...',
        '......p.p...',
        '......#.#...',
        '.....##e#...',
        '.....####p..',
        '...#####....',
        'wqq.####....',
        '.....qq.....',
      ],
      swipe2: [
        '......#.#...',
        '......p.p...',
        '......#.#...',
        '.....##e#...',
        '.....####p..',
        '...#####....',
        'w.######....',
        '.qqq.qq.....',
      ],
    },
    // a grey, a white with red eyes, a brown
    coats: [
      { q: '#f4f1ea', '#': '#b8b4ac', p: '#f2a7b5', w: '#f4f1ea', e: '#3a2a40' },
      { q: '#c4c0b8', '#': '#f0ece4', p: '#f2a7b5', w: '#ffffff', e: '#d9534f' },
      { q: '#f4f1ea', '#': '#9a6a44', p: '#e8a0a8', w: '#f4f1ea', e: '#2a1a10' },
    ],
  },
  duck: {
    frames: {
      walk1: [
        '.......###..',
        '......##e#..',
        '......####bb',
        '.......##...',
        '#.....####..',
        '##qq######..',
        '.########...',
        '....b..b....',
      ],
      walk2: [
        '.......###..',
        '......##e#..',
        '......####bb',
        '.......##...',
        '#.....####..',
        '##qq######..',
        '.########...',
        '...b....b...',
      ],
      sit: [
        '............',
        '.......###..',
        '......##e#..',
        '......####bb',
        '.......##...',
        '#.....####..',
        '##qq######..',
        '.########...',
      ],
      sleep: [
        '............',
        '............',
        '............',
        '............',
        '......###...',
        '#....#####..',
        '##qq#######.',
        '.#########..',
      ],
      lie: [
        '............',
        '............',
        '............',
        '........###.',
        '.......##e#.',
        '#.....####bb',
        '##qq#######.',
        '.#########..',
      ],
      swipe1: [
        '...q...###..',
        '...qq.##e#..',
        '....qq####bb',
        '.......##...',
        '#.....####..',
        '##########..',
        '.########...',
        '....b..b....',
      ],
      swipe2: [
        '.......###..',
        '......##e#..',
        '......####bb',
        '.qq....##...',
        '#.qqq.####..',
        '##########..',
        '.########...',
        '....b..b....',
      ],
    },
    // a yellow duckling, a white duck, a brown one
    coats: [
      { q: '#c99a2e', '#': '#f2d24b', b: '#f08a24', e: '#2b2b2b' },
      { q: '#c4c0b8', '#': '#f2efe8', b: '#f0a830', e: '#2b2b2b' },
      { q: '#6a5038', '#': '#a88a64', b: '#e8c040', e: '#2b2b2b' },
    ],
  },
  penguin: {
    frames: {
      walk1: [
        '.....####...',
        '....###e#bb.',
        '....####w...',
        '...q###ww...',
        '..qq###www..',
        '...####www..',
        '...####ww...',
        '.....b..b...',
      ],
      walk2: [
        '.....####...',
        '....###e#bb.',
        '....####w...',
        '...q###ww...',
        '...q####www.',
        '...####www..',
        '...####ww...',
        '....b...b...',
      ],
      sit: [
        '.....####...',
        '....###e#bb.',
        '....####w...',
        '...q###ww...',
        '...q###www..',
        '...####www..',
        '...####ww...',
        '....bb.bb...',
      ],
      sleep: [
        '............',
        '............',
        '............',
        '............',
        '........###.',
        '..########bb',
        '.###qq####w.',
        '..wwwwwwww..',
      ],
      lie: [
        '............',
        '............',
        '............',
        '.......####.',
        '......###e#b',
        '.###qq####w.',
        '.##########.',
        '..wwwwwwww..',
      ],
      swipe1: [
        '.....####...',
        '.q..###e#bb.',
        '..q.####w...',
        '...#####w...',
        '...####www..',
        '...####www..',
        '...####ww...',
        '.....b..b...',
      ],
      swipe2: [
        '.....####...',
        '....###e#bb.',
        '....####w...',
        'qqq####ww...',
        '...####www..',
        '...####www..',
        '...####ww...',
        '.....b..b...',
      ],
    },
    // a dark grey one, a little blue
    coats: [
      { q: '#8a93a6', '#': '#5a6170', w: '#f4f1ea', b: '#f0a830', e: '#f4f1ea' },
      { q: '#4a6a94', '#': '#6f8fb8', w: '#f4f1ea', b: '#3a3f4b', e: '#f4f1ea' },
    ],
  },
}

// A run of one look along a row of cells: a gap to leave, or characters in their colors
export type Run = { gap: number } | { ch: string; fg?: string; bg?: string }
type Cell = { ch: string; fg?: string; bg?: string; text?: boolean } | undefined

// Over a pet's head, in text, not pixels, in the top row of cells at the side it faces: a sleeping
// pet snores, z, zz, zzZ and round again, in the room every sleeping pose leaves there; a pet patted
// sits, and a heart goes up in the corner every sitting pose leaves free. Beside the transcript the
// box is taller than the pet, and the text goes in the room over it
const SNORES = ['z', 'zz', 'zzZ']
export const SNORE_COLOR = '#f9e2af'
export const HEART = '♥'
export const HEART_COLOR = '#f38ba8'
export type Over = { text: string; color: string }
export const PET_HEIGHT = 8 // a pet's pixels, high
export const LIFT_ROOM = 2 // the pixels over it, beside the transcript, for it to jump into

// The snore n steps along, from 1
export function snore(n: number): Over {
  return { text: SNORES[(n - 1) % SNORES.length] ?? 'z', color: SNORE_COLOR }
}

// Coat n of the animal, counting round: a session keeps its number when the config changes animal
export function petCoat(species: Species, n: number): Coat {
  const coats = PETS[species].coats
  return coats[Math.max(0, n) % coats.length] ?? {}
}

// The cells of a pose, facing left drawn the other way about, over over the head; see drawnCells
export function petCells(
  species: Species,
  pose: Pose,
  coat: Coat,
  faceLeft = false,
  over?: Over,
  height = PET_HEIGHT,
  lift = 0,
): Run[][] {
  const drawn = PETS[species].frames[pose].map(r => (faceLeft ? [...r].reverse().join('') : r))
  return drawnCells(drawn, coat, height, lift, over, faceLeft)
}

// The cells of a drawing, a row of runs for each two rows of pixels: ▀ in the upper pixel's color on
// the lower one's, ▄ when only the lower one is there, a gap when neither is. height is the box's, in
// pixels (even, at least the drawing's); the drawing stands on its bottom, lift pixels up it when it
// jumps. over goes in the top row, at the right, or at the left when overLeft.
export function drawnCells(drawn: string[], coat: Coat, height = PET_HEIGHT, lift = 0, over?: Over, overLeft = false): Run[][] {
  const faceLeft = overLeft
  const width = drawn[0]?.length ?? 0
  const above = Math.max(0, height - drawn.length - lift)
  const empty = '.'.repeat(width)
  const frame = [...Array<string>(above).fill(empty), ...drawn, ...Array<string>(Math.max(0, height - above - drawn.length)).fill(empty)]
  const at = (y: number, x: number) => coat[frame[y]?.[x] ?? '.']
  const grid: Cell[][] = []
  for (let y = 0; y < frame.length; y += 2) {
    grid.push(
      Array.from({ length: width }, (_, x) => {
        const top = at(y, x)
        const bottom = at(y + 1, x)
        if (top === undefined && bottom === undefined) return undefined
        return top === undefined ? { ch: '▄', fg: bottom } : { ch: '▀', fg: top, bg: bottom }
      }),
    )
  }
  const top = grid[0]
  if (over !== undefined && top !== undefined) {
    // From the right edge leftwards facing right, from the left edge facing left
    const text = faceLeft ? [...over.text].reverse().join('') : over.text
    const from = faceLeft ? 0 : width - text.length
    ;[...text].forEach((ch, i) => (top[from + i] = { ch, fg: over.color, text: true }))
  }
  return grid.map(row => {
    const runs: Run[] = []
    for (const cell of row) {
      const last = runs[runs.length - 1]
      if (cell === undefined) {
        if (last && 'gap' in last) last.gap++
        else runs.push({ gap: 1 })
      } else if (cell.text && last && 'ch' in last && last.fg === cell.fg && last.bg === undefined) {
        last.ch += cell.ch
      } else runs.push({ ch: cell.ch, fg: cell.fg, bg: cell.bg })
    }
    return runs
  })
}

// What a pet does, beside Claude. While Claude works it wanders: it picks a place along its walk
// and goes there, running now and then when it is far, and there it sits a while, hops, or goes
// straight on to the next, or now and then swipes at the air (user, 2026-10-14: to and fro alone was
// dull). It climbed the ends of its walk for a while; every way of drawing that looked odd, and it went
// (user, 2026-10-15). When Claude is done it stops and sits; idle it winds down, sitting, lying down
// after two minutes and asleep after ten, now and then strolling a little, looking the other way or
// swiping, and wakes the moment anything goes on. A tool failing makes it jump; a reply done, a
// compaction or a pat gets a line out of it; a pat sits it down a moment, a heart over its head, and
// wakes it. A ball thrown in (/ccdrawu ball) comes from its far side, bounces and rolls to a stop;
// it runs over, paws at it a while, and has it (user, 2026-10-15, after vscode-pets). None of it
// touches `$`: the hooks feed it events, a tick and a random source.

export const PET_TICK = 200 // ms between ticks
export const TRACK = 20 // steps from one end of its walk to the other
const LINE_MS = 8_000
const HEART_MS = 3_000
const JUMP_MS = 600
const LIE_AFTER = 2 * 60_000
const SLEEP_AFTER = 10 * 60_000
const SNORE_MS = 800
const SWIPE_MS = 1_200 // a swipe at the air
const NEAREST = 4 // the next place is at least this far from where it is
const IDLE_CHANCE = 1 / 150 // a tick's chance, idle, of a stroll, a look round or a swipe: about every 30 s
const ROLL_MS = 1_200 // a ball rolls this long to a stop
const PLAY_MS = 1_600 // the pet paws at it this long
const BALL_MAX_MS = 20_000 // a ball not had by then goes
export const BALL_REACH = 5 // how far from the ball the pet stands to paw at it, along its walk
export const BALL_LEFT = 4 // the ball's columns from the left of the box a pet at its place would have

// The ball, 4 x 4 pixels on the ground: red, a shine on it
export const BALL = ['.rr.', 'rhrr', 'rrrr', '.rr.']
export const BALL_INK: Coat = { r: '#f38ba8', h: '#fde2e7' }
const BOUNCE = [2, 1, 0, 1, 0] // its pixels off the ground, a tick at a time, as it comes in

export const PET_START: PetState = {
  coat: -1,
  busy: false,
  active: 0,
  mode: 'sit',
  x: 0,
  target: 0,
  run: false,
  stride: 0,
  restUntil: 0,
  lift: 0,
  faceLeft: false,
  snore: 0,
  heart: false,
  lineUntil: 0,
  heartUntil: 0,
  jumpFrom: 0,
  jumpUntil: 0,
  ballPos: 0,
  ballLift: 0,
  playUntil: 0,
}

// The next place to go: a few tries for one far enough, else the far end
function nextTarget(x: number, rand: () => number): number {
  for (let i = 0; i < 4; i++) {
    const t = Math.round(rand() * TRACK)
    if (Math.abs(t - x) >= NEAREST) return t
  }
  return x < TRACK / 2 ? TRACK : 0
}

// busy: Claude is at work; done, stop, failed: its turn ended with an answer, stopped by the user, or
// on an error; error: a tool failed while it works on; pat, ball: the user's
export type PetEvent = 'busy' | 'done' | 'stop' | 'failed' | 'error' | 'compact' | 'pat' | 'ball'

// Lines are plain ASCII: a character two cells wide would push the row past the pane
const LINES: Record<'done' | 'error' | 'compact', string[]> = {
  done: ['All done!', 'Nice work.', 'Ta-da!', 'Finished!', 'That went well.'],
  error: ['Uh oh...', 'Yikes!', 'Oops!', 'That broke.', 'Hmm...'],
  compact: ['Tidied up!', 'Squeezed it in.', 'Packed it up.', 'Lighter now.'],
}
export const PAT_LINES: Record<Species, string[]> = {
  cat: ['Purr...', 'Mrrp!', 'Meow.'],
  dog: ['Woof!', '*wag wag*', 'Arf!'],
  rabbit: ['*happy hop*', '*nose wiggle*', 'Eep!'],
  duck: ['Quack!', 'Quack quack!', '*waddle*'],
  penguin: ['*flap flap*', 'Squawk!', '*happy slide*'],
}
export const BALL_LINES: Record<Species, string[]> = {
  cat: ['*pounce*', 'Got it!', 'Mine!'],
  dog: ['Fetch!', 'Got it!', '*tail wags*'],
  rabbit: ['*boing*', 'Got it!', 'Hop hop!'],
  duck: ['Quack! Got it!', 'Mine!', '*waddle waddle*'],
  penguin: ['*belly slide*', 'Got it!', 'Squawk!'],
}

// Where a ball is along the walk, and how far off the ground, as it rolls in to a stop
export function ballAt(b: PetBall, now: number): { pos: number; lift: number } {
  const t = Math.min(1, Math.max(0, (now - b.from) / ROLL_MS))
  // Slowing as it goes
  const pos = Math.round(b.start + (b.rest - b.start) * (1 - (1 - t) ** 2))
  const lift = BOUNCE[Math.floor((now - b.from) / PET_TICK)] ?? 0
  return { pos, lift }
}

const pick = (lines: readonly string[], rand: () => number) => lines[Math.floor(rand() * lines.length) % lines.length] ?? ''

// A jump from now
const jumps = (now: number) => ({ jumpFrom: now, jumpUntil: now + JUMP_MS })

// What an event does to the pet; the tick after it shows it
export function petEvent(s: PetState, event: PetEvent, now: number, species: Species, rand: () => number): PetState {
  const said = (line: string) => ({ line, lineUntil: now + LINE_MS })
  // A turn over, or a pat: it stops where it is
  const stops = { target: s.x, restUntil: 0 }
  switch (event) {
    case 'busy':
      return { ...s, busy: true, active: now }
    case 'done':
      return { ...s, ...stops, busy: false, active: now, ...said(pick(LINES.done, rand)) }
    case 'stop':
      return { ...s, ...stops, busy: false, active: now }
    case 'failed':
      return { ...s, ...stops, busy: false, active: now, ...jumps(now), ...said(pick(LINES.error, rand)) }
    case 'error':
      return { ...s, active: now, ...jumps(now), ...said(pick(LINES.error, rand)) }
    case 'compact':
      return { ...s, active: now, ...said(pick(LINES.compact, rand)) }
    case 'pat':
      return { ...s, ...stops, active: now, heartUntil: now + HEART_MS, ...said(pick(PAT_LINES[species], rand)) }
    case 'ball': {
      // From its far side, to a stop in the middle stretch; it will stand on the nearer side of it
      const start = s.x < TRACK / 2 ? TRACK : 0
      const rest = BALL_REACH + Math.floor(rand() * (TRACK - 2 * BALL_REACH + 1))
      const sides = [rest - BALL_REACH, rest + BALL_REACH]
      const goal = sides.reduce((a, b) => (Math.abs(b - s.x) < Math.abs(a - s.x) ? b : a))
      const line = pick(BALL_LINES[species], rand)
      return { ...s, ...stops, active: now, heartUntil: 0, playUntil: 0, ball: { from: now, start, rest, goal, line } }
    }
  }
}

// The pet a tick later. The same object back when nothing in it changed, so a pet sitting still
// costs no redraw
export function petTick(s: PetState, now: number, rand: () => number = Math.random): PetState {
  const heart = now < s.heartUntil
  const asleep = !s.busy && now - s.active >= SLEEP_AFTER
  let { x, target, run, stride, restUntil, jumpFrom, jumpUntil, faceLeft, ball, playUntil, line, lineUntil } = s
  let moving = false
  // A ball: it runs to it, paws at it once it has stopped rolling, and has it; one not had in time goes
  if (ball !== undefined && now - ball.from > BALL_MAX_MS) [ball, playUntil] = [undefined, 0]
  if (ball !== undefined && !heart) {
    if (playUntil > 0 && now >= playUntil) {
      ;[line, lineUntil] = [ball.line, now + LINE_MS]
      ;[ball, playUntil] = [undefined, 0]
    } else if (playUntil === 0 && x === ball.goal && now >= ball.from + ROLL_MS) {
      playUntil = now + PLAY_MS
      faceLeft = ball.goal > ball.rest
    } else if (playUntil === 0) {
      ;[target, run, restUntil] = [ball.goal, true, 0]
    }
  }
  const playing = now < playUntil
  if (playing && !heart) stride++
  if (!heart && !asleep && !playing && now >= jumpUntil && now >= restUntil) {
    if (x !== target) {
      // On its way: a step, two when it runs, facing where it goes
      const step = Math.min(run ? 2 : 1, Math.abs(target - x))
      faceLeft = target < x
      x += faceLeft ? -step : step
      stride++
      moving = true
    } else if (ball !== undefined) {
      // At the ball's place before it has stopped: it waits for it
    } else if (s.busy) {
      // There: it swipes at the air, sits a while, hops, or goes straight on; the next place, and a run
      // if it is far
      const r = rand()
      if (r < 0.15) playUntil = now + SWIPE_MS
      else if (r < 0.45) restUntil = now + 1_000 + Math.floor(rand() * 2_000)
      else if (r < 0.65) ({ jumpFrom, jumpUntil } = jumps(now))
      target = nextTarget(x, rand)
      run = Math.abs(target - x) > TRACK / 2 && rand() < 0.4
    } else if (rand() < IDLE_CHANCE) {
      // Idle: now and then a stroll, a look the other way, or a swipe
      const r = rand()
      if (r < 0.4) [target, run] = [nextTarget(x, rand), false]
      else if (r < 0.7) faceLeft = !faceLeft
      else playUntil = now + SWIPE_MS
    }
  }
  // A swipe begun this tick is pawing at once
  const pawing = playing || (!heart && now < playUntil)
  if (pawing && !playing) stride++
  const lying = !s.busy && now - s.active >= LIE_AFTER
  const mode: PetMode =
    now < jumpUntil
      ? 'jump'
      : heart
        ? 'sit'
        : pawing
          ? 'play'
          : asleep
            ? 'sleep'
            : moving
              ? 'walk'
              : lying
                ? 'lie'
                : 'sit'
  // Off the ground in a jump: up and down a pixel a tick, 1, 2, 1
  const lift =
    mode === 'jump'
      ? Math.min(LIFT_ROOM, Math.floor((now - jumpFrom) / PET_TICK) + 1, Math.ceil((jumpUntil - now) / PET_TICK))
      : 0
  const snoreAt = mode === 'sleep' ? (Math.floor(now / SNORE_MS) % 3) + 1 : 0
  if (line !== undefined && now >= lineUntil) line = undefined
  const rolled = ball === undefined ? { pos: 0, lift: 0 } : ballAt(ball, now)
  const next: PetState = {
    ...s,
    mode,
    x,
    target,
    run,
    stride,
    restUntil,
    jumpFrom,
    jumpUntil,
    lift,
    faceLeft,
    snore: snoreAt,
    heart,
    line,
    lineUntil,
    ball,
    ballPos: rolled.pos,
    ballLift: rolled.lift,
    playUntil,
  }
  return (Object.keys(next) as (keyof PetState)[]).every(k => next[k] === s[k]) ? s : next
}

// How to draw it: the pose (its legs take turns as it goes), what goes over its head, and how far off
// the ground it is
export function petLook(s: PetState): { pose: Pose; over?: Over; lift: number } {
  switch (s.mode) {
    case 'walk':
      return { pose: s.stride % 2 === 0 ? 'walk1' : 'walk2', lift: 0 }
    case 'jump':
      // Its legs apart, as in a stride
      return { pose: 'walk2', lift: s.lift }
    case 'play':
      // Pawing at the ball, or at the air, one way and the other
      return { pose: s.stride % 2 === 0 ? 'swipe1' : 'swipe2', lift: 0 }
    case 'lie':
      return { pose: 'lie', lift: 0 }
    case 'sleep':
      return { pose: 'sleep', over: s.snore > 0 ? snore(s.snore) : undefined, lift: 0 }
    case 'sit':
      return { pose: 'sit', over: s.heart ? { text: HEART, color: HEART_COLOR } : undefined, lift: 0 }
  }
}

// The pet the config names, cat when it names none or one there is not
export function petSpecies(name: unknown): Species {
  return SPECIES.find(s => s === name) ?? 'cat'
}
