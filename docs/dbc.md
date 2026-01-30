# DBC (Database Client Files)

DBC files are client-side database files used by World of Warcraft. They contain game data like items, spells, creatures, maps, and more. Thorium provides tools to extract, edit, and rebuild DBC files.

## Overview

The DBC workflow in Thorium:

1. **Extract** - Pull DBC files from the WoW client's MPQ archives
2. **Import** - Load DBC data into a MySQL database for easy editing
3. **Edit** - Modify data using SQL migrations in your mods
4. **Export** - Generate new DBC files from the database
5. **Build** - Package DBCs into an MPQ for the client (and copy to server)

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
│   └── dbc/           # Extracted DBC files (source of truth)
└── mods/
    └── my-mod/
        └── dbc_sql/   # SQL migrations for DBC changes
            ├── 20250129_120000_add_custom_spell.sql
            └── 20250129_120000_add_custom_spell.rollback.sql
```

## Database Tables

When DBCs are imported, each DBC file becomes a table in the `dbc` database. For example:

- `Item.dbc` → `item` table
- `Spell.dbc` → `spell` table
- `ItemDisplayInfo.dbc` → `itemdisplayinfo` table

Column names and types are derived from meta files that describe each DBC's structure.

## Meta Files

Thorium uses JSON meta files to understand DBC structure. These define:

- Column names and types
- String localization columns
- Primary key fields

Meta files are located in `thorium/internal/dbc/meta/` and cover all standard 3.3.5a DBCs.

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
