// The git boxes (user, 2026-11-06): beside the transcript, at the right end of the pet's walk (at the
// left the pet, which starts there, sat in front of them), a box for each file not committed, and a box
// taped shut, ready to go, for each commit not pushed. A commit takes boxes of files away and adds a
// taped one; a push takes the taped ones away. Boxes in three inks of a dark brown (user, 2026-11-08): a
// taped one shut under a lid, a pixel of it down the middle where the flaps meet, one not taped open in
// the middle of the top (user, 2026-11-08: told apart by how they look, not by their color); lighter
// inside, and round the sides and the bottom a frame, darker still. The earth colors of cardboard, shaded
// down to the lightness of a dark grey tried first: as light as the pets, they ran into the orange and
// brown ones in front of them (user, 2026-11-06: an orange cat). Nine at most, three a layer, three
// layers (user, 2026-11-08: two left the room over them in the strip empty), a column apart, from the
// right: the taped ones first along the bottom, by the end. Kits go on any layer of them but the top of
// the third, with no room over it (hooks/pets.ts). Drawn in pixels as a pet is (drawnCells).

import type { GitInfo } from '../types'
import type { Coat } from './pets'

const BOX_WIDTH = 5 // a box's pixels, across: odd, for the seam down the middle
export const BOX_HEIGHT = 4 // a layer of boxes, in pixels: kits stand on the top of one (hooks/pets.ts)
const PER_LAYER = 3
const LAYERS = 3
const SLOT = BOX_WIDTH + 1 // a box and the column between it and the next
export const PILE_WIDTH = PER_LAYER * SLOT - 1
// Taped boxes take the bottom layer at most, so the files not committed still show
const PACKED_MAX = PER_LAYER

// Open in the middle of the top, or shut under the lid over the frame; the same inks
const LOOSE = ['kkckk', 'kccck', 'kccck', 'kkkkk']
const PACKED = ['lllll', 'kclck', 'kccck', 'kkkkk']
export const BOX_INK: Coat = { l: '#392310', c: '#49331c', k: '#2a1d14' }

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
const placeOf = (i: number) => ({ x: PILE_WIDTH - BOX_WIDTH - (i % PER_LAYER) * SLOT, layer: Math.floor(i / PER_LAYER) })

// The pile in pixels, PILE_WIDTH wide and a layer of BOX_HEIGHT high for each layer it has; none when it is empty
export function pileDrawing(p: Pile): string[] {
  const boxes = [...Array<string[]>(p.packed).fill(PACKED), ...Array<string[]>(p.loose).fill(LOOSE)]
  const height = Math.ceil(boxes.length / PER_LAYER) * BOX_HEIGHT
  const rows = Array.from({ length: height }, () => Array<string>(PILE_WIDTH).fill('.'))
  boxes.forEach((art, i) => {
    const { x, layer } = placeOf(i)
    const top = height - (layer + 1) * BOX_HEIGHT
    art.forEach((line, dy) => [...line].forEach((mark, dx) => (rows[top + dy]![x + dx] = mark)))
  })
  return rows.map(r => r.join(''))
}

// How high the pile is over each of its columns, in pixels
export function pileHeights(p: Pile): number[] {
  const heights = Array<number>(PILE_WIDTH).fill(0)
  for (let i = 0; i < p.packed + p.loose; i++) {
    const { x, layer } = placeOf(i)
    for (let dx = 0; dx < BOX_WIDTH; dx++) heights[x + dx] = (layer + 1) * BOX_HEIGHT
  }
  return heights
}

// How high the ground is under width columns from left: the highest box under any of them
export function groundUnder(heights: readonly number[], left: number, width: number): number {
  let h = 0
  for (let c = left; c < left + width; c++) h = Math.max(h, heights[c] ?? 0)
  return h
}
