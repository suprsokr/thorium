# DBC (Database Client Files)

DBC files are client-side database files used by World of Warcraft. They contain game data like items, spells, creatures, maps, and more. Thorium provides tools to import, edit, and rebuild DBC files.

## Overview

The DBC workflow in Thorium:

1. **Import** - Load DBC files into a MySQL database (and save baseline copies)
2. **Edit** - Modify data using SQL migrations in your mods
3. **Export** - Generate new DBC files from modified tables
4. **Build** - Package only modified DBCs into an MPQ for the client (and copy to server)

## Why Use a Database?

Editing DBC files directly is tedious and error-prone. By importing them into MySQL:

- Use familiar SQL to query and modify data
- Track changes with version-controlled migration files
- Merge changes from multiple mods automatically
- Avoid binary file conflicts in git

## Directory Structure

```
mods/
├── shared/
│   └── dbc/
│       ├── dbc_source/    # Baseline DBC files (copied during import)
│       └── dbc_out/       # Exported modified DBC files
└── mods/
    └── my-mod/
        └── dbc_sql/       # SQL migrations for DBC changes
            ├── 20250129_120000_add_custom_spell.sql
            └── 20250129_120000_add_custom_spell.rollback.sql
```

## Importing DBCs

Before you can edit DBCs, you need to import them into the database. Thorium supports importing from any directory containing DBC files (e.g., TrinityCore's extracted DBCs).

```bash
# Import from TrinityCore's extracted DBCs
thorium import dbc --source /path/to/trinitycore/bin/dbc

# Import from WoW client extraction
thorium import dbc --source /path/to/extracted/DBFilesClient
```

The import command:
1. Creates a MySQL table for each DBC file
2. Loads all records into the table
3. Copies source DBCs to `mods/shared/dbc/dbc_source/` as baseline
4. Stores checksums to track modifications

**Note:** Import is typically run once during initial setup. The baseline files in `dbc_source/` are used later to determine which DBCs have been modified and need to be packaged.

## Database Tables

When DBCs are imported, each DBC file becomes a table in the `dbc` database. For example:

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

Meta files are embedded in the Thorium binary and cover all standard 3.3.5a DBCs.

## Common Use Cases

### Adding a Custom Spell

```sql
-- In: mods/my-mod/dbc_sql/20250129_add_fireball.sql
INSERT INTO `spell` (`ID`, `Name_Lang_enUS`, `Description_Lang_enUS`, ...)
VALUES (90001, 'Super Fireball', 'Launches a massive fireball', ...);
```

### Modifying an Existing Item Display

```sql
-- In: mods/my-mod/dbc_sql/20250129_fix_sword_model.sql
UPDATE `itemdisplayinfo` 
SET `ModelName_0` = 'MySword.m2'
WHERE `ID` = 12345;
```

### Adding a New Zone

```sql
-- In: mods/my-mod/dbc_sql/20250129_new_zone.sql
INSERT INTO `areatable` (`ID`, `AreaName_Lang_enUS`, `MapID`, ...)
VALUES (9001, 'Custom Valley', 0, ...);
```

### Enabling Flying in a Zone

```sql
-- In: mods/my-mod/dbc_sql/20250129_enable_flying.sql
-- Add AREA_FLAG_OUTLAND (0x1000 = 4096) to enable flying
UPDATE `areatable` 
SET `flags` = `flags` | 4096 
WHERE `id` = 12;  -- Elwynn Forest
```

## Exporting Modified DBCs

After applying migrations, export the modified tables back to DBC files:

```bash
thorium export
```

The export command:
1. Compares current table checksums against stored baselines
2. Only exports tables that have been modified
3. Writes DBC files to `mods/shared/dbc/dbc_out/`

You can also run export as part of the full build pipeline with `thorium build`.

## Building and Packaging

The build command handles the full pipeline:

```bash
thorium build                    # Build all mods
thorium build --mod my-mod       # Build specific mod only
```

During packaging, Thorium:
1. Compares exported DBCs (`dbc_out/`) against baseline (`dbc_source/`)
2. Only packages DBCs that actually differ
3. Creates a patch MPQ for the client (e.g., `patch-T.MPQ`)
4. Copies modified DBCs to the server's data folder

This ensures your patch MPQ only contains the DBCs you've actually modified, keeping file sizes minimal.

## Complete Example Workflow

```bash
# 1. Initialize workspace (if not already done)
thorium init

# 2. Import DBCs from TrinityCore
thorium import dbc --source /path/to/trinitycore/bin/dbc

# 3. Create a mod
thorium create-mod flying-zones

# 4. Create a migration
thorium create-migration --mod flying-zones --db dbc "enable flying elwynn"

# 5. Edit the migration SQL file to add your changes

# 6. Build (applies migrations, exports, packages)
thorium build --mod flying-zones

# 7. Test in-game with the new patch MPQ
```
