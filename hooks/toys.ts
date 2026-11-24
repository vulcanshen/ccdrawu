// Snails and owls (user, 2026-11-19): beside the transcript, for each shell left in the background a
// snail on the pet's strip, a sea snail (user, 2026-11-20: a shell's) or the look the settings pick
// (SHELL_LOOKS), and for each monitor an owl, either kind none at all if the settings say so, as many
// as there are, over one another, the git boxes and the pet as they come, each in colors made up at
// random when it comes, in a hue away from
// those of its kind about (user, 2026-11-20: their colors too often alike). Toys, as the ball
// is: idle, the pet now and then goes over and paws at one (hooks/pets.ts), and a snail draws into its
// shell, an owl spreads its wings. A snail crawls in from beyond an end of the strip, about a column a
// second, wanders slowly, and crawls off past the nearer end once its shell is done (user: no hurry to
// show it there or gone); an owl flies down onto the strip, rests ten seconds, blinking and looking
// the other way now and then (user, 2026-11-24), flies ten seconds, its wings beating, looking where
// it goes, along the top of the strip, and comes down where it is, on the
// ground or on the git boxes as high as there is room over them, as a kit goes (user, 2026-11-19), to
// rest again (user, 2026-11-20), and flies off up. Known from the background work a reply's end lists.
// An earlier screen for a shell and radar for a monitor, beside the boxes, went (user, 2026-11-06: too
// hard to show).

import type { Toy } from '../types'
import { drawnCells, TRACK, WALK_HEIGHT, type Coat, type Run, type Under } from './pets'
import { BOX_HEIGHT } from './pile'

const PET_WIDTH = 12 // a pet's columns: a toy is placed along the walk as a pet is
export const TOY_LEFT = 3 // a toy's columns from the left of the box a pet at its place would have
const SNAIL_STEP_MS = 1_000
const SNAIL_REST_MS = 4_000 // and up to as long again, where it got to
export const SNAIL_HIDE_MS = 2_000 // in its shell after the pet has done pawing at it
export const OWL_WIDTH = 5 // its wings in
const OWL_HEIGHT = 6
const OWL_FLY = 2 // pixels a tick, up or down
// Down, it rests ten seconds, then flies ten (user, 2026-11-20: every 10 s off, 10 s flying, 10 s of rest)
const OWL_REST_MS = 10_000
const OWL_FLIGHT_MS = 10_000
// The highest it perches: on a layer of the boxes with room over it, as a kit goes (hooks/pets.ts)
export const OWL_TOP = Math.floor((WALK_HEIGHT - OWL_HEIGHT) / BOX_HEIGHT) * BOX_HEIGHT
const OWL_CRUISE = WALK_HEIGHT - OWL_HEIGHT // how high it flies: the top of the strip
const BLINK_CHANCE = 1 / 25 // a tick's: about every five seconds
const BLINK_MS = 400
const GLANCE_CHANCE = 1 / 15 // a tick's, resting: looking the other way about every three seconds

