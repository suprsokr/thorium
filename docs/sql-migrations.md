# SQL Migrations

Thorium uses SQL migrations to manage database changes. Migrations are versioned SQL files that apply changes to either the DBC database or the World database.

## Overview

Each mod can contain migrations for:

- **DBC database** - Client data like spells, items, creatures display info
- **World database** - Server data like NPCs, quests, loot tables, item stats

Migrations are timestamped and applied in order. Each migration has a rollback counterpart.

## DBC Migrations

DBC migrations are applied to the `dbc` database, then exported to DBC files during `thorium build`, which packages them into client MPQs. See [DBC Workflow](dbc.md) for complete details on setting up and working with DBC migrations.

## World Migrations

World migrations modify server-side data in your TrinityCore world database. These migrations:

- **Use existing database** - Work with your TrinityCore `world` database (no special setup needed)
- **Modify server data** - Changes affect server behavior (NPCs, quests, loot, item stats, etc.)
- **Applied directly** - SQL runs directly against the world database
- **No export needed** - Changes are in the database, no file generation required

**Workflow:** World migrations are applied directly to your world database and take effect immediately after applying.

## Directory Structure

```
mods/mods/my-mod/
├── dbc_sql/                                    # DBC database migrations
│   ├── 20250129_120000_add_custom_spell.sql
│   └── 20250129_120000_add_custom_spell.rollback.sql
└── world_sql/                                  # World database migrations
    ├── 20250129_130000_add_custom_npc.sql
    └── 20250129_130000_add_custom_npc.rollback.sql
```

## Creating Migrations

Use the `create-migration` command to generate migration file pairs:

```bash
# Create a DBC migration
thorium create-migration --mod my-mod --db dbc "add custom spell"

# Create a World migration  
thorium create-migration --mod my-mod --db world "add custom npc"
```

This generates timestamped files:

```
mods/mods/my-mod/dbc_sql/
├── 20250129_143052_add_custom_spell.sql          # Apply migration
└── 20250129_143052_add_custom_spell.rollback.sql # Rollback migration
```

## Writing Migrations

### Apply Migration

The apply file contains SQL to make your changes:

```sql
-- Migration: add_custom_spell
-- Database: DBC

INSERT INTO `spell` (`ID`, `Name_Lang_enUS`, `Description_Lang_enUS`)
VALUES (90001, 'Super Fireball', 'A powerful fireball spell');
```

### Rollback Migration

The rollback file undoes the apply changes:

```sql
-- Rollback: add_custom_spell
-- Database: DBC

DELETE FROM `spell` WHERE `ID` = 90001;
```

## Applying Migrations

Use `thorium build` with component selection to apply migrations:

```bash
# Apply all pending migrations for all mods
thorium build dbc_sql world_sql

# Apply migrations for a specific mod
thorium build --mod my-mod dbc_sql world_sql

# Apply only DBC migrations
thorium build dbc_sql

# Apply only World migrations
thorium build world_sql
```

**Note:** The `build` command applies migrations as part of its pipeline. Use component selection (`dbc_sql`, `world_sql`) to run only the migration steps without building other components.

## Rolling Back

```bash
# Rollback the last migration for all mods
thorium rollback

# Rollback all migrations for a mod
thorium rollback --mod my-mod --all

# Rollback only DBC migrations
thorium rollback --db dbc
```

## Migration Tracking

Applied migrations are tracked in:

```
mods/shared/migrations_applied/
└── my-mod/
    ├── dbc/
    │   └── 20250129_120000_add_custom_spell.sql
    └── world/
        └── 20250129_130000_add_custom_npc.sql
```

When a migration is applied, it's copied here. This prevents re-applying the same migration and enables rollback tracking.

## Best Practices

1. **Use high entry IDs** - Start custom entries at 90000+ to avoid conflicts with existing data

2. **Always write rollbacks** - Even if you don't plan to use them, rollbacks document what the migration changes

3. **Keep migrations small** - One logical change per migration makes debugging easier

4. **Use DELETE before INSERT** - Makes migrations idempotent:
   ```sql
   DELETE FROM `spell` WHERE `ID` = 90001;
   INSERT INTO `spell` (`ID`, ...) VALUES (90001, ...);
   ```
