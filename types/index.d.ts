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
// alias: one more name for /ccdrawu
export type Config = { dock?: string[]; inline?: string[]; alias?: string }

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
    }
  }
}