// The looks a shell's toy takes, facing left, the user's pick in the settings (pet.shell; user,
// 2026-11-20: every one tried on the way to the sea snail a choice, or none). The first, the default:
// a sea snail with a domed shell, its spire a nub on top, resting on the ground at the back, its body
// two columns in under it and out at the front, a tentacle up (user: half the shell on the ground, the
// body in under it two cells; 6 x 3, smaller than a kit). The sea snails are all so: as they go, what
// touches the ground moves, the front of the foot, out and in, and the tentacle stays; in the shell,
// the lid (d) shuts the opening. A hermit crab's legs move; the snail, as it was, leans its stalk
export type ShellLook = Record<'walk1' | 'walk2' | 'shell', string[]>
export const SHELL_LOOKS = {
  'sea-snail': { walk1: ['b..sd.', '.bsdss', 'bbbbss'], walk2: ['b..sd.', '.bsdss', '.bbbss'], shell: ['...sd.', '..sdss', '..ddss'] },
  'sea-snail-7x4': {
    walk1: ['b...d..', '.b.sss.', '.bsdsds', 'bbbbsss'],
    walk2: ['b...d..', '.b.sss.', '.bsdsds', '.bbbsss'],
    shell: ['....d..', '...sss.', '..sdsds', '..ddsss'],
  },
  'sea-snail-6x4': {
    walk1: ['b...d.', '.b.ss.', '.bsdss', 'bbbbss'],
    walk2: ['b...d.', '.b.ss.', '.bsdss', '.bbbss'],
    shell: ['....d.', '...ss.', '..sdss', '..ddss'],
  },
  'sea-snail-5x3': { walk1: ['b.sd.', '.bsds', 'bbbbs'], walk2: ['b.sd.', '.bsds', '.bbbs'], shell: ['..sd.', '..sds', '..dds'] },
  // Its spire up and back to a point
  spire: {
    walk1: ['......d', '.....sd', 'b..ssds', '.bsdsss', 'bbbbdss'],
    walk2: ['......d', '.....sd', 'b..ssds', '.bsdsss', '.bbbdss'],
    shell: ['......d', '.....sd', '...ssds', '..sdsss', '..dddss'],
  },
  // A wide opening, back to a horn's point
  horn: {
    walk1: ['b..ss..', '.bsdss.', '..sdsdd', '..ssss.', 'bbbbss.'],
    walk2: ['b..ss..', '.bsdss.', '..sdsdd', '..ssss.', '.bbbss.'],
    shell: ['...ss..', '..sdss.', '..sdsdd', '..ssss.', '..ddss.'],
  },
  // A long cone lying back, its point up
  cone: {
    walk1: ['b.ss...', '.bsdss.', '..ssdsd', 'bbbbdss'],
    walk2: ['b.ss...', '.bsdss.', '..ssdsd', '.bbbdss'],
    shell: ['..ss...', '..sdss.', '..ssdsd', '..dddss'],
  },
  // Its eye (p) on a stalk, a claw (c) out
  'hermit-crab': {
    walk1: ['p.sss.', 'c.sdds', 'ccsdds', '.c.c..'],
    walk2: ['p.sss.', 'c.sdds', 'ccsdds', 'c.c.c.'],
    shell: ['..sss.', '..sdds', '..sdds', '..sss.'],
  },
  snail: {
    walk1: ['b..ss.', 'b.sdss', 'bbsdds', 'bbbbbb'],
    walk2: ['.b.ss.', '.bsdss', 'bbsdds', 'bbbbbb'],
    shell: ['...ss.', '..sdss', '..sdds', '..sss.'],
  },
} satisfies Record<string, ShellLook>
// pet.shell's values, as the settings list them (config.PetShells), and pet.monitor's
export const SHELLS = [...Object.keys(SHELL_LOOKS), 'none']
export const MONITORS = ['owl', 'none']

// The look pet.shell names, the sea snail for one it does not name; none for none
export function shellLook(name: unknown): ShellLook | undefined {
  if (name === 'none') return undefined
  return Object.hasOwn(SHELL_LOOKS, name as string) ? SHELL_LOOKS[name as keyof typeof SHELL_LOOKS] : SHELL_LOOKS['sea-snail']
}

// The toys the config shows: a kind it says none for is not drawn, and the pet neither goes over to
// one of it nor stays up for it; it is kept, and shows again, where it got to, once the config names
// a look again
export function shownToys(toys: Toy[], pet?: { shell?: string; monitor?: string }): Toy[] {
  const hidden = (t: Toy) => (t.kind === 'snail' ? shellLook(pet?.shell) === undefined : pet?.monitor === 'none')
  return toys.some(hidden) ? toys.filter(t => !hidden(t)) : toys
}

// Steps beyond an end of the walk where a snail is out of sight: past the left the widest look's width
// as well
export const TOY_OFF = Math.max(PET_WIDTH - TOY_LEFT, TOY_LEFT + Math.max(...Object.values(SHELL_LOOKS).map(l => l.walk1[0]!.length)))
// Facing out; its wings out a column either side, pawed at; flying, they beat up, out, down and out, a
// frame a tick (user, 2026-11-19: out and still, they did not beat)
const OWL: Record<'stand' | 'blink' | 'wings' | 'up' | 'down', string[]> = {
  stand: ['o...o', 'ooooo', 'ykayk', 'obbbo', 'obbbo', '.a.a.'],
  blink: ['o...o', 'ooooo', 'ooaoo', 'obbbo', 'obbbo', '.a.a.'],
  wings: ['.o...o.', '.ooooo.', 'oykayko', 'oobbboo', '.obbbo.', '..a.a..'],
  up: ['o.o...o.o', '.ooooooo.', '..ykayk..', '..obbbo..', '..obbbo..', '...a.a...'],
  down: ['..o...o..', '..ooooo..', '..ykayk..', '.oobbboo.', 'oo.bbb.oo', '...a.a...'],
}
const WINGBEAT = [OWL.up, OWL.wings, OWL.down, OWL.wings]

