// Where ccdrawu keeps its files (terminu design E3), the same places the ccdrawu TUI reads and writes.
// None touches `$`, so tests can call it.

export type Env = { home: string; data?: string; config?: string; xdgConfig?: string }

// data: CCDRAWU__DATA, else ~/.ccdrawu (the status line bridge, the catalog, vim's state per session);
// configFile: config.json in CCDRAWU__CONFIG, else $XDG_CONFIG_HOME/ccdrawu, else ~/.config/ccdrawu.
// A variable set but empty counts as unset
export function locations(env: Env): { data: string; configFile: string } {
  const data = env.data || `${env.home}/.ccdrawu`
  const config = env.config || (env.xdgConfig ? `${env.xdgConfig}/ccdrawu` : `${env.home}/.config/ccdrawu`)
  return { data, configFile: `${config}/config.json` }
}
