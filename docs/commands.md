# Commands

Quick reference for all Thorium CLI commands. See other docs for detailed coverage of [DBC](dbc.md), [LuaXML](luaxml.md), [SQL Migrations](sql-migrations.md), [Distribution](distribution.md), and [Client Patcher](client-patcher.md).

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
├── dbc_sql/       # DBC database migrations
├── world_sql/     # World database migrations
└── luaxml/        # Interface file overrides
```

## Build Pipeline

### `build`

Full build pipeline: applies pending migrations, exports DBCs, and packages MPQs.

```bash
thorium build                    # Build all mods
thorium build --mod my-mod       # Build specific mod only
```

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

### `package`

Package files into MPQ archives without running migrations or exports.

```bash
thorium package                  # Package all (DBC + LuaXML)
thorium package --client         # Package client MPQs only
```

## Extraction

### `extract`

Extract files from the WoW client's MPQ archives.

```bash
thorium extract                  # Extract both DBC and LuaXML
thorium extract --dbc            # Extract DBC files only
thorium extract --luaxml         # Extract LuaXML files only
```

Extracted files go to `mods/shared/dbc/` and `mods/shared/luaxml/`.

## Distribution

### `dist`

Create a distributable zip containing client MPQs and server SQL. See [distribution.md](distribution.md) for details.

```bash
thorium dist                     # Package all mods
thorium dist --mod my-mod        # Package specific mod
thorium dist --output release.zip
```

## Client Patching

### `patch`

Apply binary patches to the WoW client executable. See [client-patcher.md](client-patcher.md) for details.

```bash
thorium patch                    # Apply all patches
thorium patch --list             # List available patches
thorium patch --dry-run          # Preview what would be applied
thorium patch --restore          # Restore from backup
```

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
