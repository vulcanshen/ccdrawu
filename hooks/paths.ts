// Where ccdrawu keeps its files (terminu design E3), the same places the ccdrawu TUI reads and writes.
// None touches `$`, so tests can call it.

// What the environment says, a variable set but empty counted as unset: `os` is darwin, windows or else
export type Env = {
  home: string
  os: string
  config?: string
  cache?: string
  xdgConfig?: string
  xdgCache?: string
  localAppData?: string
}

// configFile: config.yaml in CCDRAWU__CONFIG, else $XDG_CONFIG_HOME/ccdrawu, else ~/.config/ccdrawu; oldConfigFile
// is the config.json of 0.1.4 and before beside it, read until ccdrawu moves it.
// cache: CCDRAWU__CACHE, else ccdrawu in the system's cache folder, as Go's os.UserCacheDir has it (%LOCALAPPDATA%,
// ~/Library/Caches, an absolute $XDG_CACHE_HOME, else ~/.cache): the catalog and the statusline's values, which this
// mod writes anew, so a cache
export function locations(env: Env): { configFile: string; oldConfigFile: string; cache: string } {
  const config = env.config || (env.xdgConfig ? `${env.xdgConfig}/ccdrawu` : `${env.home}/.config/ccdrawu`)
  return { configFile: `${config}/config.yaml`, oldConfigFile: `${config}/config.json`, cache: cacheDir(env) }
}

function cacheDir(env: Env): string {
  if (env.cache) return env.cache
  if (env.os === 'windows') return `${env.localAppData || `${env.home}/AppData/Local`}/ccdrawu`
  if (env.os === 'darwin') return `${env.home}/Library/Caches/ccdrawu`
  if (env.xdgCache?.startsWith('/')) return `${env.xdgCache}/ccdrawu`
  return `${env.home}/.cache/ccdrawu`
}
