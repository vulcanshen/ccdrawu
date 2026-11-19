// Snails and owls (user, 2026-11-19): beside the transcript, for each shell left in the background a
// snail on the pet's strip, and for each monitor an owl, as many as there are, over one another, the
// git boxes and the pet as they come, each in colors made up at random when it comes. Toys, as the ball
// is: idle, the pet now and then goes over and paws at one (hooks/pets.ts), and a snail draws into its
// shell, an owl spreads its wings. A snail crawls in from beyond an end of the strip, about a column a
// second, wanders slowly, and crawls off past the nearer end once its shell is done (user: no hurry to
// show it there or gone); an owl flies down onto the strip, blinks now and then, now and then flies
// off to perch somewhere else, on the ground or on the git boxes as high as there is room over them, as
// a kit goes (user, 2026-11-19), and flies off up. Known from the background work a reply's end lists.
// An earlier screen for a shell and radar for a monitor, beside the boxes, went (user, 2026-11-06: too
// hard to show).

import type { Toy } from '../types'
import { drawnCells, TRACK, WALK_HEIGHT, type Coat, type Run, type Under } from './pets'
import { BOX_HEIGHT } from './pile'

const PET_WIDTH = 12 // a pet's columns: a toy is placed along the walk as a pet is
export const TOY_LEFT = 3 // a toy's columns from the left of the box a pet at its place would have
export const TOY_OFF = PET_WIDTH - TOY_LEFT // steps beyond an end of the walk where a snail is out of sight
const SNAIL_STEP_MS = 1_000
const SNAIL_REST_MS = 4_000 // and up to as long again, where it got to
export const SNAIL_HIDE_MS = 2_000 // in its shell after the pet has done pawing at it
export const OWL_WIDTH = 5 // its wings in
const OWL_HEIGHT = 6
const OWL_FLY = 2 // pixels a tick, up or down
const OWL_FLIGHT_CHANCE = 1 / 50 // a tick's, perched, of flying off to perch somewhere else: about every 10 s (user, 2026-11-19: 30 was too seldom)
// The highest it perches: on a layer of the boxes with room over it, as a kit goes (hooks/pets.ts)
export const OWL_TOP = Math.floor((WALK_HEIGHT - OWL_HEIGHT) / BOX_HEIGHT) * BOX_HEIGHT
const OWL_CRUISE = WALK_HEIGHT - OWL_HEIGHT // how high it flies, from one perch to another: the top of the strip
const BLINK_CHANCE = 1 / 25 // a tick's: about every five seconds
const BLINK_MS = 400

// Facing left; its eye stalk leans as it goes
const SNAIL: Record<'walk1' | 'walk2' | 'shell', string[]> = {
  walk1: ['b..ss.', 'b.sdss', 'bbsdds', 'bbbbbb'],
  walk2: ['.b.ss.', '.bsdss', 'bbsdds', 'bbbbbb'],
  shell: ['...ss.', '..sdss', '..sdds', '..sss.'],
}
// Facing out; its wings out a column either side, flying or pawed at
const OWL: Record<'stand' | 'blink' | 'wings', string[]> = {
  stand: ['o...o', 'ooooo', 'ykayk', 'obbbo', 'obbbo', '.a.a.'],
  blink: ['o...o', 'ooooo', 'ooaoo', 'obbbo', 'obbbo', '.a.a.'],
  wings: ['.o...o.', '.ooooo.', 'oykayko', 'oobbboo', '.obbbo.', '..a.a..'],
}

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

