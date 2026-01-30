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

### `world_sql/`

SQL migrations that modify server-side data (NPC stats, quests, loot tables, item stats, etc.). These are applied directly to the TrinityCore world database.

```
world_sql/
├── 20250129_130000_add_custom_npc.sql
└── 20250129_130000_add_custom_npc.rollback.sql
```

### `luaxml/`

Interface file overrides. Files here follow the same structure as the WoW `Interface/` directory and override the defaults.

```
luaxml/
└── Interface/
    └── GlueXML/
        └── AccountLogin.lua    # Custom login screen logic
```

## The `shared/` Directory

The `shared/` directory contains resources that are common across all mods. Thorium creates this folder for you. It probably should not be touched.

### `shared/dbc/`

Extracted DBC files from the WoW client. These serve as the baseline data that gets imported into the DBC database.

```bash
thorium extract --dbc    # Populates shared/dbc/
```

After extraction, the DBC files are imported into MySQL tables where your mod migrations can modify them.

### `shared/luaxml/`

Extracted interface files from the WoW client. Use these as reference when creating overrides in your mod's `luaxml/` folder.

```bash
thorium extract --luaxml    # Populates shared/luaxml/
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
# 1. Initialize workspace
thorium init

# 2. Extract client data
thorium extract

# 3. Create a mod
thorium create-mod custom-sword

# 4. Create migrations
thorium create-migration --mod custom-sword --db dbc "add sword display"
thorium create-migration --mod custom-sword --db world "add sword item"

# 5. Edit the generated SQL files

# 6. Build everything
thorium build

# 7. Test in-game

# 8. Create distribution package
thorium dist --mod custom-sword
```
