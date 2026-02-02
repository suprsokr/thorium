# DBC (Database Client Files)

DBC files are client-side database files used by World of Warcraft. They contain game data like items, spells, creatures, maps, and more. Thorium provides tools to import, edit, and rebuild DBC files.

## Overview

The DBC workflow in Thorium:

1. **Extract** - Extract DBC files from WoW client MPQs to `shared/dbc/dbc_source/`
2. **Import** - Load DBC files into MySQL databases (`dbc_source` and `dbc`)
3. **Edit** - Modify data using SQL migrations in your mods (applied to `dbc`)
4. **Export** - Generate new DBC files from modified tables in `dbc`
5. **Build** - Package only modified DBCs into an MPQ for the client (and copy to server)

## Just-In-Time Setup

**You only need to set up the DBC workflow if your mod modifies DBCs.** Many mods (like those with only addons, binary edits, or assets) don't need this at all!

When you create your first DBC migration and run `thorium build`, Thorium will detect that the DBC databases aren't set up yet and guide you through the process.

## Setting Up DBC Workflow

When you're ready to work with DBCs, it's a simple one-command setup.

### Prerequisites

**Important:** Make sure `config.json` has correct settings before proceeding:

```json
{
  "wotlk": {
    "path": "${WOTLK_PATH:-/wotlk}",
    "locale": "enUS"
  },
  "databases": {
    "dbc": {
      "user": "trinity",
      "password": "trinity",
      "host": "${MYSQL_HOST:-127.0.0.1}",
      "port": "${MYSQL_PORT:-3306}",
      "name": "dbc"
    },
    "dbc_source": {
      "user": "trinity",
      "password": "trinity",
      "host": "${MYSQL_HOST:-127.0.0.1}",
      "port": "${MYSQL_PORT:-3306}",
      "name": "dbc_source"
    }
  }
}
```

**Key properties:**
- `wotlk.path` - Path to your WoW 3.3.5 client directory (for extracting DBCs)
- `databases.dbc` - Development database where mod migrations are applied and tested
- `databases.dbc_source` - Pristine baseline database, never modified by migrations
- The MySQL user must have permissions to create databases and tables

### One-Command Setup

```bash
thorium init db
```

This command does everything:
1. Creates the `dbc` and `dbc_source` databases
2. Extracts DBCs from your WoW client to `shared/dbc/dbc_source/`
3. Imports DBCs to the `dbc_source` database
4. Copies `dbc_source` to `dbc` (your development database)

That's it! Your DBC workflow is ready.

**Note:** The `world` database is created and managed by TrinityCore, not Thorium.

### Why Two Databases?

The separation allows `thorium dist` to create minimal distribution packages:

1. Apply your migrations to a temporary copy of `dbc_source`
2. Export only the modified DBCs
3. Rollback migrations to restore `dbc_source` to pristine state

This ensures distributed packages only contain the DBCs you actually modified.

## Why Use a Database?

Editing DBC files directly is tedious and error-prone. By importing them into MySQL:

- Use familiar SQL and tooling to query and modify data.
- Track changes with version-controlled migration files
- Avoid binary file conflicts in git
- LLMs are exceptional at navigating, modifying and scripting sql; not so for DBCs.

## Directory Structure

```
.
├── shared/
│   └── dbc/
│       ├── dbc_source/    # Baseline DBC files (extracted)
│       └── dbc_out/       # Exported modified DBC files
└── mods/
    └── my-mod/
        └── dbc_sql/       # SQL migrations for DBC changes
            ├── 20250129_120000_add_custom_spell.sql
            └── 20250129_120000_add_custom_spell.rollback.sql
```

## Database Tables

When DBCs are imported, each DBC file becomes a table in both the `dbc_source` and `dbc` databases. For example:

- `Item.dbc` → `item` table
- `Spell.dbc` → `spell` table  
- `AreaTable.dbc` → `areatable` table
- `ItemDisplayInfo.dbc` → `itemdisplayinfo` table

Column names and types are derived from embedded meta files that describe each DBC's structure.

## Meta Files

Thorium uses JSON meta files to understand DBC structure. These define:

- Column names and types
- String localization columns
- Primary key fields

Meta files are embedded in the Thorium binary and cover all standard 3.3.5a DBCs. Credits to [Foereaper's DBCTool](https://github.com/foereaper/dbctool) for the meta files.

## Checksums

Thorium uses MySQL's `CHECKSUM TABLE` feature to track which DBC tables have been modified by SQL migrations. This allows efficient export of only modified tables.

### How It Works

1. **During Import**: When DBCs are imported into the database, Thorium calculates and stores a baseline checksum for each table in the `dbc_checksum` table:
   ```sql
   CREATE TABLE dbc_checksum (
       table_name VARCHAR(255) PRIMARY KEY,
       checksum BIGINT UNSIGNED NOT NULL
   );
   ```

2. **During Migrations**: When you apply SQL migrations that modify DBC tables (e.g., `INSERT`, `UPDATE`, `DELETE`), the table's data changes, which changes its checksum.

3. **During Export**: When exporting DBCs:
   - Thorium calculates the current checksum for each table using `CHECKSUM TABLE`
   - Compares it against the stored baseline checksum in `dbc_checksum`
   - If checksums differ, the table has been modified and gets exported
   - If checksums match, the table is skipped (no changes)

4. **After Export**: The checksum is updated in `dbc_checksum` to reflect the new state, so subsequent exports only detect new changes.

## Building and Packaging
