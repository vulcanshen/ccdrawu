// The git boxes (user, 2026-11-06): beside the transcript, at the right end of the pet's walk (at the
// left the pet, which starts there, sat in front of them), a box for each file not committed, and a box
// taped shut, ready to go, for each commit not pushed. A commit takes boxes of files away and adds a
// taped one; a push takes the taped ones away. Square boxes, outlined in near black so they read apart
// from a pet of the same color in front of them (user: an orange cat ran into them), cardboard inside,
// the taped ones light inside (user: flaps, and more colors, were too much). Six at most, three a layer, a
// column apart, from the right: the taped ones first along the bottom, by the end. Kits jump up onto
// them and walk over them (hooks/pets.ts). Drawn in pixels as a pet is (drawnCells).

import type { GitInfo } from '../types'
import type { Coat } from './pets'

export const BOX = 4 // a box's pixels a side
const PER_LAYER = 3
const LAYERS = 2
const SLOT = BOX + 1 // a box and the column between it and the next
export const PILE_WIDTH = PER_LAYER * SLOT - 1
// Taped boxes take the bottom layer at most, so the files not committed still show
const PACKED_MAX = PER_LAYER

// Outlined all round, so a box on another reads apart from it too
const LOOSE = ['kkkk', 'kcck', 'kcck', 'kkkk']
const PACKED = ['kkkk', 'kttk', 'kttk', 'kkkk']
export const BOX_INK: Coat = { c: '#d39a5f', t: '#f5e3c0', k: '#2b2320' }

export type Pile = { packed: number; loose: number }

// The boxes for where git is: taped ones for the commits ahead of the upstream (none without one: what is
// not pushed is not known then), then one for each file changed, as many as there is room for
export function pileOf(git: GitInfo | null): Pile {
  if (git === null) return { packed: 0, loose: 0 }
  const packed = Math.min(git.ahead ?? 0, PACKED_MAX)
  const loose = Math.min(git.files ?? 0, PER_LAYER * LAYERS - packed)
  return { packed, loose }
}

// Where box i goes: its column, from the right, and its layer from the bottom
const placeOf = (i: number) => ({ x: PILE_WIDTH - BOX - (i % PER_LAYER) * SLOT, layer: Math.floor(i / PER_LAYER) })

// The pile in pixels, PILE_WIDTH wide and a layer of BOX high for each layer it has; none when it is empty
export function pileDrawing(p: Pile): string[] {
  const boxes = [...Array<string[]>(p.packed).fill(PACKED), ...Array<string[]>(p.loose).fill(LOOSE)]
  const height = Math.ceil(boxes.length / PER_LAYER) * BOX
  const rows = Array.from({ length: height }, () => Array<string>(PILE_WIDTH).fill('.'))
  boxes.forEach((art, i) => {
    const { x, layer } = placeOf(i)
    const top = height - (layer + 1) * BOX
    art.forEach((line, dy) => [...line].forEach((mark, dx) => (rows[top + dy]![x + dx] = mark)))
  })
  return rows.map(r => r.join(''))
}

// How high the pile is over each of its columns, in pixels
export function pileHeights(p: Pile): number[] {
  const heights = Array<number>(PILE_WIDTH).fill(0)
  for (let i = 0; i < p.packed + p.loose; i++) {
    const { x, layer } = placeOf(i)
    for (let dx = 0; dx < BOX; dx++) heights[x + dx] = (layer + 1) * BOX
  }
  return heights
}

// How high the ground is under width columns from left: the highest box under any of them
export function groundUnder(heights: readonly number[], left: number, width: number): number {
  let h = 0
  for (let c = left; c < left + width; c++) h = Math.max(h, heights[c] ?? 0)
  return h
}
