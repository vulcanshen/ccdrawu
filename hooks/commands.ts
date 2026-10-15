import type { CommandSpec } from 'claude-code'

import { commandNames } from './format'

// Registers /ccdrawu and the alias the config names, each once, through $.command.register. The engine
// cannot take a command back, so a name stays for the rest of the session: an alias changed or cleared
// still works until then
export async function registerCommands(
  register: (spec: CommandSpec) => Promise<unknown>,
  registered: Set<string>,
  alias: unknown,
) {
  for (const name of commandNames(alias)) {
    if (registered.has(name)) continue
    const description =
      name === 'ccdrawu'
        ? 'Toggle the status drawer; pat or ball to play with the pet'
        : 'Toggle the status drawer; pat or ball to play with the pet (alias of /ccdrawu)'
    await register({ name, description, immediate: true })
    registered.add(name)
  }
}
