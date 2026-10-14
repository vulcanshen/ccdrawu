export type Limit = { kind: string; percentUsed: number; resetsAt?: string }

export type Usage = {
  startedAt: number
  contextPercent?: number
  limits: Limit[]
  costUsd?: number
}

export type ModelInfo = { name: string; effort?: string }

export type GitInfo = { branch: string; added: number; removed: number }

// The main conversation's last request: tokens read from the prompt cache, and its whole input
export type CacheInfo = { read: number; total: number }

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
