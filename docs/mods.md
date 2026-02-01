# Mods

Thorium organizes your customizations into mods. Each mod is a self-contained set of changes that can be developed, tested, and distributed independently.

## Workspace Initialization

Before creating mods, initialize a Thorium workspace:

```bash
thorium init
```

This creates:

```
mods/
├── config.json         # Workspace configuration
├── shared/             # Shared resources (extracted client data)
│   ├── dbc/            # Extracted DBC files
│   ├── luaxml/         # Extracted interface files
│   └── migrations_applied/  # Tracking for applied migrations
└── mods/               # Your mods go here
    └── .gitkeep
```

## Creating a Mod

```bash
thorium create-mod my-awesome-mod
```

This creates:

```
mods/mods/my-awesome-mod/
├── dbc_sql/            # SQL migrations for DBC database
├── world_sql/          # SQL migrations for World database
├── scripts/            # TrinityCore C++ scripts
├── server-patches/     # TrinityCore source patches (.patch files)
├── binary-edits/       # Client binary patches (.json files)
├── assets/             # Files to copy to client directory
└── luaxml/             # Interface file overrides (Lua/XML)
```

## Mod Structure

### `dbc_sql/`

SQL migrations that modify client-side data (spells, items display info, creature models, etc.). These changes are exported to DBC files and packaged into the client MPQ.

```
dbc_sql/
├── 20250129_120000_add_custom_spell.sql
└── 20250129_120000_add_custom_spell.rollback.sql
```

See [sql-migrations.md](sql-migrations.md) for details.

### `world_sql/`

SQL migrations that modify server-side data (NPC stats, quests, loot tables, item stats, etc.). These are applied directly to the TrinityCore world database.

```
world_sql/
├── 20250129_130000_add_custom_npc.sql
└── 20250129_130000_add_custom_npc.rollback.sql
```

See [sql-migrations.md](sql-migrations.md) for details.

### `scripts/`

TrinityCore C++ scripts for custom behavior. Scripts are deployed to the TrinityCore source during `thorium build` and compiled into the server binary.

```
scripts/
├── spell_fire_blast.cpp
├── npc_custom_vendor.cpp
└── server_my_hooks.cpp
```

See [scripts.md](scripts.md) for details.

### `server-patches/`

Git patch files (`.patch`) for TrinityCore source code. Applied automatically during `thorium build` if `trinitycore.source_path` is configured. Tracked in `shared/server_patches_applied.json`.

```
server-patches/
├── my-custom-hook.patch
└── extended-api.patch
```

See [server-patches.md](server-patches.md) for details on creating and managing patches.

### `binary-edits/`

JSON files that describe binary patches for `Wow.exe`. Applied automatically during `thorium build`. Tracked in `shared/binary_edits_applied.json`.

```
binary-edits/
└── load-clientextensions.json
```

See [binary-edits.md](binary-edits.md) for the JSON format and examples.

### `assets/`

Files to copy to the WoW client directory. Requires `assets/config.json` to specify destinations.

```
assets/
├── config.json
└── ClientExtensions.dll
```

See [assets.md](assets.md) for configuration details.

### `luaxml/`

Interface file overrides and custom addons. Files here follow the same structure as the WoW `Interface/` directory.

```
luaxml/
└── Interface/
    ├── GlueXML/
    │   └── AccountLogin.lua    # Custom login screen logic
    └── AddOns/
        └── MyAddon/            # Custom addon (created with thorium create-addon)
            ├── MyAddon.toc
            └── main.lua
```

Create addons with:

```bash
thorium create-addon --mod my-mod MyAddon
```

See [luaxml.md](luaxml.md) for details.

## The `shared/` Directory

The `shared/` directory contains resources that are common across all mods. Thorium creates this folder for you. It generally should not be modified directly.

### `shared/dbc/`

Contains baseline and exported DBC files:

```
shared/dbc/
├── dbc_source/    # Baseline DBC files (copied during import)
└── dbc_out/       # Exported modified DBC files
```

```bash
# Import DBCs from TrinityCore into MySQL (also copies to dbc_source/)
thorium import dbc --source /path/to/trinitycore/bin/dbc
```

After import, the DBC data is in MySQL tables where your mod migrations can modify them. Modified tables are exported to `dbc_out/` during build.

### `shared/luaxml/`

