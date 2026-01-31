# Scripts

Thorium supports TrinityCore C++ scripts as part of your mods. Scripts enable custom server-side behavior that goes beyond what SQL and DBC modifications can achieve.

## Overview

Scripts are C++ files that compile into the TrinityCore server binary. They use TrinityCore's scripting API to hook into game events, modify spell behavior, create custom AI, and more.

## Directory Structure

```
mods/mods/my-mod/
└── scripts/
    ├── spell_fire_blast.cpp
    ├── aura_regeneration.cpp
    ├── npc_custom_vendor.cpp
    └── server_my_hooks.cpp
```

## Creating Scripts

Use the `create-script` command to generate script templates:

```bash
thorium create-script --mod my-mod --type <type> <name>
```

Available script types:

| Type | Purpose | Example |
|------|---------|---------|
| `spell` | Custom spell behavior | Damage calculations, proc effects |
| `aura` | Custom aura/buff effects | Periodic effects, stat modifications |
| `creature` | NPC AI and behavior | Boss mechanics, custom vendors |
| `server` | Server-wide hooks | Login events, initialization |
| `packet` | Custom packet handlers | Client-server protocols |

## Script Types

### SpellScript

Hooks into spell casting, targeting, and effects.

```bash
thorium create-script --mod my-mod --type spell fire_blast
```

Common use cases:
- Custom damage or healing calculations
- Conditional spell effects based on target or caster state
- Multi-hit or chain effects
- Spell proc triggers

**Registration**: Scripts must specify which spell ID they apply to in the `Register()` function.

### AuraScript

Handles aura/buff application, periodic ticks, and removal.

```bash
thorium create-script --mod my-mod --type aura regeneration
```

Common use cases:
- Custom periodic effects (damage, healing, resource generation)
- Stat modifications based on conditions
- Aura stacking behavior
- On-apply/on-remove effects

**Registration**: Scripts must specify which spell ID's aura they handle.

### CreatureScript

Controls NPC behavior, AI, and interactions.

```bash
thorium create-script --mod my-mod --type creature custom_vendor
```

Common use cases:
- Boss encounter mechanics
- Custom creature AI patterns
- Gossip menu interactions
- Special vendor or trainer logic

**Registration**: Scripts are registered by creature script name (not entry ID). The name is then assigned to creatures via `creature_template.ScriptName` in SQL.

### ServerScript

Provides server-wide hooks for initialization, configuration, and global events.

```bash
thorium create-script --mod my-mod --type server my_hooks
```

Common use cases:
- Server initialization logic
- Configuration loading
- Global event handling
- World state management

**Registration**: ServerScripts run globally and don't target specific entities.

### Packet Handler (ServerScript)

Handles custom client-server packets for advanced communication.

```bash
thorium create-script --mod my-mod --type packet my_protocol
```

Common use cases:
- Custom UI data synchronization
- Real-time features (chat systems, mini-games)
- Advanced client-server protocols
- Custom authentication or session handling

**Registration**: Packet handlers check incoming packet opcodes and process matching packets.

**Note**: Custom packets require client-side addon support. See [client-patcher.md](client-patcher.md) for information about the custom packets extension.

## Deployment

Scripts are automatically deployed during `thorium build`:

1. **Collection**: All `.cpp` files from `mods/*/scripts/` are collected
2. **Deployment**: Files are copied to `TrinityCore/src/server/scripts/Custom/`
3. **Registration**: A loader file is auto-generated to register all scripts
4. **CMake**: `CMakeLists.txt` is updated to include all script files

After deployment, you must rebuild TrinityCore:

```bash
cd /path/to/TrinityCore/build
make -j$(nproc)
```

## Script Registration

Each script must have an `AddSC_*()` function that Thorium will automatically call during server startup:

```cpp
void AddSC_spell_fire_blast()
{
    RegisterSpellScript(spell_fire_blast);
}
```

The function name must match the pattern:
- SpellScript: `AddSC_spell_<name>()`
- AuraScript: `AddSC_aura_<name>()`
- CreatureScript: `AddSC_npc_<name>()` or `AddSC_creature_<name>()`
- ServerScript: `AddSC_<name>_server()` or `AddSC_<name>_packet()`

## Example Workflow

