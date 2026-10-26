// A small YAML reader for config.yaml (tdp E3): what the ccdrawu command writes, and what a
// person writes there by hand in YAML's usual ways. Block mappings and sequences; flow [..] and {..}, and so
// JSON too; plain, single- and double-quoted scalars, and block scalars (| and >); comments. Anchors, tags
// and documents after the first are not read: they throw, as text that is not YAML does.
// None touches `$`, so tests can call it.

type Line = { indent: number; text: string }

export function parseYaml(source: string): unknown {
  const lines: Line[] = []
  for (const raw of source.replace(/^﻿/, '').split(/\r?\n/)) {
    if (/^ *\t/.test(raw) && raw.trim() !== '') throw new Error('a tab indents a line')
    lines.push({ indent: raw.length - raw.trimStart().length, text: raw.trimEnd() })
  }
  let at = 0
  const blank = (l: Line) => l.text.trim() === '' || l.text.trimStart().startsWith('#')
  const skip = () => {
    while (at < lines.length && blank(lines[at]!)) at++
  }
  // the next line with something on it, at `indent` or deeper
  const next = (indent: number): Line | undefined => {
    skip()
    const line = lines[at]
    if (line === undefined || line.indent < indent) return undefined
    const t = line.text.trim()
    if (t === '---') {
      if (at === 0 || lines.slice(0, at).every(blank)) {
        at++
        return next(indent)
      }
      throw new Error('more than one document')
    }
    if (t === '...') return undefined
    return line
  }

  function node(indent: number): unknown {
    const line = next(indent)
    if (line === undefined) return null
    const t = line.text.trim()
    if (t === '-' || t.startsWith('- ')) return sequence(line.indent)
    if (keyOf(t) !== undefined) return mapping(line.indent)
    at++
    return inline(t, line.indent)
  }

  function sequence(indent: number): unknown[] {
    const out: unknown[] = []
    for (let line = next(indent); line && line.indent === indent; line = next(indent)) {
      const t = line.text.trim()
      if (t !== '-' && !t.startsWith('- ')) break
      const rest = t.slice(1).trimStart()
      if (rest === '' || rest.startsWith('#')) {
        at++
        out.push(node(indent + 1))
        continue
      }
      // `- key: value` starts a mapping as deep as the key
      const inner = indent + (t.length - rest.length)
      if (rest.startsWith('- ') || keyOf(rest) !== undefined) {
        lines[at] = { indent: inner, text: ' '.repeat(inner) + rest }
        out.push(node(inner))
        continue
      }
      at++
      out.push(inline(rest, inner))
    }
    return out
  }

  function mapping(indent: number): Record<string, unknown> {
    const out: Record<string, unknown> = {}
    for (let line = next(indent); line && line.indent === indent; line = next(indent)) {
      const t = line.text.trim()
      const key = keyOf(t)
      if (key === undefined) throw new Error(`not a key: ${t}`)
      at++
      const rest = stripComment(t.slice(key.end).trimStart())
      if (rest === '') {
        // a sequence may sit as deep as its key
        const below = next(indent)
        out[key.name] = below && below.indent === indent && /^-( |$)/.test(below.text.trim()) ? sequence(indent) : node(indent + 1)
      } else if (/^[|>]/.test(rest)) {
        out[key.name] = blockScalar(rest, indent)
      } else {
        out[key.name] = inline(rest, indent)
      }
    }
    return out
  }

  // | keeps the lines, > folds them into one; - drops the last newline, + keeps every one
  function blockScalar(head: string, indent: number): string {
    const m = /^([|>])([-+]?)\d*([-+]?)\s*$/.exec(head)
    if (!m) throw new Error(`a block scalar's head: ${head}`)
    const chomp = m[2] || m[3]
    const body: string[] = []
    let deep: number | undefined
    while (at < lines.length) {
      const line = lines[at]!
      if (line.text.trim() === '') {
        body.push('')
        at++
        continue
      }
      if (line.indent <= indent) break
      deep ??= line.indent
      if (line.indent < deep) break
      body.push(line.text.slice(deep))
      at++
    }
    let text = m[1] === '|' ? body.join('\n') : body.join('\n').replace(/([^\n])\n(?!\n)/g, '$1 ')
    const trailing = /\n*$/.exec(text)![0]
    text = text.slice(0, text.length - trailing.length)
    if (chomp === '+') return text + '\n' + trailing
    return chomp === '-' ? text : text + '\n'
  }

  // A value on the line of its key or dash: a flow collection or a scalar, which a quoted or a flow one may
  // carry on over the lines deeper than `indent`
  function inline(text: string, indent: number): unknown {
    let src = text
    const open = () => {
      if (src.startsWith('"')) return !/^"(?:[^"\\]|\\.)*"/.test(src)
      if (src.startsWith("'")) return !/^'(?:[^']|'')*'/.test(src)
      if (src.startsWith('[') || src.startsWith('{')) return !balanced(src)
      return false
    }
    while (open()) {
      const line = lines[at]
      if (line === undefined || (line.text.trim() !== '' && line.indent <= indent)) throw new Error('unclosed')
      src += '\n' + line.text.trim()
      at++
    }
    const flow = new Flow(src)
    const value = flow.value()
    flow.end()
    return value
  }

  const value = node(0)
  if (next(0) !== undefined) throw new Error(`left over: ${lines[at]!.text}`)
  return value
}

