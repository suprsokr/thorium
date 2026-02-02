# Commands

## Workspace Setup

### `init`

Initialize a new Thorium workspace. See [init.md](init.md) for details.

```bash
thorium init                    # Initialize workspace
thorium init --force            # Overwrite existing files
thorium init db                 # Set up DBC workflow (see [dbc.md](dbc.md))
```

### `create-mod <name>`

Create a new mod with the standard folder structure. See [mods.md](mods.md) for details.

```bash
thorium create-mod my-custom-items
```

## Modding

### `status`

Show the current state of migrations and mods. See [status.md](status.md) for details.

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

### `create-addon`

Create a new WoW addon in a mod's luaxml folder. See [luaxml.md](luaxml.md) for details.

```bash
thorium create-addon --mod my-mod MyAddon
```

### `build`

Full build pipeline: applies migrations, binary edits, server patches, copies assets, exports DBCs, deploys scripts, and packages MPQs. See [build.md](build.md) for complete details.

```bash
thorium build                           # Build all mods (all 8 steps)
thorium build --mod my-mod              # Build specific mod only
# Many more options!
```

### `rollback`

Undo applied migrations. See [sql-migrations.md](sql-migrations.md) for details.

```bash
thorium rollback                 # Rollback last migration
thorium rollback --mod my-mod    # Rollback for specific mod
thorium rollback --all           # Rollback all migrations
```

## Distribution

### `dist`

Create a player-ready distribution package containing client MPQs and optionally patched wow.exe. For mod source distribution, host on GitHub and use `thorium get`, see [Thorium Get - Installing Mods from GitHub](get.md). See [distribution.md](distribution.md) for details.

```bash
thorium dist                     # Package all mods
thorium dist --mod my-mod        # Package specific mod
thorium dist --no-exe            # Skip including wow.exe
thorium dist --output ./releases/v1.0.0.zip
```

## Mod Sharing and Discovery

### `get <url>`

Install a mod from a GitHub repository. See [get.md](get.md) for details.

```bash
thorium get https://github.com/user/repo
thorium get https://github.com/user/repo --name custom-name
```

### `search [query]`

Search the mod registry for available mods. See [get.md](get.md) for details.

```bash
thorium search                    # List all available mods
thorium search custom-items       # Search for mods
```

## Extraction & Import

**Note:** These commands are primarily for internal/advanced usage. Most users should use `thorium init db` which automatically handles DBC extraction (`thorium extract --dbc`) and import. See [dbc.md](dbc.md) and [luaxml.md](luaxml.md) for details.

### `import`

Import DBC files into the database.

```bash
thorium import dbc                       # Import from default path
thorium import dbc --source /path/to/dbc
```

### `export`

Export DBC files from the database. See [dbc.md](dbc.md) for details.

```bash
thorium export                   # Export all modified DBCs
```

## Help

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