### Creating a Custom Spell

```bash
# 1. Create DBC entry for spell display info
thorium create-migration --mod my-mod --db dbc "add fire blast spell"

# Edit: mods/my-mod/dbc_sql/20250130_*_add_fire_blast_spell.sql
# INSERT INTO Spell (...) VALUES (90001, 'Fire Blast', ...);

# 2. Create spell script for custom behavior
thorium create-script --mod my-mod --type spell fire_blast

# Edit: mods/my-mod/scripts/spell_fire_blast.cpp
# Implement custom damage calculation, effects, etc.

# 3. Build and deploy
thorium build

# 4. Rebuild TrinityCore
cd /path/to/TrinityCore/build && make -j$(nproc)

# 5. Restart server and test
```

### Creating a Custom NPC

```bash
# 1. Create world database entry
thorium create-migration --mod my-mod --db world "add custom boss"

# Edit: mods/my-mod/world_sql/20250130_*_add_custom_boss.sql
# INSERT INTO creature_template (entry, name, ScriptName, ...) 
# VALUES (90001, 'Fire Lord', 'npc_fire_lord', ...);

# 2. Create creature script
thorium create-script --mod my-mod --type creature fire_lord

# Edit: mods/my-mod/scripts/npc_fire_lord.cpp
# Implement boss mechanics in UpdateAI()

# 3. Build and deploy
thorium build

# 4. Rebuild TrinityCore
cd /path/to/TrinityCore/build && make -j$(nproc)

# 5. Restart server and test
```

## Best Practices

### Script Organization

- **One script per file** - Keep files focused and maintainable
- **Descriptive names** - Use clear naming: `spell_fire_blast.cpp`, `npc_custom_vendor.cpp`
- **Group related scripts** - Keep scripts for the same feature in the same mod

### Script Development

- **Start simple** - Begin with basic functionality, then add complexity
- **Use logging** - Add `TC_LOG_DEBUG` statements for debugging
- **Test incrementally** - Test each change before moving to the next
- **Comment your code** - Explain non-obvious logic and calculations

### Performance

- **Avoid heavy operations in UpdateAI()** - This runs every tick (~100ms)
- **Cache calculations** - Don't recalculate constants every tick
- **Use appropriate hooks** - Only hook events you actually need
- **Clean up resources** - Release any allocated memory or handles

### Compatibility

- **Use TrinityCore APIs** - Don't access private internals
- **Check for null** - Always validate pointers before use
- **Handle edge cases** - Account for players logging out, NPCs despawning, etc.
- **Avoid global state** - Use instance/player-specific data instead

## Troubleshooting

### "TrinityCore scripts path not configured"

Set `trinitycore.scripts_path` in your `config.json`:

```json
{
  "trinitycore": {
    "source_path": "/path/to/TrinityCore",
    "scripts_path": "/path/to/TrinityCore/src/server/scripts/Custom"
  }
}
```

### "No AddSC function found"

Ensure your script has the correctly named registration function:

```cpp
void AddSC_spell_my_spell()  // Must match this pattern
{
    RegisterSpellScript(spell_my_spell);
}
```

### Scripts not loading in-game

1. Verify scripts were deployed: Check `TrinityCore/src/server/scripts/Custom/`
2. Check loader was generated: Look for `custom_scripts_loader.cpp`
3. Rebuild completely: `make clean && make -j$(nproc)`
4. Check server logs: Look for script registration messages on startup

### Compilation errors

- **Missing includes**: Add required headers (`#include "Player.h"`, etc.)
- **API changes**: Verify you're using the correct TrinityCore API for your version
- **Syntax errors**: Check for typos, missing semicolons, mismatched braces

## Distribution

When distributing mods with scripts, include the C++ source files. Recipients must:

1. Copy scripts to their TrinityCore `Custom/` directory
2. Rebuild TrinityCore
3. Restart the server

See [distribution.md](distribution.md) for details.

## Further Reading

- [TrinityCore Script API Documentation](https://www.azerothcore.org/pages/wiki/script-api.html)
- [sql-migrations.md](sql-migrations.md) - SQL migrations for database changes
- [mods.md](mods.md) - Mod structure and organization
- [commands.md](commands.md) - CLI command reference