// OKLCH to sRGB, its chroma brought into the gamut, lightness and hue kept
const gamma = (v: number) => (v <= 0.0031308 ? 12.92 * v : 1.055 * v ** (1 / 2.4) - 0.055)
function linear(L: number, C: number, h: number): number[] {
  const a = C * Math.cos(h)
  const b = C * Math.sin(h)
  const l = (L + 0.3963377774 * a + 0.2158037573 * b) ** 3
  const m = (L - 0.1055613458 * a - 0.0638541728 * b) ** 3
  const s = (L - 0.0894841775 * a - 1.291485548 * b) ** 3
  return [
    4.0767416621 * l - 3.3077115913 * m + 0.2309699292 * s,
    -1.2684380046 * l + 2.6097574011 * m - 0.3413193965 * s,
    -0.0041960863 * l - 0.7034186147 * m + 1.707614701 * s,
  ]
}
export function oklch(L: number, C: number, h: number): string {
  let c = C
  while (c > 0 && linear(L, c, h).some(v => v < 0 || v > 1)) c -= 0.001
  const hex = (v: number) =>
    Math.round(gamma(Math.min(1, Math.max(0, v))) * 255)
      .toString(16)
      .padStart(2, '0')
  return '#' + linear(L, Math.max(0, c), h).map(hex).join('')
}

// A hermit crab's claw and eye, the same on every one: drawn under a shell's own colors, so one from
// before there were hermit crabs has them too
const SHELL_INK: Coat = { c: oklch(0.7, 0.13, 0.55), p: '#f5f0e6' }

const TURN = 2 * Math.PI

// Colors made up at random, a lightness that shows on a dark background: a snail's shell of hue, its
// whorls and lid darker, a pale body of a hue any way round; an owl's feathers of hue, muted, its breast paler,
// its eyes, beak and feet an owl's
function inkOf(kind: Toy['kind'], hue: number, rand: () => number): Coat {
  if (kind === 'snail') {
    return { s: oklch(0.76, 0.09, hue), d: oklch(0.55, 0.08, hue), b: oklch(0.82, 0.035, rand() * TURN) }
  }
  return { o: oklch(0.64, 0.06, hue), b: oklch(0.86, 0.035, hue), y: '#f9e2af', k: '#11111b', a: '#fab387' }
}

// A hue at random away from those of the toys about: in the middle third of the widest gap between them,
// any way round when there are none. A toy from before hues were kept has none
function hueAmong(toys: readonly Toy[], rand: () => number): number {
  const hues = toys.flatMap(t => (Number.isFinite(t.hue) ? [((t.hue % TURN) + TURN) % TURN] : [])).sort((a, b) => a - b)
  if (hues.length === 0) return rand() * TURN
  let from = 0
  let width = -1
  hues.forEach((h, i) => {
    const before = i === 0 ? hues[hues.length - 1]! - TURN : hues[i - 1]!
    if (h - before > width) [from, width] = [before, h - before]
  })
  return (from + (width * (1 + rand())) / 3 + TURN) % TURN
}

const KINDS: Record<string, Toy['kind']> = { shell: 'snail', monitor: 'owl' }

