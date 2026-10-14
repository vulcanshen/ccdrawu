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

// permission: the permission mode (auto, plan, ...); editor: settings' editorMode (vim / normal);
// vim: vim's live state (INSERT, NORMAL, ...), passed on by the status line bridge
export type Modes = { permission?: string; editor?: string; vim?: string }

declare module 'claude-code' {
  interface PluginState {
    statusdrawer: {
      usage: Usage | null
      model: ModelInfo | null
      git: GitInfo | null
      path: string
      turns: number
      cache: CacheInfo | null
      compacts: number
      agents: number
      version: VersionInfo | null
      modes: Modes
      now: number
    }
  }
}
