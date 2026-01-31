# Thorium

A single, unified workflow for creating WoW 3.3.5 server and client-side mods.

## Why Thorium?

**Works with stock TrinityCore.** No forked core required. Your mods are portable and easy to share.

**SQL-first workflow.** Edit DBCs and world data using SQL you already know. No learning curve for new file formats.

**Handles the boring stuff.** Extraction, packaging, distribution. Focus on your mod, not the pipeline.

**Organized by design.** Each mod is self-contained with its own migrations, scripts, and assets. Enable, disable, or share individual mods without touching others.

**LLM-friendly.** Thorium's SQL and Lua-based files lets AI assistants help you build mods without getting lost in binary formats or complex toolchains.

## Features

### DBC Editing via SQL
Modify client data files—items, spells, creatures, maps—using familiar SQL syntax. Thorium imports DBCs into MySQL tables, lets you edit with migrations, then exports and distributes automatically to both client and server.

### World Database Migrations
Same migration workflow for TrinityCore's world database. Version-controlled, with apply and rollback for every change.

### LuaXML & Custom Addons
Extract and modify client interface files. Create custom addons that get packaged into MPQs automatically. Your UI changes distribute with a single command.

### Custom Packets
Build features that need real-time client-server communication. Send custom data between your addons and server scripts with a simple Lua API. See [docs/custom-packets.md](docs/custom-packets.md).

### C++ Script Scaffolding
Generate TrinityCore script templates for spells, creatures, and packet handlers. Thorium sets up the boilerplate scripts and distributes into your Trinity Core server so you can focus on logic.

### Binary Patches
Optional client patches for development: custom login screens, extended Lua APIs, custom packet support, and more. See [docs/client-patcher.md](docs/client-patcher.md).

### Distribution Packages
Bundle your mod for sharing: MPQs, SQL files, and scripts packaged together with install instructions.

## Install

See [docs/install.md](docs/install.md) for pre-built binaries or building from source.

## Quick Start

Create a new item mod from scratch:

```bash
# Initialize workspace
thorium init

# Configure paths and database connections
vim config.json
# Update these required fields:
#   "wotlk": {
#     "path": "/path/to/your/wotlk/client"  # or "${WOTLK_PATH}" to use env var
#   },
#   "databases": {
#     "dbc": {
#       "user": "root",        # DBC database user
#       "password": "",        # DBC database password
#       "host": "127.0.0.1",   # DBC database host
#       "port": "3306",        # DBC database port
#       "name": "dbc"          # DBC database name
#     },
#     "world": {
#       "user": "trinity",     # World database user
#       "password": "trinity", # World database password
#       "host": "127.0.0.1",  # World database host
#       "port": "3306",       # World database port
#       "name": "world"        # World database name
#     }
#   },
#   "server": {
#     "dbc_path": "/path/to/trinitycore/server/data/dbc"  # or "${TC_SERVER_PATH}/data/dbc"
#   }

# Create a new mod (without LuaXML since we're just adding an item)
thorium create-mod my-custom-item --no-luaxml

# Create DBC migration for the item definition
thorium create-migration --mod my-custom-item --db dbc add_legendary_sword

# Edit the DBC migration file (apply)
vim mods/my-custom-item/dbc_sql/*_add_legendary_sword.sql
# DELETE FROM `Item` WHERE `id` = 90001;
# INSERT INTO `Item` (`id`, `class`, `subclass`, `sound_override_subclass`, `material`, `display_id`, `inventory_type`, `sheath`) 
# VALUES (90001, 2, 7, -1, 1, 32254, 17, 1);

# Edit the DBC rollback file
vim mods/my-custom-item/dbc_sql/*_add_legendary_sword.rollback.sql
# DELETE FROM `Item` WHERE `id` = 90001;

# Create world database migration for server-side item data
thorium create-migration --mod my-custom-item --db world add_legendary_sword_stats

# Edit the world migration file (apply)
vim mods/my-custom-item/world_sql/*_add_legendary_sword_stats.sql
# DELETE FROM `item_template` WHERE `entry` = 90001;
# INSERT INTO `item_template` (
#     `entry`, `class`, `subclass`, `SoundOverrideSubclass`, `name`, `displayid`,
#     `Quality`, `InventoryType`, `ItemLevel`, `RequiredLevel`,
#     `StatsCount`, `stat_type1`, `stat_value1`,
#     `dmg_min1`, `dmg_max1`, `dmg_type1`, `delay`,
#     `bonding`, `description`, `Material`, `sheath`, `MaxDurability`,
#     `RequiredDisenchantSkill`, `DisenchantID`
# ) VALUES (
#     90001, 2, 7, -1, 'Legendary Sword of Power', 32254,
#     5, 17, 80, 70,
#     1, 7, 50,
#     200, 300, 0, 3500,
#     1, 'A legendary weapon of immense power.', 1, 1, 120,
#     -1, 60
# );

# Edit the world rollback file
vim mods/my-custom-item/world_sql/*_add_legendary_sword_stats.rollback.sql
# DELETE FROM `item_template` WHERE `entry` = 90001;

# Build everything: apply migrations, export DBCs, package MPQs, distribute to server and client.
thorium build
```

Launch your Trinty Core 335 server and client. `.additem 90001` in game from a GM account.