// The toys for the background work a reply's end lists: one comes for each shell and monitor new in it,
// and one whose task is gone from it leaves. The same array back when none changed
export function toysFor(toys: Toy[], tasks: readonly { id: string; type: string }[], rand: () => number): Toy[] {
  let changed = false
  const next = toys.map(t => {
    if (t.leaving || tasks.some(task => task.id === t.id)) return t
    changed = true
    const target = t.kind === 'snail' ? (t.x < TRACK / 2 ? -TOY_OFF : TRACK + TOY_OFF) : t.x
    return { ...t, leaving: true, target, stepAt: 0 }
  })
  for (const task of tasks) {
    const kind = KINDS[task.type]
    if (kind === undefined || next.some(t => t.id === task.id)) continue
    const spot = Math.round(rand() * TRACK)
    // A snail from beyond an end; an owl over the spot it lands on, out of sight above the strip
    const x = kind === 'snail' ? (rand() < 0.5 ? -TOY_OFF : TRACK + TOY_OFF) : spot
    // Away from its own kind's: a snail and an owl already look apart, and the owls' muted colors want
    // the room (user, 2026-11-20: two of a kind in colors alike)
    const hue = hueAmong(next.filter(t => t.kind === kind), rand)
    const toy: Toy = {
      id: task.id,
      kind,
      hue,
      ink: inkOf(kind, hue, rand),
      x,
      target: spot,
      faceLeft: spot < x,
      y: kind === 'owl' ? WALK_HEIGHT : 0,
      air: kind === 'owl',
      leaving: false,
      stepAt: 0,
      flyUntil: 0,
      stride: 0,
      blinkUntil: 0,
      hitUntil: 0,
    }
    next.push(toy)
    changed = true
  }
  return changed ? next : toys
}

// The toys a tick later; one leaving goes once it is out of sight. ground is how high the git boxes are
// under an owl at a place along the walk, in pixels. The same array back when none changed
export function toysTick(toys: Toy[], now: number, rand: () => number = Math.random, ground: (x: number) => number = () => 0): Toy[] {
  let changed = false
  const next: Toy[] = []
  for (const t of toys) {
    const moved = t.kind === 'snail' ? snailStep(t, now, rand) : owlStep(t, now, rand, ground)
    if (moved !== t) changed = true
    if (moved !== undefined) next.push(moved)
  }
  return changed ? next : toys
}

function snailStep(t: Toy, now: number, rand: () => number): Toy | undefined {
  // In its shell: it comes out once that is over
  if (t.hitUntil > 0) return now < t.hitUntil ? t : { ...t, hitUntil: 0, stepAt: now + SNAIL_STEP_MS }
  if (t.leaving && t.x === t.target) return undefined
  if (now < t.stepAt) return t
  if (t.x !== t.target) {
    const faceLeft = t.target < t.x
    return { ...t, faceLeft, x: t.x + (faceLeft ? -1 : 1), stride: t.stride + 1, stepAt: now + SNAIL_STEP_MS }
  }
  // There: a rest, then somewhere else along the walk
  return { ...t, target: Math.round(rand() * TRACK), stepAt: now + SNAIL_REST_MS + Math.floor(rand() * SNAIL_REST_MS) }
}

function owlStep(t: Toy, now: number, rand: () => number, ground: (x: number) => number): Toy | undefined {
  // Flying, a wingbeat's frame a tick
  const stride = t.stride + 1
  if (t.leaving) return t.y + OWL_FLY >= WALK_HEIGHT ? undefined : { ...t, y: t.y + OWL_FLY, air: true, stride }
  // In its flight: a step a tick, up along the top of the strip, on to somewhere else along the walk
  // each time it gets where it was going, till the flight is over; then down where it is
  if (t.flyUntil > 0) {
    if (now >= t.flyUntil) return { ...t, flyUntil: 0, target: t.x, stride }
    const target = t.x !== t.target ? t.target : elsewhere(t.x, rand)
    const faceLeft = target < t.x
    return { ...t, target, faceLeft, x: t.x + (faceLeft ? -1 : 1), y: Math.min(OWL_CRUISE, t.y + OWL_FLY), air: true, stride }
  }
  // Down onto the boxes, as high as there is room over them, or the ground, its rest begun; up onto
  // boxes that came under it, down where they went, its rest going on
  const perch = Math.min(ground(t.x), OWL_TOP)
  if (t.y !== perch) {
    const y = t.y > perch ? Math.max(perch, t.y - OWL_FLY) : Math.min(perch, t.y + OWL_FLY)
    const stepAt = y === perch && t.stepAt <= now ? now + OWL_REST_MS : t.stepAt
    return { ...t, y, air: y !== perch, stepAt, stride }
  }
  if (t.air) return { ...t, air: false, stepAt: t.stepAt <= now ? now + OWL_REST_MS : t.stepAt }
  if (t.hitUntil > 0) return now < t.hitUntil ? t : { ...t, hitUntil: 0 }
  if (t.blinkUntil > 0) return now < t.blinkUntil ? t : { ...t, blinkUntil: 0 }
  // Rested: off on a flight; resting, now and then a blink, or a look the other way
  if (now >= t.stepAt) return { ...t, flyUntil: now + OWL_FLIGHT_MS, target: elsewhere(t.x, rand), air: true, stride }
  const r = rand()
  if (r < BLINK_CHANCE) return { ...t, blinkUntil: now + BLINK_MS }
  return r < BLINK_CHANCE + GLANCE_CHANCE ? { ...t, faceLeft: !t.faceLeft } : t
}