// Colors made up at random: a hue any way round, a lightness that shows on a dark background. A snail's
// shell of one hue, its whorl darker, a pale body of another; an owl's feathers muted, its breast paler,
// its eyes, beak and feet an owl's
function inkOf(kind: Toy['kind'], rand: () => number): Coat {
  const hue = () => rand() * 2 * Math.PI
  if (kind === 'snail') {
    const shell = hue()
    return { s: oklch(0.76, 0.09, shell), d: oklch(0.55, 0.08, shell), b: oklch(0.82, 0.035, hue()) }
  }
  const feathers = hue()
  return { o: oklch(0.64, 0.06, feathers), b: oklch(0.86, 0.035, feathers), y: '#f9e2af', k: '#11111b', a: '#fab387' }
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
    const toy: Toy = {
      id: task.id,
      kind,
      ink: inkOf(kind, rand),
      x,
      target: spot,
      faceLeft: spot < x,
      y: kind === 'owl' ? WALK_HEIGHT : 0,
      air: kind === 'owl',
      leaving: false,
      stepAt: 0,
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
  if (t.leaving) return t.y + OWL_FLY >= WALK_HEIGHT ? undefined : { ...t, y: t.y + OWL_FLY, air: true }
  // On its way to another perch: a step a tick, up along the top of the strip
  if (t.x !== t.target) {
    const faceLeft = t.target < t.x
    return { ...t, faceLeft, x: t.x + (faceLeft ? -1 : 1), y: Math.min(OWL_CRUISE, t.y + OWL_FLY), air: true, blinkUntil: 0 }
  }
  // There: down onto the boxes, as high as there is room over them, or the ground; up onto boxes that
  // came under it, down where they went
  const perch = Math.min(ground(t.x), OWL_TOP)
  if (t.y !== perch) {
    const y = t.y > perch ? Math.max(perch, t.y - OWL_FLY) : Math.min(perch, t.y + OWL_FLY)
    return { ...t, y, air: y !== perch }
  }
  if (t.air) return { ...t, air: false }
  if (t.hitUntil > 0) return now < t.hitUntil ? t : { ...t, hitUntil: 0 }
  if (t.blinkUntil > 0) return now < t.blinkUntil ? t : { ...t, blinkUntil: 0 }
  // Perched: now and then off to somewhere else along the walk, now and then a blink
  const r = rand()
  if (r < OWL_FLIGHT_CHANCE) {
    const spot = Math.round(rand() * TRACK)
    return { ...t, target: spot !== t.x ? spot : t.x < TRACK / 2 ? TRACK : 0 }
  }
  return r < OWL_FLIGHT_CHANCE + BLINK_CHANCE ? { ...t, blinkUntil: now + BLINK_MS } : t
}

// The pet has begun pawing at a toy, until then: a snail draws into its shell a while longer, an owl's
// wings are out till then
export function toyHit(toys: Toy[], id: string, until: number): Toy[] {
  return toys.map(t => (t.id === id ? { ...t, hitUntil: t.kind === 'snail' ? until + SNAIL_HIDE_MS : until } : t))
}

// A toy's pixels as it is now
export function toyDrawing(t: Toy): string[] {
  if (t.kind === 'owl') return t.air || t.hitUntil > 0 ? OWL.wings : t.blinkUntil > 0 ? OWL.blink : OWL.stand
  const art = t.hitUntil > 0 ? SNAIL.shell : t.stride % 2 === 0 ? SNAIL.walk1 : SNAIL.walk2
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

// The cells of a toy in a strip columns wide, standing on its bottom, under behind it, and the column
// they start at: its columns inside the strip alone, and its rows below the top, flying in or off
export function toyCells(t: Toy, columns: number, under?: Under): { left: number; width: number; cells: Run[][] } | undefined {
  const left = toyLeft(t, columns)
  const art = toyDrawing(t)
  const from = Math.max(0, -left)
  const to = Math.min(art[0]!.length, columns - left)
  const shown = art.slice(Math.max(0, art.length + t.y - WALK_HEIGHT)).map(r => r.slice(from, to))
  if (to <= from || shown.length === 0) return undefined
  const cells = drawnCells(shown, t.ink, WALK_HEIGHT, t.y, undefined, false, (y, x) => under?.(y, left + from + x))
  return { left: left + from, width: to - from, cells }
}
