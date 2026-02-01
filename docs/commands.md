# Commands

Quick reference for all Thorium CLI commands. See other docs for detailed coverage of [Mods](mods.md), [DBC](dbc.md), [LuaXML](luaxml.md), [SQL Migrations](sql-migrations.md), [Scripts](scripts.md), [Binary Edits](binary-edits.md), [Server Patches](server-patches.md), [Assets](assets.md), [Distribution](distribution.md), and [Custom Packets](custom-packets.md).

## Workspace Setup

### `init`

Initialize a new Thorium workspace in the current directory. Creates the standard directory structure and a `config.json` file.

```bash
thorium init
```

### `create-mod <name>`

Create a new mod with the standard folder structure.

```bash
thorium create-mod my-custom-items
```

Creates:
```
mods/mods/my-custom-items/
├── dbc_sql/         # DBC database migrations
├── world_sql/       # World database migrations
├── scripts/         # TrinityCore C++ scripts
├── server-patches/  # TrinityCore source patches (.patch files)
├── binary-edits/    # Client binary patches (.json files)
├── assets/          # Files to copy to client directory
└── luaxml/          # Interface file overrides (Lua/XML)
```

## Build Pipeline

### `build`

Full build pipeline: applies migrations, binary edits, server patches, copies assets, exports DBCs, deploys scripts, and packages MPQs.

```bash
thorium build                    # Build all mods
thorium build --mod my-mod       # Build specific mod only
thorium build --force            # Reapply patches/edits even if tracked as applied
```

**Flags:**
- `--skip-migrations` - Skip SQL migrations
- `--skip-export` - Skip DBC export
- `--skip-package` - Skip MPQ packaging
- `--skip-server` - Skip copying DBCs to server
- `--force` - Reapply binary-edits and server-patches even if already applied

### `apply`

Apply pending SQL migrations without running the full build.

```bash
thorium apply                    # Apply all pending migrations
thorium apply --mod my-mod       # Apply for specific mod
thorium apply --db dbc           # Apply only DBC migrations
thorium apply --db world         # Apply only World migrations
```

### `rollback`

Undo applied migrations.

```bash
thorium rollback                 # Rollback last migration
thorium rollback --mod my-mod    # Rollback for specific mod
thorium rollback --all           # Rollback all migrations
```

### `export`

Export DBC files from the database without running the full build.

```bash
thorium export                   # Export all modified DBCs
```

## Extraction & Import

### `extract`

Extract files from the WoW client's MPQ archives, or copy baseline files to a mod.

```bash
# Extract from WoW client to shared baseline
thorium extract                  # Extract both DBC and LuaXML
thorium extract --dbc            # Extract DBC files only
thorium extract --luaxml         # Extract LuaXML files only

# Extract only specific folders (recommended for LuaXML)
thorium extract --luaxml --filter Interface/FrameXML
thorium extract --luaxml --filter Interface/AddOns/Blizzard_CombatText

# Copy files from baseline to a mod for editing
thorium extract --mod my-mod --dest Interface/FrameXML/ChatFrame.lua
thorium extract --mod my-mod --dest Interface/AddOns/Blizzard_AchievementUI
```

**Flags:**
- `--dbc` - Extract DBC files only
- `--luaxml` - Extract LuaXML files only
- `--filter <path>` - Only extract files matching this path prefix (useful for large extractions)
- `--mod <name>` - Copy files to a mod's `luaxml/` directory instead of extracting
- `--dest <path>` - Path of file/folder to copy (required with `--mod`)

**Output locations:**
- `shared/dbc/dbc_source/` - Baseline DBC files
- `shared/luaxml/luaxml_source/` - Baseline LuaXML files
- `mods/<mod>/luaxml/` - Mod-specific LuaXML overrides

### `import`

Import data files into the MySQL database. Currently supports DBC files.

```bash
thorium import dbc                           # Import from default path
thorium import dbc --source /path/to/dbc     # Import from custom path
thorium import dbc --database custom_dbc     # Import to specific database
thorium import dbc --skip-existing           # Skip tables that already exist
```

The import command:
1. Loads each DBC file and creates a corresponding MySQL table
2. Copies source DBCs to `mods/shared/dbc/dbc_source/` as baseline for comparison
3. Stores checksums to track which tables have been modified

This is typically run once during initial setup. See [dbc.md](dbc.md) for the full DBC workflow.

## Distribution

### `dist`

Create a distributable zip containing client MPQs and server SQL. See [distribution.md](distribution.md) for details.

```bash
thorium dist                     # Package all mods
thorium dist --mod my-mod        # Package specific mod
thorium dist --client-only       # Client files only (for players)
thorium dist --include-exe       # Include patched wow.exe
thorium dist --output release.zip
```

Flags:
- `--mod <name>` - Package specific mod only
- `--client-only` - Omit server SQL (for player distribution)
- `--include-exe` - Include patched wow.exe
- `--output <path>` - Custom output path

## Utilities

### `status`

Show the current state of migrations and mods.

```bash
thorium status                   # Show all status
thorium status --mod my-mod      # Show status for specific mod
```

### `create-migration`

Create a new timestamped migration file pair. See [sql-migrations.md](sql-migrations.md) for details.

```bash
thorium create-migration --mod my-mod --db world "add custom npc"
thorium create-migration --mod my-mod --db dbc "add custom spell"
```

### `create-script`

Create a new TrinityCore script in a mod. See [scripts.md](scripts.md) for details.

```bash
thorium create-script --mod my-mod --type spell fire_blast
thorium create-script --mod my-mod --type creature custom_vendor
thorium create-script --mod my-mod --type server my_hooks
thorium create-script --mod my-mod --type packet my_protocol
thorium create-script --mod my-mod --type aura regeneration_aura
```

Available script types:
- `spell` - SpellScript for spell behavior
- `aura` - AuraScript for aura/buff effects
- `creature` - CreatureScript for NPC AI and behavior
- `server` - ServerScript for server-wide hooks
- `packet` - ServerScript for custom packet handlers

### `create-addon`

Create a new WoW addon in a mod's luaxml folder.

```bash
thorium create-addon --mod my-mod MyAddon
```

Creates:
```
mods/my-mod/luaxml/Interface/AddOns/MyAddon/
├── MyAddon.toc    # Addon metadata
└── main.lua       # Main addon code with slash commands
```

The addon is automatically packaged into the LuaXML MPQ when you run `thorium build`.

### `version`

Show Thorium version information.

```bash
thorium version
```

### `help`

Show help message with all commands and examples.

```bash
thorium help
```