Contains baseline interface files extracted from the WoW client:

```
shared/luaxml/
└── luaxml_source/                # Baseline files
    └── Interface/
        ├── FrameXML/             # Core UI (extracted)
        ├── GlueXML/              # Login screen UI (extracted)
        └── AddOns/               # Any extracted addons
```

```bash
# Extract all interface files (~500MB)
thorium extract --luaxml

# Or extract only what you need (recommended)
thorium extract --luaxml --filter Interface/FrameXML
```

### `shared/migrations_applied/`

Tracks which migrations have been applied. This prevents re-running migrations and enables rollback functionality.

```
shared/migrations_applied/
└── my-awesome-mod/
    ├── dbc/
    │   └── 20250129_120000_add_custom_spell.sql
    └── world/
        └── 20250129_130000_add_custom_npc.sql
```

When a migration is applied, it's copied here. Rollback removes it.

### `shared/server_patches_applied.json`

Tracks which server patches have been applied to TrinityCore source.

```json
{
  "applied": [
    {
      "name": "custom-packets/custom-packets.patch",
      "version": "1.0.0",
      "applied_at": "2025-01-31T12:00:00Z",
      "applied_by": "thorium build"
    }
  ]
}
```

### `shared/binary_edits_applied.json`

Tracks which binary edits have been applied to the WoW client.

```json
{
  "applied": [
    {
      "name": "custom-packets/load-clientextensions.json",
      "applied_at": "2025-01-31T12:00:00Z",
      "applied_by": "thorium build"
    }
  ]
}
```

### `shared/assets_applied.json`

Tracks which assets have been copied to the client, by MD5 hash.

```json
{
  "applied": [
    {
      "name": "custom-packets/ClientExtensions.dll",
      "md5": "a1b2c3d4e5f6...",
      "applied_at": "2025-01-31T12:00:00Z",
      "applied_by": "thorium build"
    }
  ]
}
```

## Multiple Mods

You can have multiple mods in a workspace:

```
mods/mods/
├── custom-items/
├── custom-quests/
├── ui-improvements/
└── balance-changes/
```

Benefits:
- **Modularity** - Enable/disable individual mods
- **Organization** - Keep related changes together
- **Distribution** - Package mods separately with `thorium dist --mod <name>`
- **Collaboration** - Different team members work on different mods

## Mod Priority and Conflicts

When multiple mods edit the same database tables, **mods are processed in alphabetical order**. Later mods can overwrite changes made by earlier ones.

For example, if both `balance-changes` and `custom-items` modify the same spell:

1. `balance-changes` runs first (alphabetically)
2. `custom-items` runs second and its changes take precedence

### Controlling Priority

To control which mods run first, prefix mod names with numbers:

```
mods/mods/
├── 01-core-framework/      # Runs first - base tables and IDs
├── 02-custom-items/        # Runs second - adds items
├── 02-custom-spells/       # Same priority as items (alphabetical within tier)
└── 99-balance-tweaks/      # Runs last - final adjustments
```

### Best Practices for Multi-Mod Workspaces

- **Use distinct ID ranges** - Assign each mod a range (e.g., mod A uses 90000-90999, mod B uses 91000-91999)
- **Avoid UPDATE on same rows** - If two mods UPDATE the same entry, the last one wins
- **Use DELETE before INSERT** - Makes migrations idempotent and avoids duplicate key errors
- **Document dependencies** - Note in your mod if it requires another mod to run first

## Example Workflow

```bash
# 1. Initialize workspace (also creates CustomPackets addon)
thorium init

# 2. Extract client data
thorium extract

# 3. Create a mod
thorium create-mod custom-sword

# 4. Create migrations
thorium create-migration --mod custom-sword --db dbc "add sword display"
thorium create-migration --mod custom-sword --db world "add sword item"

# 5. Create a script (optional)
thorium create-script --mod custom-sword --type spell sword_strike

# 6. Create an addon (optional)
thorium create-addon --mod custom-sword SwordUI

# 7. Edit the generated SQL, C++, and Lua files

# 8. Build everything (packages addons into MPQ)
thorium build

# 9. Rebuild TrinityCore (if using scripts)
cd /path/to/TrinityCore/build && make -j$(nproc)

# 10. Test in-game

# 11. Create distribution package
thorium dist --mod custom-sword
```
