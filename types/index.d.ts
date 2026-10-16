export type Limit = { kind: string; percentUsed: number; resetsAt?: string }

// contextTokens over contextWindow is contextPercent, in tokens
export type Usage = {
  startedAt: number
  contextPercent?: number
  contextTokens?: number
  contextWindow?: number
  limits: Limit[]
  costUsd?: number
}

export type ModelInfo = { name: string; effort?: string }

// The branch with its line counts; then, from `git status --porcelain=v2`, the commit, how far it is
// from its upstream (none without one) and the files staged, unstaged and untracked; repo is origin's
// owner/name, worktree the linked worktree's folder when the session is in one
export type GitInfo = {
  branch: string
  added: number
  removed: number
  sha?: string
  ahead?: number
  behind?: number
  staged?: number
  unstaged?: number
  untracked?: number
  repo?: string
  worktree?: string
}

// The main conversation's last request: tokens read from the prompt cache, and its whole input
export type CacheInfo = { read: number; total: number }

// The main conversation's tokens since the drawer loaded, and its last reply's output tokens a second
export type TokenInfo = { input: number; output: number; cacheRead: number; cacheWrite: number; speed?: number }

// This session: its id, the account signed in, the output style
export type SessionInfo = { id?: string; account?: string; outputStyle?: string }

// status.claude.com's overall status: indicator none, minor, major or critical
export type ServiceStatus = { indicator: string; description: string }

// Skills run in this session, and the last one's name
export type SkillsInfo = { count: number; last?: string }

export type VersionInfo = { current: string; latest?: string }

// permission: the permission mode (auto, plan, ...)
export type Modes = { permission?: string }

// The background work in flight as of Claude's last reply, counted by type (shell, monitor, workflow, ...);
// subagents are counted live from $.agent.list() instead
export type Background = Record<string, number>

// The items on in each placement, as the TUI wrote them; one left out takes the catalog's defaults.
// alias: one more name for /ccdrawu; pet: which animal; drawer: whether it shows beside the transcript
// and above the prompt; statusline: whether `ccdrawu statusline` is Claude Code's status line (the rest
// of its block is the TUI's and the command's)
export type Config = {
  dock?: string[]
  inline?: string[]
  alias?: string
  pet?: { species?: string }
  bar?: string
  drawer?: { sidebarOff: boolean; inlineOff: boolean }
  statusline?: { on: boolean }
}

// The pet (hooks/pets.ts): what it is doing, where along its walk, and what it says, a session's
export type PetMode = 'walk' | 'sit' | 'lie' | 'sleep' | 'jump' | 'play'
// A ball thrown in (/ccdrawu ball): it rolls from start to rest, from the time thrown on; the pet
// stands at goal to play with it, and says line when it has it
export type PetBall = { from: number; start: number; rest: number; goal: number; line: string }
export type PetState = {
  coat: number // which coat, picked at random once a session; -1 until then
  busy: boolean // Claude is working
  active: number // when Claude or the user last did something
  mode: PetMode
  x: number // 0 to TRACK, how far along its walk
  target: number // where along it it is going
  run: boolean // running there, two steps a tick
  stride: number // its steps, for its legs to take turns
  restUntil: number // sitting a while where it got to, before it goes on
  lift: number // how many pixels off the ground it is, in a jump
  faceLeft: boolean
  snore: number // how far along its snore, while asleep
  heart: boolean
  line?: string
  lineUntil: number
  heartUntil: number
  jumpFrom: number // in the air from then until jumpUntil
  jumpUntil: number
  ball?: PetBall
  ballPos: number // where the ball is along the walk, and how far off the ground, a tick's
  ballLift: number
  playUntil: number // pawing at the ball until then
}

declare module 'claude-code' {
  interface PluginState {
    ccdrawu: {
      usage: Usage | null
      model: ModelInfo | null
      git: GitInfo | null
      path: string
      turns: number
      cache: CacheInfo | null
      tokens: TokenInfo
      session: SessionInfo
      status: ServiceStatus | null
      skills: SkillsInfo
      compacts: number
      agents: number
      background: Background
      version: VersionInfo | null
      drawer: string | null
      modes: Modes
      now: number
      config: Config
      pet: PetState
      manual: 'on' | 'off' | null
    }
  }
}