// A spot along the walk, at random, other than x: the far end when the draw is x
function elsewhere(x: number, rand: () => number): number {
  const spot = Math.round(rand() * TRACK)
  return spot !== x ? spot : x < TRACK / 2 ? TRACK : 0
}

// The pet has begun pawing at a toy, until then: a snail draws into its shell a while longer, an owl's
// wings are out till then
export function toyHit(toys: Toy[], id: string, until: number): Toy[] {
  return toys.map(t => (t.id === id ? { ...t, hitUntil: t.kind === 'snail' ? until + SNAIL_HIDE_MS : until } : t))
}

// A toy's pixels as it is now, a shell's in the look the config picks
export function toyDrawing(t: Toy, look: ShellLook = SHELL_LOOKS['sea-snail']): string[] {
  if (t.kind === 'owl') {
    const art = t.hitUntil > 0 ? OWL.wings : t.air ? WINGBEAT[t.stride % WINGBEAT.length]! : t.blinkUntil > 0 ? OWL.blink : OWL.stand
    // Its eyes the way it faces: the pupils (k) on the right of the eyes (y), or on the left
    return t.faceLeft ? art.map(r => r.replace(/yk/g, 'ky')) : art
  }
  const art = t.hitUntil > 0 ? look.shell : t.stride % 2 === 0 ? look.walk1 : look.walk2
  return t.faceLeft ? art : art.map(r => [...r].reverse().join(''))
}

// A toy's left column in a strip columns wide, where a pet at x along its walk would have its middle;
// beyond an end, a column a step, out of sight TOY_OFF steps on. An owl's wings out start a column left
export function toyLeft(t: Toy, columns: number): number {
  const wider = t.kind === 'owl' ? (toyDrawing(t)[0]!.length - OWL_WIDTH) / 2 : 0
  return toyAt(t.x, columns) - wider
}

// The left column of a toy at x along the walk, in a strip columns wide, an owl's wings in
export function toyAt(x: number, columns: number): number {
  const span = Math.max(0, columns - PET_WIDTH)
  return (x < 0 ? x : x > TRACK ? span + x - TRACK : Math.round((x / TRACK) * span)) + TOY_LEFT
}

// A toy's colors, a shell's over a hermit crab's claw and eye
export function toyInk(t: Toy): Coat {
  return t.kind === 'snail' ? { ...SHELL_INK, ...t.ink } : t.ink
}

// The cells of a toy in a strip columns wide, a shell's in look, standing on its bottom, under behind
// it, and the column they start at: its columns inside the strip alone, and its rows below the top,
// flying in or off
export function toyCells(
  t: Toy,
  columns: number,
  under?: Under,
  look?: ShellLook,
): { left: number; width: number; cells: Run[][] } | undefined {
  const left = toyLeft(t, columns)
  const art = toyDrawing(t, look)
  const from = Math.max(0, -left)
  const to = Math.min(art[0]!.length, columns - left)
  const shown = art.slice(Math.max(0, art.length + t.y - WALK_HEIGHT)).map(r => r.slice(from, to))
  if (to <= from || shown.length === 0) return undefined
  const cells = drawnCells(shown, toyInk(t), WALK_HEIGHT, t.y, undefined, false, (y, x) => under?.(y, left + from + x))
  return { left: left + from, width: to - from, cells }
}
