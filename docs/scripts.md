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
