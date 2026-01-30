# Thorium

A TrinityCore 335 modding framework.

Thorium handles extracting, applying, repackaging and distributing DBCs, world database and client interface files. You write SQL for both DBC and world database edits, or edit extracted client interface files. It also optionally applies some binary patches to wow.exe that aide development.

Thorium does not help you:
* Do anything with models, textures or maps (unless perhaps they are DBC or luaxml related).
* Make Trinity Core edits.
* Reverse engineer the client.

Thorium is nice for a more minimalist modding framework that does not require a forked TrinityCore, but is more complete that combining fragmented tools. Additionally, this framework plays will with LLMs by allowing them to do what they are strong at (view/edit SQL, write luaxml code) without getting lost on what they simply don't know (extraction, application, repacking and distribution of those files).

## Features

**Edit DBCs via SQL, auto-distribute to client and server.** Modify client data files using familiar SQL syntax. Thorium automatically exports your changes to DBC files and distributes them to both server and client.

**Minimalist SQL migration system.** One workflow for all SQL edits, including DBC changes. Apply and rollback with version control built-in.

**Unpack and modify LuaXML files.** Edit only what you need. Thorium finds your changes and packages just those files, automatically distributing them to your client.

**Optional, development-friendly binary patches.** See [docs/client-patcher.md](docs/client-patcher.md).

## Install

See [docs/install.md](docs/install.md) for detailed installation instructions, requirements, and configuration.

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
