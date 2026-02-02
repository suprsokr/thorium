# Initialization

The `thorium init` command creates a new Thorium workspace with everything you need to start building mods.

## What It Does

When you run `thorium init`, Thorium creates:

1. **Directory structure** for organizing your mods and shared resources
2. **config.json** with environment variable placeholders
3. **.gitignore** with sensible defaults for version control
4. **README.md** with workspace documentation

That's it! No complex setup required. You can start creating mods immediately.

## Usage

```bash
# Initialize a new workspace
thorium init

# Overwrite existing files (if reinitializing)
thorium init --force
```

### Flags

- `--force` - Overwrite existing config.json, .gitignore, and README.md

## Directory Structure

After `thorium init`, you'll have:

```
.
├── config.json              # Workspace configuration
├── .gitignore               # Git ignore rules
├── README.md                # Workspace documentation
├── shared/
│   ├── dbc/
│   │   ├── dbc_source/      # Original DBCs (if using DBC workflow)
│   │   └── dbc_out/         # Modified DBCs (if using DBC workflow)
│   ├── luaxml/
│   │   └── luaxml_source/   # Original LuaXML (if extracted)
│   └── migrations_applied/  # Tracks which SQL migrations have been applied
└── mods/                    # Your mods go here
```

## Next Steps

### 1. Configure Your Workspace

Edit `config.json` to set your paths:

```bash
vim config.json
```

Or use [Peacebloom](https://github.com/suprsokr/peacebloom) to manage your environment with Docker.

### 2. Create Your First Mod

```bash
thorium create-mod my-mod
```

### 3. Start Building

Depending on what you want to do:

**For simple mods (addons, binary edits, assets):**
```bash
thorium create-addon --mod my-mod MyAddon
# Edit your addon files
thorium build
```

**For mods with DBC changes:**

See [DBC Workflow](dbc.md#just-in-time-setup) for setting up DBC databases when you need them.

**For mods with World database changes:**
```bash
thorium create-migration --mod my-mod --db world "add custom npc"
# Edit the SQL file
thorium build world_sql
```

## Configuration

The generated `config.json` uses environment variables with sensible defaults:

```json
{
  "wotlk": {
    "path": "${WOTLK_PATH:-/wotlk}",
    "locale": "enUS"
  },
  "databases": {
    "dbc": { ... },
    "dbc_source": { ... },
    "world": { ... }
  },
  "server": {
    "dbc_path": "${TC_SERVER_PATH:-/home/peacebloom/server}/bin/dbc"
  },
  "trinitycore": {
    "source_path": "${TC_SOURCE_PATH:-/home/peacebloom/TrinityCore}",
    "scripts_path": "${TC_SOURCE_PATH:-/home/peacebloom/TrinityCore}/src/server/scripts/Custom"
  },
  "output": {
    "dbc_mpq": "patch-T.MPQ",
    "luaxml_mpq": "patch-{locale}-T.MPQ"
  }
}
```

Environment variables like `${WOTLK_PATH}` are automatically expanded. See the [full configuration reference](config.md) for details.

## Reinitializing

If you need to recreate the config or other files:

```bash
thorium init --force
```

This will overwrite `config.json`, `.gitignore`, and `README.md`, but won't delete any data in `shared/` or `mods/`.

## Database Initialization

**You don't need to set up databases during init!** Thorium will guide you through database setup when you create your first migration that needs it.

However, if you want to initialize databases manually:

```bash
thorium init db
```

This creates the `dbc`, `dbc_source`, and `world` databases in MySQL. See [DBC Workflow](dbc.md) for when and why you need these.