// The key a line of a block mapping starts with, and where its value starts; undefined for a line that is not
function keyOf(t: string): { name: string; end: number } | undefined {
  if (t.startsWith('"') || t.startsWith("'")) {
    const m = t.startsWith('"') ? /^"((?:[^"\\]|\\.)*)"\s*:(?=\s|$)/.exec(t) : /^'((?:[^']|'')*)'\s*:(?=\s|$)/.exec(t)
    if (!m) return undefined
    return { name: t.startsWith('"') ? unescape(m[1]!) : m[1]!.replace(/''/g, "'"), end: m[0].length }
  }
  if (/^[[{#&*!|>%@`,]/.test(t)) return undefined
  const m = /^([^#:]*?|[^#]*?[^\s#:]):(?=\s|$)/.exec(t)
  if (!m || m[1]!.trim() === '') return undefined
  return { name: m[1]!.trim(), end: m[0].length }
}

// A line's text without its comment: a # after a space, outside quotes
function stripComment(t: string): string {
  let quote = ''
  for (let i = 0; i < t.length; i++) {
    const c = t[i]!
    if (quote) {
      if (c === '\\' && quote === '"') i++
      else if (c === quote) quote = ''
    } else if (c === '"' || c === "'") quote = c
    else if (c === '#' && (i === 0 || /\s/.test(t[i - 1]!))) return t.slice(0, i).trimEnd()
  }
  return t
}

function balanced(t: string): boolean {
  let depth = 0
  let quote = ''
  for (let i = 0; i < t.length; i++) {
    const c = t[i]!
    if (quote) {
      if (c === '\\' && quote === '"') i++
      else if (c === quote) quote = ''
    } else if (c === '"' || c === "'") quote = c
    else if (c === '[' || c === '{') depth++
    else if (c === ']' || c === '}') {
      depth--
      if (depth === 0) return true
    }
  }
  return false
}

// One value in flow style: [..], {..}, a quoted or a plain scalar
class Flow {
  private i = 0
  constructor(private readonly s: string) {}

  private space() {
    while (this.i < this.s.length && /\s/.test(this.s[this.i]!)) this.i++
    if (this.s[this.i] === '#' && (this.i === 0 || /\s/.test(this.s[this.i - 1]!))) this.i = this.s.length
  }

  end() {
    this.space()
    if (this.i < this.s.length) throw new Error(`left over: ${this.s.slice(this.i)}`)
  }

  value(inFlow = false): unknown {
    this.space()
    const c = this.s[this.i]
    if (c === '[') return this.list()
    if (c === '{') return this.map()
    if (c === '"' || c === "'") return this.quoted()
    if (c === '&' || c === '*' || c === '!') throw new Error('anchors and tags are not read')
    return resolve(this.plain(inFlow))
  }

  private list(): unknown[] {
    const out: unknown[] = []
    this.i++
    for (;;) {
      this.space()
      if (this.s[this.i] === ']') {
        this.i++
        return out
      }
      out.push(this.value(true))
      this.space()
      if (this.s[this.i] === ',') this.i++
      else if (this.s[this.i] !== ']') throw new Error('a list unclosed')
    }
  }

  private map(): Record<string, unknown> {
    const out: Record<string, unknown> = {}
    this.i++
    for (;;) {
      this.space()
      if (this.s[this.i] === '}') {
        this.i++
        return out
      }
      const key = this.value(true)
      this.space()
      if (this.s[this.i] !== ':') throw new Error('a key with no value')
      this.i++
      out[String(key)] = this.value(true)
      this.space()
      if (this.s[this.i] === ',') this.i++
      else if (this.s[this.i] !== '}') throw new Error('a mapping unclosed')
    }
  }

  private quoted(): string {
    const q = this.s[this.i]!
    let j = this.i + 1
    let raw = ''
    for (;;) {
      if (j >= this.s.length) throw new Error('a quote unclosed')
      const c = this.s[j]!
      if (q === "'" && c === "'" && this.s[j + 1] === "'") {
        raw += "''"
        j += 2
      } else if (q === '"' && c === '\\') {
        raw += c + (this.s[j + 1] ?? '')
        j += 2
      } else if (c === q) break
      else {
        raw += c
        j++
      }
    }
    this.i = j + 1
    // a line break inside quotes folds to a space, an empty line to a newline
    const folded = raw.replace(/[ \t]*\n[ \t]*/g, '\n').replace(/([^\n])\n(?!\n)/g, '$1 ').replace(/\n(\n*)/g, '$1')
    return q === "'" ? folded.replace(/''/g, "'") : unescape(folded)
  }

  private plain(inFlow: boolean): string {
    const start = this.i
    while (this.i < this.s.length) {
      const c = this.s[this.i]!
      if (c === '#' && /\s/.test(this.s[this.i - 1] ?? ' ')) break
      if (inFlow && (c === ',' || c === ']' || c === '}')) break
      if (c === ':' && (inFlow ? /[\s,\]}]/.test(this.s[this.i + 1] ?? ' ') : /\s/.test(this.s[this.i + 1] ?? ' '))) break
      this.i++
    }
    return this.s.slice(start, this.i).replace(/\s*\n\s*/g, ' ').trim()
  }
}

// A plain scalar as YAML 1.2's core schema reads it
function resolve(t: string): unknown {
  if (/^(~|null|Null|NULL)?$/.test(t)) return null
  if (/^(true|True|TRUE)$/.test(t)) return true
  if (/^(false|False|FALSE)$/.test(t)) return false
  if (/^[-+]?\d+$/.test(t)) return Number(t)
  if (/^[-+]?(\d+\.\d*|\.\d+|\d+)([eE][-+]?\d+)?$/.test(t)) return Number(t)
  return t
}

const ESCAPES: Record<string, string> = {
  '0': '\0', a: '\x07', b: '\b', t: '\t', '\t': '\t', n: '\n', v: '\v', f: '\f', r: '\r', e: '\x1b', ' ': ' ',
  '"': '"', '/': '/', '\\': '\\', N: '\x85', _: '\xa0', L: ' ', P: ' ',
}

function unescape(s: string): string {
  return s.replace(/\\(x[0-9a-fA-F]{2}|u[0-9a-fA-F]{4}|U[0-9a-fA-F]{8}|.)/g, (_, e: string) => {
    if (e.length > 1) return String.fromCodePoint(parseInt(e.slice(1), 16))
    const out = ESCAPES[e]
    if (out === undefined) throw new Error(`an escape YAML has not: \\${e}`)
    return out
  })
}
